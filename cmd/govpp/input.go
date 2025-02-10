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

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen/vppapi"
)

func resolveVppInput(input string) (*vppapi.VppInput, error) {
	if input == "" {
		logrus.Tracef("VPP input is not set, trying to detect automatically")
		input = detectVppApiInput()
	}

	logrus.Tracef("resolving VPP input: %q\n%s", input, strings.Repeat("-", 100))
	t := time.Now()

	vppInput, err := vppapi.ResolveVppInput(input)
	if err != nil {
		return nil, fmt.Errorf("resolve VPP input: %w", err)
	}

	tookSec := time.Since(t).Seconds()

	logrus.WithFields(map[string]interface{}{
		"took":    fmt.Sprintf("%.3fs", tookSec),
		"version": vppInput.Schema.Version,
		"files":   len(vppInput.Schema.Files),
		"apiDir":  len(vppInput.Schema.Files),
	}).Tracef("resolved VPP input %q\n%s\n - API dir: %s\n - VPP Version: %s\n - Files: %v",
		input, strings.Repeat("-", 100), vppInput.ApiDirectory, vppInput.Schema.Version, len(vppInput.Schema.Files))

	return vppInput, nil
}

func detectVppApiInput() string {
	var (
		relPathSrcVpp      = filepath.Join(".", "src", "vpp")
		relPathBuildVppApi = filepath.Join(".", "build-root", "install-vpp-native", "vpp", "share", "vpp", "api")
	)
	// check if VPP API files are built
	if dirExists(relPathBuildVppApi) {
		return relPathBuildVppApi
	}
	// check if within the VPP repository
	if dirExists(relPathSrcVpp) {
		return relPathBuildVppApi
	}
	// check if within VPP API directory
	if dirExists("core", "plugins") {
		return "."
	}
	// check if VPP is installed on the system
	if dirExists(vppapi.DefaultDir) {
		return vppapi.DefaultDir
	}
	// if none of the above conditions are met, return the current working directory
	return "."
}

func dirExists(dir ...string) bool {
	for _, d := range dir {
		if _, err := os.Stat(d); err != nil {
			return false
		}
	}
	return true
}

func filterFilesByPaths(allapifiles []vppapi.File, paths []string) []vppapi.File {
	var apifiles []vppapi.File

	// filter files
	added := make(map[string]bool)
	for _, p := range paths {
		var found bool
		for _, apifile := range allapifiles {
			if added[apifile.Path] {
				continue
			}
			if fileMatchesPath(apifile, p) {
				apifiles = append(apifiles, apifile)
				found = true
				added[apifile.Path] = true
			}
		}
		if !found {
			logrus.Debugf("path %q did not match any file", p)
		}
	}
	return apifiles
}

func fileMatchesPath(apifile vppapi.File, arg string) bool {
	if apifile.Name == strings.TrimSuffix(arg, ".api") {
		return true
	}
	if apifile.Path == arg {
		return true
	}
	dir, file := path.Split(apifile.Path)
	return file == arg || path.Clean(dir) == arg
}
