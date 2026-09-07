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

package statsclient

import (
	"sync/atomic"
	"testing"

	"go.fd.io/govpp/adapter"
)

// countingSegment counts data reads so a test can assert that a symlink fan
// costs ONE read of its target rather than one per symlink.
type countingSegment struct {
	statSegment
	copies int64
}

func (c *countingSegment) CopyEntryData(segment dirSegment, index uint32) adapter.Stat {
	atomic.AddInt64(&c.copies, 1)
	return c.statSegment.CopyEntryData(segment, index)
}

func countingClient(f *fakeSegment) (*StatsClient, *countingSegment) {
	sc := f.client()
	c := &countingSegment{statSegment: sc.statSegment}
	sc.statSegment = c
	return sc, c
}

// TestUpdateDirReadsSymlinkTargetOnce is the property the whole grouping exists
// for. VPP names every item of /node/errors with its own /err/<node>/<reason>
// symlink, so resolving them one at a time re-reads the same backing vector once
// per item - thousands of times per refresh on a real box.
func TestUpdateDirReadsSymlinkTargetOnce(t *testing.T) {
	const n = 64
	f := newFakeSegment(t, fakeGroupValues(n))
	sc, counter := countingClient(f)

	dir, err := sc.PrepareDir("^/err/")
	if err != nil {
		t.Fatal("PrepareDir failed:", err)
	}
	if len(dir.Entries) != n {
		t.Fatalf("prepared %d entries, want %d symlinks", len(dir.Entries), n)
	}

	atomic.StoreInt64(&counter.copies, 0)
	if err := sc.UpdateDir(dir); err != nil {
		t.Fatal("UpdateDir failed:", err)
	}

	// All n symlinks alias the same vector, so one read serves them all.
	if got := atomic.LoadInt64(&counter.copies); got != 1 {
		t.Errorf("refreshing %d symlinks over one target did %d data reads, want 1", n, got)
	}
}

// TestUpdateDirGroupedMatchesIndividual pins the fan-out against the
// one-at-a-time resolution it replaces: same values, same shape. An off-by-one
// here would mislabel every error counter while looking entirely plausible.
func TestUpdateDirGroupedMatchesIndividual(t *testing.T) {
	values := fakeGroupValues(24)
	f := newFakeSegment(t, values)
	sc := f.client()

	dir, err := sc.PrepareDir("^/err/")
	if err != nil {
		t.Fatal("PrepareDir failed:", err)
	}
	// Move the counters so the refresh has to do real work.
	for i := range values {
		values[i] += 7777
		f.setCounter(i, values[i])
	}
	if err := sc.UpdateDir(dir); err != nil {
		t.Fatal("UpdateDir failed:", err)
	}

	// Reference: resolve each symlink on its own, as DumpStats does.
	individual, err := sc.DumpStats("^/err/")
	if err != nil {
		t.Fatal("DumpStats failed:", err)
	}
	want := make(map[string]uint64, len(individual))
	for _, e := range individual {
		want[string(e.Name)] = symlinkValue(t, e)
	}

	for _, e := range dir.Entries {
		got := symlinkValue(t, e)
		if w, ok := want[string(e.Name)]; !ok {
			t.Errorf("%s: not returned by DumpStats", e.Name)
		} else if got != w {
			t.Errorf("%s: grouped refresh read %d, individual resolution %d", e.Name, got, w)
		}
		if item := fakeErrItem(t, e.Name); got != values[item] {
			t.Errorf("%s: read %d, backing item %d holds %d", e.Name, got, item, values[item])
		}
	}
}

// TestUpdateDirSymlinkRefreshDoesNotAllocate — the fan-out writes into the
// slices the prepared entries already hold.
func TestUpdateDirSymlinkRefreshDoesNotAllocate(t *testing.T) {
	f := newFakeSegment(t, fakeGroupValues(128))
	sc := f.client()

	dir, err := sc.PrepareDir("^/err/")
	if err != nil {
		t.Fatal("PrepareDir failed:", err)
	}
	if err := sc.UpdateDir(dir); err != nil { // warm up
		t.Fatal(err)
	}
	allocs := testing.AllocsPerRun(20, func() {
		if err := sc.UpdateDir(dir); err != nil {
			t.Fatal(err)
		}
	})
	// What remains is one name clone per prepared entry, from GetStatDirOnIndex
	// in updateStatOnIndex, plus one read of the target vector. The fan-out
	// itself must add nothing per symlink: it writes into the slices the
	// prepared entries already hold.
	const n = 128
	if allocs > n+40 {
		t.Errorf("refreshing %d symlinks allocated %.0f times, want at most one per entry "+
			"(the name clone) plus the target read; the fan-out is not reusing the prepared slices",
			n, allocs)
	}
	t.Logf("allocations per refresh of %d symlinks: %.0f", n, allocs)
}

func fakeGroupValues(n int) []uint64 {
	v := make([]uint64, n)
	for i := range v {
		v[i] = uint64(i)*13 + 500
	}
	return v
}

// BenchmarkUpdateDirSymlinkFan measures the refresh at the fan size a real box
// carries: a capture from a live device has 4223 /err/* symlinks over a
// 4223-item /node/errors.
func BenchmarkUpdateDirSymlinkFan(b *testing.B) {
	f := newFakeSegment(b, fakeGroupValues(4223))
	sc := f.client()
	dir, err := sc.PrepareDir("^/err/")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := sc.UpdateDir(dir); err != nil {
			b.Fatal(err)
		}
	}
}
