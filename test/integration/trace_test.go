//  Copyright (c) 2023 Cisco and/or its affiliates.
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
	"fmt"
	"testing"

	"go.fd.io/govpp/api"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/core"
	"go.fd.io/govpp/test/vpptesting"
)

func TestTrace(t *testing.T) {
	test := vpptesting.SetupVPP(t)

	trace := core.NewTrace(test.Conn, 50)

	runTraceRequests(t, test)

	records := trace.GetRecords()

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	printTraceRecords(t, records)
}

func printTraceRecords(t *testing.T, records []*api.Record) {
	t.Logf("API trace (records: %d):\n", len(records))
	t.Logf("--------------------\n")
	for _, item := range records {
		h, m, s := item.Timestamp.Clock()
		reply := ""
		if item.IsReceived {
			reply = "(reply)"
		}
		fmt.Printf("%dh:%dm:%ds:%dns %s %s\n", h, m, s,
			item.Timestamp.Nanosecond(), item.Message.GetMessageName(), reply)
	}
	t.Logf("--------------------\n")
}

func runTraceRequests(t *testing.T, test *vpptesting.TestCtx) {
	vpeRPC := vpe.NewServiceClient(test.Conn)

	_, err := vpeRPC.ShowVersion(context.Background(), &vpe.ShowVersion{})
	if err != nil {
		t.Fatalf("getting version failed: %v", err)
	}
}
