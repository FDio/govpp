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
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	VPPVersionEnvVar = "VPP_VERSION"
	VPPDirEnvVar     = "VPP_DIR"

	versionScriptPath = "src/scripts/version"
	localBuildRoot    = "build-root/install-vpp-native/vpp/share/vpp/api"
)

// ResolveApiDir checks if parameter dir is a path to directory of local VPP
// repository and returns path to directory with VPP API JSON files under
// build-root. It will execute `make json-api-files` in case the folder with
// VPP API JSON files does not exist yet.
func ResolveApiDir(dir string) string {
	logrus.Tracef("resolving api dir %q", dir)

	_, err := os.Stat(path.Join(dir, "build-root"))
	if err == nil {
		logrus.Tracef("build-root exists, checking %q", localBuildRoot)
		// local VPP build
		_, err := os.Stat(path.Join(dir, localBuildRoot))
		if err == nil {
			logrus.Tracef("returning %q as api dir", localBuildRoot)
			return path.Join(dir, localBuildRoot)
		} else if errors.Is(err, os.ErrNotExist) {
			logrus.Tracef("folder %q does not exist, running 'make json-api-files'", localBuildRoot)
			cmd := exec.Command("make", "json-api-files")
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Dir = dir
			err := cmd.Run()
			if err != nil {
				logrus.Warnf("make json-api-files failed: %v", err)
			} else {
				return path.Join(dir, localBuildRoot)
			}
		} else {
			logrus.Tracef("error occurred when checking %q: %v'", localBuildRoot, err)
		}
	}

	return dir
}

// ResolveVPPVersion resolves version of the VPP for target directory.
//
// Version resolved here can be overriden by setting VPP_VERSION env var.
func ResolveVPPVersion(apidir string) string {
	// check env variable override
	if ver := os.Getenv(VPPVersionEnvVar); ver != "" {
		logrus.Debugf("VPP version was manually set to %q via %s env var", ver, VPPVersionEnvVar)
		return ver
	}

	// check if inside VPP repo
	repoDir, err := findGitRepoRootDir(apidir)
	if err != nil {
		logrus.Debugf("ERR: failed to check VPP git repo: %v", err)
	} else {
		logrus.Debugf("resolved git repo root directory: %v", repoDir)

		version, err := GetVPPVersionRepo(repoDir)
		if err != nil {
			logrus.Debugf("ERR: failed to resolve  VPP version from version script: %v", err)
		} else {
			logrus.Debugf("resolved VPP version from version script: %v", version)
			return version
		}

		// try to read VPP_VERSION file
		data, err := os.ReadFile(path.Join(repoDir, "VPP_VERSION"))
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	// assuming VPP package is installed
	if _, err := exec.LookPath("vpp"); err == nil {
		version, err := GetVPPVersionInstalled()
		if err != nil {
			logrus.Debugf("ERR: failed to resolve VPP version from installed package: %v", err)
		} else {
			logrus.Debugf("resolved VPP version from installed package: %v", version)
			return version
		}
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

// GetVPPVersionRepo retrieves VPP version using script in repo directory.
func GetVPPVersionRepo(repoDir string) (string, error) {
	scriptPath := path.Join(repoDir, versionScriptPath)
	if _, err := os.Stat(scriptPath); err != nil {
		return "", err
	}
	cmd := exec.Command(scriptPath)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("version script failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func findGitRepoRootDir(dir string) (string, error) {
	if conf := os.Getenv(VPPDirEnvVar); conf != "" {
		logrus.Infof("VPP directory was manually set to %q via %s env var", conf, VPPDirEnvVar)
		return conf, nil
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}
