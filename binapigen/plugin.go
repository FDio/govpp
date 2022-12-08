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

type Plugin struct {
	Name         string
	GenerateFile GenerateFileFn
}

type GenerateFileFn = func(*Generator, *File) *GenFile

var plugins []*Plugin
var pluginsByName = map[string]*Plugin{}

func RegisterPlugin(name string, genfn GenerateFileFn) {
	if name == "" {
		panic("plugin name empty")
	}
	if _, ok := pluginsByName[name]; ok {
		panic("duplicate plugin name: " + name)
	}
	p := &Plugin{
		Name:         name,
		GenerateFile: genfn,
	}
	plugins = append(plugins, p)
	pluginsByName[name] = p
}

func RunPlugin(name string, gen *Generator, file *File) error {
	p, err := getPlugin(name)
	if err != nil {
		return fmt.Errorf("plugin %s not found: %w", name, err)
	}

	p.GenerateFile(gen, file)

	return nil
}

func getPlugin(name string) (*Plugin, error) {
	var err error

	// find name in registered plugins
	p, ok := pluginsByName[name]
	if !ok {
		// name might be the path to an external plugin
		p, err = loadExternalPlugin(name)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func loadExternalPlugin(name string) (*Plugin, error) {
	plg, err := plugin.Open(name)
	if err != nil {
		return nil, err
	}

	symGenerateFile, err := plg.Lookup("GenerateFile")
	if err != nil {
		return nil, err
	}

	return &Plugin{
		Name:         name,
		GenerateFile: symGenerateFile.(GenerateFileFn),
	}, nil
}

/*
func RunPlugin(name string, gen *Generator, file *File) error {
	p, ok := pluginsByName[name]
	if !ok {
		return fmt.Errorf("plugin not found: %q", name)
	}
	p.GenerateFile(gen, file)
	return nil
}
*/
