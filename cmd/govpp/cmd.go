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
	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp/internal/version"
)

const logo = `
<fg=lightCyan;bg=black;op=bold>   ______         _    _  _____   _____   </>  <fg=lightWhite;op=bold>%s</>
<fg=lightCyan;bg=black;op=bold>  |  ____  _____   \  /  |_____] |_____]  </>  <fg=lightBlue>%s</>
<fg=lightCyan;bg=black;op=bold>  |_____| [_____]   \/   |       |        </>  <fg=blue>%s</>
<fg=lightCyan;bg=black;op=bold>                                          </> 
`

func newRootCmd(cli Cli) *cobra.Command {
	var (
		glob GlobalOptions
	)

	cmd := &cobra.Command{
		Use:     "govpp [OPTIONS] COMMAND",
		Short:   "GoVPP CLI tool",
		Long:    color.Sprintf(logo, version.Short(), version.BuiltBy(), version.BuildTime()),
		Version: version.String(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			InitOptions(cli, &glob)
			logrus.Tracef("global options: %+v", glob)
			logrus.Tracef("args: %+v", args)
			return nil
		},
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	// Setup options
	cmd.Flags().SortFlags = false
	cmd.PersistentFlags().SortFlags = false

	// Global options
	glob.InstallFlags(cmd.PersistentFlags())

	// Version option
	cmd.InitDefaultVersionFlag()
	cmd.Flags().Lookup("version").Shorthand = ""
	// Help option
	cmd.InitDefaultHelpFlag()
	cmd.Flags().Lookup("help").Hidden = true

	// Commands
	cmd.AddCommand(
		newGenerateCmd(cli),
		newVppapiCmd(cli),
		newHttpCmd(cli),
		newCliCommand(cli),
	)

	// Help command
	cmd.InitDefaultHelpCmd()
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			c.Hidden = true
		}
	}

	return cmd
}
