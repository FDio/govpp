// Copyright (c) 2018 Cisco and/or its affiliates.
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

// Package socketclient is a pure Go implementation of adapter.VppAPI, which uses
// unix domain sockets as the transport for connecting to the VPP binary API.
//
// The current implementation only supports VPP binary API, the VPP stats API
// is not supported and clients still have to use vppapiclient for retrieving stats.
//
// # Requirements
//
// The socketclient connects to unix domain socket defined in VPP configuration.
//
// It is enabled by default at /run/vpp/api.sock by the following config section:
//
//	socksvr {
//		socket-name default
//	}
//
// If you want to use custom socket path:
//
//	socksvr {
//		socket-name /run/vpp/api.sock
//	}
package socketclient
