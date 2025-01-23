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

package performance

import (
	"testing"

	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/adapter/statsclient"
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/core"
	"go.fd.io/govpp/test/vpptesting"
)

func newStatsClient() adapter.StatsAPI {
	return statsclient.NewStatsClient("")
}

func BenchmarkStatClientNodeStatsGet(b *testing.B) {
	vpptesting.SetupVPP(b)

	b.Run("1", func(b *testing.B) {
		benchStatClientNodeStatsGet(b, 1)
	})
	b.Run("10", func(b *testing.B) {
		benchStatClientNodeStatsGet(b, 10)
	})
}

func benchStatClientNodeStatsGet(b *testing.B, repeatN int) {
	client := newStatsClient()
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

func BenchmarkStatClientNodeStatsUpdate(b *testing.B) {
	vpptesting.SetupVPP(b)

	b.Run("1", func(b *testing.B) {
		benchStatClientNodeStatsLoad(b, 1)
	})
	b.Run("10", func(b *testing.B) {
		benchStatClientNodeStatsLoad(b, 10)
	})
}

func benchStatClientNodeStatsLoad(b *testing.B, repeatN int) {
	client := newStatsClient()
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

func BenchmarkStatClientStatsUpdate(b *testing.B) {
	vpptesting.SetupVPP(b)

	b.Run("1", func(b *testing.B) {
		benchStatClientStatsUpdate(b, 1)
	})
	b.Run("10", func(b *testing.B) {
		benchStatClientStatsUpdate(b, 10)
	})
	b.Run("100", func(b *testing.B) {
		benchStatClientStatsUpdate(b, 100)
	})
}

func benchStatClientStatsUpdate(b *testing.B, repeatN int) {
	client := newStatsClient()
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
