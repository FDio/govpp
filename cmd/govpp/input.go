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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen/vppapi"
)

func resolveInput(input string) (*vppapi.VppInput, error) {
	if input == "" {
		logrus.Tracef("input empty, trying to detect automatically")
		input = detectVppApiInput()
	}

	logrus.Tracef("resolving VPP input: %q", input)

	vppInput, err := vppapi.ResolveVppInput(input)
	if err != nil {
		return nil, err
	}

	logrus.Tracef("resolved VPP input:\n - API dir: %s\n - VPP Version: %s\n - Files: %v",
		vppInput.ApiDirectory, vppInput.Schema.Version, len(vppInput.Schema.Files))

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
	if len(paths) == 0 {
		return allapifiles
	}
	added := make(map[string]bool)
	// filter files
	for _, arg := range paths {
		var found bool
		for _, apifile := range allapifiles {
			if added[apifile.Path] {
				continue
			}
			dir, file := path.Split(apifile.Path)
			if apifile.Name == strings.TrimSuffix(arg, ".api") || apifile.Path == arg || file == arg || path.Clean(dir) == arg {
				apifiles = append(apifiles, apifile)
				found = true
				added[apifile.Path] = true
			}
		}
		if !found {
			logrus.Warnf("path %q did not match any file", arg)
		}
	}
	return apifiles
}
