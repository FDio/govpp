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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen/vppapi"
)

// TODO:
//  - table format for differences
//  - option to exit with non-zero status on breaking changes

type VppApiDiffCmdOptions struct {
	*VppApiCmdOptions

	Against     string
	Differences []string
}

func newVppApiDiffCmd(vppapiOpts *VppApiCmdOptions) *cobra.Command {
	var (
		opts = VppApiDiffCmdOptions{VppApiCmdOptions: vppapiOpts}
	)
	cmd := &cobra.Command{
		Use:     "diff INPUT --against=AGAINST",
		Aliases: []string{"cmp", "compare"},
		Short:   "Compare VPP API schemas",
		Long:    "Compares two VPP API schemas and lists the differences.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Input = args[0]
			}
			return runVppApiDiffCmd(cmd.OutOrStdout(), opts)
		},
	}

	cmd.PersistentFlags().StringSliceVar(&opts.Differences, "differences", nil, "List only specific differences")
	cmd.PersistentFlags().StringVar(&opts.Against, "against", "", "The VPP API schema to compare against.")
	must(cobra.MarkFlagRequired(cmd.PersistentFlags(), "against"))

	return cmd
}

var (
	clrWhite    = color.Style{color.White}
	clrCyan     = color.Style{color.Cyan}
	clrDiffFile = color.Style{color.Yellow}
)

func runVppApiDiffCmd(out io.Writer, opts VppApiDiffCmdOptions) error {
	if opts.Format != "" {
		color.Disable()
	}

	vppInput, err := resolveInput(opts.Input)
	if err != nil {
		return err
	}

	vppAgainst, err := vppapi.ResolveVppInput(opts.Against)
	if err != nil {
		return err
	}
	logrus.Tracef("VPP against:\n - API dir: %s\n - VPP Version: %s\n - Files: %v",
		vppAgainst.ApiDirectory, vppAgainst.Schema.Version, len(vppAgainst.Schema.Files))

	// compare schemas
	schema1 := vppInput.Schema
	schema2 := vppAgainst.Schema

	logrus.Tracef("comparing schemas:\n\tSCHEMA 1: %+v\n\tSCHEMA 2: %+v\n", schema1, schema2)

	diffs := CompareSchemas(&schema1, &schema2)

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

func printDifferencesSimple(out io.Writer, diffs []Difference) {
	if len(diffs) == 0 {
		fmt.Fprintln(out, "No differences found.")
		return
	}

	fmt.Fprintf(out, "Listing %d differences:\n", len(diffs))
	for _, d := range diffs {
		color.Fprintf(out, " - %s\n", d)
	}
}

func filterDiffs(diffs []Difference, differences []string) ([]Difference, error) {
	wantDiffs := map[DifferenceType]bool{}
	for _, d := range differences {
		var ok bool
		for _, diff := range differenceTypes {
			if string(diff) == d {
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("unknown difference type: %s", d)
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
