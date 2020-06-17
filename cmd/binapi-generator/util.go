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

package main

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	versionScriptPath = "./src/scripts/version"
	defaultVppApiDir  = "/usr/share/vpp/api"
)

func ResolveVppVersion(inputDir string) string {
	// assuming VPP package is installed
	if inputDir == defaultVppApiDir {
		// resolve VPP version using dpkg
		cmd := exec.Command("dpkg-query", "-f", "${Version}", "-W", "vpp")
		out, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Warnf("resolving VPP version from installed package failed: %v", err)
			logrus.Warnf("command output: %s", out)
		} else {
			version := strings.TrimSpace(string(out))
			logrus.Debugf("resolved VPP version from installed package: %v", version)
			return version
		}
	}
	// check if inside VPP git repo
	if inputDir != "" {
		repo := findVppGitRepo(inputDir)
		if repo != "" {
			cmd := exec.Command(versionScriptPath)
			cmd.Dir = repo
			out, err := cmd.CombinedOutput()
			if err != nil {
				logrus.Warnf("resolving VPP version from version script failed: %v", err)
				logrus.Warnf("command output: %s", out)
			} else {
				version := strings.TrimSpace(string(out))
				logrus.Debugf("resolved VPP version from version script: %v", version)
				return version
			}
		}
		file, err := ioutil.ReadFile(path.Join(inputDir, "VPP_VERSION"))
		if err == nil {
			return strings.TrimSpace(string(file))
		}
	}
	logrus.Warnf("VPP version could not be resolved, you can set it manually using VPP_API_VERSION env var")
	return "unknown"
}

func findVppGitRepo(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Warnf("checking VPP git repo failed: %v", err)
		logrus.Warnf("command output: %s", out)
		return ""
	}
	return strings.TrimSpace(string(out))
}
