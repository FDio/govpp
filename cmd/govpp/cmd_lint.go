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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

type LintCmdOptions struct {
	Input string
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
			return runLintCmd(opts, args)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.Input, "input", "", "Input for VPP API (e.g. path to VPP API directory, local VPP repo)")

	return cmd
}

func runLintCmd(opts LintCmdOptions, args []string) error {
	vppInput, err := resolveInput(opts.Input)
	if err != nil {
		return err
	}

	checks := defautLintChecks()

	logrus.Tracef("running %d linter checks", len(checks))

	schema := vppInput.Schema

	err = checkSchema(&schema, checks)
	if err != nil {
		return err
	}

	return nil
}

/*type Linter interface {
	Enable(rules ...*LintRule)
	Disable(rules ...*LintRule)
	Run(schema *vppapi.Schema) error
}

var rules = []LintRule{
	{
		Code:        "F000",
		Name:        "missing-crc",
		Description: "Must have CRC defined",
		Category:    "file",
		Check: func(schema *vppapi.Schema) error {
			return nil
		},
	},
}*/

func checkSchema(schema *vppapi.Schema, checks []LintChecker) error {
	var issues LintErrors

	for _, check := range checks {
		if err := check.Check(schema); err != nil {
			switch e := err.(type) {
			case LintError:
				issues = append(issues, e)
			case LintErrors:
				issues = append(issues, e...)
			default:
				issues = append(issues, LintError{
					Message: err.Error(),
				})
			}
		}
	}

	if len(issues) > 0 {
		logrus.Debugf("Found %d issues in the VPP API schema", len(issues))
		return issues
	}
	logrus.Debugln("No issues found in the VPP API schema")
	return nil
}

type LintChecker interface {
	Check(schema *vppapi.Schema) error
}

type CheckFunc func(schema *vppapi.Schema) error

func (c CheckFunc) Check(schema *vppapi.Schema) error {
	return c(schema)
}

func defautLintChecks() []LintChecker {
	return []LintChecker{
		CheckFunc(CheckMissingCRC),
		CheckFunc(CheckDeprecatedMessages),
	}
}

type LintRule struct {
	Code        string
	Name        string
	Description string
	Category    string
	Check       func(schema *vppapi.Schema) error
}

type LintError struct {
	Rule    LintRule
	File    string
	Line    int
	Object  any
	Message string
}

func (l LintError) Error() string {
	if l.Line == 0 {
		return fmt.Sprintf("%s:%v ", l.File, l.Message)
	}
	return fmt.Sprintf("%s:%d:%v ", l.File, l.Line, l.Message)
}

type LintErrors []LintError

func (le LintErrors) Error() string {
	var sb strings.Builder
	for _, e := range le {
		sb.WriteString(e.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

func CheckMissingCRC(schema *vppapi.Schema) error {
	var issues LintErrors
	for _, file := range schema.Files {
		if file.CRC == "" {
			issues = append(issues, LintError{
				File:    file.Path,
				Message: fmt.Sprintf("CRC is missing for file: %s", file.Name),
			})
		}
	}
	if len(issues) > 0 {
		return issues
	}
	return nil
}

func CheckDeprecatedMessages(schema *vppapi.Schema) error {
	var issues LintErrors
	for _, file := range schema.Files {
		messageVersions := extractFileMessageVersions(file)
		for _, message := range file.Messages {
			baseName, version := extractBaseNameAndVersion(message.Name)
			if version < messageVersions[baseName] {
				if _, ok := message.Options["in_progress"]; ok {
					continue
				}
				if _, ok := message.Options["deprecated"]; !ok {
					issues = append(issues, LintError{
						File:    file.Path,
						Message: fmt.Sprintf("Message %s is missing the deprecated option for older version", message.Name),
					})
				}
			}
		}
	}
	if len(issues) > 0 {
		return issues
	}
	return nil
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

func extractFileMessageVersions(file vppapi.File) map[string]int {
	messageVersions := make(map[string]int)
	for _, message := range file.Messages {
		baseName, version := extractBaseNameAndVersion(message.Name)
		if version > messageVersions[baseName] {
			messageVersions[baseName] = version
		}
	}
	return messageVersions
}

/*func extractMessageVersions(file vppapi.File) map[string][]string {
	messageVersions := make(map[string][]string)
	for _, message := range file.Messages {
		baseName, _ := extractBaseNameAndVersion(message.Name)
		messageVersions[baseName] = append(messageVersions[baseName], message.Name)
	}
	return messageVersions
}*/
