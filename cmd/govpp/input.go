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
	"path/filepath"

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
	// check if the current working dir is within the VPP repository
	if _, err := os.Stat(filepath.Join(".", "src", "vpp")); err == nil {
		// if true, return the path to the VPP API directory within the repository
		return filepath.Join(".", "build-root", "install-vpp-native", "vpp", "share", "vpp", "api")
	}

	// check if VPP is installed on the system
	if _, err := os.Stat(vppapi.DefaultDir); err == nil {
		// if true, return the path to the VPP API directory
		return vppapi.DefaultDir
	}

	// if none of the above conditions are met, return the current working directory
	return "."
}
