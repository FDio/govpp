// Copyright (c) 2020 Cisco and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package govpp

//go:generate bash -c "docker run -e PKGPREFIX=$(go list)/binapi -v $(go list -f '{{ .Dir }}'):/gen $(docker build . -q --build-arg GOVPP_VERSION=$(go list -m -f '{{ .Version }}'  go.fd.io/govpp))"
