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
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/binapigen"
	"go.fd.io/govpp/binapigen/vppapi"
)

type GenerateOptions struct {
	Input         string
	Output        string
	ImportPrefix  string
	NoVersionInfo bool
	NoSourceInfo  bool
	RunPlugins    []string
}

func newGenerateCmd(cli Cli) *cobra.Command {
	var (
		opts = GenerateOptions{
			RunPlugins: []string{"rpc"},
		}
	)
	cmd := &cobra.Command{
		Use:     "generate [apifile...]",
		Aliases: []string{"gen"},
		Short:   "Generate code",
		Long:    "Generates bindings for VPP API",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Input == "" {
				opts.Input = detectVppApiInput()
			}
			return runGenerator(opts, args)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.Input, "input", "", "Input for VPP API (e.g. VPP API dir, VPP repository, ..)")
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "Output location for generated files")
	cmd.PersistentFlags().StringVar(&opts.ImportPrefix, "import-prefix", "", "Prefix imports in the generated go code. \nE.g. other API Files (e.g. api_file.ba.go) will be imported with :\nimport (\n  api_file \"<import-prefix>/api_file\"\n)")
	cmd.PersistentFlags().BoolVar(&opts.NoVersionInfo, "no-version-info", false, "Exclude version info.")
	cmd.PersistentFlags().BoolVar(&opts.NoSourceInfo, "no-source-info", false, "Exclude source info.")
	cmd.PersistentFlags().StringSliceVar(&opts.RunPlugins, "gen", opts.RunPlugins, "List of generator plugins to run for files.")

	return cmd
}

const DefaultOutputDir = "binapi"

func runGenerator(cmdOpts GenerateOptions, args []string) error {

	var filesToGen []string
	filesToGen = append(filesToGen, args...)

	opts := binapigen.Options{
		ImportPrefix:     cmdOpts.ImportPrefix,
		OutputDir:        cmdOpts.Output,
		NoVersionInfo:    cmdOpts.NoVersionInfo,
		NoSourcePathInfo: cmdOpts.NoSourceInfo,
		GenerateFiles:    filesToGen,
	}

	// generate in same directory when current dir is binapi
	if opts.OutputDir == "" {
		opts.OutputDir = DefaultOutputDir
	}
	if wd, _ := os.Getwd(); opts.OutputDir == DefaultOutputDir && filepath.Base(wd) == DefaultOutputDir {
		opts.OutputDir = "."
	}

	vppInput, err := vppapi.ResolveVppInput(cmdOpts.Input)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("resolved VPP input: %+v", vppInput)

	genPlugins := cmdOpts.RunPlugins
	binapigen.Run(vppInput, opts, binapigen.GeneratePlugins(genPlugins))

	return nil
}
