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
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
)

const (
	VPPVersionEnvVar  = "VPP_VERSION"
	VPPDirEnvVar      = "VPP_DIR"
	versionScriptPath = "./src/scripts/version"
	generatedJsonPath = "build-root/install-vpp-native/vpp/share/vpp/api/"
)

// ResolveVPPVersion resolves version of the VPP for target directory.
//
// Version resolved here can be overriden by setting VPP_VERSION env var.
func ResolveVPPVersion(apidir string) string {
	// check env variable override
	if ver := os.Getenv(VPPVersionEnvVar); ver != "" {
		logrus.Infof("VPP version was manually set to %q via %s env var", ver, VPPVersionEnvVar)
		return ver
	}

	// assuming VPP package is installed
	if _, err := exec.LookPath("vpp"); err == nil {
		version, err := GetVPPVersionInstalled()
		if err != nil {
			logrus.Warnf("resolving VPP version from installed package failed: %v", err)
		} else if version != "" {
			logrus.Infof("resolved VPP version from installed package: %v", version)
			return version
		}
	}

	// check if inside VPP repo
	repoDir, err := FindGitRepoRootDir(apidir)
	if err != nil {
		logrus.Warnf("checking VPP git repo failed: %v", err)
	} else {
		logrus.Debugf("resolved git repo root directory: %v", repoDir)
		version, err := GetVPPVersionRepo(repoDir)
		if err != nil {
			logrus.Warnf("resolving VPP version from version script failed: %v", err)
		} else if version != "" {
			logrus.Infof("resolved VPP version from version script: %v", version)
			return version
		}
	}

	// try to read VPP_VERSION file
	data, err := ioutil.ReadFile(path.Join(repoDir, "VPP_VERSION"))
	if err == nil {
		return strings.TrimSpace(string(data))
	}

	logrus.Warnf("VPP version could not be resolved, you can set it manually using %s env var", VPPVersionEnvVar)
	return "unknown"
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

func FindGitRepoRootDir(dir string) (string, error) {
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

// resolveImportPath tries to resolve import path for a directory.
func ResolveImportPath(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	modRoot := FindGoModuleRoot(absPath)
	if modRoot == "" {
		return "", err
	}
	modPath, err := ReadModulePath(path.Join(modRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	relDir, err := filepath.Rel(modRoot, absPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(modPath, relDir), nil
}

// FindGoModuleRoot looks for enclosing Go module.
func FindGoModuleRoot(dir string) (root string) {
	dir = filepath.Clean(dir)
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}

var modulePathRE = regexp.MustCompile(`module[ \t]+([^ \t\r\n]+)`)

// ReadModulePath reads module path from go.mod file.
func ReadModulePath(gomod string) (string, error) {
	data, err := ioutil.ReadFile(gomod)
	if err != nil {
		return "", err
	}
	m := modulePathRE.FindSubmatch(data)
	if m == nil {
		return "", err
	}
	return string(m[1]), nil
}

// SplitAndStrip takes a string and splits it any separator
func SplitAndStrip(s string) []string {
	return strings.FieldsFunc(s, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_'
	})
}

func ExpandPaths(paths ...string) string {
	if strings.HasPrefix(paths[0], "~/") {
		dirname, _ := os.UserHomeDir()
		paths[0] = filepath.Join(dirname, paths[0][2:])
	}
	return os.ExpandEnv(filepath.Join(paths...))
}

func GetApiFileDirectory(apiSrcDir, vppSrcDir string) string {
	if apiSrcDir != "" {
		return ExpandPaths(apiSrcDir)
	} else if vppSrcDir != "" {
		// Get directory containing the binAPI package
		cmd := exec.Command("make", "json-api-files")
		cmd.Dir = ExpandPaths(vppSrcDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			logrus.Fatalf("Failed to 'make json-api-files' : %s", err)
			return ""
		}
		return ExpandPaths(vppSrcDir, generatedJsonPath)
	} else {
		return DefaultDir
	}
}

func getGoModuleFromPath(path string) string {
	cmd := exec.Command("go", "list")
	cmd.Dir = path
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		logrus.Debugf("Failed to 'go list' : %s", err)
		return ""
	}
	soutput := strings.TrimSpace(string(output))
	if soutput == "" {
		logrus.Debugf("'go list' did not return anything")
		return ""
	}
	return soutput
}

func GetTargetPackagePath(userPackageName string, outputPath string) string {
	if userPackageName != "" {
		return userPackageName
	}
	modPath := getGoModuleFromPath(outputPath)
	if modPath != "" {
		return modPath
	} else if filepath.IsAbs(outputPath) {
		logrus.Fatalf("Failed find go module name for %s", outputPath)
	}
	modPath = getGoModuleFromPath(".")
	if modPath == "" {
		logrus.Fatalf("Failed find relative go module name for %s", outputPath)
	}

	return filepath.Join(modPath, outputPath)
}
