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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

type LintCmdOptions struct {
	Input  string
	Output string
}

func newLintCmd() *cobra.Command {
	var (
		opts = LintCmdOptions{}
	)
	cmd := &cobra.Command{
		Use:     "lint [apifile...]",
		Aliases: []string{"gen"},
		Short:   "Lint VPP API files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Input == "" {
				opts.Input = resolveVppApiInput()
			}
			return runLintCmd(opts, args)
		},
		Hidden: true,
	}

	cmd.PersistentFlags().StringVar(&opts.Input, "input", "", "Input for VPP API (e.g. path to VPP API directory, local VPP repo)")
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "Output location for linter")

	return cmd
}

type LintRule struct {
	Name        string
	Code        string
	Description string
	Check       func(obj any) error
}

type LintError struct {
	Obj     any
	Message string
}

func (l LintError) Error() string {
	return l.Message
}

func runLintCmd(opts LintCmdOptions, args []string) error {
	// Input
	vppInput, err := vppapi.ResolveVppInput(opts.Input)
	if err != nil {
		return err
	}

	logrus.Tracef("VPP input:\n - API dir: %s\n - VPP Version: %s\n - Files: %v",
		vppInput.ApiDirectory, vppInput.VppVersion, len(vppInput.ApiFiles))

	checks := defautChecks()
	schema := &vppapi.Schema{
		Files:   vppInput.ApiFiles,
		Version: vppInput.VppVersion,
	}

	err = checkSchema(schema, checks)
	if err != nil {
		return err
	}

	/*var rules []LintRule

	rules := defaultRules()

	err := checkRules(vppInput, rules)
	if err != nil {
		return err
	}*/

	return nil
}

/*func checkRules(input *vppapi.VppInput, rules []LintRule) error {

	return nil
}

func defaultRules() []LintRule {
	return []LintRule{
		{
			Name:        "",
			Code:        "",
			Description: "",
			Check: func(obj any) error {

			},
		},
	}
}*/

func defautChecks() []Check {
	return []Check{
		CheckMissingCRC{},
		CheckDeprecatedMessages{},
	}
}

type Check interface {
	Check(schema *vppapi.Schema) (issuesFound bool, issueDescription string)
}

type CheckMissingCRC struct{}

func (c CheckMissingCRC) Check(schema *vppapi.Schema) (issuesFound bool, issueDescription string) {
	for _, file := range schema.Files {
		if file.CRC == "" {
			return true, fmt.Sprintf("CRC is missing for file: %s\n", file.Name)
		}
	}
	return false, ""
}

type CheckDeprecatedMessages struct{}

func (c CheckDeprecatedMessages) Check(schema *vppapi.Schema) (issuesFound bool, issueDescription string) {
	messageVersions := make(map[string]int)
	for _, file := range schema.Files {
		for _, message := range file.Messages {
			baseName, version := extractBaseNameAndVersion(message.Name)
			if version > messageVersions[baseName] {
				messageVersions[baseName] = version
			}
		}
	}

	var issueBuilder strings.Builder
	for _, file := range schema.Files {
		for _, message := range file.Messages {
			baseName, version := extractBaseNameAndVersion(message.Name)
			if version < messageVersions[baseName] {
				if _, deprecated := message.Options["deprecated"]; !deprecated {
					issuesFound = true
					issueBuilder.WriteString(fmt.Sprintf("Message %s.%s is missing the deprecated option\n", file.Name, message.Name))
				}
			}
		}
	}

	return issuesFound, issueBuilder.String()
}

func extractBaseNameAndVersion(messageName string) (string, int) {
	re := regexp.MustCompile(`^(.+)_v(\d+)$`)
	matches := re.FindStringSubmatch(messageName)

	if len(matches) == 3 {
		return matches[1], 2
	} else {
		return messageName, 1
	}
}

func checkSchema(schema *vppapi.Schema, checks []Check) error {
	var err error
	issuesFound := false

	for _, check := range checks {
		checkIssuesFound, issueDescription := check.Check(schema)
		if checkIssuesFound {
			issuesFound = true
			fmt.Print(issueDescription)
			err = fmt.Errorf("check error: %v", issueDescription)
		}
	}

	if issuesFound {
		fmt.Println("Linting issues found in the VPP API schema")
	} else {
		fmt.Println("No issues found in the VPP API schema")
	}

	return err
}
