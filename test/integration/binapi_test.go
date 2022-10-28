//  Copyright (c) 2022 Cisco and/or its affiliates.
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

package integration

import (
	"context"
	"testing"

	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/test/vpptesting"
)

func TestVersion(t *testing.T) {
	test := vpptesting.SetupVPP(t)

	vpeRPC := vpe.NewServiceClient(test.Conn)

	versionInfo, err := vpeRPC.ShowVersion(context.Background(), &vpe.ShowVersion{})
	if err != nil {
		t.Fatalf("getting version failed: %v", err)
	}

	t.Logf("VPP version: %v", versionInfo.Version)
	if versionInfo.Version == "" {
		t.Fatal("expected VPP version to not be empty")
	}
}
