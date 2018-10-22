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

// +build windows darwin

package vppapiclient

import (
	"git.fd.io/govpp.git/adapter"
)

// StatClient is just an stub adapter that does nothing. It builds only on Windows and OSX, where the real
// VPP stats API client adapter does not build. Its sole purpose is to make the compiler happy on Windows and OSX.
type StatClient struct{}

func NewStatClient(socketName string) *StatClient {
	return new(StatClient)
}

func (*StatClient) Connect() error {
	return adapter.ErrNotImplemented
}

func (*StatClient) Disconnect() error {
	return nil
}

func (*StatClient) ListStats(patterns ...string) (statNames []string, err error) {
	return nil, adapter.ErrNotImplemented
}

func (*StatClient) DumpStats(patterns ...string) ([]*adapter.StatEntry, error) {
	return nil, adapter.ErrNotImplemented
}
