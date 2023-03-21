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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/version"
)

const logo = `
 _________     ___    _________________
 __  ____/_______ |  / /__  __ \__  __ \
 _  / __ _  __ \_ | / /__  /_/ /_  /_/ /  %s
 / /_/ / / /_/ /_ |/ / _  ____/_  ____/   %s
 \____/  \____/_____/  /_/     /_/        %s
`

func Execute() {
	root := newRootCmd()

	if err := root.Execute(); err != nil {
		logrus.Fatalf("ERROR: %v", err)
	}
}

func newRootCmd() *cobra.Command {
	var (
		glob GlobalOptions
	)
	cmd := &cobra.Command{
		Use:               "govpp",
		Short:             "GoVPP CLI",
		Long:              fmt.Sprintf(logo, version.Short(), version.BuildTime(), version.BuiltBy()),
		Version:           version.String(),
		SilenceUsage:      true,
		SilenceErrors:     true,
		TraverseChildren:  true,
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			InitOptions(glob)

			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.PersistentFlags().SortFlags = false

	// Global options
	glob.InstallFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		newGenerateCmd(),
		newVppapiCmd(),
		newServerCmd(),
		newCliCommand(),
		//newExportCmd(),
		newDiffCmd(),
	)

	// Help
	cmd.InitDefaultHelpFlag()
	cmd.Flags().Lookup("help").Hidden = true
	cmd.InitDefaultHelpCmd()
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			c.Hidden = true
		}
	}

	// Version
	cmd.InitDefaultVersionFlag()

	return cmd
}
