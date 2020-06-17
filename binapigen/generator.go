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
	"io"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

type Generator struct {
	Options

	files       []*File
	filesByPath map[string]*File
}

func New(opts Options, apifiles []*vppapi.File) (*Generator, error) {
	g := &Generator{
		Options: opts,
	}
	if err := g.AddFiles(apifiles...); err != nil {
		return nil, err
	}
	for _, genfile := range opts.FilesToGenerate {
		for _, file := range g.files {
			if genfile == file.Name {
				file.Generate = true
			}
		}
	}
	return g, nil
}

func (g *Generator) AddFiles(apifiles ...*vppapi.File) error {
	for _, m := range apifiles {
		if err := g.addFile(*m); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) addFile(apifile vppapi.File) error {
	// check for duplicate add
	for _, f := range g.files {
		if f.Name == apifile.Name {
			return fmt.Errorf("file %s already added", apifile.Name)
		}
	}

	file, err := newFile(g, apifile)
	if err != nil {
		return err
	}
	g.files = append(g.files, file)

	return nil
}

func (g *Generator) Generate(outputDir string, w io.Writer) error {
	for _, file := range g.files {
		if !file.Generate {
			logrus.Infof("skip generate for file %q", file.Path)
		}

		if err := g.generateFile(file, outputDir, w); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateFile(file *File, outputDir string, w io.Writer) error {
	/*logf("----------------------------")
	logf("generating binapi for %q", file.PackageName)
	logf("----------------------------")

	generateHeader(file, w)
	generateImports(file, w)*/

	ctx := &Context{
		moduleName:  file.Name,
		packageName: file.PackageName,

		RefMap: file.refmap,
	}
	packageDir := filepath.Join(outputDir, ctx.packageName)
	ctx.outputFile = filepath.Join(packageDir, ctx.packageName+outputFileExt)
	ctx.outputFileRPC = filepath.Join(packageDir, ctx.packageName+rpcFileSuffix+outputFileExt)

	if err := generatePackage(ctx, w); err != nil {
		return err
	}

	return nil
}
