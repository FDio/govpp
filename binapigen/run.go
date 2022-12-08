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
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen/vppapi"
)

/*type Config struct {
	Version string // version of config

	Input string // API input (local VPP dir, ...)

	VppReference      string   // commit id / tag / branch
	Patches           []string // list of custom patches to include
	RegenerateJsonAPI bool

	GenerateFiles []string // list of API files to generate
	Plugins       []string // enabled generator plugins
}*/

type VppInput struct {
	ApiFiles   []*vppapi.File
	VppVersion string
}

func ResolveVppInput(input string) (*VppInput, error) {
	vppInput := &VppInput{}

	if input == "" {
		input = vppapi.DefaultDir
	}

	u, err := url.Parse(input)
	if err != nil {
		logrus.Warnf("parsing url error: %v", err)
	} else {
		switch u.Scheme {
		case "", "file":
			i, err := os.Stat(input)
			if err != nil {
				return nil, fmt.Errorf("file error: %v", err)
			} else {
				if i.IsDir() {
					apidir := vppapi.ResolveApiDir(u.Path)
					logrus.Debugf("path %q resolved to api dir: %v", u.Path, apidir)
					apiFiles, err := vppapi.ParseDir(apidir)
					if err != nil {
						logrus.Warnf("vppapi parsedir error: %v", err)
					} else {
						vppInput.ApiFiles = apiFiles
						logrus.Infof("resolved %d apifiles", len(apiFiles))
					}
					vppInput.VppVersion = vppapi.ResolveVPPVersion(u.Path)
					if vppInput.VppVersion == "" {
						vppInput.VppVersion = "unknown"
					}
				} else {
					return nil, fmt.Errorf("files not supported")
				}
			}
		case "http", "https":
			return nil, fmt.Errorf("http(s) not yet supported")
		case "git", "ssh":
			return nil, fmt.Errorf("ssh/git not yet supported")
		default:
			return nil, fmt.Errorf("unsupported scheme: %v", u.Scheme)
		}
	}

	return vppInput, nil
}

func Run(vppInput *VppInput, opts Options, f func(*Generator) error) {
	if err := run(vppInput, opts, f); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}

func run(vppInput *VppInput, opts Options, genFn func(*Generator) error) error {
	var err error

	//
	if opts.OutputDir == "binapi" {
		if wd, _ := os.Getwd(); filepath.Base(wd) == "binapi" {
			opts.OutputDir = "."
		}
	}

	if opts.ImportPrefix == "" {
		opts.ImportPrefix, err = ResolveImportPath(opts.OutputDir)
		if err != nil {
			return fmt.Errorf("cannot resolve import path for output dir %s: %w", opts.OutputDir, err)
		}
		logrus.Debugf("resolved import path prefix: %s", opts.ImportPrefix)
	}

	gen, err := New(opts, vppInput)
	if err != nil {
		return err
	}

	if genFn == nil {
		genFn = GenerateDefault
	}
	if err := genFn(gen); err != nil {
		return err
	}

	if err = gen.Generate(); err != nil {
		return err
	}

	return nil
}

func GeneratePlugins(genPlugins []string) func(*Generator) error {
	return func(gen *Generator) error {
		for _, file := range gen.Files {
			if !file.Generate {
				continue
			}
			GenerateAPI(gen, file)
			for _, p := range genPlugins {
				if err := RunPlugin(p, gen, file); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func GenerateDefault(gen *Generator) error {
	for _, file := range gen.Files {
		if !file.Generate {
			continue
		}
		GenerateAPI(gen, file)
		GenerateRPC(gen, file)
	}
	return nil
}

// ResolveImportPath tries to resolve import path for a directory.
func ResolveImportPath(dir string) (string, error) {
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
	data, err := os.ReadFile(gomod)
	if err != nil {
		return "", err
	}
	m := modulePathRE.FindSubmatch(data)
	if m == nil {
		return "", err
	}
	return string(m[1]), nil
}
