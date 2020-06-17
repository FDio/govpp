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

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen"
	"git.fd.io/govpp.git/binapigen/vppapi"
	"git.fd.io/govpp.git/version"
)

func main() {
	var (
		theInputFile = flag.String("input-file", "", "Input VPP API file.")
		theInputDir  = flag.String("input-dir", vppapi.DefaultAPIDir, "Directory with VPP API files.")
		theOutputDir = flag.String("output-dir", ".", "Output directory where package folders will be generated.")

		importPrefix = flag.String("import-prefix", "", "Define import path prefix to be used to import types.")
		//theInputTypes = flag.String("input-types", "", "Types input file with VPP API in JSON format. (split by comma)")

		includeAPIVer      = flag.Bool("include-apiver", true, "Include APIVersion constant for each module.")
		includeServices    = flag.Bool("include-services", true, "Include RPC service api and client implementation.")
		includeComments    = flag.Bool("include-comments", false, "Include JSON API source in comments for each object.")
		includeBinapiNames = flag.Bool("include-binapi-names", false, "Include binary API names in struct tag.")
		includeVppVersion  = flag.Bool("include-vpp-version", true, "Include version of the VPP that provided input files.")

		continueOnError = flag.Bool("continue-onerror", false, "Continue with next file on error.")
		debugMode       = flag.Bool("debug", os.Getenv("DEBUG_GOVPP") != "", "Enable debug mode.")

		printVersion = flag.Bool("version", false, "Prints version and exits.")
	)
	flag.Parse()

	if *printVersion {
		fmt.Fprintln(os.Stdout, version.Info())
		os.Exit(0)
	}

	if flag.NArg() > 0 {
		switch cmd := flag.Arg(0); cmd {
		case "version":
			fmt.Fprintln(os.Stdout, version.Verbose())
			os.Exit(0)

		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
			flag.Usage()
			os.Exit(2)
		}
	}

	var opts binapigen.Options

	if ver := os.Getenv("VPP_API_VERSION"); ver != "" {
		// use version from env var if set
		opts.VPPVersion = ver
	} else {
		opts.VPPVersion = ResolveVppVersion(*theInputDir)
	}

	// prepare options
	opts.IncludeAPIVersion = *includeAPIVer
	opts.IncludeComments = *includeComments
	opts.IncludeBinapiNames = *includeBinapiNames
	opts.IncludeServices = *includeServices
	opts.IncludeVppVersion = *includeVppVersion
	opts.ImportPrefix = *importPrefix

	if *debugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debug mode enabled")
	}

	if err := run(*theInputFile, *theInputDir, *theOutputDir, *continueOnError, opts); err != nil {
		logrus.Errorln("binapi-generator:", err)
		os.Exit(1)
	}
}

func run(inputFile, inputDir string, outputDir string, continueErr bool, opts binapigen.Options) (err error) {
	if inputFile == "" && inputDir == "" {
		return fmt.Errorf("input file or input dir must be specified")
	}

	var typesPkgs []*binapigen.Context
	/*if *theInputTypes != "" {
		types := strings.Split(*theInputTypes, ",")
		typesPkgs, err = binapigen.LoadTypesPackages(types...)
		if err != nil {
			return fmt.Errorf("loading types input failed: %v", err)
		}
	}*/

	if inputFile != "" {
		// process one input file
		if err := binapigen.GenerateFromFile(inputFile, outputDir, typesPkgs, opts); err != nil {
			return fmt.Errorf("code generation from %s failed: %v\n", inputFile, err)
		}
	} else {
		// process all files in specified directory
		dir, err := filepath.Abs(inputDir)
		if err != nil {
			return fmt.Errorf("invalid input directory: %v\n", err)
		}
		files, err := vppapi.FindFiles(inputDir, 1)
		if err != nil {
			return fmt.Errorf("problem getting files from input directory: %v\n", err)
		} else if len(files) == 0 {
			return fmt.Errorf("no input files found in input directory: %v\n", dir)
		}
		for _, file := range files {
			if err := binapigen.GenerateFromFile(file, outputDir, typesPkgs, opts); err != nil {
				if continueErr {
					logrus.Warnf("code generation from %s failed: %v (error ignored)\n", file, err)
					continue
				} else {
					return fmt.Errorf("code generation from %s failed: %v\n", file, err)
				}
			}
		}
	}

	return nil
}
