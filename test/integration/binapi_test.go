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
	"io"
	"testing"

	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/memclnt"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/test/vpptesting"
)

// TestVersion tests getting VPP version works.
func TestVersion(t *testing.T) {
	test := vpptesting.SetupVPP(t)
	ctx := test.Context

	c := vpe.NewServiceClient(test.Conn)

	reply, err := c.ShowVersion(ctx, &vpe.ShowVersion{})
	if err != nil {
		t.Fatalf("getting version failed: %v", err)
	}

	t.Logf("VPP version: %v", reply.Version)
	if reply.Version == "" {
		t.Fatal("expected VPP version to not be empty")
	}
}

// TestAPIJSON tests getting VPP API JSON.
func TestAPIJSON(t *testing.T) {
	test := vpptesting.SetupVPP(t)
	ctx := test.Context

	if test.VPPVersion() < "24.06" {
		t.Skip("memclnt.GetAPIJSON is not supported in VPP < 24.06")
	}

	c := memclnt.NewServiceClient(test.Conn)

	reply, err := c.GetAPIJSON(ctx, &memclnt.GetAPIJSON{})
	if err != nil {
		t.Fatalf("getting version failed: %v", err)
	}

	t.Logf("JSON:\n%s", reply.JSON)
}

// TestInterfacesLoopback tests creating a loopback interface works and returns
// non-zero ID and that the interface will be listed in interface dump.
func TestInterfacesLoopback(t *testing.T) {
	test := vpptesting.SetupVPP(t)
	ctx := test.Context

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
