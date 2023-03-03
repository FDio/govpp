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
	"plugin"
)

// Plugin is an extension of the Generator. Plugins can be registered in
// application with RegisterPlugin or loaded from an external file compiled
// when calling RunPlugin.
type Plugin struct {
	Name         string
	GenerateAll  GenerateAllFn
	GenerateFile GenerateFileFn
	External     bool
}

type GenerateAllFn = func(*Generator) []*GenFile
type GenerateFileFn = func(*Generator, *File) *GenFile

var plugins []*Plugin
var pluginsByName = map[string]*Plugin{}

// RegisterPlugin registers a new plugin with name and generate
// func. Name must not be empty or already taken.
func RegisterPlugin(name string, genfn GenerateFileFn) {
	if name == "" {
		panic("plugin name is empty")
	}
	if _, ok := pluginsByName[name]; ok {
		panic("duplicate plugin name: " + name)
	}
	p := &Plugin{
		Name:         name,
		GenerateFile: genfn,
	}
	addPlugin(p)
}

// RunPlugin executes plugin with given name, if name is not found it attempts
// to load plugin from a filesystem, using name as path to the file. The file
// must be Go binary compiled using "plugin" buildmode and must contain exported
// func with the signature of GenerateFileFn.
func RunPlugin(name string, gen *Generator, file *File) error {
	p, err := getPlugin(name)
	if err != nil {
		return fmt.Errorf("plugin %s not found: %w", name, err)
	}

	if file != nil && p.GenerateFile != nil {
		p.GenerateFile(gen, file)
	} else if file == nil && p.GenerateAll != nil {
		p.GenerateAll(gen)
	}

	return nil
}

func addPlugin(p *Plugin) {
	plugins = append(plugins, p)
	pluginsByName[p.Name] = p
}

func getPlugin(name string) (*Plugin, error) {
	var err error

	// check in registered plugins
	p, ok := pluginsByName[name]
	if !ok {
		// name might be the path to an external plugin
		p, err = loadExternalPlugin(name)
		if err != nil {
			return nil, err
		}
		addPlugin(p)
	}

	return p, nil
}

func loadExternalPlugin(name string) (*Plugin, error) {
	plg, err := plugin.Open(name)
	if err != nil {
		return nil, err
	}

	p := &Plugin{
		Name:     name,
		External: true,
	}

	symGenerateFile, err := plg.Lookup("GenerateFile")
	if err != nil {
		symGenerateAll, err := plg.Lookup("GenerateAll")
		if err != nil {
			return nil, err
		} else {
			p.GenerateAll = symGenerateAll.(GenerateAllFn)
		}
	} else {
		p.GenerateFile = symGenerateFile.(GenerateFileFn)
	}

	return p, nil
}
