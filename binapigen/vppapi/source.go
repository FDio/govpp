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
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	FormatDir = "dir"
	FormatGit = "git"
	FormatTar = "tar"
	FormatZip = "zip"

	//FormatJson = "json"
)

// Input is a
type Input struct {
	Path    string
	Format  string
	Options map[string]string
}

// ParseInput parses an input string and returns Input or an error if a problem
// occurs durig parsing. The input string consists of path, which can be followed
// by '#' and one or more options separated by comma. The actual format can be
// specified explicitely by an option 'format', if that is not the case then the
// format will be detected from path automatically if possible.
//
//	path
//	path#key=val,key2=val
//
// Supported formats are:
//
// * Directory (dir)
//   - /usr/share/vpp/api (absolute)
//   - ./api (relative)
//
// * Repository (git)
//   - .git (local)
//   - https://github.com/FDio/vpp.git (remote)
//
// * Archive (tar, zip)
//   - api.tar.gz (gzipped tar)
//   - api.zip (zip)
//   - https://example.com/api.tar.gz (remote)
//
// Supported format options:
//
//   - Directory
//     -
//   - Repository
//   - branch: name of branch to checkout
//   - tag: tag to checkout
//   - ref: ref to checkout
//   - Archive
//     -
func ParseInput(inputStr string) (*Input, error) {
	logrus.Tracef("parsing input string: %q", inputStr)

	path, options := parseInputOptions(inputStr)

	format, ok := options["format"]
	if ok {
		delete(options, "format")
	} else {
		format = detectFormat(path)
	}

	if format == FormatDir && path == "" {
		path = "."
	}

	input := &Input{
		Format:  format,
		Path:    path,
		Options: options,
	}

	logrus.Tracef("Input: %+v", input)

	return input, nil

}

func parseInputOptions(input string) (path string, options map[string]string) {
	// Split input string into path and options
	parts := strings.Split(input, "#")
	path = parts[0]
	options = make(map[string]string)

	if len(parts) > 1 {
		// Split options into key-value pairs
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

func detectFormat(path string) string {
	// archive
	if strings.HasSuffix(path, ".tar") || strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		return FormatTar
	}
	if strings.HasSuffix(path, ".zip") {
		return FormatZip
	}
	// git
	if strings.HasSuffix(path, ".git") || strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "ssh://") || strings.HasPrefix(path, "git://") {
		return FormatGit
	}
	// directory
	return FormatDir
}
