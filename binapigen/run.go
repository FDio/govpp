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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

type Options struct {
	OutputDir     string // output directory for generated files
	ImportPrefix  string // prefix for import paths
	NoVersionInfo bool   // disables generating version info
}

func Run(apiDir string, filesToGenerate []string, opts Options, f func(*Generator) error) {
	if err := run(apiDir, filesToGenerate, opts, f); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}

func run(apiDir string, filesToGenerate []string, opts Options, fn func(*Generator) error) error {
	apifiles, err := vppapi.ParseDir(apiDir)
	if err != nil {
		return err
	}

	if opts.ImportPrefix == "" {
		opts.ImportPrefix = resolveImportPath(opts.OutputDir)
		logrus.Debugf("resolved import prefix: %s", opts.ImportPrefix)
	}

	gen, err := New(opts, apifiles, filesToGenerate)
	if err != nil {
		return err
	}

	gen.vppVersion = vppapi.ResolveVPPVersion(apiDir)
	if gen.vppVersion == "" {
		gen.vppVersion = "unknown"
	}

	if fn == nil {
		GenerateDefault(gen)
	} else {
		if err := fn(gen); err != nil {
			return err
		}
	}

	if err = gen.Generate(); err != nil {
		return err
	}

	return nil
}

func GenerateDefault(gen *Generator) {
	for _, file := range gen.Files {
		if !file.Generate {
			continue
		}
		GenerateAPI(gen, file)
		GenerateRPC(gen, file)
	}
}

var Logger = logrus.New()

func init() {
	if debug := os.Getenv("DEBUG_GOVPP"); strings.Contains(debug, "binapigen") {
		Logger.SetLevel(logrus.DebugLevel)
		logrus.SetLevel(logrus.DebugLevel)
	} else if debug != "" {
		Logger.SetLevel(logrus.InfoLevel)
	} else {
		Logger.SetLevel(logrus.WarnLevel)
	}
}

func logf(f string, v ...interface{}) {
	Logger.Debugf(f, v...)
}

func resolveImportPath(dir string) string {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	modRoot := findGoModuleRoot(absPath)
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

func findGoModuleRoot(dir string) (root string) {
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
