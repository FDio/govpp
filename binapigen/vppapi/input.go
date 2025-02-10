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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// ResolveVppInput parses an input string and returns VppInput or an error if a problem
// occurs durig parsing. The input string consists of path, which can be followed
// by '#' and one or more options separated by comma. The actual format can be
// specified explicitely by an option 'format', if that is not the case then the
// format will be detected from path automatically if possible.
//
// - `path`
// - `path#option1=val,option2=val`
//
// Available formats:
//
// * Directory: `dir`
//   - absolute: `/usr/share/vpp/api`
//   - relative: `path/to/apidir`
//
// * Git repository: `git`
//   - local repository: `.git`
//   - remote repository: `https://github.com/FDio/vpp.git`
//
// * Tarball/Zip Archive (`tar`/`zip`)
//   - local archive:  `api.tar.gz`
//   - remote archive: `https://example.com/api.tar.gz`
//   - standard input: `-`
//
// Format options:
//
// * Git repository
//   - `branch`: name of branch
//   - `tag`:    specific git tag
//   - `ref`:    git reference
//   - `depth`:  git depth
//   - `subdir`: subdirectory to use as base directory
//
// * Tarball/ZIP Archive
//   - `compression`: compression to use (`gzip`)
//   - `subdir`:      subdirectory to use as base directory
//   - 'strip':       strip first N directories, applied before `subdir`
//
// Returns VppInput on success.
func ResolveVppInput(input string) (*VppInput, error) {
	inputRef, err := ParseInputRef(input)
	if err != nil {
		return nil, err
	}
	v, err := inputRef.Retrieve()
	if err != nil {
		return v, fmt.Errorf("retrieve error: %w", err)
	}
	return v, nil
}

// VppInput defines VPP API input source.
type VppInput struct {
	ApiDirectory string
	Schema       Schema
}

func resolveVppInputFromDir(path string) (*VppInput, error) {
	vppInput := new(VppInput)

	apidir := ResolveApiDir(path)
	vppInput.ApiDirectory = apidir

	logrus.WithField("path", path).Tracef("resolved API dir: %q", apidir)

	apiFiles, err := ParseDir(apidir)
	if err != nil {
		//logrus.Warnf("vppapi parsedir error: %v", err)
		return nil, fmt.Errorf("parsing API dir %s failed: %w", apidir, err)
	}
	vppInput.Schema.Files = apiFiles
	logrus.Tracef("resolved %d apifiles", len(apiFiles))

	vppVersion := ResolveVPPVersion(path)
	if vppVersion == "" {
		vppVersion = "unknown"
	}
	vppInput.Schema.Version = vppVersion

	return vppInput, nil
}

const (
	cacheDir = "./.cache"
)

func cloneRepoLocally(repo string, commit string, branch string, depth int) (string, error) {
	repoPath := strings.ReplaceAll(repo, "/", "_")
	repoPath = strings.ReplaceAll(repoPath, ":", "_")
	cachePath := filepath.Join(cacheDir, repoPath)

	// Clone the repository or use cached one
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create cache directory: %w", err)
		}

		args := []string{"--single-branch"}
		if depth > 0 {
			args = append(args, fmt.Sprintf("--depth=%d", depth))
		}
		args = append(args, repo, cachePath)
		logrus.Debugf("cloning repo: %v", args)
		cmd := exec.Command("git", append([]string{"clone"}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, output)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to check if cache exists: %w", err)
	}
	logrus.Debugf("using local repo dir: %q, fetching %q", cachePath, commit)

	cmd := exec.Command("git", "fetch", "--tags", "origin")
	cmd.Dir = cachePath
	if output, err := cmd.CombinedOutput(); err != nil {
		logrus.Debugf("ERROR: failed to fetch tags: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "fetch", "-f", "origin", commit)
	cmd.Dir = cachePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to fetch commit: %w\nOutput: %s", err, output)
	}

	// Resolve the commit hash for the given branch/tag
	ref := commit
	if branch != "" {
		ref = "origin/" + branch
	}
	commitHash, err := resolveCommitHash(cachePath, ref)
	if err != nil {
		return "", fmt.Errorf("failed to resolve commit hash: %w", err)
	}

	// Check out the repository at the resolved commit
	cmd = exec.Command("git", "checkout", commitHash)
	cmd.Dir = cachePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to check out repository: %w\nOutput: %s", err, output)
	}

	return cachePath, nil
}

func resolveCommitHash(repoPath, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	cmd.Dir = repoPath

	outputBytes, err := cmd.Output()
	if err != nil {
		logrus.Tracef("[ERR] command %q output: %s", cmd, outputBytes)
		return "", fmt.Errorf("failed to run command: %w", err)
	} else {
		logrus.Tracef("[OK] command %q output: %s", cmd, outputBytes)
	}

	return strings.TrimSpace(string(outputBytes)), nil
}

func extractTar(reader io.Reader, gzipped bool, strip int) (string, error) {
	tempDir, err := os.MkdirTemp("", "govpp-vppapi-archive-")
	if err != nil {
		return "", err
	}

	if gzipped {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return "", err
		}
		defer gzReader.Close()

		reader = gzReader
	}

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		nameList := strings.Split(header.Name, "/")
		if len(nameList) < strip {
			return "", fmt.Errorf("cannot strip first %d directories for file: %s", strip, header.Name)
		}
		fileName := filepath.Join(nameList[strip:]...)
		filePath := filepath.Join(tempDir, fileName)
		fileInfo := header.FileInfo()

		logrus.Tracef(" - extracting tar file: %s into: %s", header.Name, filePath)

		if fileInfo.IsDir() {
			err = os.MkdirAll(filePath, fileInfo.Mode())
			if err != nil {
				return "", err
			}
			continue
		} else {
			err = os.MkdirAll(filepath.Dir(filePath), 0750)
			if err != nil {
				return "", err
			}
		}
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileInfo.Mode())
		if err != nil {
			return "", fmt.Errorf("open file error: %w", err)
		}
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return "", fmt.Errorf("copy data error: %w", err)
		}
		if err := file.Close(); err != nil {
			return "", fmt.Errorf("close file error: %w", err)
		}
	}
	return tempDir, nil
}
