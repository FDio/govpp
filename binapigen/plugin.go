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

import "fmt"

type Plugin struct {
	Name         string
	GenerateFile func(*Generator, *File) *GenFile
}

var Plugins = map[string]*Plugin{}
var plugins []*Plugin

func RegisterPlugin(name string, genfn func(*Generator, *File) *GenFile) {
	if name == "" {
		panic("plugin name empty")
	}
	for _, p := range plugins {
		if p.Name == name {
			panic("duplicate plugin name: " + name)
		}
	}
	plugin := &Plugin{
		Name:         name,
		GenerateFile: genfn,
	}
	plugins = append(plugins, plugin)
	Plugins[name] = plugin
}

func RunPlugin(name string, gen *Generator, file *File) error {
	p, ok := Plugins[name]
	if !ok {
		return fmt.Errorf("plugin not found: %q", name)
	}
	p.GenerateFile(gen, file)
	return nil
}
