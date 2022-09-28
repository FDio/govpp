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

// Context of Go structs out of the VPP binary API definitions in JSON format.
//
// The JSON input can be specified as a single file (using the `input-file`
// CLI flag), or as a directory that will be scanned for all `.json` files
// (using the `input-dir` CLI flag). The generated Go bindings will  be
// placed into `output-dir` (by default the current working directory),
// where each Go package will be placed into its own separate directory,
// for example:
//
//	binapi-generator --input-file=/usr/share/vpp/api/core/interface.api.json --output-dir=.
package main
