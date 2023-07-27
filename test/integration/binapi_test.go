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
	"io"
	"testing"

	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/test/vpptesting"
)

// TestVersion tests that getting VPP version works.
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

// TestInterfacesLoopback tests that creating a loopback interface works and returns
// non-zero ID and that the interface will be listed in interface dump.
func TestInterfacesLoopback(t *testing.T) {
	test := vpptesting.SetupVPP(t)
	ctx := context.Background()

	ifacesRPC := interfaces.NewServiceClient(test.Conn)

	// create loopback interface
	reply, err := ifacesRPC.CreateLoopback(ctx, &interfaces.CreateLoopback{})
	if err != nil {
		t.Fatal("interfaces.CreateLoopback error:", err)
	}
	loopId := reply.SwIfIndex
	t.Logf("loopback interface created (id: %v)", reply.SwIfIndex)

	// list interfaces
	stream, err := ifacesRPC.SwInterfaceDump(ctx, &interfaces.SwInterfaceDump{})
	if err != nil {
		t.Fatal("interfaces.SwInterfaceDump error:", err)
	}

	t.Log("Dumping interfaces")
	foundLoop := false
	numIfaces := 0
	for {
		iface, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal("interfaces.SwInterfaceDump/Recv error:", err)
		}
		numIfaces++
		t.Logf("- interface[%d]: %q\n", iface.SwIfIndex, iface.InterfaceName)
		if iface.SwIfIndex == loopId {
			foundLoop = true
		}
	}

	// verify expected
	if !foundLoop {
		t.Fatalf("loopback interface (id: %v) not found", loopId)
	}
	if numIfaces != 2 {
		t.Errorf("expected 2 interfaces in dump, got %d", numIfaces)
	}
}
