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
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/binapi/vlib"
	"go.fd.io/govpp/core"
)

// TODO:
//  - add option to allow connecting via CLI socket
//  - try several ways to connect to VPP if not specified

const exampleCliCommand = `
  <cyan># Execute 'show version' command</>
  govpp cli show version

  <cyan># Enter REPL mode to send commands interactively</>
  govpp cli

  <cyan># Read CLI command(s) from stdin</>
  echo "show errors" | govpp cli

  <cyan># Execute commands and write output to file</>
  govpp cli -o cli.log show version
`

type CliOptions struct {
	ApiSocket string
	Force     bool
	Output    string

	Stdin  io.Reader
	Stderr io.Writer
	Stdout io.Writer
}

func newCliCommand(cli Cli) *cobra.Command {
	var (
		opts = CliOptions{
			ApiSocket: socketclient.DefaultSocketName,
		}
	)
	cmd := &cobra.Command{
		Use:                   "cli [COMMAND]",
		Aliases:               []string{"c"},
		Short:                 "Send CLI via VPP API",
		Long:                  "Send VPP CLI command(s) via VPP API",
		Example:               color.Sprint(exampleCliCommand),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Stdin = cli.In()
			opts.Stderr = cli.Err()

			// Setup output
			if opts.Output != "" {
				info, err := os.Stat(opts.Output)
				if err != nil && !os.IsNotExist(err) {
					return err
				} else if err == nil {
					if info.IsDir() {
						return fmt.Errorf("output cannot be a directory")
					}
					if !opts.Force {
						return fmt.Errorf("output file already exists (use --force to overwrite)")
					}
				}
				file, err := os.Create(opts.Output)
				if err != nil {
					return fmt.Errorf("failed to create output file: %v", err)
				}
				opts.Stdout = file
			} else {
				opts.Stdout = cmd.OutOrStdout()
			}

			var cmds []string

			// Check if there is any data sent to the program via stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				logrus.Debugf("stdin has input")

				if len(args) > 0 {
					return fmt.Errorf("cannot use both arguments and stdin for the CLI commands")
				}

				input, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("error reading stdin: %w", err)
				}

				// Split the input into lines and execute each line as a command
				lines := strings.Split(string(input), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					cmds = append(cmds, line)
				}
			} else if len(args) > 0 {
				logrus.Debugf("provided %d args", len(args))

				cmdArgs := strings.Join(args[:], " ")
				lines := strings.Split(cmdArgs, "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					cmds = append(cmds, line)
				}
			}

			return runCliCmd(opts, cmds)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.ApiSocket, "apisock", opts.ApiSocket, "Path to VPP API socket")
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "Output location for the CLI reply")
	cmd.PersistentFlags().BoolVarP(&opts.Force, "force", "f", false, "Force overwriting output file")

	return cmd
}

const vppcliPrompt = "vpp# "

func runCliCmd(opts CliOptions, cmds []string) (err error) {
	var cli VppCli

	cli, err = connectBinapiVppCLI(opts.ApiSocket)
	if err != nil {
		return fmt.Errorf("connecting to CLI failed: %w", err)
	}
	defer cli.Close()

	if len(cmds) == 0 {
		logrus.Debugf("entering REPL mode")

		fmt.Fprintln(opts.Stdout)

		scanner := bufio.NewScanner(opts.Stdin)
		for {
			fmt.Fprint(opts.Stdout, vppcliPrompt)

			if !scanner.Scan() {
				return nil
			}
			line := scanner.Text()
			if line == "" {
				continue
			}

			logrus.Debugf("executing CLI command: %v", line)

			reply, err := cli.Execute(line)
			if err != nil {
				logrus.Errorf("command error: %v", err)
			}

			fmt.Fprint(opts.Stdout, reply)
			fmt.Fprintln(opts.Stdout)
		}
	} else {
		for _, cmd := range cmds {
			fmt.Fprintln(opts.Stdout, vppcliPrompt+cmd)

			logrus.Debugf("executing CLI command: %v", cmd)

			reply, err := cli.Execute(cmd)
			if err != nil {
				return err
			}

			fmt.Fprint(opts.Stdout, reply)
			fmt.Fprintln(opts.Stdout)
		}
	}

	return nil
}

func connectBinapiVppCLI(apiSock string) (VppCli, error) {
	cli, err := newBinapiVppCli(apiSock)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

type VppCli interface {
	Execute(args ...string) (string, error)
	Close() error
}

type vppcliBinapi struct {
	apiSocket string
	conn      *core.Connection
	client    vlib.RPCService
}

func newBinapiVppCli(apiSock string) (*vppcliBinapi, error) {
	logrus.Tracef("connecting to VPP API socket %q", apiSock)

	conn, err := govpp.Connect(apiSock)
	if err != nil {
		return nil, fmt.Errorf("connecting to VPP failed: %w", err)
	}

	ch, err := conn.NewAPIChannel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	if err := ch.CheckCompatiblity(vlib.AllMessages()...); err != nil {
		return nil, fmt.Errorf("compatibility check failed: %w", err)
	}

	cli := &vppcliBinapi{
		apiSocket: apiSock,
		conn:      conn,
		client:    vlib.NewServiceClient(conn),
	}
	return cli, nil
}

func (b *vppcliBinapi) Execute(args ...string) (string, error) {
	cmd := strings.Join(args, " ")

	logrus.Tracef("sending CLI command: %q", cmd)

	reply, err := b.client.CliInband(context.Background(), &vlib.CliInband{
		Cmd: cmd,
	})
	if err != nil {
		return "", err
	}

	return reply.Reply, nil
}

func (b *vppcliBinapi) Close() error {
	if b.conn != nil {
		logrus.Debugf("disconnecting VPP API connection")
		b.conn.Disconnect()
		b.conn = nil
	}
	return nil
}
