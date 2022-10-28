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
	"testing"

	"go.fd.io/govpp/api"
	"go.fd.io/govpp/test/vpptesting"
)

func TestStatClientAll(t *testing.T) {
	test := vpptesting.SetupVPP(t)

	c := test.StatsConn()

	var err error
	t.Run("SystemStats", func(t *testing.T) {
		stats := new(api.SystemStats)
		if err = c.GetSystemStats(stats); err != nil {
			t.Fatal("getting stats failed:", err)
		}
		t.Logf("%+v", stats)
	})
	t.Run("NodeStats", func(t *testing.T) {
		stats := new(api.NodeStats)
		if err = c.GetNodeStats(stats); err != nil {
			t.Fatal("getting stats failed:", err)
		}
		t.Logf("%d node stats", len(stats.Nodes))
	})
	t.Run("ErrorStats", func(t *testing.T) {
		stats := new(api.ErrorStats)
		if err = c.GetErrorStats(stats); err != nil {
			t.Fatal("getting stats failed:", err)
		}
		t.Logf("%d error stats", len(stats.Errors))
	})
	t.Run("InterfaceStats", func(t *testing.T) {
		stats := new(api.InterfaceStats)
		if err = c.GetInterfaceStats(stats); err != nil {
			t.Fatal("getting stats failed:", err)
		}
		t.Logf("%d interface stats", len(stats.Interfaces))
	})
	t.Run("MemoryStats", func(t *testing.T) {
		stats := new(api.MemoryStats)
		if err = c.GetMemoryStats(stats); err != nil {
			t.Fatal("getting stats failed:", err)
		}
		t.Logf("%d main, %d stat memory stats", len(stats.Main), len(stats.Stat))
	})
	t.Run("BufferStats", func(t *testing.T) {
		stats := new(api.BufferStats)
		if err = c.GetBufferStats(stats); err != nil {
			t.Fatal("getting stats failed:", err)
		}
		t.Logf("%d buffers stats", len(stats.Buffer))
	})
}

func TestStatClientNodeStats(t *testing.T) {
	test := vpptesting.SetupVPP(t)

	c := test.StatsConn()

	stats := new(api.NodeStats)

	if err := c.GetNodeStats(stats); err != nil {
		t.Fatal("getting node stats failed:", err)
	}
}

func TestStatClientNodeStatsAgain(t *testing.T) {
	test := vpptesting.SetupVPP(t)

	c := test.StatsConn()

	stats := new(api.NodeStats)

	if err := c.GetNodeStats(stats); err != nil {
		t.Fatal("getting node stats failed:", err)
	}
	if err := c.GetNodeStats(stats); err != nil {
		t.Fatal("getting node stats failed:", err)
	}
}
