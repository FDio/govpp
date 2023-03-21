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
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// VppInput defines VPP input parameters for the Generator.
type VppInput struct {
	ApiDirectory string
	ApiFiles     []File
	VppVersion   string
}

// ResolveVppInput resolves given input string into VppInput.
//
// Supported input formats are:
//   - directory with VPP API JSON files (e.g. `/usr/share/vpp/api/`)
//   - directory with VPP repository (runs `make json-api-files`)
func ResolveVppInput(input string) (*VppInput, error) {

	vppInput := &VppInput{}

	if input == "" {
		input = DefaultDir
		vppInput.ApiDirectory = DefaultDir
	}

	u, err := url.Parse(input)
	if err != nil {
		logrus.Tracef("VPP input %q is invalid: %v", input, err)
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	switch u.Scheme {
	case "":
		fallthrough // assume file by default

	case "file":
		info, err := os.Stat(input)
		if err != nil {
			return nil, fmt.Errorf("file error: %v", err)
		} else {
			if info.IsDir() {
				return resolveVppInputFromDir(u.Path)
			} else {
				return nil, fmt.Errorf("files not supported")
			}
		}

	case "http", "https":
		fallthrough

	case "git", "ssh":
		commit := u.Fragment
		u.Fragment = ""
		repo := u.String()

		dir, err := checkoutRepoLocally(repo, commit)
		if err != nil {
			return nil, err
		}

		return resolveVppInputFromDir(dir)

	default:
		return nil, fmt.Errorf("unsupported scheme: %v", u.Scheme)
	}

}

func resolveVppInputFromDir(path string) (*VppInput, error) {
	vppInput := new(VppInput)

	apidir := ResolveApiDir(path)
	vppInput.ApiDirectory = apidir

	logrus.Debugf("path %q resolved to api dir: %v", path, apidir)

	apiFiles, err := ParseDir(apidir)
	if err != nil {
		logrus.Warnf("vppapi parsedir error: %v", err)
	} else {
		vppInput.ApiFiles = apiFiles
		logrus.Debugf("resolved %d apifiles", len(apiFiles))
	}

	vppInput.VppVersion = ResolveVPPVersion(path)
	if vppInput.VppVersion == "" {
		vppInput.VppVersion = "unknown"
	}

	return vppInput, nil
}

const cacheDir = "./.cache"

func checkoutRepoLocally(repo string, commit string /*, command string, args ...string*/) (string, error) {
	repoPath := strings.ReplaceAll(repo, "/", "_")
	repoPath = strings.ReplaceAll(repoPath, ":", "_")
	cachePath := filepath.Join(cacheDir, repoPath)

	// Clone the repository or use cached one
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create cache directory: %w", err)
		}
		logrus.Infof("CLONING")
		cmd := exec.Command("git", "clone", repo, cachePath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, output)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to check if cache exists: %w", err)
	} else {
		logrus.Infof("FETCHING")
		cmd := exec.Command("git", "fetch", "origin", commit)
		cmd.Dir = cachePath
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to fetch repository: %w\nOutput: %s", err, output)
		}
	}

	// Resolve the commit hash for the given branch/tag
	commitHash, err := resolveCommitHash(cachePath, commit)
	if err != nil {
		return "", fmt.Errorf("failed to resolve commit hash: %w", err)
	}

	// Check out the repository at the resolved commit
	cmd := exec.Command("git", "checkout", commitHash)
	cmd.Dir = cachePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to check out repository: %w\nOutput: %s", err, output)
	}

	return cachePath, nil
}

func resolveCommitHash(repoPath, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	cmd.Dir = repoPath

	output, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create output pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}
	outputBytes, err := io.ReadAll(output)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	return strings.TrimSpace(string(outputBytes)), nil
}
