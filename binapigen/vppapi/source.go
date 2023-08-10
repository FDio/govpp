//  Copyright (c) 2023 Cisco and/or its affiliates.
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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// TODO:
//  - validate format-specific options parsed by input ref
//  - add support for zip format
//  - add support for json format to allow using JSON-formatted schema as input

type InputFormat string

const (
	FormatDir InputFormat = "dir"
	FormatGit InputFormat = "git"
	FormatTar InputFormat = "tar"
	FormatZip InputFormat = "zip"
)

const (
	OptionFormat = "format"

	OptionGitBranch = "branch"
	OptionGitTag    = "tag"
	OptionGitRef    = "ref"
	OptionGitDepth  = "depth"
	OptionGitSubdir = "subdir"

	OptionArchiveCompression = "compression"
	OptionArchiveSubdir      = "subdir"
	OptionArchiveStrip       = "strip"
)

const (
	defaultGitRef = "master"
)

// InputRef is used to specify reference to VPP API input.
type InputRef struct {
	Path    string
	Format  InputFormat
	Options map[string]string
}

func ParseInputRef(input string) (*InputRef, error) {
	logrus.Tracef("parsing input string: %q", input)

	path, options := parsePathAndOptions(input)

	var inputFormat InputFormat
	format, ok := options[OptionFormat]
	if ok {
		inputFormat = InputFormat(format)
		delete(options, OptionFormat)
	} else {
		inputFormat = detectFormatFromPath(path)
		logrus.Tracef("detected format: %q", inputFormat)
	}

	// Use current working dir by default
	if inputFormat == FormatDir && path == "" {
		path = "."
	}

	ref := &InputRef{
		Format:  inputFormat,
		Path:    path,
		Options: options,
	}
	logrus.Tracef("parsed input ref: %+v", ref)

	return ref, nil

}

func parsePathAndOptions(input string) (path string, options map[string]string) {
	// Split into path and options
	parts := strings.Split(input, "#")
	path = parts[0]
	options = make(map[string]string)
	if len(parts) > 1 {
		// Split into key-value pairs
		optionsList := strings.Split(parts[1], ",")
		for _, option := range optionsList {
			// Split each option into key and value
			keyValue := strings.SplitN(option, "=", 2)
			key := keyValue[0]
			value := ""
			if len(keyValue) > 1 {
				value = keyValue[1]
			}
			options[key] = value
		}
	}
	return path, options
}

func detectFormatFromPath(path string) InputFormat {
	// By suffix
	if strings.HasSuffix(path, ".tar") || strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		return FormatTar
	}
	if strings.HasSuffix(path, ".zip") {
		return FormatZip
	}
	if strings.HasSuffix(path, ".git") {
		return FormatGit
	}

	// By prefix
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "ssh://") || strings.HasPrefix(path, "git://") {
		return FormatGit
	}

	// By default
	return FormatDir
}

func (ref *InputRef) Retrieve() (*VppInput, error) {
	if ref.Path == "" {
		return nil, fmt.Errorf("missing path")
	}
	if ref.Format == "" {
		return nil, fmt.Errorf("undefined format")
	}

	logrus.Tracef("retrieving from ref: %+v", ref)

	switch ref.Format {
	case FormatDir:
		info, err := os.Stat(ref.Path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("directory path %q does not exist", ref.Path)
			} else if errors.Is(err, fs.ErrPermission) {
				return nil, fmt.Errorf("directory path %q does not have sufficient permission", ref.Path)
			}
			return nil, fmt.Errorf("path error: %v", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("path must be a directory")
		}

		return resolveVppInputFromDir(ref.Path)

	case FormatGit:
		branch := ref.Options[OptionGitBranch]
		tag := ref.Options[OptionGitTag]
		rev := ref.Options[OptionGitRef]
		if branch != "" && tag != "" {
			return nil, fmt.Errorf("cannot set both branch and tag")
		} else if branch != "" || tag != "" {
			if rev != "" {
				return nil, fmt.Errorf("cannot set rev if branch or tag is set")
			}
			if branch != "" {
				rev = branch
			} else if tag != "" {
				rev = tag
			}
		}

		commit := rev
		if commit == "" {
			commit = defaultGitRef
		}

		cloneDepth := 0
		if depth := ref.Options[OptionGitDepth]; depth != "" {
			d, err := strconv.Atoi(depth)
			if err != nil {
				return nil, fmt.Errorf("invalid depth: %w", err)
			}
			cloneDepth = d
		}

		logrus.Debugf("updating local repo %s to %s", ref.Path, commit)

		repoDir, err := cloneRepoLocally(ref.Path, commit, branch, cloneDepth)
		if err != nil {
			return nil, err
		}
		dir := repoDir

		subdir, hasSubdir := ref.Options[OptionGitSubdir]
		if hasSubdir {
			dir = filepath.Join(repoDir, subdir)
			if info, err := os.Stat(dir); err != nil {
				if os.IsNotExist(err) {
					return nil, fmt.Errorf("subdirectory %q does not exists", subdir)
				}
				return nil, fmt.Errorf("subdirectory %q err: %w", subdir, err)
			} else if !info.IsDir() {
				return nil, fmt.Errorf("subdirectory must be a directory")
			}
		}

		return resolveVppInputFromDir(dir)

	case FormatTar:
		stripN := 0
		strip, hasStrip := ref.Options[OptionArchiveStrip]
		if hasStrip {
			parsedStrip, err := strconv.Atoi(strip)
			if err != nil {
				return nil, fmt.Errorf("invalid strip value: %s", strip)
			}
			if parsedStrip < 0 {
				return nil, fmt.Errorf("strip must be a non-negative integer")
			}
			stripN = parsedStrip
		}

		gzipped := false
		compression, hasCompression := ref.Options[OptionArchiveCompression]
		if hasCompression {
			if compression == "gzip" {
				gzipped = true
			} else {
				return nil, fmt.Errorf("unknown compression: %s", compression)
			}
		} else {
			if strings.HasSuffix(ref.Path, ".gz") || strings.HasSuffix(ref.Path, ".tgz") {
				gzipped = true
			}
		}

		reader, err := getInputPathReader(ref.Path)
		if err != nil {
			return nil, fmt.Errorf("input path: %w", err)
		}

		logrus.Tracef("extracting tarball %s (gzip: %v)", ref.Path, gzipped)

		tempDir, err := extractTar(reader, gzipped, stripN)
		if err != nil {
			return nil, fmt.Errorf("failed to extract tarball: %w", err)
		}

		dir := tempDir

		subdir, hasSubdir := ref.Options[OptionArchiveSubdir]
		if hasSubdir {
			dir = filepath.Join(tempDir, subdir)
			if info, err := os.Stat(dir); err != nil {
				if os.IsNotExist(err) {
					return nil, fmt.Errorf("subdirectory %q does not exists", subdir)
				}
				return nil, fmt.Errorf("subdirectory %q err: %w", subdir, err)
			} else if !info.IsDir() {
				return nil, fmt.Errorf("subdirectory must be a directory")
			}
		}

		return resolveVppInputFromDir(dir)

	case FormatZip:
		return nil, fmt.Errorf("zip format is not yet supported")

	default:
		return nil, fmt.Errorf("unknown format: %q", ref.Format)
	}
}

func getInputPathReader(path string) (io.ReadCloser, error) {
	var reader io.ReadCloser
	if path == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		reader = file
	}
	return reader, nil
}
