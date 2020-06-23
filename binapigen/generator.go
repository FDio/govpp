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
	"path"
	"path/filepath"
	"regexp"

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

	if len(opts.FilesToGenerate) > 0 {
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
				// generate all imported files
				for _, impFile := range file.importedFiles(g) {
					impFile.Generate = true
				}
			}
		}
	} else {
		logrus.Debugf("Files to generate not specified, marking all %d files to generate", len(g.Files))
		for _, file := range g.Files {
			file.Generate = true
		}
	}

	logrus.Debugf("Resolving imported types")
	for _, file := range g.Files {
		if !file.Generate {
			// skip resolving for non-generated files
			continue
		}
		var importedFiles []*File
		for _, impFile := range file.importedFiles(g) {
			if !impFile.Generate {
				// exclude imports of non-generated files
				continue
			}
			importedFiles = append(importedFiles, impFile)
		}
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
		if err := writeSourceTo(genfile.filename, genfile.Content()); err != nil {
			return fmt.Errorf("writing source for RPC package %s failed: %v", genfile.filename, err)
		}
	}
	return nil
}

type GenFile struct {
	*Generator
	filename  string
	file      *File
	outputDir string
	buf       bytes.Buffer
}

func (g *Generator) NewGenFile(filename string) *GenFile {
	f := &GenFile{
		Generator: g,
		filename:  filename,
	}
	g.genfiles = append(g.genfiles, f)
	return f
}

func (f *GenFile) Content() []byte {
	return f.buf.Bytes()
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

func listImports(genfile *GenFile) map[string]string {
	var importPath = genfile.ImportPrefix
	if importPath == "" {
		importPath = resolveImportPath(genfile.outputDir)
		logrus.Debugf("resolved import path: %s", importPath)
	}
	imports := map[string]string{}
	for _, imp := range genfile.file.imports {
		if _, ok := imports[imp]; !ok {
			imports[imp] = path.Join(importPath, imp)
		}
	}
	return imports
}

func resolveImportPath(outputDir string) string {
	absPath, err := filepath.Abs(outputDir)
	if err != nil {
		panic(err)
	}
	modRoot := findModuleRoot(absPath)
	if modRoot == "" {
		logrus.Fatalf("module root not found at: %s", absPath)
	}
	modPath := findModulePath(path.Join(modRoot, "go.mod"))
	if modPath == "" {
		logrus.Fatalf("module path not found")
	}
	relDir, err := filepath.Rel(modRoot, absPath)
	if err != nil {
		panic(err)
	}
	return filepath.Join(modPath, relDir)
}

func findModuleRoot(dir string) (root string) {
	if dir == "" {
		panic("dir not set")
	}
	dir = filepath.Clean(dir)

	// Look for enclosing go.mod.
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}

var (
	modulePathRE = regexp.MustCompile(`module[ \t]+([^ \t\r\n]+)`)
)

func findModulePath(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	m := modulePathRE.FindSubmatch(data)
	if m == nil {
		return ""
	}
	return string(m[1])
}
