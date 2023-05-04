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
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

type LintCmdOptions struct {
	Input    string
	Format   string
	Enable   []string
	Disable  []string
	ExitCode bool
}

func newLintCmd() *cobra.Command {
	var (
		opts = LintCmdOptions{}
	)
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint VPP API files",
		Long:  "Run linter checks for VPP API files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLintCmd(cmd.OutOrStdout(), opts, args)
		},
	}

	cmd.PersistentFlags().StringSliceVarP(&opts.Enable, "enable", "E", nil, "Enable previously disabled linters.\n")
	cmd.PersistentFlags().StringSliceVar(&opts.Disable, "disable", nil, "Disable previously enabled linters.")
	cmd.PersistentFlags().StringVar(&opts.Input, "input", "", "Input for VPP API (e.g. path to VPP API directory, local VPP repo)")
	cmd.PersistentFlags().StringVar(&opts.Format, "format", "", "The format for lint issues")
	cmd.PersistentFlags().BoolVar(&opts.ExitCode, "exit-code", false, "Exit with non-zero exit code if there are any issues")

	return cmd
}

func runLintCmd(out io.Writer, opts LintCmdOptions, args []string) error {
	vppInput, err := resolveInput(opts.Input)
	if err != nil {
		return err
	}

	schema := vppInput.Schema

	linter := NewLinter()

	if err := linter.Lint(&schema); err != nil {
		if errs, ok := err.(LintErrors); ok {
			if opts.Format == "" {
				printLintErrorsAsTable(out, errs)
			} else {
				formatAsTemplate(out, opts.Format, errs)
			}
		}
		if opts.ExitCode {
			return err
		} else {
			logrus.Errorln(err)
		}
	}

	return nil
}

func printLintErrorsAsTable(out io.Writer, errs LintErrors) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"#", "Rule", "File", "Issue",
	})
	table.SetAutoMergeCells(true)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetBorder(false)

	for i, e := range errs {
		index := i + 1
		file := e.File
		if e.Line > 0 {
			file += fmt.Sprintf(":%d", e.Line)
		}

		row := []string{
			fmt.Sprint(index), e.RuleId, file, e.Message,
		}
		table.Append(row)
	}
	table.Render()
}

var (
	FILE_BASIC = LintRule{
		Id:      "FILE_BASIC",
		Purpose: "File must have basic info defined",
		check: checkFiles(func(file *vppapi.File) error {
			var errs LintErrors
			if file.CRC == "" {
				errs = errs.Append(fmt.Errorf("file %s does not have CRC defined", file.Name))
			}
			if file.Options != nil && file.Options[vppapi.OptFileVersion] == "" {
				errs = errs.Append(fmt.Errorf("file %s does not have version defined", file.Name))
			}
			return errs
		}),
	}
	MESSAGE_DEPRECATE_OLD = LintRule{
		Id:      "MESSAGE_DEPRECATE_OLD",
		Purpose: "Message should be marked as deprecated if newer version exists",
		check: checkFiles(func(file *vppapi.File) error {
			var errs LintErrors
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
					if _, ok := message.Options["deprecated"]; !ok {
						obj := map[string]any{
							"Message": message,
						}
						errs = errs.Append(LintError{
							File: file.Path,
							// TODO: Line: ?,
							Object:  obj,
							Message: fmt.Sprintf("message %s should be marked as deprecated (newer version available)", message.Name),
						})
					}
				}
			}
			return errs
		}),
	}
)

var defaultLintRules = []LintRule{
	FILE_BASIC,
	MESSAGE_DEPRECATE_OLD,
}

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
		found := false
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

func (l *Linter) Disable(rules ...*LintRule) {
	for _, rule := range rules {
		for i, r := range l.rules {
			if r.Id == rule.Id {
				l.rules = append(l.rules[:i], l.rules[i+1:]...)
				break
			}
		}
	}
}

func (l *Linter) Lint(schema *vppapi.Schema) error {
	logrus.Debugf("running linter for schema version: %v (%d files)", schema.Version, len(schema.Files))

	var errs LintErrors
	for _, rule := range l.rules {
		logrus.Debugf("running check for lint rule: %v (%v)", rule.Id, rule.Purpose)

		err := rule.check(schema)
		if err != nil {
			switch e := err.(type) {
			case LintError:
				if e.RuleId == "" {
					e.RuleId = rule.Id
				}
				errs = errs.Append(e)
			case LintErrors:
				for _, le := range e {
					if le.RuleId == "" {
						le.RuleId = rule.Id
					}
					errs = errs.Append(le)
				}
			default:
				errs = append(errs, LintError{
					RuleId:  rule.Id,
					Message: err.Error(),
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

type LintRule struct {
	Id      string
	Purpose string
	//Categories []string
	check func(schema *vppapi.Schema) error
}

func checkFiles(checkFn func(file *vppapi.File) error) func(*vppapi.Schema) error {
	return func(schema *vppapi.Schema) error {
		errs := LintErrors{}
		logrus.Tracef("running checkFiles for %d files", len(schema.Files))
		for _, file := range schema.Files {
			errs = errs.Append(checkFn(&file))
			logrus.Tracef("checkFile: %v", file.Name)
		}
		return errs
	}
}

type LintError struct {
	RuleId string

	File string
	Line int `json:",omitempty"`

	Object  any `json:",omitempty"`
	Message string
}

func (l LintError) Error() string {
	return l.Message
}

type LintErrors []LintError

func (le LintErrors) Error() string {
	if len(le) == 0 {
		return "found no lint issues"
	}
	return fmt.Sprintf("found %d lint issues", len(le))
}

func (le LintErrors) Append(errs ...error) LintErrors {
	var r = le
	for _, err := range errs {
		if err == nil {
			continue
		}
		switch e := err.(type) {
		case LintError:
			r = append(r, e)
		case LintErrors:
			r = append(r, e...)
		default:
			r = append(r, LintError{
				Message: err.Error(),
			})
		}
	}
	return r
}

const statusInProgress = "in_progress"

func isMessageInProgress(message vppapi.Message) bool {
	if _, ok := message.Options[statusInProgress]; ok {
		return true
	}
	if val, ok := message.Options["status"]; ok && strings.ToLower(val) == statusInProgress {
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
