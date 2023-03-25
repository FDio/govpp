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

	"github.com/spf13/cobra"
)

type ExportCmdOptions struct {
	Input  string
	Output string
}

func newExportCmd() *cobra.Command {
	var (
		opts = ExportCmdOptions{}
	)
	cmd := &cobra.Command{
		Use:     "export [apifile...]",
		Aliases: []string{"gen"},
		Short:   "Export VPP API schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Input == "" {
				opts.Input = resolveVppApiInput()
			}
			return runExportCmd(opts, args)
		},
		Hidden: true,
	}

	cmd.PersistentFlags().StringVar(&opts.Input, "input", "", "Input for VPP API (e.g. path to VPP API directory, local VPP repo)")
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "Output location for generated files")

	return cmd
}

func runExportCmd(opts ExportCmdOptions, args []string) error {
	// TODO: implement this
	return fmt.Errorf("not implemented")
}
