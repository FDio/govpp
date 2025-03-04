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

	// APIFileExtension is a VPP API file extension suffix.
	APIFileExtension = ".api.json"
)

// FindFiles searches for API files in specified directory or in a subdirectory
// that is at most 1-level deeper than dir. This effectively finds all API files
// under core & plugins directories inside API directory. The returned list of
// files will contain paths relative to dir.
func FindFiles(dir string) (files []string, err error) {
	return FindFilesRecursive(dir, 1)
}

// FindFilesRecursive recursively searches for API files inside specified directory
// or a subdirectory that is not nested more than deep. The returned list of files
// will contain paths relative to dir.
func FindFilesRecursive(dir string, deep int) (files []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %q error: %v", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() && deep > 0 {
			nestedFiles, err := FindFilesRecursive(filepath.Join(dir, e.Name()), deep-1)
			if err != nil {
				return nil, err
			}
			for _, nestedFile := range nestedFiles {
				files = append(files, filepath.Join(e.Name(), nestedFile))
			}
		} else if !e.IsDir() && strings.HasSuffix(e.Name(), APIFileExtension) {
			files = append(files, e.Name())
		}
	}
	return files, nil
}

// ParseDefault parses API files in the directory DefaultDir, which is a default
// location of the API files for VPP installed on the host system, and returns list
// of File or an error if any occurs.
func ParseDefault() ([]File, error) {
	// check if DefaultDir directory exists
	if _, err := os.Stat(DefaultDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("default API directory %s does not exist", DefaultDir)
	} else if err != nil {
		return nil, err
	}
	return ParseDir(DefaultDir)
}

// ParseDir searches for API files in apiDir, parses them and returns list of
// File or an error if any occurs during searching or parsing.
// The returned files will have Path field set to a path relative to apiDir.
//
// API files must have suffix `.api.json` and must be formatted as JSON.
func ParseDir(apiDir string) ([]File, error) {
	// prepare list of files to parse
	list, err := FindFiles(apiDir)
	if err != nil {
		return nil, err
	}

	logf("found %d files in API dir %q", len(list), apiDir)

	var files []File
	for _, f := range list {
		file, err := ParseFile(filepath.Join(apiDir, f))
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
// during parsing. The retrurned file will have Path field set to apiFile.
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

	logf("Parsing file: %q (%d bytes)", base, len(content))

	file, err := ParseRaw(content)
	if err != nil {
		return nil, fmt.Errorf("parsing API file %q failed: %w", base, err)
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

	file, err = parseApiJsonFile(content)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return file, nil
}
