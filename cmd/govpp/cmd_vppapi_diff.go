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

// TODO:
//  - table format for differences
//  - option to exit with non-zero status on breaking changes

const exampleVppApiDiffCommand = `
  <cyan># Compare VPP API in current directory against master on upstream</>
  govpp vppapi compare --against https://github.com/FDio/vpp.git

  <cyan># Compare only specific differences of VPP API schemas</>
  govpp vppapi compare --against https://github.com/FDio/vpp.git --differences=FileVersion,FileCRC
  govpp vppapi compare --against https://github.com/FDio/vpp.git --differences=MessageAdded

  <cyan># List all types of differences</>
  govpp vppapi lint --list-differences
`

type VppApiDiffCmdOptions struct {
	*VppApiCmdOptions

	Format          string
	Against         string
	Differences     []string
	CommentDiffs    bool
	ListDifferences bool
}

func newVppApiDiffCmd(cli Cli, vppapiOpts *VppApiCmdOptions) *cobra.Command {
	var (
		opts = VppApiDiffCmdOptions{VppApiCmdOptions: vppapiOpts}
	)
	cmd := &cobra.Command{
		Use:     "compare [INPUT] --against AGAINST [--differences DIFF]... | [--list-differences]",
		Aliases: []string{"cmp", "diff", "comp"},
		Short:   "Compare VPP API schemas",
		Long:    "Compares two VPP API schemas and lists the differences.",
		Example: color.Sprint(exampleVppApiDiffCommand),
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !opts.ListDifferences {
				must(cmd.MarkPersistentFlagRequired("against"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Input = args[0]
			}
			return runVppApiDiffCmd(cli.Out(), opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.Format, "format", "f", "", "Format for the output (json, yaml, go-template..)")
	cmd.PersistentFlags().StringVar(&opts.Against, "against", "", "The VPP API schema to compare against.")
	cmd.PersistentFlags().BoolVar(&opts.CommentDiffs, "comments", false, "Include message comment differences")
	cmd.PersistentFlags().StringSliceVarP(&opts.Differences, "differences", "d", nil, "List only specific differences")
	cmd.PersistentFlags().BoolVar(&opts.ListDifferences, "list-differences", false, "List all types of differences")
	cmd.MarkFlagsMutuallyExclusive("list-differences", "against")

	return cmd
}

func runVppApiDiffCmd(out io.Writer, opts VppApiDiffCmdOptions) error {
	if opts.ListDifferences {
		diffs := defaultDifferenceTypes
		if opts.Format == "" {
			return printDiffsAsTable(out, diffs)
		} else {
			return formatAsTemplate(out, opts.Format, diffs)
		}
	}

	vppInput, err := resolveVppInput(opts.Input)
	if err != nil {
		return err
	}

	vppAgainst, err := resolveVppInput(opts.Against)
	if err != nil {
		return fmt.Errorf("resolving --against failed: %w", err)
	}
	logrus.Tracef("VPP against:\n - API dir: %s\n - VPP Version: %s\n - Files: %v",
		vppAgainst.ApiDirectory, vppAgainst.Schema.Version, len(vppAgainst.Schema.Files))

	// compare schemas
	schema1 := vppInput.Schema
	schema2 := vppAgainst.Schema

	diffs := CompareSchemas(&schema1, &schema2)

	if !opts.CommentDiffs {
		var filtered []Difference
		for _, diff := range diffs {
			if diff.Type != MessageCommentDifference {
				filtered = append(filtered, diff)
			}
		}
		diffs = filtered
	}

	if len(opts.Differences) > 0 {
		diffs, err = filterDiffs(diffs, opts.Differences)
		if err != nil {
			return err
		}
	}

	if opts.Format == "" {
		printDifferencesSimple(out, diffs)
	} else {
		return formatAsTemplate(out, opts.Format, diffs)
	}

	return nil
}

func printDiffsAsTable(out io.Writer, diffs []DifferenceType) error {
	table := tablewriter.NewTable(
		out,
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.Off,
				},
			},
		}),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
		tablewriter.WithRowMergeMode(tw.MergeNone),
	)
	table.Header("#", "Difference Type")
	for i, d := range diffs {
		err := table.Append(fmt.Sprint(i+1), fmt.Sprint(d))
		if err != nil {
			return err
		}
	}
	return table.Render()
}

func printDifferencesSimple(out io.Writer, diffs []Difference) {
	if len(diffs) == 0 {
		fmt.Fprintln(out, "No differences found.")
		return
	}

	var lastFile string
	fmt.Fprintf(out, "Listing %d differences:\n", len(diffs))
	for _, diff := range diffs {
		if diff.File != lastFile {
			color.Fprintf(out, " %s\n", clrDiffFile.Sprint(diff.File))
		}
		color.Fprintf(out, " - [%v] %v\n", clrWhite.Sprint(diff.Type), diff.Description)
		lastFile = diff.File
	}
}

func filterDiffs(diffs []Difference, differences []string) ([]Difference, error) {
	wantDiffs := map[DifferenceType]bool{}
	for _, d := range differences {
		var ok bool
		for _, diff := range defaultDifferenceTypes {
			if string(diff) == d {
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("unknown difference type: %q", d)
		}
		wantDiffs[DifferenceType(d)] = true
	}
	var list []Difference
	for _, d := range diffs {
		if want := wantDiffs[d.Type]; want {
			list = append(list, d)
		}
	}
	return list, nil
}
