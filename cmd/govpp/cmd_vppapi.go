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
  <cyan># Specify input source forl VPP API</>
  govpp vppapi COMMAND [INPUT]
  govpp vppapi COMMAND ./vpp
  govpp vppapi COMMAND /usr/share/vpp/api
  govpp vppapi COMMAND vppapi.tar.gz 
  govpp vppapi COMMAND http://github.com/FDio/vpp.git

  <cyan># List VPP API contents</>
  govpp vppapi list [INPUT]
  govpp vppapi list --show-contents

  <cyan># Export VPP API files</>
  govpp vppapi export [INPUT] --output vppapi
  govpp vppapi export [INPUT] --output vppapi.tar.gz

  <cyan># Lint VPP API definitions</>
  govpp vppapi lint [INPUT]

  <cyan># Compare VPP API schemas</>
  govpp vppapi diff [INPUT] --against http://github.com/FDio/vpp.git
`

type VppApiCmdOptions struct {
	Input string
	Paths []string
}

func newVppapiCmd(cli Cli) *cobra.Command {
	var (
		opts VppApiCmdOptions
	)
	cmd := &cobra.Command{
		Use:   "vppapi",
		Short: "Manage VPP API development",
		Long: "Manage the development of VPP API using basic commands to list, browse and export API files,\n" +
			"or with more advanced commands to lint API definitions or compare different API versions.",
		Example:          color.Sprint(exampleVppapi),
		TraverseChildren: true,
	}

	cmd.PersistentFlags().StringSliceVar(&opts.Paths, "path", nil, "Limit to specific files or directories.\n"+
		"For example: \"vpe\" or \"core/\".\n"+
		"If specified multiple times, the union is taken.")

	cmd.AddCommand(
		newVppApiLsCmd(cli, &opts),
		newVppApiDiffCmd(cli, &opts),
		newVppApiExportCmd(cli, &opts),
		newVppApiLintCmd(cli, &opts),
	)

	return cmd
}

func prepareVppApiFiles(allapifiles []vppapi.File, paths []string, includeImported, sortByName bool) ([]vppapi.File, error) {
	logrus.Tracef("preparing %d files", len(allapifiles))

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
			return nil, fmt.Errorf("no files matched paths: %q", paths)
		}
		logrus.Tracef("%d/%d files matched paths: %q", len(apifiles), len(allapifiles), paths)
	}

	// sort files
	if sortByName {
		binapigen.SortFilesByName(apifiles)
	}

	return apifiles, nil
}
