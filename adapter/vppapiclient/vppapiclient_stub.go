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

// +build windows darwin

package vppapiclient

import (
	"git.fd.io/govpp.git/adapter"
)

// VppClient is just an stub adapter that does nothing. It builds only on Windows and OSX, where the real
// VPP binary API client adapter does not build. Its sole purpose is to make the compiler happy on Windows and OSX.
type BinapiClient struct{}

func NewVppAdapter(string) *BinapiClient {
	return &BinapiClient{}
}

func (a *BinapiClient) Connect() error {
	return adapter.ErrNotImplemented
}

func (a *BinapiClient) Disconnect() error {
	return nil
}

func (a *BinapiClient) GetMsgID(msgName string, msgCrc string) (uint16, error) {
	return 0, nil
}

func (a *BinapiClient) SendMsg(clientID uint32, data []byte) error {
	return nil
}

func (a *BinapiClient) SetMsgCallback(cb adapter.MsgCallback) {
	// no op
}

func (a *BinapiClient) WaitReady() error {
	return adapter.ErrNotImplemented
}
