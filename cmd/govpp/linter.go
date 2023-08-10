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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen/vppapi"
)

// TODO:
//  - consider adding categories for linter rules

const (
	FILE_BASIC                       = "FILE_BASIC"
	MESSAGE_DEPRECATE_OLDER_VERSIONS = "MESSAGE_DEPRECATE_OLDER_VERSIONS"
	MESSAGE_SAME_STATUS              = "MESSAGE_SAME_STATUS"
	UNUSED_MESSAGE                   = "UNUSED_MESSAGE"
)

var defaultLintRules = []string{
	FILE_BASIC,
	MESSAGE_DEPRECATE_OLDER_VERSIONS,
	MESSAGE_SAME_STATUS,
	UNUSED_MESSAGE,
}

type LintRule struct {
	Id      string
	Purpose string

	check func(schema *vppapi.Schema) LintIssues
}

var lintRules = map[string]LintRule{
	FILE_BASIC: {
		Id:      FILE_BASIC,
		Purpose: "File must have basic info defined",
		check:   checkFiles(checkFileBasic),
	},
	MESSAGE_DEPRECATE_OLDER_VERSIONS: {
		Id:      MESSAGE_DEPRECATE_OLDER_VERSIONS,
		Purpose: "Message should be marked as deprecated if newer version exists",
		check:   checkFiles(checkFileMessageDeprecateOldVersions),
	},
	MESSAGE_SAME_STATUS: {
		Id:      MESSAGE_SAME_STATUS,
		Purpose: "Messages that are related must have the same status",
		check:   checkFiles(checkFileMessageSameStatus),
	},
	UNUSED_MESSAGE: {
		Id:      UNUSED_MESSAGE,
		Purpose: "Messages should be used in services",
		check:   checkFiles(checkFileMessageUsed),
	},
}

func GetLintRule(id string) *LintRule {
	rule, ok := lintRules[id]
	if ok {
		return &rule
	}
	return nil
}

func ListLintRules(ids ...string) []*LintRule {
	var rules []*LintRule
	for _, id := range ids {
		rule := GetLintRule(id)
		if rule != nil {
			rules = append(rules, rule)
		}
	}
	return rules
}

func checkFiles(checkFn func(file *vppapi.File) LintIssues) func(*vppapi.Schema) LintIssues {
	return func(schema *vppapi.Schema) LintIssues {
		var issues LintIssues

		logrus.Tracef("running checkFiles for %d files", len(schema.Files))

		for _, file := range schema.Files {
			e := checkFn(&file)
			if e != nil {
				logrus.Tracef("checked file: %v (%v)", file.Name, e)
			}
			issues = append(issues, e...)
		}

		return issues
	}
}

func checkFileBasic(file *vppapi.File) LintIssues {
	var issues LintIssues
	if file.CRC == "" {
		issues = issues.Append(fmt.Errorf("file %s must have CRC defined", file.Name))
	}
	if file.Options != nil && file.Options[vppapi.OptFileVersion] == "" {
		issues = issues.Append(fmt.Errorf("file %s must have version defined", file.Name))
	}
	return issues
}

func checkFileMessageDeprecateOldVersions(file *vppapi.File) LintIssues {
	var issues LintIssues

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
				issues = issues.Append(LintIssue{
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

	return issues
}

func checkFileMessageSameStatus(file *vppapi.File) LintIssues {
	var issues LintIssues

	for _, message := range file.Messages {
		status := getMessageStatus(message)
		relatedMsgs := getRelatedMessages(file, message.Name)

		for _, relatedMsg := range relatedMsgs {
			relMsg, ok := getFileMessage(file, relatedMsg)
			if !ok {
				logrus.Warnf("could not find related message %s in file %s", relatedMsg, file.Path)
				continue
			}
			if relStatus := getMessageStatus(relMsg); relStatus != status {
				issues = issues.Append(LintIssue{
					File: file.Path,
					Object: map[string]any{
						"Message":       message,
						"Related":       relMsg,
						"Status":        status,
						"RelatedStatus": relStatus,
					},
					Violation: color.Sprintf("message %s does not have consistent status (%v) with related message: %v (%v)",
						color.Cyan.Sprint(message.Name), clrWhite.Sprint(status), color.Cyan.Sprint(relatedMsg), clrWhite.Sprint(relStatus)),
				})
			}
		}
	}

	return issues
}

func checkFileMessageUsed(file *vppapi.File) LintIssues {
	var issues LintIssues

	rpcMsgs := extractFileMessagesToRPC(file)

	for _, message := range file.Messages {
		if _, ok := rpcMsgs[message.Name]; !ok {
			issues = issues.Append(LintIssue{
				File: file.Path,
				Object: map[string]any{
					"Message": message,
				},
				Violation: color.Sprintf("message %s is not used by services", color.Cyan.Sprint(message.Name)),
			})
		}
	}

	return issues
}

type Linter struct {
	rules []*LintRule
}

func NewLinter() *Linter {
	return &Linter{
		rules: ListLintRules(defaultLintRules...),
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

func (l *Linter) Lint(schema *vppapi.Schema) (LintIssues, error) {
	logrus.Debugf("running %d rule checks for %d files (schema version: %v)", len(l.rules), len(schema.Files), schema.Version)

	var allIssues LintIssues

	for _, rule := range l.rules {
		log := logrus.WithField("rule", rule.Id)
		log.Tracef("running rule check (purpose: %v)", rule.Purpose)

		issues := rule.check(schema)
		if len(issues) > 0 {
			log.Tracef("rule check found %d issues", len(issues))

			for _, issue := range issues {
				if issue.RuleId == "" {
					issue.RuleId = rule.Id
				}
				allIssues = allIssues.Append(issue)
			}
		} else {
			log.Tracef("rule check passed")
		}
	}

	if len(allIssues) > 0 {
		logrus.Tracef("found %d issues in %d files", len(allIssues), len(schema.Files))
	} else {
		logrus.Tracef("no issues found in %d files", len(schema.Files))
	}

	return allIssues, nil
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
	RuleId    string
	File      string
	Violation string

	Object any `json:",omitempty"`
	Line   int `json:",omitempty"`
}

func (l LintIssue) Error() string {
	return l.Violation
}

type LintIssues []LintIssue

func (le LintIssues) Append(errs ...error) LintIssues {
	var r = le
	for _, err := range errs {
		if err == nil {
			continue
		}
		var e LintIssue
		switch {
		case errors.As(err, &e):
			r = append(r, e)
		default:
			r = append(r, LintIssue{
				RuleId:    "",
				Violation: err.Error(),
			})
		}
	}
	return r
}

const (
	optionStatus           = "status"
	optionStatusInProgress = "in_progress"
	optionStatusDeprecated = "deprecated"

	noReply = "null"

	noStatus = "n/a"
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
	return noStatus
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

func extractFileMessagesToRPC(file *vppapi.File) map[string]vppapi.RPC {
	if file.Service == nil {
		return nil
	}
	messagesRPC := make(map[string]vppapi.RPC)
	for _, rpc := range file.Service.RPCs {
		for _, m := range extractRPCMessages(rpc) {
			messagesRPC[m] = rpc
		}
	}
	return messagesRPC
}

func extractMessageRequestsToRPC(file *vppapi.File) map[string]vppapi.RPC {
	if file.Service == nil {
		return nil
	}
	messagesRPC := make(map[string]vppapi.RPC)
	for _, rpc := range file.Service.RPCs {
		if m := rpc.Request; m != "" {
			messagesRPC[m] = rpc
		}
	}
	return messagesRPC
}

func extractRPCMessages(rpc vppapi.RPC) []string {
	var messages []string
	if m := rpc.Request; m != "" {
		messages = append(messages, m)
	}
	if m := rpc.Reply; m != "" && m != noReply {
		messages = append(messages, m)
	}
	if m := rpc.StreamMsg; m != "" {
		messages = append(messages, m)
	}
	messages = append(messages, rpc.Events...)
	return messages
}

func getRelatedMessages(file *vppapi.File, msg string) []string {
	msgsRPC := extractMessageRequestsToRPC(file)
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
