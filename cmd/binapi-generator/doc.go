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

// The binapi-generator parses VPP API definitions in JSON format and generates Go code.
//
// The VPP API input can be specified using --input=<INPUT> option, where INPUT
// is one of the following:
//
// - path to directory containing `*.api.json` files (these may be nested under core/plugins)
// - path to local VPP repository (uses files under`build-root/install-vpp-native/vpp/share/vpp/api`)
//
// The generated Go code will be placed into directory specified using
// `--output-dir=<OUTPUT>` option (defaults to `binapi`).
// Each VPP API file will be generated as a separate Go package.
//
// Example:
//
//	binapi-generator --input=/usr/share/vpp/api --output-dir=binapi
package main
