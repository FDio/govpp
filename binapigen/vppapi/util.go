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
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	VPPVersionEnvVar = "VPP_VERSION"
	VPPVersionFile   = "VPP_VERSION"
	VPPDirEnvVar     = "VPP_DIR"

	versionScriptPath   = "src/scripts/version"
	localVPPBuildApiDir = "build-root/install-vpp-native/vpp/share/vpp/api"
)

// ResolveApiDir checks if parameter dir is a path to directory of local VPP
// repository and returns path to directory with VPP API JSON files under
// build-root. It will execute `make json-api-files` in case the folder with
// VPP API JSON files does not exist yet.
func ResolveApiDir(dir string) string {
	log := logrus.WithField("dir", dir)
	log.Tracef("trying to resolve VPP API directory")

	apiDirPath := path.Join(dir, localVPPBuildApiDir)

	// assume dir is a local VPP repository
	_, err := os.Stat(path.Join(dir, "build-root"))
	if err == nil {
		logrus.Tracef("build-root exists in %s, checking %q", dir, localVPPBuildApiDir)

		// check if the API directory exists
		_, err := os.Stat(apiDirPath)
		if err == nil {
			logrus.Tracef("api dir %q exists, running 'make json-api-files'", localVPPBuildApiDir)
			if err := makeJsonApiFiles(dir); err != nil {
				logrus.Warnf("make json-api-files failed: %v", err)
			}
			return apiDirPath
		} else if errors.Is(err, os.ErrNotExist) {
			logrus.Tracef("api dir %q does not exist, running 'make json-api-files'", localVPPBuildApiDir)
			if err := makeJsonApiFiles(dir); err != nil {
				logrus.Warnf("make json-api-files failed: %v", err)
			} else {
				return apiDirPath
			}
		} else {
			logrus.Tracef("error occurred when checking %q: %v'", localVPPBuildApiDir, err)
		}
	}
	return dir
}

func makeJsonApiFiles(dir string) error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("make", "json-api-files")
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	t := time.Now()
	if err := cmd.Run(); err != nil {
		logrus.Debugf("command `%v` failed: %v", cmd, err)
		if stdout.Len() > 0 {
			logrus.Debugf("# STDOUT:\n%v", stdout.String())
		}
		if stderr.Len() > 0 {
			logrus.Debugf("# STDERR:\n%v", stderr.String())
		}
		return fmt.Errorf("command `%v` failed: %v", cmd, err)
	}
	logrus.Debugf("command `%v` done (took %.3fs)\n", cmd, time.Since(t).Seconds())
	return nil
}

// ResolveVPPVersion resolves version of the VPP for target directory.
//
// Version resolved here can be overriden by setting VPP_VERSION env var.
func ResolveVPPVersion(apidir string) string {
	// check if using default dir
	if filepath.Clean(apidir) == DefaultDir {
		// assuming VPP package is installed
		version, err := GetVPPVersionInstalled()
		if err != nil {
			logrus.Tracef("ERR: failed to resolve VPP version from installed package: %v", err)
		} else {
			logrus.Tracef("resolved VPP version from installed package: %v", version)
			return version
		}
	}

	// check if inside VPP repo
	repoDir, err := findGitRepoRootDir(apidir)
	if err != nil {
		logrus.Tracef("failed to check VPP git repo: %v", err)
	} else {
		logrus.Tracef("resolved git repo root directory: %v", repoDir)

		version, err := GetVPPVersionRepo(repoDir)
		if err != nil {
			logrus.Tracef("ERR: failed to resolve  VPP version from version script: %v", err)
		} else {
			logrus.Tracef("resolved VPP version from version script: %v", version)
			return version
		}

		// try to read VPP_VERSION file
		data, err := os.ReadFile(path.Join(repoDir, VPPVersionFile))
		if err == nil {
			ver := strings.TrimSpace(string(data))
			logrus.Tracef("VPP version was resolved to %q from %s file", ver, VPPVersionFile)
			return ver
		}
	}

	// try to read VPP_VERSION file
	data, err := os.ReadFile(path.Join(apidir, VPPVersionFile))
	if err == nil {
		ver := strings.TrimSpace(string(data))
		logrus.Tracef("VPP version was resolved to %q from %s file", ver, VPPVersionFile)
		return ver
	}

	// check env variable override
	if ver := os.Getenv(VPPVersionEnvVar); ver != "" {
		logrus.Tracef("VPP version was manually set to %q via %s env var", ver, VPPVersionEnvVar)
		return ver
	}

	logrus.Tracef("VPP version could not be resolved, you can set it manually using %s env var", VPPVersionEnvVar)
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
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command failed: %v\noutput: %s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}
