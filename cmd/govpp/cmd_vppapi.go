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
	"github.com/spf13/cobra"
)

const exampleVppapi = `
 # Specify VPP API input source
 govpp vppapi --input=. COMMAND
 govpp vppapi --input=http://github.com/FDio/vpp.git COMMAND
 govpp vppapi --input=vppapi.tar.gz COMMAND

 # Browse VPP API files
 govpp vppapi --input=INPUT ls
 govpp vppapi --input=INPUT ls --show-messages

 # Export VPP API files
 govpp vppapi --input=INPUT export --output=vppapi 

 # Lint VPP API
 govpp vppapi --input=INPUT lint

 # Compare VPPPI schemas
 govpp vppapi -input=INPUT diff --against=http://github.com/FDio/vpp.git
`

type VppApiCmdOptions struct {
	Input  string
	Format string
}

func newVppapiCmd() *cobra.Command {
	var (
		opts VppApiCmdOptions
	)
	cmd := &cobra.Command{
		Use:              "vppapi",
		Short:            "Manage VPP API",
		Long:             "Manage VPP API development",
		Example:          exampleVppapi,
		TraverseChildren: true,
	}

	cmd.PersistentFlags().StringVar(&opts.Input, "input", opts.Input, "Input for VPP API (e.g. path to VPP API directory, local VPP repo)")
	cmd.PersistentFlags().StringVar(&opts.Format, "format", "", "Output format (json, yaml, go-template..)")

	cmd.AddCommand(
		newVppApiLsCmd(&opts),
		newVppApiExportCmd(&opts),
		newVppApiDiffCmd(&opts),
		newVppApiLintCmd(&opts),
	)

	return cmd
}
