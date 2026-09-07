//  Copyright (c) 2022 Cisco and/or its affiliates.
//  Copyright (c) 2026 Meter, Inc.
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

	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/adapter/statsclient"
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

// TestStatClientSymlinkRefresh checks that refreshing a prepared dir over a
// symlink fan agrees with resolving each symlink individually, against a real
// VPP directory.
//
// UpdateDir groups symlinks by the entry they alias and reads each target once,
// rather than resolving every symlink separately. /err/<node>/<reason> is the
// case that motivates it: every one is a symlink into a single /node/errors
// vector, and an item's position there comes from a heap allocation in
// vlib_register_errors, so the grouping has to recover it from the directory
// rather than compute it.
//
// The unit tests in adapter/statsclient cover the pointer walking and the
// fan-out deterministically against a synthetic segment; this test is about
// agreeing with a real VPP's layout.
func TestStatClientSymlinkRefresh(t *testing.T) {
	test := vpptesting.SetupVPP(t)

	// Create an interface so the directory carries per-interface symlink entries
	// (/interfaces/* aliasing into /if/*) alongside the /err/* ones.
	test.MustCli("create loopback interface", "set interface state loop0 up")

	client := statsclient.NewStatsClient("")
	if err := client.Connect(); err != nil {
		t.Fatal("connecting stats client failed:", err)
	}
	defer func() { _ = client.Disconnect() }()

	const pattern = "^/err/"

	dir, err := client.PrepareDir(pattern)
	if err != nil {
		t.Fatal("PrepareDir failed:", err)
	}
	if len(dir.Entries) == 0 {
		t.Skip("no /err/ symlinks in this VPP's stats directory")
	}
	if err := client.UpdateDir(dir); err != nil {
		t.Fatal("UpdateDir failed:", err)
	}

	// Reference: resolve each symlink on its own, which is what DumpStats does.
	individual, err := client.DumpStats(pattern)
	if err != nil {
		t.Fatal("DumpStats failed:", err)
	}
	want := make(map[string]adapter.Stat, len(individual))
	for _, e := range individual {
		want[string(e.Name)] = e.Data
	}

	var checked int
	for _, e := range dir.Entries {
		w, ok := want[string(e.Name)]
		if !ok {
			t.Errorf("%s: refreshed by UpdateDir but not returned by DumpStats", e.Name)
			continue
		}
		got, gotOK := e.Data.(adapter.SimpleCounterStat)
		exp, expOK := w.(adapter.SimpleCounterStat)
		if !gotOK || !expOK {
			// Older segments report error counters as ErrorStat; the shapes are
			// compared only where both sides are counter vectors.
			continue
		}
		if len(got) != len(exp) {
			t.Errorf("%s: grouped refresh has %d worker rows, individual resolution %d",
				e.Name, len(got), len(exp))
			continue
		}
		for i := range exp {
			if len(got[i]) != len(exp[i]) {
				t.Errorf("%s: worker %d has %d items, want %d", e.Name, i, len(got[i]), len(exp[i]))
				break
			}
			for j := range exp[i] {
				if got[i][j] != exp[i][j] {
					t.Errorf("%s: worker %d item %d is %d, individual resolution gives %d",
						e.Name, i, j, got[i][j], exp[i][j])
				}
			}
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no symlink entries were comparable between the two paths")
	}
	t.Logf("%d /err/ symlinks agree between grouped refresh and individual resolution", checked)
}
