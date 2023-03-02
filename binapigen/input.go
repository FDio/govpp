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

package binapigen

import (
	"fmt"
	"net/url"
	"os"

	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/binapigen/vppapi"
)

// VppInput defines VPP input parameters for the Generator.
type VppInput struct {
	ApiFiles   []*vppapi.File
	VppVersion string
}

// ResolveVppInput resolves given input string into VppInput.
//
// Supported input formats are:
//   - directory with VPP API JSON files (e.g. `/usr/share/vpp/api/`)
//   - directory with VPP repository (runs `make json-api-files`)
func ResolveVppInput(input string) (*VppInput, error) {
	vppInput := &VppInput{}

	if input == "" {
		input = vppapi.DefaultDir
	}

	u, err := url.Parse(input)
	if err != nil {
		logrus.Warnf("parsing url error: %v", err)
	} else {
		switch u.Scheme {
		case "", "file":
			info, err := os.Stat(input)
			if err != nil {
				return nil, fmt.Errorf("file error: %v", err)
			} else {
				if info.IsDir() {
					apidir := vppapi.ResolveApiDir(u.Path)
					logrus.Debugf("path %q resolved to api dir: %v", u.Path, apidir)

					apiFiles, err := vppapi.ParseDir(apidir)
					if err != nil {
						logrus.Warnf("vppapi parsedir error: %v", err)
					} else {
						vppInput.ApiFiles = apiFiles
						logrus.Infof("resolved %d apifiles", len(apiFiles))
					}

					vppInput.VppVersion = vppapi.ResolveVPPVersion(u.Path)
					if vppInput.VppVersion == "" {
						vppInput.VppVersion = "unknown"
					}
				} else {
					return nil, fmt.Errorf("files not supported")
				}
			}
		case "http", "https":
			return nil, fmt.Errorf("http(s) not yet supported")
		case "git", "ssh":
			return nil, fmt.Errorf("ssh/git not yet supported")
		default:
			return nil, fmt.Errorf("unsupported scheme: %v", u.Scheme)
		}
	}

	return vppInput, nil
}
