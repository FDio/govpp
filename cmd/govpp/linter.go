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

var defaultLintRules = []LintRule{
	FILE_BASIC,
	MESSAGE_DEPRECATE_OLDER_VERSIONS,
}

var (
	FILE_BASIC = LintRule{
		Id:      "FILE_BASIC",
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
	}
	MESSAGE_DEPRECATE_OLDER_VERSIONS = LintRule{
		Id:      "MESSAGE_DEPRECATE_OLDER_VERSIONS",
		Purpose: "Message should be marked as deprecated if newer version exists",
		check: checkFiles(func(file *vppapi.File) error {
			var errs LintIssues
			messageVersions := extractFileMessageVersions(file)
			versionMessages := extractMessageVersions(file)
			for _, message := range file.Messages {
				baseName, version := extractBaseNameAndVersion(message.Name)
				// if this is not the latest version of a message
				if version < messageVersions[baseName] {
					// check if newer message version is in progress
					if vers, ok := versionMessages[baseName]; ok {
						if newVer, ok := vers[version+1]; ok && isMessageInProgress(newVer) {
							continue
						}
					}
					// otherwise the message should be marked as deprecated
					if !isMessageDeprecated(message) {
						obj := map[string]any{
							"Message": message,
						}
						errs = errs.Append(LintIssue{
							File: file.Path,
							// TODO: Line: ?,
							Object:    obj,
							Violation: color.Sprintf("message %s has newer version available and should be marked as deprecated", color.Cyan.Sprint(message.Name)),
						})
					}
				}
			}
			return errs.ToErr()
		}),
	}
)

type Linter struct {
	rules []LintRule
}

func NewLinter() *Linter {
	return &Linter{
		rules: defaultLintRules,
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
			l.rules = append(l.rules, *rule)
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
	logrus.Debugf("running linter for schema version: %v (%d files)", schema.Version, len(schema.Files))

	var errs LintIssues
	for _, rule := range l.rules {
		logrus.Debugf("running check for lint rule: %v (%v)", rule.Id, rule.Purpose)

		err := rule.check(schema)
		if err != nil {
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
				errs = append(errs, LintIssue{
					RuleId:    rule.Id,
					Violation: err.Error(),
				})
			}
		}
	}
	if len(errs) > 0 {
		logrus.Debugf("linter found %d issues in %d files", len(errs), len(schema.Files))
		return errs
	}
	logrus.Debugln("linter found no issues found in the VPP API schema")

	return nil
}

func (l *Linter) SetRules(rules []string) {
	var list []LintRule
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

type LintRule struct {
	Id      string
	Purpose string
	//Categories []string
	check func(schema *vppapi.Schema) error
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

type LintIssue struct {
	RuleId string

	File string
	Line int `json:",omitempty"`

	Object    any `json:",omitempty"`
	Violation string
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
			r = append(r, LintIssue{
				Violation: err.Error(),
			})
		}
	}
	return r
}

const (
	optionStatusInProgress = "in_progress"
	optionStatusDeprecated = "deprecated"
)

func isMessageDeprecated(message vppapi.Message) bool {
	if _, ok := message.Options[optionStatusDeprecated]; ok {
		return true
	}
	if val, ok := message.Options["status"]; ok && strings.ToLower(val) == optionStatusDeprecated {
		return true
	}
	return false
}

func isMessageInProgress(message vppapi.Message) bool {
	if _, ok := message.Options[optionStatusInProgress]; ok {
		return true
	}
	if val, ok := message.Options["status"]; ok && strings.ToLower(val) == optionStatusInProgress {
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
