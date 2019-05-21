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

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/core"
)

var (
	// VPP binary API adapter that will be used in the subsequent Connect calls
	vppAdapter adapter.VppAPI
)

func getVppAdapter(addr string) adapter.VppAPI {
	if vppAdapter == nil {
		vppAdapter = socketclient.NewVppClient(addr)
	}
	return vppAdapter
}

// SetVppAdapter sets the adapter that will be used for connections to VPP in the subsequent `Connect` calls.
func SetVppAdapter(a adapter.VppAPI) {
	vppAdapter = a
}

// Connect connects the govpp core to VPP either using the default VPP Adapter, or using the adapter previously
// set by SetAdapter (useful mostly just for unit/integration tests with mocked VPP adapter).
// This call blocks until VPP is connected, or an error occurs. Only one connection attempt will be performed.
func Connect(shm string) (*core.Connection, error) {
	return core.Connect(getVppAdapter(shm))
}

// AsyncConnect asynchronously connects the govpp core to VPP either using the default VPP Adapter,
// or using the adapter previously set by SetAdapter.
// This call does not block until connection is established, it returns immediately. The caller is
// supposed to watch the returned ConnectionState channel for Connected/Disconnected events.
// In case of disconnect, the library will asynchronously try to reconnect.
func AsyncConnect(shm string, attempts int, interval time.Duration) (*core.Connection, chan core.ConnectionEvent, error) {
	return core.AsyncConnect(getVppAdapter(shm), attempts, interval)
}
