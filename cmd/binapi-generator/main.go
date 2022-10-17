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
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen"
	"go.fd.io/govpp/binapigen/vppapi"
	"go.fd.io/govpp/version"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Generating Go bindings for the VPP API\n")
		fmt.Fprintf(os.Stderr, "--------------------------------------\n\n")
		fmt.Fprintf(os.Stderr, "This generates VPP api Go bindings based on .api.json files or a VPP repository\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLES:\n")
		fmt.Fprintf(os.Stderr, "  Generate bindings from VPP API files (*.api.json):\n")
		fmt.Fprintf(os.Stderr, "   %s --api /usr/share/vpp/api --output ./myapp/binding --filter interface,ip\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Generate bindings from the VPP repository :\n")
		fmt.Fprintf(os.Stderr, "   %s --vpp ~/vpp --output ./myapp/binding\n", os.Args[0])
	}
}

func printErrorAndExit(msg string) {
	color.New(color.FgRed).Fprintf(os.Stderr, "Error:\n%s\n\n", msg)
	flag.Usage()
	os.Exit(1)
}

func main() {
	apiSrcDir := flag.String("input-dir", vppapi.DefaultDir, "[DEPRECATED, use --api] Input directory containing API files.")
	flag.StringVar(apiSrcDir, "api", "", "Generate based on .api files in this directory.")
	vppSrcDir := flag.String("vpp", "", "Generate based on a vpp cloned in this directory.")

	// Filtering API files
	filterList := flag.String("input-file", "", "[DEPRECATED: Use --filter] defines apis to generate.")
	flag.StringVar(filterList, "filter", "", "Comma separated list of api to generate (e.g. ipip,ipsec, ...)")

	// Where to output the files
	outputDir := flag.String("output-dir", ".", "[DEPRECATED, use --output] Output directory where code will be generated.")
	flag.StringVar(outputDir, "output", "", "Output directory for generated files.")
	flag.StringVar(outputDir, "o", "", "Output directory for generated files.")

	// Package name to use
	importPrefix := flag.String("import-prefix", "", "[DEPRECATED, use --package] Prefix imports in the generated go code.")
	flag.StringVar(importPrefix, "package", "", "Package path to generate to e.g. myapp.me.com/myapp/bindings. If omitted, we'll try to guess it from go modules and the output directory")

	// Plugins
	generatorPlugins := flag.String("gen", "", "[DEPRECATED, use --plugins] List of generator plugins to run for files.")
	flag.StringVar(generatorPlugins, "plugins", "", fmt.Sprintf("List of generator plugins to run for files. (%s or a path to an external plugin)", binapigen.GetAvailablePluginNames()))

	printVersion := flag.Bool("version", false, "Prints version and exits.")
	debugLog := flag.Bool("debug", false, "Enable verbose logging.")
	noVersionInfo := flag.Bool("no-version-info", false, "Disable version info in generated files.")
	noSourcePathInfo := flag.Bool("no-source-path-info", false, "Disable source path info in generated files.")

	flag.Parse()

	if *printVersion {
		fmt.Fprintln(os.Stdout, version.Info())
		os.Exit(0)
	}

	if *debugLog {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if *vppSrcDir == "" && *apiSrcDir == "" {
		printErrorAndExit("Please provide either --api or --vpp")
	}

	fileFilter := binapigen.SplitAndStrip(*filterList)
	if flag.NArg() > 0 {
		logrus.Warnf("Deprecated, use --filter instead to pass the API names")
		fileFilter = append(fileFilter, flag.Args()...)
	}

	generator, err := binapigen.New(binapigen.Options{
		ImportPrefix:      binapigen.GetTargetPackagePath(*importPrefix, *outputDir),
		OutputDir:         binapigen.ExpandPaths(*outputDir),
		NoVersionInfo:     *noVersionInfo,
		NoSourcePathInfo:  *noSourcePathInfo,
		ActivePluginNames: binapigen.SplitAndStrip(*generatorPlugins),
		ApiDir:            binapigen.GetApiFileDirectory(*apiSrcDir, *vppSrcDir),
		FileFilter:        fileFilter,
	})
	if err != nil {
		logrus.Fatalf("error creating generator %s", err)
	}

	err = generator.Generate()
	if err != nil {
		logrus.Fatalf("error generating %s", err)
	}
}
