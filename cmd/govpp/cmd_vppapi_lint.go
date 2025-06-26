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
	"github.com/olekukonko/tablewriter/tw"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const exampleVppApiLintCmd = `
  <cyan># Lint VPP master using default rules</>
  govpp vppapi lint https://github.com/FDio/vpp

  <cyan># Lint using only specific rules</>
  govpp vppapi lint https://github.com/FDio/vpp --rules=MESSAGE_DEPRECATE_OLDER_VERSIONS

  <cyan># List all linter rules</>
  govpp vppapi lint --list-rules
`

type VppApiLintCmdOptions struct {
	*VppApiCmdOptions

	Format    string
	Rules     []string
	Except    []string
	ExitCode  bool
	ListRules bool
}

func newVppApiLintCmd(cli Cli, vppapiOpts *VppApiCmdOptions) *cobra.Command {
	var (
		opts = VppApiLintCmdOptions{VppApiCmdOptions: vppapiOpts}
	)
	cmd := &cobra.Command{
		Use:     "lint [INPUT] [--rules RULE]... [--except RULE]... [--exit-code] | [--list-rules]",
		Short:   "Lint VPP API definitions",
		Long:    "Lint VPP API definitions by running linter with rule checks to detect any violations.",
		Example: color.Sprint(exampleVppApiLintCmd),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Input = args[0]
			}
			return runVppApiLintCmd(cli.Out(), opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.Format, "format", "f", "", "Format for the output (json, yaml, go-template..)")
	cmd.PersistentFlags().StringSliceVar(&opts.Rules, "rules", nil, "Check only specific rules")
	cmd.PersistentFlags().StringSliceVar(&opts.Except, "except", nil, "Exclude specific rules")
	cmd.PersistentFlags().BoolVar(&opts.ExitCode, "exit-code", false, "Exit with non-zero exit code if any issue is found")
	cmd.PersistentFlags().BoolVar(&opts.ListRules, "list-rules", false, "List all known linter rules")

	return cmd
}

func runVppApiLintCmd(out io.Writer, opts VppApiLintCmdOptions) error {
	if opts.ListRules {
		rules := ListLintRules(defaultLintRules...)
		if opts.Format == "" {
			return printLintRulesAsTable(out, rules)
		} else {
			return formatAsTemplate(out, opts.Format, rules)
		}
	}

	vppInput, err := resolveVppInput(opts.Input)
	if err != nil {
		return err
	}

	schema := vppInput.Schema

	apifiles, err := prepareVppApiFiles(schema.Files, opts.Paths, false, true)
	if err != nil {
		return err
	}
	schema.Files = apifiles

	linter := NewLinter()

	if len(opts.Rules) > 0 {
		logrus.Debugf("setting rules to: %v", opts.Rules)
		linter.SetRules(opts.Rules)
	}
	if len(opts.Except) > 0 {
		logrus.Debugf("disabling rules: %v", opts.Except)
		linter.Disable(opts.Except...)
	}

	lintIssues, err := linter.Lint(&schema)
	if err != nil {
		return fmt.Errorf("linter failure: %w", err)
	}

	if opts.Format == "" {
		err := printLintErrorsAsTable(out, lintIssues)
		if err != nil {
			return err
		}
	} else {
		if err := formatAsTemplate(out, opts.Format, lintIssues); err != nil {
			return err
		}
	}

	if len(lintIssues) > 0 {
		if opts.ExitCode {
			return fmt.Errorf("found %d issues", len(lintIssues))
		} else {
			logrus.Errorf("found %d issues", len(lintIssues))
		}
	}

	return nil
}

func printLintRulesAsTable(out io.Writer, rules []*LintRule) error {
	table := tablewriter.NewTable(
		out,
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenRows: tw.Off},
			},
		}),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
		tablewriter.WithRowMergeMode(tw.MergeNone),
	)
	table.Header("#", "Id", "Purpose")
	for i, r := range rules {
		err := table.Append(fmt.Sprint(i+1), r.Id, r.Purpose)
		if err != nil {
			return err
		}
	}
	return table.Render()
}

func printLintErrorsAsTable(out io.Writer, issues LintIssues) error {
	if len(issues) == 0 {
		fmt.Fprintln(out, "No issues found")
		return nil
	}

	table := tablewriter.NewTable(
		out,
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenRows: tw.On},
			},
		}),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
		tablewriter.WithRowMergeMode(tw.MergeVertical),
	)

	table.Header("#", "Rule", "Location", "Violation")
	for i, e := range issues {
		loc := e.File
		if e.Line > 0 {
			loc += fmt.Sprintf(":%d", e.Line)
		}
		err := table.Append(fmt.Sprint(i+1), e.RuleId, loc, e.Violation)
		if err != nil {
			return err
		}
	}
	return table.Render()
}
