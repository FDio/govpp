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

// Package vppapi parses VPP API files without any additional processing.
package vppapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultDir is default location of API files.
	DefaultDir = "/usr/share/vpp/api"

	// APIFileExtension is a VPP API file extension suffix
	APIFileExtension = ".api.json"
)

// FindFiles finds API files located in dir or in a nested directory that is not nested deeper than deep.
func FindFiles(dir string, deep int) (files []string, err error) {
	entries, err := os.ReadDir(dir)
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
		} else if !e.IsDir() && strings.HasSuffix(e.Name(), APIFileExtension) {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files, nil
}

// Parse parses API files in directory DefaultDir.
func Parse() ([]File, error) {
	return ParseDir(DefaultDir)
}

// ParseDir finds and parses API files in given directory and returns parsed files.
// Supports API files in JSON format (.api.json) only.
func ParseDir(apiDir string) ([]File, error) {
	list, err := FindFiles(apiDir, 1)
	if err != nil {
		return nil, err
	}

	logf("found %d files in API dir %q", len(list), apiDir)

	var files []File
	for _, f := range list {
		file, err := ParseFile(f)
		if err != nil {
			return nil, err
		}
		if path, err := filepath.Rel(apiDir, file.Path); err == nil {
			file.Path = path
		}
		files = append(files, *file)
	}
	return files, nil
}

// ParseFile parses API file and returns File.
func ParseFile(apiFile string) (*File, error) {
	if !strings.HasSuffix(apiFile, APIFileExtension) {
		return nil, fmt.Errorf("unsupported file format: %q", apiFile)
	}

	data, err := os.ReadFile(apiFile)
	if err != nil {
		return nil, fmt.Errorf("reading file %s failed: %v", apiFile, err)
	}

	base := filepath.Base(apiFile)
	name := base[:strings.Index(base, ".")]

	logf("parsing file %q", base)

	module, err := ParseRaw(data)
	if err != nil {
		return nil, fmt.Errorf("parsing file %s failed: %v", base, err)
	}
	module.Name = name
	module.Path = apiFile

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
