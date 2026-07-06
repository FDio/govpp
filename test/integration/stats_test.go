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

	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/adapter/statsclient"
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/test/vpptesting"
)

// statsSocket is the default VPP stats segment socket, matching the path the
// vpptesting harness launches VPP with.
const statsSocket = "/run/vpp/stats.sock"

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

// TestStatClientSymlinkRefresh exercises the lower-level statsclient adapter against
// a live VPP: the symlink-target metadata on StatEntry, the Epoch() accessor, and
// that UpdateDir refreshes a prepared dir (and rejects one prepared under a stale
// epoch). A loopback is created up front so per-interface symlinks (/interfaces/*
// aliasing into /if/*) are present.
func TestStatClientSymlinkRefresh(t *testing.T) {
	test := vpptesting.SetupVPP(t)

	// create an interface so the directory carries per-interface symlink entries
	test.MustCli("create loopback interface", "set interface state loop0 up")

	client := statsclient.NewStatsClient(statsSocket)
	if err := client.Connect(); err != nil {
		t.Fatal("connecting stats client failed:", err)
	}
	defer func() { _ = client.Disconnect() }()

	t.Run("Epoch", func(t *testing.T) {
		epoch, _, err := client.Epoch()
		if err != nil {
			t.Fatal("Epoch failed:", err)
		}
		if epoch == 0 {
			t.Fatal("expected a non-zero stats epoch")
		}
		t.Logf("stats epoch = %d", epoch)
	})

	t.Run("SymlinkTargetsResolve", func(t *testing.T) {
		// A full dump lets us resolve each symlink's target index (a directory index,
		// the same space as StatIdentifier.Index) back to a real backing entry.
		all, err := client.DumpStats()
		if err != nil {
			t.Fatal("DumpStats failed:", err)
		}
		nameByIndex := make(map[uint32]string, len(all))
		symlinkByIndex := make(map[uint32]bool, len(all))
		var symlinks []adapter.StatEntry
		for _, e := range all {
			nameByIndex[e.Index] = string(e.Name)
			symlinkByIndex[e.Index] = e.Symlink
			if e.Symlink {
				symlinks = append(symlinks, e)
			}
		}
		if len(symlinks) == 0 {
			t.Fatal("expected at least one symlink entry in the stats directory")
		}
		for _, e := range symlinks {
			target, ok := nameByIndex[e.SymlinkTarget]
			if !ok {
				t.Fatalf("%s: symlink target index %d not present in directory", e.Name, e.SymlinkTarget)
			}
			if symlinkByIndex[e.SymlinkTarget] {
				t.Fatalf("%s: symlink target %q is itself a symlink", e.Name, target)
			}
		}
		t.Logf("validated %d symlink entries resolve to real backing vectors", len(symlinks))
	})

	t.Run("UpdateDirRefresh", func(t *testing.T) {
		dir, err := client.PrepareDir("/if", "/interfaces")
		if err != nil {
			t.Fatal("PrepareDir failed:", err)
		}
		// Under an unchanged epoch UpdateDir must succeed and re-resolve symlink
		// entries (the change under test) rather than leaving them stale.
		if err := client.UpdateDir(dir); err != nil {
			t.Fatal("UpdateDir failed:", err)
		}
		for i := range dir.Entries {
			if e := &dir.Entries[i]; e.Symlink && e.Data == nil {
				t.Fatalf("%s: symlink entry has nil data after UpdateDir", e.Name)
			}
		}
	})

	t.Run("EpochChangeInvalidatesDir", func(t *testing.T) {
		dir, err := client.PrepareDir("/if", "/interfaces")
		if err != nil {
			t.Fatal("PrepareDir failed:", err)
		}
		before, _, err := client.Epoch()
		if err != nil {
			t.Fatal("Epoch failed:", err)
		}

		// Adding an interface changes the directory layout, which bumps the epoch.
		test.MustCli("create loopback interface")

		after, _, err := client.Epoch()
		if err != nil {
			t.Fatal("Epoch failed:", err)
		}
		if after == before {
			t.Fatalf("expected epoch to change after adding an interface (still %d)", before)
		}
		// A dir prepared under the old epoch is now stale.
		if err := client.UpdateDir(dir); err != adapter.ErrStatsDirStale {
			t.Fatalf("expected ErrStatsDirStale after epoch change, got %v", err)
		}
		// Re-preparing under the new epoch works again.
		if _, err := client.PrepareDir("/if", "/interfaces"); err != nil {
			t.Fatal("re-PrepareDir after epoch change failed:", err)
		}
	})
}
