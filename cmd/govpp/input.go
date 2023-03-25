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
)

func resolveVppApiInput() string {
	// check if the current working dir is within the VPP repository
	if _, err := os.Stat(filepath.Join(".", "src", "vpp")); err == nil {
		// if true, return the path to the VPP API directory within the repository
		return filepath.Join(".", "build-root", "install-vpp-native", "vpp", "share", "vpp", "api")
	}

	// check if VPP is installed on the system
	if _, err := os.Stat(filepath.Join("/", "usr", "share", "vpp")); err == nil {
		// if true, return the path to the VPP API directory
		return filepath.Join("/", "usr", "share", "vpp", "api")
	}

	// if none of the above conditions are met, return the current working directory
	return "."
}
