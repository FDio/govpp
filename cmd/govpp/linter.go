//  Copyright (c) 2023 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen/vppapi"
)

const (
	FILE_BASIC                       = "FILE_BASIC"
	MESSAGE_DEPRECATE_OLDER_VERSIONS = "MESSAGE_DEPRECATE_OLDER_VERSIONS"
	MESSAGE_SAME_STATUS              = "MESSAGE_SAME_STATUS"
)

var defaultLintRules = []string{
	FILE_BASIC,
	MESSAGE_DEPRECATE_OLDER_VERSIONS,
	MESSAGE_SAME_STATUS,
}

type LintRule struct {
	Id      string
	Purpose string
	//Categories []string
	check func(schema *vppapi.Schema) error
}

func LintRules(ids ...string) []*LintRule {
	var rules []*LintRule
	for _, id := range ids {
		rule := GetLintRule(id)
		if rule != nil {
			rules = append(rules, rule)
		}
	}
	return rules
}

func GetLintRule(id string) *LintRule {
	rule, ok := lintRules[id]
	if ok {
		return &rule
	}
	return nil
}

var lintRules = map[string]LintRule{
	FILE_BASIC: {
		Id:      FILE_BASIC,
		Purpose: "File must have basic info defined",
		check: checkFiles(func(file *vppapi.File) error {
			var errs LintIssues
			if file.CRC == "" {
				errs = errs.Append(fmt.Errorf("file %s must have CRC defined", file.Name))
			}
			if file.Options != nil && file.Options[vppapi.OptFileVersion] == "" {
				errs = errs.Append(fmt.Errorf("file %s must have version defined", file.Name))
			}
			return errs.ToErr()
		}),
	},
	MESSAGE_DEPRECATE_OLDER_VERSIONS: {
		Id:      MESSAGE_DEPRECATE_OLDER_VERSIONS,
		Purpose: "Message should be marked as deprecated if newer version exists",
		check: checkFiles(func(file *vppapi.File) error {
			var errs LintIssues
			messageVersions := extractFileMessageVersions(file)
			versionMessages := extractMessageVersions(file)
			for _, message := range file.Messages {
				baseName, version := extractBaseNameAndVersion(message.Name)
				// if this is not the latest version of a message
				if version < messageVersions[baseName] {
					var newer vppapi.Message
					if vers, ok := versionMessages[baseName]; ok {
						if newVer, ok := vers[version+1]; ok {
							newer = newVer
						}
					}
					// older messages should be marked as deprecated (if newer message version is not in progress)
					if !isMessageDeprecated(message) && !isMessageInProgress(newer) {
						errs = errs.Append(LintIssue{
							File: file.Path,
							// TODO: Line: ?,
							Object: map[string]any{
								"Message": message,
								"Base":    baseName,
								"Version": version,
								"Latest":  messageVersions[baseName],
							},
							Violation: color.Sprintf("message %s has newer version available (%v) but is not marked as deprecated",
								color.Cyan.Sprint(message.Name), color.Cyan.Sprint(newer.Name)),
						})
					}
				}
			}
			return errs.ToErr()
		}),
	},
	MESSAGE_SAME_STATUS: {
		Id:      MESSAGE_SAME_STATUS,
		Purpose: "Message request and reply must have the same status",
		check: checkFiles(func(file *vppapi.File) error {
			var errs LintIssues
			for _, message := range file.Messages {
				status := getMessageStatus(message)
				related := getRelatedMessages(file, message.Name)
				for _, rel := range related {
					if relMsg, ok := getFileMessage(file, rel); ok {
						if relStatus := getMessageStatus(relMsg); relStatus != status {
							errs = errs.Append(LintIssue{
								File: file.Path,
								Object: map[string]any{
									"Message":       message,
									"Related":       relMsg,
									"Status":        status,
									"RelatedStatus": relStatus,
								},
								Violation: color.Sprintf("message %s does not have consistent status (%v) with related message: %v (%v)",
									color.Cyan.Sprint(message.Name), clrWhite.Sprint(status), color.Cyan.Sprint(rel), clrWhite.Sprint(relStatus)),
							})
						}
					}
				}
			}
			return errs.ToErr()
		}),
	},
}

func checkFiles(checkFn func(file *vppapi.File) error) func(*vppapi.Schema) error {
	return func(schema *vppapi.Schema) error {
		errs := LintIssues{}
		logrus.Tracef("running checkFiles for %d files", len(schema.Files))
		for _, file := range schema.Files {
			e := checkFn(&file)
			if e != nil {
				logrus.Tracef("checked file: %v (%v)", file.Name, e)
			}
			errs = errs.Append(e)
		}
		return errs
	}
}

type Linter struct {
	rules []*LintRule
}

func NewLinter() *Linter {
	return &Linter{
		rules: LintRules(defaultLintRules...),
	}
}

func (l *Linter) Enable(rules ...*LintRule) {
	for _, rule := range rules {
		var found bool
		for _, r := range l.rules {
			if r.Id == rule.Id {
				found = true
				break
			}
		}
		if !found {
			l.rules = append(l.rules, rule)
		}
	}
}

func (l *Linter) Disable(rules ...string) {
	for _, rule := range rules {
		for i, r := range l.rules {
			if r.Id == rule {
				l.rules = append(l.rules[:i], l.rules[i+1:]...)
				break
			}
		}
	}
}

func (l *Linter) Lint(schema *vppapi.Schema) error {
	logrus.Debugf("linter will run %d rules for %d files (schema version: %v)", len(l.rules), len(schema.Files), schema.Version)

	var errs LintIssues

	for _, rule := range l.rules {
		log := logrus.WithField("rule", rule.Id)
		log.Debugf("running linter check for rule (purpose: %v)", rule.Purpose)

		err := rule.check(schema)
		if err != nil {
			log.Tracef("linter check failed: %v", err)

			switch e := err.(type) {
			case LintIssue:
				if e.RuleId == "" {
					e.RuleId = rule.Id
				}
				errs = errs.Append(e)
			case LintIssues:
				for _, le := range e {
					if le.RuleId == "" {
						le.RuleId = rule.Id
					}
					errs = errs.Append(le)
				}
			default:
				errs = append(errs, createLintIssue(rule.Id, err))
			}
		} else {
			log.Tracef("linter check passed")
		}
	}
	if len(errs) > 0 {
		logrus.Debugf("found %d issues in %d files", len(errs), len(schema.Files))
		return errs
	} else {
		logrus.Debugf("no issues in %d files", len(schema.Files))
	}

	return nil
}

func (l *Linter) SetRules(rules []string) {
	var list []*LintRule
	for _, rule := range l.rules {
		keep := false
		for _, r := range rules {
			if rule.Id == r {
				keep = true
				break
			}
		}
		if keep {
			list = append(list, rule)
		}
	}
	l.rules = list
}

type LintIssue struct {
	RuleId string

	File string
	Line int `json:",omitempty"`

	Object    any `json:",omitempty"`
	Violation string
}

func createLintIssue(id string, err error) LintIssue {
	return LintIssue{
		RuleId:    id,
		Violation: err.Error(),
	}
}

func (l LintIssue) Error() string {
	return l.Violation
}

type LintIssues []LintIssue

func (le LintIssues) ToErr() error {
	if len(le) == 0 {
		return nil
	}
	return le
}

func (le LintIssues) Error() string {
	if len(le) == 0 {
		return "no issues"
	}
	return fmt.Sprintf("%d issues", len(le))
}

func (le LintIssues) Append(errs ...error) LintIssues {
	var r = le
	for _, err := range errs {
		if err == nil {
			continue
		}
		switch e := err.(type) {
		case LintIssue:
			r = append(r, e)
		case LintIssues:
			r = append(r, e...)
		default:
			r = append(r, createLintIssue("", err))
		}
	}
	return r
}

const (
	optionStatus           = "status"
	optionStatusInProgress = "in_progress"
	optionStatusDeprecated = "deprecated"
)

func getMessageStatus(message vppapi.Message) string {
	if isMessageDeprecated(message) {
		return optionStatusDeprecated
	}
	if isMessageInProgress(message) {
		return optionStatusInProgress
	}
	if status, ok := message.Options[optionStatus]; ok {
		return status
	}
	return "n/a"
}

func isMessageDeprecated(message vppapi.Message) bool {
	if _, ok := message.Options[optionStatusDeprecated]; ok {
		return true
	}
	if val, ok := message.Options[optionStatus]; ok && strings.ToLower(val) == optionStatusDeprecated {
		return true
	}
	return false
}

func isMessageInProgress(message vppapi.Message) bool {
	if _, ok := message.Options[optionStatusInProgress]; ok {
		return true
	}
	if val, ok := message.Options[optionStatus]; ok && strings.ToLower(val) == optionStatusInProgress {
		return true
	}
	return false
}

func extractBaseNameAndVersion(messageName string) (string, int) {
	re := regexp.MustCompile(`^(.+)_v(\d+)(_(?:reply|dump|details))?$`)
	matches := re.FindStringSubmatch(messageName)
	if len(matches) == 4 {
		name := matches[1] + matches[3]
		version, _ := strconv.Atoi(matches[2])
		return name, version
	} else {
		return messageName, 1
	}
}

func getFileMessage(file *vppapi.File, msg string) (vppapi.Message, bool) {
	for _, message := range file.Messages {
		if message.Name == msg {
			return message, true
		}
	}
	return vppapi.Message{}, false
}

func extractFileMessageVersions(file *vppapi.File) map[string]int {
	messageVersions := make(map[string]int)
	for _, message := range file.Messages {
		baseName, version := extractBaseNameAndVersion(message.Name)
		if version > messageVersions[baseName] {
			messageVersions[baseName] = version
		}
	}
	return messageVersions
}

func extractMessageVersions(file *vppapi.File) map[string]map[int]vppapi.Message {
	messageVersions := make(map[string]map[int]vppapi.Message)
	for _, message := range file.Messages {
		baseName, version := extractBaseNameAndVersion(message.Name)
		if messageVersions[baseName] == nil {
			messageVersions[baseName] = make(map[int]vppapi.Message)
		}
		messageVersions[baseName][version] = message
	}
	return messageVersions
}

func extractMessagesRPC(file *vppapi.File) map[string]vppapi.RPC {
	messagesRPC := make(map[string]vppapi.RPC)
	if file.Service != nil {
		for _, rpc := range file.Service.RPCs {
			if m := rpc.Request; m != "" {
				messagesRPC[m] = rpc
			}
			/*if m := rpc.Reply; m != "" {
				messagesRPC[m] = rpc
			}
			if m := rpc.StreamMsg; m != "" {
				messagesRPC[m] = rpc
			}
			for _, m := range rpc.Events {
				messagesRPC[m] = rpc
			}*/
		}
	}
	return messagesRPC
}

func extractRPCMessages(rpc vppapi.RPC) []string {
	var messages []string
	if m := rpc.Request; m != "" {
		messages = append(messages, m)
	}
	if m := rpc.Reply; m != "" {
		messages = append(messages, m)
	}
	if m := rpc.StreamMsg; m != "" {
		messages = append(messages, m)
	}
	for _, m := range rpc.Events {
		messages = append(messages, m)
	}
	return messages
}

func getRelatedMessages(file *vppapi.File, msg string) []string {
	msgsRPC := extractMessagesRPC(file)
	var related []string
	if rpc, ok := msgsRPC[msg]; ok {
		for _, m := range extractRPCMessages(rpc) {
			if m != msg {
				related = append(related, m)
			}
		}
	}
	return related
}
