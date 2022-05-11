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
	"path/filepath"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen"
	"git.fd.io/govpp.git/binapigen/vppapi"
	"git.fd.io/govpp.git/internal/version"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  Parse API_FILES and generate Go bindings\n")
		fmt.Fprintf(os.Stderr, "  Provide API_FILES by file name, or with full path including extension.\n")
		fmt.Fprintf(os.Stderr, "  %s [OPTION] API_FILES\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLES:\n")
		fmt.Fprintf(os.Stderr, "  %s \\\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "    --input-dir=$VPP/build-root/install-vpp-native/vpp/share/vpp/api/ \\\n")
		fmt.Fprintf(os.Stderr, "    --output-dir=~/output \\\n")
		fmt.Fprintf(os.Stderr, "    interface ip\n")
		fmt.Fprintf(os.Stderr, "  Assuming --input-dir contains interface.api.json & ip.api.json\n")
	}
}

func printErrorAndExit(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n", msg)
	flag.Usage()
	os.Exit(1)
}

func main() {
	var (
		theApiDir        = flag.String("input-dir", vppapi.DefaultDir, "Input directory containing API files. (e.g. )")
		theOutputDir     = flag.String("output-dir", "binapi", "Output directory where code will be generated.")
		importPrefix     = flag.String("import-prefix", "", "Prefix imports in the generated go code. \nE.g. other API Files (e.g. api_file.ba.go) will be imported with :\nimport (\n  api_file \"<import-prefix>/api_file\"\n)")
		generatorPlugins = flag.String("gen", "rpc", "List of generator plugins to run for files.")
		theInputFile     = flag.String("input-file", "", "DEPRECATED: Use program arguments to define files to generate.")

		printVersion     = flag.Bool("version", false, "Prints version and exits.")
		debugLog         = flag.Bool("debug", false, "Enable verbose logging.")
		noVersionInfo    = flag.Bool("no-version-info", false, "Disable version info in generated files.")
		noSourcePathInfo = flag.Bool("no-source-path-info", false, "Disable source path info in generated files.")
	)
	flag.Parse()

	if *printVersion {
		fmt.Fprintln(os.Stdout, version.Info())
		os.Exit(0)
	}

	if *debugLog {
		logrus.SetLevel(logrus.DebugLevel)
	}

	var filesToGenerate []string
	if *theInputFile != "" {
		if flag.NArg() > 0 {
			printErrorAndExit("input-file cannot be combined with files to generate in arguments")
		}
		filesToGenerate = append(filesToGenerate, *theInputFile)
	} else {
		filesToGenerate = append(filesToGenerate, flag.Args()...)
	}

	opts := binapigen.Options{
		ImportPrefix:     *importPrefix,
		OutputDir:        *theOutputDir,
		NoVersionInfo:    *noVersionInfo,
		NoSourcePathInfo: *noSourcePathInfo,
	}
	if opts.OutputDir == "binapi" {
		if wd, _ := os.Getwd(); filepath.Base(wd) == "binapi" {
			opts.OutputDir = "."
		}
	}
	apiDir := *theApiDir
	genPlugins := strings.FieldsFunc(*generatorPlugins, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	binapigen.Run(apiDir, filesToGenerate, opts, func(gen *binapigen.Generator) error {
		for _, file := range gen.Files {
			if !file.Generate {
				continue
			}
			binapigen.GenerateVersion(gen, file)
			binapigen.GenerateAPI(gen, file)
			for _, p := range genPlugins {
				if err := binapigen.RunPlugin(p, gen, file); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
