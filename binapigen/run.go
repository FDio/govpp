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
)

var debugMode = true

func logf(f string, v ...interface{}) {
	if debugMode {
		logrus.Debugf(f, v...)
	}
}

func GenerateFromFile(inputFile, outputDir string, typesPkgs []*Context, opts Options) error {

	return
}

// GenerateFromFile generates Go package from one input JSON file
func GenerateFromFile(inputFile, outputDir string, typesPkgs []*Context, opts Options) error {
	// create generator context
	ctx, err := newContext(inputFile, outputDir)
	if err != nil {
		return err
	}

	ctx.Options = opts

	logf("------------------------------------------------------------")
	logf("module: %s", ctx.moduleName)
	logf(" - input: %s", ctx.inputFile)
	logf(" - output: %s", ctx.outputFile)
	logf("------------------------------------------------------------")

	// read API definition from input file
	ctx.inputData, err = ioutil.ReadFile(ctx.inputFile)
	if err != nil {
		return fmt.Errorf("reading input file %s failed: %v", ctx.inputFile, err)
	}
	/*ctx.packageData, err = parseModule(ctx, ctx.inputData)
	if err != nil {
		return fmt.Errorf("parsing package %s failed: %v", ctx.packageName, err)
	}*/

	/*if len(typesPkgs) > 0 {
		err = loadTypeAliases(ctx.packageData, typesPkgs)
		if err != nil {
			return fmt.Errorf("loading type aliases failed: %v", err)
		}
	}*/

	// generate binapi package
	var buf bytes.Buffer
	if err := generatePackage(ctx, &buf); err != nil {
		return fmt.Errorf("generating binapi package for %s failed: %v", ctx.packageName, err)
	}

	if err := writeSourceTo(ctx.outputFile, buf.Bytes()); err != nil {
		return fmt.Errorf("writing source for binapi package %s failed: %v", ctx.packageName, err)
	}
	// generate RPC package
	if ctx.generatesRPC() {
		buf.Reset()
		if err := generatePackageRPC(ctx, &buf); err != nil {
			return fmt.Errorf("generating RPC package for %s failed: %v", ctx.packageName, err)
		}
		if err := writeSourceTo(ctx.outputFileRPC, buf.Bytes()); err != nil {
			return fmt.Errorf("writing source for RPC package %s failed: %v", ctx.packageName, err)
		}
	}

	return nil
}

/*func parseModule(ctx *Context, data []byte) (*File, error) {
	module, err := vppapi.ParseRaw(data)
	if err != nil {
		return nil, err
	}

	//ctx.RefMap = ctx.refmap
	module.Name = ctx.packageName

	//dumpModuleAPI(module)
	logf("%+v", module)

	return module, nil
}*/

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
	return nil
}

/*func LoadTypesPackages(types ...string) ([]*Context, error) {
	var ctxs []*Context
	for _, inputFile := range types {
		// create generator context
		ctx, err := newContext(inputFile, "")
		if err != nil {
			return nil, err
		}
		// read API definition from input file
		ctx.inputData, err = ioutil.ReadFile(ctx.inputFile)
		if err != nil {
			return nil, fmt.Errorf("reading input file %s failed: %v", ctx.inputFile, err)
		}
		ctx.packageData, err = parseModule(ctx, ctx.inputData)
		if err != nil {
			return nil, fmt.Errorf("parsing package %s failed: %v", ctx.packageName, err)
		}
		ctxs = append(ctxs, ctx)
	}
	return ctxs, nil
}*/

/*func loadTypeAliases(module *File, typesCtxs []*Context) error {
	for _, t := range module.Types {
		for _, c := range typesCtxs {
			if _, ok := module.Imports[t.Name]; ok {
				break
			}
			for _, at := range c.packageData.Types {
				if at.Name != t.Name {
					continue
				}
				if len(at.Fields) != len(t.Fields) {
					continue
				}
				module.Imports[t.Name] = c.packageName
			}
		}
	}
	for _, t := range module.AliasTypes {
		for _, c := range typesCtxs {
			if _, ok := module.Imports[t.Name]; ok {
				break
			}
			for _, at := range c.packageData.AliasTypes {
				if at.Name != t.Name {
					continue
				}
				if at.Length != t.Length {
					continue
				}
				if at.Type != t.Type {
					continue
				}
				module.Imports[t.Name] = c.packageName
			}
		}
	}
	for _, t := range module.EnumTypes {
		for _, c := range typesCtxs {
			if _, ok := module.Imports[t.Name]; ok {
				break
			}
			for _, at := range c.packageData.EnumTypes {
				if at.Name != t.Name {
					continue
				}
				if at.Type != t.Type {
					continue
				}
				module.Imports[t.Name] = c.packageName
			}
		}
	}
	for _, t := range module.UnionTypes {
		for _, c := range typesCtxs {
			if _, ok := module.Imports[t.Name]; ok {
				break
			}
			for _, at := range c.packageData.UnionTypes {
				if at.Name != t.Name {
					continue
				}
				module.Imports[t.Name] = c.packageName
			}
		}
	}
	return nil
}*/
