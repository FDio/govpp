// Copyright (c) 2017 Cisco and/or its affiliates.
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

import (
	"time"

	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/core"
	"go.fd.io/govpp/internal/version"
)

// Connect connects to the VPP API using a new adapter instance created with NewVppAPIAdapter.
//
// This call blocks until VPP is connected, or an error occurs.
// Only one connection attempt will be performed.
func Connect(target string) (*core.Connection, error) {
	return core.Connect(NewVppAdapter(target))
}

// AsyncConnect asynchronously connects to the VPP API using a new adapter instance
// created with NewVppAPIAdapter.
//
// This call does not block until connection is established, it returns immediately.
// The caller is supposed to watch the returned ConnectionState channel for connection events.
// In case of disconnect, the library will asynchronously try to reconnect.
func AsyncConnect(target string, attempts int, interval time.Duration) (*core.Connection, chan core.ConnectionEvent, error) {
	return core.AsyncConnect(NewVppAdapter(target), attempts, interval)
}

// NewVppAdapter returns new instance of VPP adapter for connecting to VPP API.
var NewVppAdapter = func(target string) adapter.VppAPI {
	return socketclient.NewVppClient(target)
}

// Version returns version of GoVPP.
func Version() string {
	return version.Version()
}
