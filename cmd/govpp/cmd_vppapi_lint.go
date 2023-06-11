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

	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// TODO:
//  - support file filter (include, exclude..)
//  - consider adding categories for linter rules

type VppApiLintCmdOptions struct {
	*VppApiCmdOptions

	Rules     []string
	Except    []string
	ExitCode  bool
	ListRules bool
}

func newVppApiLintCmd(vppapiOpts *VppApiCmdOptions) *cobra.Command {
	var (
		opts = VppApiLintCmdOptions{VppApiCmdOptions: vppapiOpts}
	)
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint VPP API files",
		Long:  "Run linter checks for VPP API files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVppApiLintCmd(cmd.OutOrStdout(), opts)
		},
	}

	cmd.PersistentFlags().StringSliceVar(&opts.Rules, "rules", nil, "Limit to specific linter rules")
	cmd.PersistentFlags().StringSliceVar(&opts.Except, "except", nil, "Skip specific linter rules.")
	cmd.PersistentFlags().StringVarP(&opts.Format, "format", "f", "", "The format of the output")
	cmd.PersistentFlags().BoolVar(&opts.ExitCode, "exit-code", false, "Exit with non-zero exit code if any issue is found")
	cmd.PersistentFlags().BoolVar(&opts.ListRules, "list-rules", false, "List all known linter rules")

	return cmd
}

func runVppApiLintCmd(out io.Writer, opts VppApiLintCmdOptions) error {
	if opts.Format != "" {
		color.Disable()
	}

	if opts.ListRules {
		rules := LintRules(defaultLintRules...)
		if opts.Format == "" {
			printLintRulesAsTable(out, rules)
		} else {
			return formatAsTemplate(out, opts.Format, rules)
		}
		return nil
	}

	vppInput, err := resolveInput(opts.Input)
	if err != nil {
		return err
	}

	linter := NewLinter()

	if len(opts.Rules) > 0 {
		logrus.Debugf("setting lint rules to: %v", opts.Rules)
		linter.SetRules(opts.Rules)
	}
	if len(opts.Except) > 0 {
		logrus.Debugf("disabling lint rules: %v", opts.Except)
		linter.Disable(opts.Except...)
	}

	schema := vppInput.Schema

	if err := linter.Lint(&schema); err != nil {
		if errs, ok := err.(LintIssues); ok {
			if opts.Format == "" {
				printLintErrorsAsTable(out, errs)
			} else {
				return formatAsTemplate(out, opts.Format, errs)
			}
			if opts.ExitCode {
				return err
			} else {
				logrus.Errorln("Linter found:", err)
			}
		} else {
			return fmt.Errorf("linter failure: %w", err)
		}
	} else {
		if opts.Format == "" {
			fmt.Fprintln(out, "No issues found")
		} else {
			return formatAsTemplate(out, opts.Format, nil)
		}
	}

	return nil
}

func printLintRulesAsTable(out io.Writer, rules []*LintRule) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"#", "Id", "Purpose",
	})
	table.SetAutoMergeCells(false)
	table.SetAutoWrapText(false)
	table.SetRowLine(false)
	table.SetBorder(false)
	for i, r := range rules {
		index := i + 1
		table.Append([]string{
			fmt.Sprint(index), r.Id, r.Purpose,
		})
	}
	table.Render()
}

func printLintErrorsAsTable(out io.Writer, errs LintIssues) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"#", "Rule", "Location", "Violation",
	})
	table.SetAutoMergeCells(true)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetBorder(false)
	for i, e := range errs {
		index := i + 1
		loc := e.File
		if e.Line > 0 {
			loc += fmt.Sprintf(":%d", e.Line)
		}
		table.Append([]string{
			fmt.Sprint(index), e.RuleId, loc, e.Violation,
		})
	}
	table.Render()
}
