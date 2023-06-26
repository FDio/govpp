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

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen"
	"go.fd.io/govpp/binapigen/vppapi"
)

const exampleVppapi = `
  <gray># Specify input source of VPP API</>
  govpp vppapi COMMAND [INPUT]
  govpp vppapi COMMAND ./vpp
  govpp vppapi COMMAND /usr/share/vpp/api
  govpp vppapi COMMAND vppapi.tar.gz 
  govpp vppapi COMMAND http://github.com/FDio/vpp.git

  <gray># List VPP API contents</>
  govpp vppapi ls [INPUT]
  govpp vppapi ls --show-contents

  <gray># Export VPP API files</>
  govpp vppapi export [INPUT] --output vppapi
  govpp vppapi export [INPUT] --output vppapi.tar.gz

  <gray># Lint VPP API definitions</>
  govpp vppapi lint [INPUT]

  <gray># Compare VPP API schemas</>
  govpp vppapi diff [INPUT] --against http://github.com/FDio/vpp.git
`

type VppApiCmdOptions struct {
	Input  string
	Paths  []string
	Format string
}

func newVppapiCmd(cli Cli) *cobra.Command {
	var (
		opts VppApiCmdOptions
	)
	cmd := &cobra.Command{
		Use:              "vppapi",
		Short:            "Manage VPP API",
		Long:             "Manage VPP API development.",
		Example:          color.Sprint(exampleVppapi),
		TraverseChildren: true,
	}

	cmd.PersistentFlags().StringSliceVar(&opts.Paths, "path", nil, "Limit to specific files or directories.\nFor example: \"vpe\" or \"core/\".")
	cmd.PersistentFlags().StringVarP(&opts.Format, "format", "f", "", "Format for the output (json, yaml, go-template..)")

	cmd.AddCommand(
		newVppApiLsCmd(cli, &opts),
		newVppApiExportCmd(cli, &opts),
		newVppApiDiffCmd(cli, &opts),
		newVppApiLintCmd(cli, &opts),
	)

	return cmd
}

func prepareVppApiFiles(allapifiles []vppapi.File, paths []string, includeImported, sortByName bool) ([]vppapi.File, error) {
	// remove imported types
	if !includeImported {
		binapigen.SortFilesByImports(allapifiles)
		for i, apifile := range allapifiles {
			f := apifile
			binapigen.RemoveImportedTypes(allapifiles, &f)
			allapifiles[i] = f
		}
	}

	apifiles := allapifiles

	// filter files
	if len(paths) > 0 {
		apifiles = filterFilesByPaths(allapifiles, paths)
		if len(apifiles) == 0 {
			return nil, fmt.Errorf("no files matching: %q", paths)
		}
		logrus.Tracef("filter (%d paths) matched %d/%d files", len(paths), len(apifiles), len(allapifiles))
	}

	if sortByName {
		binapigen.SortFilesByName(apifiles)
	}

	return apifiles, nil
}
