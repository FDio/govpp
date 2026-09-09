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
	"testing"
)

// A symlink's target index is read straight out of the segment, so it is only as
// trustworthy as the process writing it. Resolving one addresses the directory
// vector at 144 bytes an entry, so a large out-of-range index lands far outside
// the mapping and faults.
//
// A test cannot assert on a fault, and an index chosen to be wildly out of range
// proves nothing either: it lands in unrelated mapped memory, reads garbage as
// the directory type and falls out of the type switch as nil, which is what the
// fix produces anyway. So these point a symlink at fakeSegment.spareDir - a valid
// entry that the vector's declared length excludes. Resolving it succeeds if the
// length is not honoured and is refused if it is, which is exactly the
// distinction under test.

// corruptSymlinkTarget repoints the symlink naming item at a directory index the
// vector does not have, leaving every other entry alone.
func corruptSymlinkTarget(t *testing.T, f *fakeSegment, item uint32, target uint32) {
	t.Helper()
	want := fakeErrName(item)
	for i := 0; i < f.nDir; i++ {
		e := f.dirEntry(i)
		n := 0
		for ; n < len(e.name) && e.name[n] != 0; n++ {
		}
		if string(e.name[:n]) != want {
			continue
		}
		e.unionData = uint64(target) | uint64(item)<<32
		return
	}
	t.Fatalf("no symlink entry named %s", want)
}

func TestCopyEntryDataRejectsOutOfRangeSymlinkTarget(t *testing.T) {
	values := []uint64{10, 20, 30, 40}
	f := newFakeSegment(t, values)
	corruptSymlinkTarget(t, f, 2, uint32(f.spareDir))

	entries, err := f.client().DumpStats()
	if err != nil {
		t.Fatal("DumpStats failed:", err)
	}

	var checked int
	for _, e := range entries {
		if !e.Symlink {
			continue
		}
		item := fakeErrItem(t, e.Name)
		checked++
		if item == 2 {
			// Unresolvable, so it must carry no data rather than something
			// read from outside the segment.
			if e.Data != nil {
				t.Errorf("%s: target is out of range, want nil Data, got %v", e.Name, e.Data)
			}
			continue
		}
		// One bad entry must not disturb its neighbours.
		if got, want := symlinkValue(t, e), values[item]; got != want {
			t.Errorf("%s: value = %d, want %d", e.Name, got, want)
		}
	}
	if checked != len(values) {
		t.Fatalf("checked %d symlink entries, want %d", checked, len(values))
	}
}

// The grouped refresh path has its own guard, in updateSymlinkGroups. It skips the
// group rather than falling back to resolving each symlink on its own, because that
// fallback would hand CopyEntryData the very index just found to be out of range.
func TestUpdateDirSkipsOutOfRangeSymlinkTarget(t *testing.T) {
	values := []uint64{10, 20, 30, 40}
	f := newFakeSegment(t, values)
	sc := f.client()

	dir, err := sc.PrepareDir("/err/")
	if err != nil {
		t.Fatal("PrepareDir failed:", err)
	}
	if len(dir.Entries) != len(values) {
		t.Fatalf("prepared %d entries, want %d", len(dir.Entries), len(values))
	}

	// Corrupt after preparing, so the entries hold a resolved value that the
	// refresh must then leave alone rather than replace with the spare slot's.
	corruptSymlinkTarget(t, f, 2, uint32(f.spareDir))
	for i := range values {
		f.setCounter(i, values[i]*100)
	}

	if err := sc.UpdateDir(dir); err != nil {
		t.Fatal("UpdateDir failed:", err)
	}

	for i := range dir.Entries {
		e := &dir.Entries[i]
		item := fakeErrItem(t, e.Name)
		got := symlinkValue(t, *e)
		if item == 2 {
			// The whole group is skipped, so this keeps its PrepareDir value.
			if got != values[item] {
				t.Errorf("%s: value = %d, want the unrefreshed %d", e.Name, got, values[item])
			}
			continue
		}
		if want := values[item] * 100; got != want {
			t.Errorf("%s: value = %d, want the refreshed %d", e.Name, got, want)
		}
	}
}

// A symlink whose target names a freed directory slot - VPP zeroes an entry's name
// when it removes it - is skipped for the same reason, and must not resolve to
// whatever the slot still points at.
func TestUpdateDirSkipsSymlinkToFreedSlot(t *testing.T) {
	values := []uint64{10, 20, 30, 40}
	f := newFakeSegment(t, values)
	sc := f.client()

	dir, err := sc.PrepareDir("/err/")
	if err != nil {
		t.Fatal("PrepareDir failed:", err)
	}

	// Free the backing vector's slot the way VPP does, keeping its data pointer.
	target := f.dirEntry(fakeTargetIndex)
	for i := range target.name {
		target.name[i] = 0
	}
	for i := range values {
		f.setCounter(i, values[i]*100)
	}

	if err := sc.UpdateDir(dir); err != nil {
		t.Fatal("UpdateDir failed:", err)
	}

	// Every group aliased that one entry, so none of them refresh.
	for i := range dir.Entries {
		e := &dir.Entries[i]
		item := fakeErrItem(t, e.Name)
		if got := symlinkValue(t, *e); got != values[item] {
			t.Errorf("%s: value = %d, want the unrefreshed %d", e.Name, got, values[item])
		}
	}
}
