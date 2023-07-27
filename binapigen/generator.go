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
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen/vppapi"
)

// Options is set of input parameters for the Generator.
type Options struct {
	OutputDir     string   // output directory for generated files
	ImportPrefix  string   // prefix for package import paths
	GenerateFiles []string // list of files to generate

	NoVersionInfo    bool // disables generating version info
	NoSourcePathInfo bool // disables the 'source: /path' comment
}

// Generator processes VPP API files as input, provides API to handle content
// of generated files.
type Generator struct {
	opts Options

	vppapiSchema *vppapi.Schema

	Files       []*File
	FilesByName map[string]*File
	FilesByPath map[string]*File

	genfiles []*GenFile

	enumsByName    map[string]*Enum
	aliasesByName  map[string]*Alias
	structsByName  map[string]*Struct
	unionsByName   map[string]*Union
	messagesByName map[string]*Message
}

func New(opts Options, input *vppapi.VppInput) (*Generator, error) {
	gen := &Generator{
		opts:           opts,
		vppapiSchema:   &input.Schema,
		FilesByName:    make(map[string]*File),
		FilesByPath:    make(map[string]*File),
		enumsByName:    map[string]*Enum{},
		aliasesByName:  map[string]*Alias{},
		structsByName:  map[string]*Struct{},
		unionsByName:   map[string]*Union{},
		messagesByName: map[string]*Message{},
	}

	// normalize API files
	SortFilesByImports(gen.vppapiSchema.Files)
	for i, apiFile := range gen.vppapiSchema.Files {
		f := apiFile
		RemoveImportedTypes(gen.vppapiSchema.Files, &f)
		SortFileObjectsByName(&f)
		gen.vppapiSchema.Files[i] = f
	}

	// prepare package names and import paths
	packageNames := make(map[string]GoPackageName)
	importPaths := make(map[string]GoImportPath)
	for _, apifile := range gen.vppapiSchema.Files {
		filename := getFilename(apifile)
		packageNames[filename] = cleanPackageName(apifile.Name)
		importPaths[filename] = GoImportPath(path.Join(gen.opts.ImportPrefix, baseName(apifile.Name)))
	}

	logrus.Debugf("adding %d VPP API files to generator", len(gen.vppapiSchema.Files))

	for _, apifile := range gen.vppapiSchema.Files {
		if _, ok := gen.FilesByName[apifile.Name]; ok {
			return nil, fmt.Errorf("duplicate file: %q", apifile.Name)
		}

		filename := getFilename(apifile)
		file, err := newFile(gen, apifile, packageNames[filename], importPaths[filename])
		if err != nil {
			return nil, fmt.Errorf("loading file %s failed: %w", apifile.Name, err)
		}
		gen.Files = append(gen.Files, file)
		gen.FilesByName[apifile.Name] = file
		gen.FilesByPath[apifile.Path] = file

		logrus.Debugf("added file %q (path: %v)", apifile.Name, apifile.Path)
	}

	// mark files for generation
	if len(gen.opts.GenerateFiles) > 0 {
		logrus.Debugf("Checking %d files to generate: %v", len(gen.opts.GenerateFiles), gen.opts.GenerateFiles)
		for _, genFile := range gen.opts.GenerateFiles {
			markGen := func(file *File) {
				file.Generate = true
				// generate all imported files
				for _, impFile := range file.importedFiles(gen) {
					impFile.Generate = true
				}
			}

			if file, ok := gen.FilesByName[genFile]; ok {
				markGen(file)
				continue
			}
			logrus.Debugf("File %s was not found by name", genFile)
			if file, ok := gen.FilesByPath[genFile]; ok {
				markGen(file)
				continue
			}
			return nil, fmt.Errorf("no API file found for: %v", genFile)
		}
	} else {
		logrus.Debugf("Files to generate not specified, marking all %d files for generate", len(gen.Files))
		for _, file := range gen.Files {
			file.Generate = true
		}
	}

	return gen, nil
}

func (g *Generator) GetMessageByName(name string) *Message {
	return g.messagesByName[name]
}

func (g *Generator) GetOpts() Options { return g.opts }

func getFilename(file vppapi.File) string {
	if file.Path == "" {
		return file.Name
	}
	return file.Path
}

func (g *Generator) Generate() error {
	if len(g.genfiles) == 0 {
		return fmt.Errorf("no files to generate")
	}

	logrus.Infof("Generating %d files", len(g.genfiles))

	for _, genfile := range g.genfiles {
		content, err := genfile.Content()
		if err != nil {
			return err
		}
		logrus.Debugf("- generating file: %v (%v bytes)", genfile.filename, len(content))
		if err := WriteContentToFile(genfile.filename, content); err != nil {
			return fmt.Errorf("writing source package %s failed: %v", genfile.filename, err)
		}
	}

	return nil
}

type GenFile struct {
	gen           *Generator
	file          *File
	filename      string
	buf           bytes.Buffer
	manualImports map[GoImportPath]bool
	packageNames  map[GoImportPath]GoPackageName
}

// NewGenFile creates new generated file with
func (g *Generator) NewGenFile(filename string, file *File) *GenFile {
	f := &GenFile{
		gen:           g,
		file:          file,
		filename:      filename,
		manualImports: make(map[GoImportPath]bool),
		packageNames:  make(map[GoImportPath]GoPackageName),
	}
	g.genfiles = append(g.genfiles, f)
	return f
}

func (g *GenFile) GetFile() *File { return g.file }

func (g *GenFile) Write(p []byte) (n int, err error) {
	return g.buf.Write(p)
}

func (g *GenFile) Import(importPath GoImportPath) {
	g.manualImports[importPath] = true
}

func (g *GenFile) GoIdent(ident GoIdent) string {
	if g.file != nil && ident.GoImportPath == g.file.GoImportPath {
		return ident.GoName
	}
	if packageName, ok := g.packageNames[ident.GoImportPath]; ok {
		return string(packageName) + "." + ident.GoName
	}
	packageName := cleanPackageName(baseName(string(ident.GoImportPath)))
	g.packageNames[ident.GoImportPath] = packageName
	return string(packageName) + "." + ident.GoName
}

func (g *GenFile) P(v ...interface{}) {
	for _, x := range v {
		switch x := x.(type) {
		case GoIdent:
			fmt.Fprint(&g.buf, g.GoIdent(x))
		default:
			fmt.Fprint(&g.buf, x)
		}
	}
	fmt.Fprintln(&g.buf)
}

func (g *GenFile) Content() ([]byte, error) {
	// for *.go files we inject imports
	if strings.HasSuffix(g.filename, ".go") {
		return g.injectImports(g.buf.Bytes())
	}
	return g.buf.Bytes(), nil
}

// baseName returns the last path element of the name, with the last dotted suffix removed.
func baseName(name string) string {
	// First, find the last element
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	// Now drop the suffix
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[:i]
	}
	return name
}

func getImportClass(importPath string) int {
	if !strings.Contains(importPath, ".") {
		return 0 /* std */
	}
	return 1 /* External */
}

// injectImports parses source, injects import block declaration with all imports and return formatted
func (g *GenFile) injectImports(original []byte) ([]byte, error) {
	// Parse source code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", original, parser.ParseComments)
	if err != nil {
		var src bytes.Buffer
		s := bufio.NewScanner(bytes.NewReader(original))
		for line := 1; s.Scan(); line++ {
			fmt.Fprintf(&src, "%5d\t%s\n", line, s.Bytes())
		}
		return nil, fmt.Errorf("%v: unparsable Go source: %v\n%v", g.filename, err, src.String())
	}
	type Import struct {
		Name string
		Path string
	}
	// Prepare list of all imports
	var importPaths []Import
	for importPath := range g.packageNames {
		importPaths = append(importPaths, Import{
			Name: string(g.packageNames[importPath]),
			Path: string(importPath),
		})
	}
	for importPath := range g.manualImports {
		if _, ok := g.packageNames[importPath]; ok {
			continue
		}
		importPaths = append(importPaths, Import{
			Name: "_",
			Path: string(importPath),
		})
	}
	// Sort imports by import path
	sort.Slice(importPaths, func(i, j int) bool {
		ci := getImportClass(importPaths[i].Path)
		cj := getImportClass(importPaths[j].Path)
		if ci == cj {
			return importPaths[i].Path < importPaths[j].Path
		}
		return ci < cj
	})
	// Inject new import block into parsed AST
	if len(importPaths) > 0 {
		// Find import block position
		pos := file.Package
		tokFile := fset.File(file.Package)
		pkgLine := tokFile.Line(file.Package)
		for _, c := range file.Comments {
			if tokFile.Line(c.Pos()) > pkgLine {
				break
			}
			pos = c.End()
		}
		// Prepare the import block
		impDecl := &ast.GenDecl{Tok: token.IMPORT, TokPos: pos, Lparen: pos, Rparen: pos}
		for i, importPath := range importPaths {
			var name *ast.Ident
			if importPath.Name == "_" || strings.Contains(importPath.Path, ".") {
				name = &ast.Ident{Name: importPath.Name, NamePos: pos}
			}
			value := strconv.Quote(importPath.Path)
			if i < len(importPaths)-1 {
				if getImportClass(importPath.Path) != getImportClass(importPaths[i+1].Path) {
					value += "\n"
				}
			}
			impDecl.Specs = append(impDecl.Specs, &ast.ImportSpec{
				Name:   name,
				Path:   &ast.BasicLit{Kind: token.STRING, Value: value, ValuePos: pos},
				EndPos: pos,
			})
		}

		file.Decls = append([]ast.Decl{impDecl}, file.Decls...)
	}
	// Reformat source code
	var out bytes.Buffer
	cfg := &printer.Config{
		Mode:     printer.TabIndent | printer.UseSpaces,
		Tabwidth: 8,
	}
	if err = cfg.Fprint(&out, fset, file); err != nil {
		return nil, fmt.Errorf("cannot reformat Go code in file %q: %w", g.filename, err)
	}
	return out.Bytes(), nil
}

func WriteContentToFile(outputFile string, content []byte) error {
	// create output directory
	packageDir := filepath.Dir(outputFile)

	if err := os.MkdirAll(packageDir, 0775); err != nil {
		return fmt.Errorf("creating output dir %s failed: %v", packageDir, err)
	}

	// write generated code to output file
	if err := os.WriteFile(outputFile, content, 0666); err != nil {
		return fmt.Errorf("writing to output file %s failed: %v", outputFile, err)
	}

	lines := bytes.Count(content, []byte("\n"))
	logf("written %d lines (%d bytes) to: %q", lines, len(content), outputFile)

	return nil
}
