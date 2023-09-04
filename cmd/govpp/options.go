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
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

const (
	defaultLogLevel = logrus.InfoLevel
)

type GlobalOptions struct {
	Debug    bool
	LogLevel string
	Color    string
}

func (glob *GlobalOptions) InstallFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&glob.Debug, "debug", "D", false, "Enable debug mode")
	flags.StringVarP(&glob.LogLevel, "log-level", "L", "", "Set the logging level [trace/debug/info/warn/error]")
	flags.StringVar(&glob.Color, "color", "", "Set color mode [auto/always/never]")
}

func InitOptions(cli Cli, opts *GlobalOptions) {
	// override
	if opts.Color == "" && os.Getenv("NO_COLOR") != "" {
		// https://no-color.org/
		opts.Color = "never"
	}
	if os.Getenv("DEBUG_GOVPP") != "" || os.Getenv("GOVPP_DEBUG") != "" {
		opts.Debug = true
	}
	if loglvl := os.Getenv("GOVPP_LOGLEVEL"); loglvl != "" {
		opts.LogLevel = loglvl
	}

	switch strings.ToLower(opts.Color) {
	case "auto", "":
		if !cli.Out().IsTerminal() {
			color.Disable()
		}
	case "on", "enabled", "always", "1", "true":
		color.Enable = true
	case "off", "disabled", "never", "0", "false":
		color.Disable()
	default:
		logrus.Fatalf("invalid color mode: %q", opts.Color)
	}

	if opts.LogLevel != "" {
		if lvl, err := logrus.ParseLevel(opts.LogLevel); err == nil {
			logrus.SetLevel(lvl)
			if lvl >= logrus.TraceLevel {
				logrus.SetReportCaller(true)
			}
			logrus.Tracef("log level set to: %v", lvl)
		} else {
			logrus.Fatalf("invalid log level: %v", err)
		}
	} else if opts.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(defaultLogLevel)
	}

	logrus.Tracef("init global options: %+v", opts)
}
