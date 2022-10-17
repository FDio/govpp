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

package binapigen

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
	"go.fd.io/govpp/binapigen/vppapi"
)

const (
	generatedJsonPath = "build-root/install-vpp-native/vpp/share/vpp/api/"
)

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
	data, err := os.ReadFile(gomod)
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
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_' && c != '.' && c != '/'
	})
}

func ExpandPaths(paths ...string) string {
	if strings.HasPrefix(paths[0], "~/") {
		dirname, _ := os.UserHomeDir()
		paths[0] = filepath.Join(dirname, paths[0][2:])
	}
	return os.ExpandEnv(filepath.Join(paths...))
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
		return vppapi.DefaultDir
	}
}
