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
	"unsafe"

	"go.fd.io/govpp/adapter"
)

// dirEntryBuf lays out a directory vector holding the given names, for both
// segment versions. The vector is preceded by its length, as VPP lays it out.
func dirEntryBufV1(t *testing.T, names []string) (dirVector, func()) {
	t.Helper()
	sz := int(unsafe.Sizeof(statSegDirectoryEntryV1{}))
	buf := make([]byte, 8+sz*len(names))
	*(*uint32)(unsafe.Pointer(&buf[0])) = uint32(len(names))
	base := unsafe.Pointer(&buf[8])
	for i, n := range names {
		e := (*statSegDirectoryEntryV1)(unsafe.Add(base, i*sz))
		copy(e.name[:], n)
		e.name[len(n)] = 0
		e.directoryType = 2 // SimpleCounterVector
	}
	return dirVector(base), func() { runtimeKeepAlive(buf) }
}

func dirEntryBufV2(t *testing.T, names []string) (dirVector, func()) {
	t.Helper()
	sz := int(unsafe.Sizeof(statSegDirectoryEntryV2{}))
	buf := make([]byte, 8+sz*len(names))
	*(*uint32)(unsafe.Pointer(&buf[0])) = uint32(len(names))
	base := unsafe.Pointer(&buf[8])
	for i, n := range names {
		e := (*statSegDirectoryEntryV2)(unsafe.Add(base, i*sz))
		copy(e.name[:], n)
		e.name[len(n)] = 0
		e.directoryType = 2 // SimpleCounterVector
	}
	return dirVector(base), func() { runtimeKeepAlive(buf) }
}

func runtimeKeepAlive(b []byte) { _ = b }

// TestStatDirOnIndexMatches covers both segment versions. v1 is easy to leave
// untested - the fake segment harness is v2-only - and it is exactly where a
// regression would go unnoticed.
func TestStatDirOnIndexMatches(t *testing.T) {
	names := []string{"/if/rx", "/node/errors", "/sys/heartbeat"}

	for _, tc := range []struct {
		version string
		build   func(*testing.T, []string) (dirVector, func())
		seg     statSegment
	}{
		{"v1", dirEntryBufV1, &statSegmentV1{}},
		{"v2", dirEntryBufV2, &statSegmentV2{sharedHeader: make([]byte, 1<<12)}},
	} {
		t.Run(tc.version, func(t *testing.T) {
			v, keep := tc.build(t, names)
			defer keep()

			for i, n := range names {
				_, typ, ok := tc.seg.StatDirOnIndexMatches(v, uint32(i), []byte(n))
				if !ok {
					t.Errorf("index %d (%q): reported no match", i, n)
				}
				if typ == adapter.Unknown {
					t.Errorf("index %d (%q): matched but type is Unknown", i, n)
				}
			}

			// A name that is a prefix of the stored one must NOT match: the
			// length check is what stops /if/rx matching /if/rx-unicast.
			if _, _, ok := tc.seg.StatDirOnIndexMatches(v, 0, []byte("/if/r")); ok {
				t.Error("a prefix of the stored name matched")
			}
			// Nor an extension of it.
			if _, _, ok := tc.seg.StatDirOnIndexMatches(v, 0, []byte("/if/rx-unicast")); ok {
				t.Error("an extension of the stored name matched")
			}
			// Wrong entry at the right length.
			if _, _, ok := tc.seg.StatDirOnIndexMatches(v, 0, []byte("/if/tx")); ok {
				t.Error("a different name of equal length matched")
			}
			// Empty want against a non-empty entry.
			if _, _, ok := tc.seg.StatDirOnIndexMatches(v, 0, nil); ok {
				t.Error("an empty name matched a populated entry")
			}
			// A mismatch must report Unknown rather than a plausible type.
			if _, typ, _ := tc.seg.StatDirOnIndexMatches(v, 0, []byte("/if/tx")); typ != adapter.Unknown {
				t.Errorf("mismatch reported type %v, want Unknown", typ)
			}
		})
	}
}

// TestStatDirOnIndexMatchesAgreesWithGet pins the two accessors together: the
// in-place comparison must accept exactly the names GetStatDirOnIndex reports.
func TestStatDirOnIndexMatchesAgreesWithGet(t *testing.T) {
	f := newFakeSegment(t, fakeGroupValues(8))
	sc := f.client()
	v := sc.GetDirectoryVector()
	if v == nil {
		t.Fatal("nil directory vector")
	}
	for i := uint32(0); i < 10; i++ {
		_, name, typ := sc.GetStatDirOnIndex(v, i)
		if len(name) == 0 {
			continue
		}
		_, mTyp, ok := sc.StatDirOnIndexMatches(v, i, name)
		if !ok {
			t.Errorf("index %d: GetStatDirOnIndex says %q, Matches disagrees", i, name)
		}
		if mTyp != typ {
			t.Errorf("index %d (%q): type %v vs %v", i, name, mTyp, typ)
		}
	}
}
