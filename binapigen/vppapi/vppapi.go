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
)

const (
	// DefaultDir is default location of API files.
	DefaultDir = "/usr/share/vpp/api"
)

// FindFiles finds API files located in dir or in a nested directory that is not nested deeper than deep.
func FindFiles(dir string, deep int) (files []string, err error) {
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
				files = append(files, nested...)
			}
		} else if !e.IsDir() && strings.HasSuffix(e.Name(), ".api.json") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files, nil
}

// Parse parses API files in directory DefaultDir.
func Parse() ([]*File, error) {
	return ParseDir(DefaultDir)
}

// ParseDir finds and parses API files in given directory and returns parsed files.
// Supports API files in JSON format (.api.json) only.
func ParseDir(apidir string) ([]*File, error) {
	list, err := FindFiles(apidir, 1)
	if err != nil {
		return nil, err
	}

	logf("found %d files in API dir %q", len(list), apidir)

	var files []*File
	for _, file := range list {
		module, err := ParseFile(file)
		if err != nil {
			return nil, err
		}
		files = append(files, module)
	}
	return files, nil
}

// ParseFile parses API file and returns File.
func ParseFile(apifile string) (*File, error) {
	if !strings.HasSuffix(apifile, ".api.json") {
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

// ParseRaw parses raw API file data and returns File.
func ParseRaw(data []byte) (file *File, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic occurred: %v", e)
		}
	}()

	file, err = parseJSON(data)
	if err != nil {
		return nil, err
	}

	return file, nil
}
