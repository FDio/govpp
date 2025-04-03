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
		Use:   "govpp [OPTIONS] COMMAND",
		Short: "GoVPP CLI tool",
		Long: color.Sprintf(logo, version.Short(), version.BuiltBy(), version.BuildTime()) + "\n" +
			"GoVPP is an universal CLI tool for any VPP-related development.",
		Version: version.Version(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			InitOptions(cli, &glob)
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		//TraverseChildren: true,
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

	cobra.EnableCommandSorting = false

	// Commands
	cmd.AddCommand(
		newCliCommand(cli),
		newGenerateCmd(cli),
		newHttpCmd(cli),
		newVppapiCmd(cli),
	)

	forAllCommands(cmd, func(c *cobra.Command) {
		c.Flags().SortFlags = false
	})

	// Help command
	cmd.InitDefaultHelpCmd()
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			c.Hidden = true
		}
	}
	// Completion command
	cmd.InitDefaultCompletionCmd()

	cmd.SetUsageTemplate(color.Sprint(usageTemplate))

	return cmd
}

func forAllCommands(cmd *cobra.Command, f func(c *cobra.Command)) {
	for _, c := range cmd.Commands() {
		forAllCommands(c, f)
	}
	f(cmd)
}

const usageTemplate = `<lightWhite>USAGE:</>{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

<lightWhite>ALIASES:</>
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

<lightWhite>EXAMPLES:</>
{{- trimRightSpace .Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

<lightWhite>COMMANDS:</>{{range $cmds}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

<lightWhite>ADDITIONAL COMMANDS:</>{{range $cmds}}{{if (and (eq .GroupID "") .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

<lightWhite>OPTIONS:</>
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

<lightWhite>GLOBAL OPTIONS:</>
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

<lightWhite>TOPICS:</>{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
