//  Copyright (c) 2019 Cisco and/or its affiliates.
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

//go:build integration
// +build integration

package integration

import (
	"flag"
	"testing"

	"go.fd.io/govpp/adapter/statsclient"
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/core"
)

var (
	statsSocket = flag.String("socket", statsclient.DefaultSocketName, "Path to VPP stats socket")
)

func TestStatClientAll(t *testing.T) {
	client := statsclient.NewStatsClient(*statsSocket)

	c, err := core.ConnectStats(client)
	if err != nil {
		t.Fatal("Connecting failed:", err)
	}
	defer c.Disconnect()

	sysStats := new(api.SystemStats)
	nodeStats := new(api.NodeStats)
	errorStats := new(api.ErrorStats)
	ifaceStats := new(api.InterfaceStats)

	if err = c.GetNodeStats(nodeStats); err != nil {
		t.Fatal("updating node stats failed:", err)
	}
	if err = c.GetSystemStats(sysStats); err != nil {
		t.Fatal("updating system stats failed:", err)
	}
	if err = c.GetErrorStats(errorStats); err != nil {
		t.Fatal("updating error stats failed:", err)
	}
	if err = c.GetInterfaceStats(ifaceStats); err != nil {
		t.Fatal("updating interface stats failed:", err)
	}
}

func TestStatClientNodeStats(t *testing.T) {
	client := statsclient.NewStatsClient(*statsSocket)

	c, err := core.ConnectStats(client)
	if err != nil {
		t.Fatal("Connecting failed:", err)
	}
	defer c.Disconnect()

	stats := new(api.NodeStats)

	if err := c.GetNodeStats(stats); err != nil {
		t.Fatal("getting node stats failed:", err)
	}
}

func TestStatClientNodeStatsAgain(t *testing.T) {
	client := statsclient.NewStatsClient(*statsSocket)
	c, err := core.ConnectStats(client)
	if err != nil {
		t.Fatal("Connecting failed:", err)
	}
	defer c.Disconnect()

	stats := new(api.NodeStats)

	if err := c.GetNodeStats(stats); err != nil {
		t.Fatal("getting node stats failed:", err)
	}
	if err := c.GetNodeStats(stats); err != nil {
		t.Fatal("getting node stats failed:", err)
	}
}

func BenchmarkStatClientNodeStatsGet1(b *testing.B)  { benchStatClientNodeStatsGet(b, 1) }
func BenchmarkStatClientNodeStatsGet10(b *testing.B) { benchStatClientNodeStatsGet(b, 10) }

func benchStatClientNodeStatsGet(b *testing.B, repeatN int) {
	client := statsclient.NewStatsClient(*statsSocket)
	c, err := core.ConnectStats(client)
	if err != nil {
		b.Fatal("Connecting failed:", err)
	}
	defer c.Disconnect()

	b.ResetTimer()
	nodeStats := new(api.NodeStats)
	for i := 0; i < b.N; i++ {
		for r := 0; r < repeatN; r++ {
			if err = c.GetNodeStats(nodeStats); err != nil {
				b.Fatal("getting node stats failed:", err)
			}
		}
	}
	b.StopTimer()
}

func BenchmarkStatClientNodeStatsUpdate1(b *testing.B)  { benchStatClientNodeStatsLoad(b, 1) }
func BenchmarkStatClientNodeStatsUpdate10(b *testing.B) { benchStatClientNodeStatsLoad(b, 10) }

func benchStatClientNodeStatsLoad(b *testing.B, repeatN int) {
	client := statsclient.NewStatsClient(*statsSocket)
	c, err := core.ConnectStats(client)
	if err != nil {
		b.Fatal("Connecting failed:", err)
	}
	defer c.Disconnect()
	nodeStats := new(api.NodeStats)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for r := 0; r < repeatN; r++ {
			if err = c.GetNodeStats(nodeStats); err != nil {
				b.Fatal("getting node stats failed:", err)
			}
		}
	}
	b.StopTimer()
}

func BenchmarkStatClientStatsUpdate1(b *testing.B)   { benchStatClientStatsUpdate(b, 1) }
func BenchmarkStatClientStatsUpdate10(b *testing.B)  { benchStatClientStatsUpdate(b, 10) }
func BenchmarkStatClientStatsUpdate100(b *testing.B) { benchStatClientStatsUpdate(b, 100) }

func benchStatClientStatsUpdate(b *testing.B, repeatN int) {
	client := statsclient.NewStatsClient(*statsSocket)
	c, err := core.ConnectStats(client)
	if err != nil {
		b.Fatal("Connecting failed:", err)
	}
	defer c.Disconnect()

	sysStats := new(api.SystemStats)
	nodeStats := new(api.NodeStats)
	errorStats := new(api.ErrorStats)
	ifaceStats := new(api.InterfaceStats)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for r := 0; r < repeatN; r++ {
			if err = c.GetNodeStats(nodeStats); err != nil {
				b.Fatal("updating node stats failed:", err)
			}
			if err = c.GetSystemStats(sysStats); err != nil {
				b.Fatal("updating system stats failed:", err)
			}
			if err = c.GetErrorStats(errorStats); err != nil {
				b.Fatal("updating error stats failed:", err)
			}
			if err = c.GetInterfaceStats(ifaceStats); err != nil {
				b.Fatal("updating error stats failed:", err)
			}
		}
	}
	b.StopTimer()
}
