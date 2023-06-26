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
	"io"

	"github.com/docker/cli/cli/streams"
	"github.com/moby/term"
)

type Cli interface {
	Out() *streams.Out
	Err() io.Writer
	In() *streams.In
	Apply(...CliOption) error
}

type AppCli struct {
	out *streams.Out
	err io.Writer
	in  *streams.In
}

func NewCli(opt ...CliOption) (Cli, error) {
	cli := new(AppCli)
	if err := cli.Apply(opt...); err != nil {
		return nil, err
	}
	if cli.out == nil || cli.in == nil || cli.err == nil {
		stdin, stdout, stderr := term.StdStreams()
		if cli.in == nil {
			cli.in = streams.NewIn(stdin)
		}
		if cli.out == nil {
			cli.out = streams.NewOut(stdout)
		}
		if cli.err == nil {
			cli.err = stderr
		}
	}
	return cli, nil
}

func (cli *AppCli) Out() *streams.Out {
	return cli.out
}

func (cli *AppCli) Err() io.Writer {
	return cli.err
}

func (cli *AppCli) In() *streams.In {
	return cli.in
}

func (cli *AppCli) Apply(opt ...CliOption) error {
	for _, o := range opt {
		if err := o(cli); err != nil {
			return err
		}
	}
	return nil
}

type CliOption func(cli *AppCli) error

// WithStandardStreams sets a cli in, out and err streams with the standard streams.
func WithStandardStreams() CliOption {
	return func(cli *AppCli) error {
		// Set terminal emulation based on platform as required.
		stdin, stdout, stderr := term.StdStreams()
		cli.in = streams.NewIn(stdin)
		cli.out = streams.NewOut(stdout)
		cli.err = stderr
		return nil
	}
}

// WithCombinedStreams uses the same stream for the output and error streams.
func WithCombinedStreams(combined io.Writer) CliOption {
	return func(cli *AppCli) error {
		cli.out = streams.NewOut(combined)
		cli.err = combined
		return nil
	}
}

// WithInputStream sets a cli input stream.
func WithInputStream(in io.ReadCloser) CliOption {
	return func(cli *AppCli) error {
		cli.in = streams.NewIn(in)
		return nil
	}
}

// WithOutputStream sets a cli output stream.
func WithOutputStream(out io.Writer) CliOption {
	return func(cli *AppCli) error {
		cli.out = streams.NewOut(out)
		return nil
	}
}

// WithErrorStream sets a cli error stream.
func WithErrorStream(err io.Writer) CliOption {
	return func(cli *AppCli) error {
		cli.err = err
		return nil
	}
}
