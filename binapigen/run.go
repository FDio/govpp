//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package binapigen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

var debugMode = true

func logf(f string, v ...interface{}) {
	if debugMode {
		logrus.Debugf(f, v...)
	}
}

func GenerateBinapiFile(gen *Generator, file *File, outputDir string) *GenFile {
	packageDir := filepath.Join(outputDir, file.PackageName)
	filename := filepath.Join(packageDir, file.PackageName+outputFileExt)

	g := gen.NewGenFile(filename)
	g.file = file
	g.packageDir = filepath.Join(outputDir, file.PackageName)

	generatePackage(g, &g.buf)

	return g
}

func GenerateRPC(gen *Generator, file *File, outputDir string) *GenFile {
	packageDir := filepath.Join(outputDir, file.PackageName)
	filename := filepath.Join(packageDir, file.PackageName+rpcFileSuffix+outputFileExt)

	g := gen.NewGenFile(filename)
	g.file = file
	g.packageDir = filepath.Join(outputDir, file.PackageName)

	generatePackageRPC(g, &g.buf)

	return g
}

func Run(apiDir string, opts Options, f func(*Generator) error) {
	if err := run(apiDir, opts, f); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}

func run(apiDir string, opts Options, f func(*Generator) error) error {
	// parse API files
	apifiles, err := vppapi.ParseDir(apiDir)
	if err != nil {
		return err
	}

	g, err := New(opts, apifiles)
	if err != nil {
		return err
	}

	if err := f(g); err != nil {
		return err
	}

	if err = g.Generate(); err != nil {
		return err
	}

	return nil
}
