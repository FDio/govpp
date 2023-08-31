// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"go.fd.io/govpp/binapigen"
	"go.fd.io/govpp/binapigen/vppapi"
	"go.fd.io/govpp/internal/version"
)

const (
	DefaultOutputDir = "binapi"
)

var (
	input        = pflag.String("input", "", "Input for VPP API (e.g. path to VPP API directory, local VPP repo)")
	inputDir     = pflag.String("input-dir", "", "Input directory containing API files.")
	theOutputDir = pflag.StringP("output-dir", "o", DefaultOutputDir, "Output directory where code will be generated.")
	runPlugins   = pflag.StringSlice("gen", []string{"rpc"}, "List of generator plugins to run for files.")
	importPrefix = pflag.String("import-prefix", "", "Prefix imports in the generated go code. \nE.g. other API Files (e.g. api_file.ba.go) will be imported with :\nimport (\n  api_file \"<import-prefix>/api_file\"\n)")

	noVersionInfo    = pflag.Bool("no-version-info", false, "Disable version info in generated files.")
	noSourcePathInfo = pflag.Bool("no-source-path-info", false, "Disable source path info in generated files.")

	printVersion = pflag.Bool("version", false, "Prints version and exits.")
	enableDebug  = pflag.Bool("debug", false, "Enable debugging mode.")
)

func init() {
	pflag.Lookup("input-dir").Deprecated = "The --input-dir is deprecated, use --input instead."
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  Parse API_FILES and generate Go bindings\n")
		fmt.Fprintf(os.Stderr, "  Provide API_FILES by file name, or with full path including extension.\n")
		fmt.Fprintf(os.Stderr, "  %s [OPTION] API_FILES\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLES:\n")
		fmt.Fprintf(os.Stderr, "  %s \\\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "    --input=$VPP/build-root/install-vpp-native/vpp/share/vpp/api/ \\\n")
		fmt.Fprintf(os.Stderr, "    --output-dir=~/output \\\n")
		fmt.Fprintf(os.Stderr, "    interface ip\n")
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func main() {
	pflag.Parse()

	if *printVersion {
		fmt.Fprintln(os.Stdout, version.Info())
		os.Exit(0)
	}

	if *enableDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	var filesToGenerate []string
	filesToGenerate = append(filesToGenerate, pflag.Args()...)

	opts := binapigen.Options{
		ImportPrefix:     *importPrefix,
		OutputDir:        *theOutputDir,
		NoVersionInfo:    *noVersionInfo,
		NoSourcePathInfo: *noSourcePathInfo,
		GenerateFiles:    filesToGenerate,
	}

	// generate in same directory when current dir is binapi
	if opts.OutputDir == DefaultOutputDir {
		if wd, _ := os.Getwd(); filepath.Base(wd) == DefaultOutputDir {
			opts.OutputDir = "."
		}
	}

	theInputDir := *inputDir
	theInput := *input
	genPlugins := *runPlugins

	if theInputDir != "" {
		if theInput != "" {
			logrus.Fatalf("ignoring deprecated option 'input-dir', using 'input' instead")
		} else {
			theInput = theInputDir
		}
	}
	if theInput == "" {
		if dir := os.Getenv(vppapi.VPPDirEnvVar); dir != "" {
			logrus.Infof("VPP input directory was set to %q via %s env var", dir, vppapi.VPPDirEnvVar)
			theInput = dir
		}
	}

	vppInput, err := vppapi.ResolveVppInput(theInput)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("resolved VPP input: %+v", vppInput)

	binapigen.Run(vppInput, opts, binapigen.GeneratePlugins(genPlugins))
}
