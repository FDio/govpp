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

// FindFiles searches for API files in given directory or in a nested directory
// that is at most one level deeper than dir. This effectively finds all API files
// under core & plugins directories inside API directory.
func FindFiles(dir string) (files []string, err error) {
	return FindFilesRecursive(dir, 1)
}

// FindFilesRecursive searches for API files under dir or in a nested directory that is not
// nested deeper than deep.
func FindFilesRecursive(dir string, deep int) (files []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s failed: %v", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() && deep > 0 {
			nestedDir := filepath.Join(dir, e.Name())
			if nested, err := FindFilesRecursive(nestedDir, deep-1); err != nil {
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

// Parse parses API files in directory DefaultDir and returns collection of File
// or an error if any error occurs during parsing.
func Parse() ([]File, error) {
	return ParseDir(DefaultDir)
}

// ParseDir searches for API files in apiDir, parses the found API files and
// returns collection of File.
//
// API files must have suffix `.api.json` and must be formatted as JSON.
func ParseDir(apiDir string) ([]File, error) {
	list, err := FindFiles(apiDir)
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
		// use file path relative to apiDir
		if path, err := filepath.Rel(apiDir, file.Path); err == nil {
			file.Path = path
		}
		files = append(files, *file)
	}
	return files, nil
}

// ParseFile parses API file and returns File or an error if any error occurs
// during parsing.
func ParseFile(apiFile string) (*File, error) {
	// check API file extension
	if !strings.HasSuffix(apiFile, APIFileExtension) {
		return nil, fmt.Errorf("unsupported file: %q, file must have suffix %q", apiFile, APIFileExtension)
	}

	content, err := os.ReadFile(apiFile)
	if err != nil {
		return nil, fmt.Errorf("read file %s error: %w", apiFile, err)
	}

	// extract file name
	base := filepath.Base(apiFile)
	name := base[:strings.Index(base, ".")]

	logf("parsing file %q (%d bytes)", base, len(content))

	file, err := ParseRaw(content)
	if err != nil {
		return nil, fmt.Errorf("parsing API file %q content failed: %w", base, err)
	}
	file.Name = name
	file.Path = apiFile

	return file, nil
}

// ParseRaw parses raw API file content and returns File or an error if any error
// occurs during parsing.
func ParseRaw(content []byte) (file *File, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic occurred during parsing: %+v", e)
		}
	}()

	file, err = parseJSON(content)
	if err != nil {
		return nil, fmt.Errorf("parseJSON error: %w", err)
	}

	return file, nil
}
