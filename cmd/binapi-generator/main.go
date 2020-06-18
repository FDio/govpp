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

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen"
	"git.fd.io/govpp.git/binapigen/vppapi"
	"git.fd.io/govpp.git/version"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTION]... [API]...\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Generate code for each API.")
		fmt.Fprintf(flag.CommandLine.Output(), "Example: %s -output-dir=binapi acl interface l2\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		flag.CommandLine.PrintDefaults()
	}
}

func main() {
	var (
		theInputFile = flag.String("input-file", "", "Input VPP API file. (DEPRECATED: Use program arguments to define VPP API files)")
		theApiDir    = flag.String("input-dir", vppapi.DefaultAPIDir, "Directory with VPP API files.")
		theOutputDir = flag.String("output-dir", ".", "Output directory where code will be generated.")

		importPrefix       = flag.String("import-prefix", "", "Define import path prefix to be used to import types.")
		importTypes        = flag.Bool("import-types", false, "Generate packages for imported types.")
		includeAPIVer      = flag.Bool("include-apiver", true, "Include APIVersion constant for each module.")
		includeServices    = flag.Bool("include-services", true, "Include RPC service api and client implementation.")
		includeComments    = flag.Bool("include-comments", false, "Include JSON API source in comments for each object.")
		includeBinapiNames = flag.Bool("include-binapi-names", true, "Include binary API names in struct tag.")
		includeVppVersion  = flag.Bool("include-vpp-version", true, "Include version of the VPP that provided input files.")

		debugMode    = flag.Bool("debug", os.Getenv("DEBUG_GOVPP") != "", "Enable debug mode.")
		printVersion = flag.Bool("version", false, "Prints version and exits.")
	)
	flag.Parse()

	if *printVersion {
		fmt.Fprintln(os.Stdout, version.Info())
		os.Exit(0)
	}

	if flag.NArg() == 1 && flag.Arg(0) == "version" {
		fmt.Fprintln(os.Stdout, version.Verbose())
		os.Exit(0)
	}

	var opts binapigen.Options

	if *theInputFile != "" {
		if flag.NArg() > 0 {
			fmt.Fprintln(os.Stderr, "input-file cannot be combined with files to generate in arguments")
			os.Exit(1)
		}
		opts.FilesToGenerate = append(opts.FilesToGenerate, *theInputFile)
	} else {
		opts.FilesToGenerate = append(opts.FilesToGenerate, flag.Args()...)
	}

	// prepare options
	if ver := os.Getenv("VPP_API_VERSION"); ver != "" {
		// use version from env var if set
		opts.VPPVersion = ver
	} else {
		opts.VPPVersion = ResolveVppVersion(*theApiDir)
	}
	opts.IncludeAPIVersion = *includeAPIVer
	opts.IncludeComments = *includeComments
	opts.IncludeBinapiNames = *includeBinapiNames
	opts.IncludeServices = *includeServices
	opts.IncludeVppVersion = *includeVppVersion
	opts.ImportPrefix = *importPrefix
	opts.ImportTypes = *importTypes

	if *debugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debug mode enabled")
	}

	apiDir := *theApiDir
	outputDir := *theOutputDir

	binapigen.Run(apiDir, opts, func(g *binapigen.Generator) error {
		for _, file := range g.Files {
			if !file.Generate {
				continue
			}
			binapigen.GenerateBinapiFile(g, file, outputDir)
			if g.IncludeServices && file.Service != nil {
				binapigen.GenerateRPC(g, file, outputDir)
			}
		}
		return nil
	})
}
