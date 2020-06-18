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

package vppapi

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	DefaultAPIDir = "/usr/share/vpp/api"
)

const apifileSuffixJson = ".api.json"

// FindFiles returns all input files located in specified directory
func FindFiles(dir string, deep int) (paths []string, err error) {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s failed: %v", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() && deep > 0 {
			nestedDir := filepath.Join(dir, e.Name())
			if nested, err := FindFiles(nestedDir, deep-1); err != nil {
				return nil, err
			} else {
				paths = append(paths, nested...)
			}
		} else if strings.HasSuffix(e.Name(), apifileSuffixJson) {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	return paths, nil
}

func Parse() ([]*File, error) {
	return ParseDir(DefaultAPIDir)
}

func ParseDir(apidir string) ([]*File, error) {
	files, err := FindFiles(apidir, 1)
	if err != nil {
		return nil, err
	}

	logrus.Infof("found %d files in API dir %q", len(files), apidir)

	var modules []*File

	for _, file := range files {
		module, err := ParseFile(file)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}

	return modules, nil
}

// ParseFile parses API file contents and returns File.
func ParseFile(apifile string) (*File, error) {
	if !strings.HasSuffix(apifile, apifileSuffixJson) {
		return nil, fmt.Errorf("unsupported file format: %q", apifile)
	}

	data, err := ioutil.ReadFile(apifile)
	if err != nil {
		return nil, fmt.Errorf("reading file %s failed: %v", apifile, err)
	}

	base := filepath.Base(apifile)
	name := base[:strings.Index(base, ".")]

	logf("parsing file %q", base)

	module, err := ParseRaw(data)
	if err != nil {
		return nil, fmt.Errorf("parsing file %s failed: %v", base, err)
	}
	module.Name = name
	module.Path = apifile

	return module, nil
}

func ParseRaw(data []byte) (file *File, err error) {
	file, err = parseJSON(data)
	if err != nil {
		return nil, err
	}

	return file, nil
}
