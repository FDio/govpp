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
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

type Options struct {
	VPPVersion string // version of VPP that produced API files

	FilesToGenerate []string // list of API files to generate

	ImportPrefix       string // defines import path prefix for importing types
	ImportTypes        bool   // generate packages for import types
	IncludeAPIVersion  bool   // include constant with API version string
	IncludeComments    bool   // include parts of original source in comments
	IncludeBinapiNames bool   // include binary API names as struct tag
	IncludeServices    bool   // include service interface with client implementation
	IncludeVppVersion  bool   // include info about used VPP version
}

type Generator struct {
	Options

	Files       []*File
	FilesByPath map[string]*File
	FilesByName map[string]*File

	enumsByName   map[string]*Enum
	aliasesByName map[string]*Alias
	structsByName map[string]*Struct
	unionsByName  map[string]*Union

	genfiles []*GenFile
}

func New(opts Options, apifiles []*vppapi.File) (*Generator, error) {
	g := &Generator{
		Options:       opts,
		FilesByPath:   make(map[string]*File),
		FilesByName:   make(map[string]*File),
		enumsByName:   map[string]*Enum{},
		aliasesByName: map[string]*Alias{},
		structsByName: map[string]*Struct{},
		unionsByName:  map[string]*Union{},
	}

	logrus.Debugf("adding %d VPP API files to generator", len(apifiles))
	for _, apifile := range apifiles {
		filename := apifile.Path
		if filename == "" {
			filename = apifile.Name
		}
		if _, ok := g.FilesByPath[filename]; ok {
			return nil, fmt.Errorf("duplicate file name: %q", filename)
		}
		if _, ok := g.FilesByName[apifile.Name]; ok {
			return nil, fmt.Errorf("duplicate file: %q", apifile.Name)
		}

		file, err := newFile(g, apifile)
		if err != nil {
			return nil, err
		}
		g.Files = append(g.Files, file)
		g.FilesByPath[filename] = file
		g.FilesByName[apifile.Name] = file

		logrus.Debugf("added file %q (path: %v)", apifile.Name, apifile.Path)
		if len(file.Imports) > 0 {
			logrus.Debugf(" - %d imports: %v", len(file.Imports), file.Imports)
		}
	}

	logrus.Debugf("Checking %d files to generate: %v", len(opts.FilesToGenerate), opts.FilesToGenerate)
	for _, genfile := range opts.FilesToGenerate {
		file, ok := g.FilesByPath[genfile]
		if !ok {
			file, ok = g.FilesByName[genfile]
			if !ok {
				return nil, fmt.Errorf("no API file found for: %v", genfile)
			}
		}
		file.Generate = true
		if opts.ImportTypes {
			for _, impFile := range file.importedFiles(g) {
				impFile.Generate = true
			}
		}
	}

	logrus.Debugf("Resolving imported types")
	for _, file := range g.Files {
		if !file.Generate {
			continue
		}
		importedFiles := file.importedFiles(g)
		file.loadTypeImports(g, importedFiles)
	}

	return g, nil
}

func (g *Generator) Generate() error {
	if len(g.genfiles) == 0 {
		return fmt.Errorf("no files to generate")
	}

	logrus.Infof("Generating %d files", len(g.genfiles))
	for _, genfile := range g.genfiles {
		if err := writeSourceTo(genfile.filename, genfile.buf.Bytes()); err != nil {
			return fmt.Errorf("writing source for RPC package %s failed: %v", genfile.filename, err)
		}
	}
	return nil
}

func (g *Generator) NewGenFile(filename string) *GenFile {
	f := &GenFile{
		Generator: g,
		filename:  filename,
	}
	g.genfiles = append(g.genfiles, f)
	return f
}

func writeSourceTo(outputFile string, b []byte) error {
	// create output directory
	packageDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(packageDir, 0775); err != nil {
		return fmt.Errorf("creating output dir %s failed: %v", packageDir, err)
	}

	// format generated source code
	gosrc, err := format.Source(b)
	if err != nil {
		_ = ioutil.WriteFile(outputFile, b, 0666)
		return fmt.Errorf("formatting source code failed: %v", err)
	}

	// write generated code to output file
	if err := ioutil.WriteFile(outputFile, gosrc, 0666); err != nil {
		return fmt.Errorf("writing to output file %s failed: %v", outputFile, err)
	}

	lines := bytes.Count(gosrc, []byte("\n"))
	logf("wrote %d lines (%d bytes) of code to: %q", lines, len(gosrc), outputFile)

	return nil
}
