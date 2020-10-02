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
	OutputDir        string // output directory for generated files
	ImportPrefix     string // prefix for import paths
	NoVersionInfo    bool   // disables generating version info
	NoSourcePathInfo bool   // disables the 'source: /path' comment
}

func Run(apiDir string, filesToGenerate []string, opts Options, f func(*Generator) error) {
	if err := run(apiDir, filesToGenerate, opts, f); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}

func run(apiDir string, filesToGenerate []string, opts Options, fn func(*Generator) error) error {
	apiFiles, err := vppapi.ParseDir(apiDir)
	if err != nil {
		return err
	}

	if opts.ImportPrefix == "" {
		opts.ImportPrefix, err = resolveImportPath(opts.OutputDir)
		if err != nil {
			return fmt.Errorf("cannot resolve import path for output dir %s: %w", opts.OutputDir, err)
		}
		logrus.Debugf("resolved import path prefix: %s", opts.ImportPrefix)
	}

	gen, err := New(opts, apiFiles, filesToGenerate)
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
	}
}

func logf(f string, v ...interface{}) {
	Logger.Debugf(f, v...)
}

// resolveImportPath tries to resolve import path for a directory.
func resolveImportPath(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	modRoot := findGoModuleRoot(absPath)
	if modRoot == "" {
		return "", err
	}
	modPath, err := readModulePath(path.Join(modRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	relDir, err := filepath.Rel(modRoot, absPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(modPath, relDir), nil
}

// findGoModuleRoot looks for enclosing Go module.
func findGoModuleRoot(dir string) (root string) {
	dir = filepath.Clean(dir)
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

var modulePathRE = regexp.MustCompile(`module[ \t]+([^ \t\r\n]+)`)

// readModulePath reads module path from go.mod file.
func readModulePath(gomod string) (string, error) {
	data, err := ioutil.ReadFile(gomod)
	if err != nil {
		return "", err
	}
	m := modulePathRE.FindSubmatch(data)
	if m == nil {
		return "", err
	}
	return string(m[1]), nil
}
