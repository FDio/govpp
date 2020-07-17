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
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	VPPVersionEnvVar = "VPP_VERSION"
)

// ResolveVPPVersion resolves version of the VPP for target directory.
//
// Version resolved here can be overriden by setting VPP_VERSION env var.
func ResolveVPPVersion(apidir string) string {
	// check env variable override
	if ver := os.Getenv(VPPVersionEnvVar); ver != "" {
		logrus.Debugf("VPP version was manually set to %q via %s env var", ver, VPPVersionEnvVar)
		return ver
	}

	// assuming VPP package is installed
	if path.Clean(apidir) == DefaultDir {
		version, err := GetVPPVersionInstalled()
		if err != nil {
			logrus.Warnf("resolving VPP version from installed package failed: %v", err)
		} else {
			logrus.Debugf("resolved VPP version from installed package: %v", version)
			return version
		}
	}

	// check if inside VPP repo
	repoDir, err := findGitRepoRootDir(apidir)
	if err != nil {
		logrus.Warnf("checking VPP git repo failed: %v", err)
	} else {
		logrus.Debugf("resolved git repo root directory: %v", repoDir)
		version, err := GetVPPVersionRepo(repoDir)
		if err != nil {
			logrus.Warnf("resolving VPP version from version script failed: %v", err)
		} else {
			logrus.Debugf("resolved VPP version from version script: %v", version)
			return version
		}
	}

	// try to read VPP_VERSION file
	data, err := ioutil.ReadFile(path.Join(repoDir, "VPP_VERSION"))
	if err == nil {
		return strings.TrimSpace(string(data))
	}

	logrus.Warnf("VPP version could not be resolved, you can set it manually using %s env var", VPPVersionEnvVar)
	return ""
}

// GetVPPVersionInstalled retrieves VPP version of installed package using dpkg-query.
func GetVPPVersionInstalled() (string, error) {
	cmd := exec.Command("dpkg-query", "-f", "${Version}", "-W", "vpp")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("dpkg-query command failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

const versionScriptPath = "./src/scripts/version"

// GetVPPVersionRepo retrieves VPP version using script in repo directory.
func GetVPPVersionRepo(repoDir string) (string, error) {
	if _, err := os.Stat(versionScriptPath); err != nil {
		return "", err
	}
	cmd := exec.Command(versionScriptPath)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("version script failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func findGitRepoRootDir(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}
