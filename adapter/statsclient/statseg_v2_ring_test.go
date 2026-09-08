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
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"testing"
	"unsafe"

	"go.fd.io/govpp/adapter"
)

// A synthetic v2 segment holding one ring-buffer stat, laid out as VPP lays one
// out, so the windowed reader's slot arithmetic can be exercised without a
// running VPP.
//
//	index 0: /sys/fake-scalar   filler, so the ring is not at index 0
//	index 1: /fake/records      the ring
//
// Each entry's first eight bytes are the producer sequence that wrote it, which
// is what lets a test assert ordering and unwrapping without reimplementing
// either.
const (
	fakeRingIndex      = 1
	fakeTypeRingBuffer = 8
	fakeRingSchema     = `{"name":"fake.record","version":1,"entry_size":16,"fields":[]}`
)

type fakeRing struct {
	buf        []byte
	ringBase   int // offset of the ring buffer's own header
	dataOff    int // relative to ringBase
	metaOff    int // relative to ringBase
	entrySize  uint32
	ringSize   uint32
	nThreads   uint32
	metaStride int
	head       []uint32
	seq        []uint64
}

func newFakeRing(t testing.TB, entrySize, ringSize, nThreads uint32) *fakeRing {
	t.Helper()
	return newFakeRingStride(t, entrySize, ringSize, nThreads, ringBufferMetaSize)
}

// newFakeRingStride lays the metadata array out with the given stride, which on
// a real segment is VPP's CLIB_CACHE_LINE_BYTES.
func newFakeRingStride(t testing.TB, entrySize, ringSize, nThreads uint32, metaStride int) *fakeRing {
	t.Helper()

	const (
		hdrSize   = 64
		vecHdr    = 8
		trailer   = 8
		nDir      = fakeRingIndex + 1
		dirEntLen = int(unsafe.Sizeof(statSegDirectoryEntryV2{}))
	)

	dirLenOff := hdrSize
	dirOff := dirLenOff + vecHdr
	ringBase := align8(dirOff + nDir*dirEntLen)

	// Offsets below are relative to ringBase, which is what the header declares.
	metaOff := 64 // past the ring header, cache-line aligned as VPP has it
	schemaOff := metaOff + int(nThreads)*metaStride
	dataOff := align8(schemaOff + len(fakeRingSchema) + 1)
	total := ringBase + dataOff + int(nThreads)*int(ringSize)*int(entrySize) + trailer

	f := &fakeRing{
		buf:        make([]byte, total),
		ringBase:   ringBase,
		dataOff:    dataOff,
		metaOff:    metaOff,
		entrySize:  entrySize,
		ringSize:   ringSize,
		nThreads:   nThreads,
		metaStride: metaStride,
		head:       make([]uint32, nThreads),
		seq:        make([]uint64, nThreads),
	}

	f.putU64(fakeOffVersion, 2)
	f.putU64(fakeOffBase, fakeBase)
	f.putU64(fakeOffEpoch, 1)
	f.putU64(fakeOffInProgress, 0)
	f.putU64(fakeOffDirVector, fakeBase+uint64(dirOff))
	f.putU64(fakeOffErrorVector, 0)
	f.putU64(dirLenOff, nDir)

	f.putDirEntry(dirOff, 0, fakeTypeScalarIndex, 7, "/sys/fake-scalar")
	f.putDirEntry(dirOff, fakeRingIndex, fakeTypeRingBuffer, fakeBase+uint64(ringBase), "/fake/records")

	// The ring's own header.
	f.putU32(ringBase+0, entrySize)
	f.putU32(ringBase+4, ringSize)
	f.putU32(ringBase+8, nThreads)
	f.putU32(ringBase+12, uint32(len(fakeRingSchema)+1))
	f.putU32(ringBase+16, 1) // schema version
	f.putU32(ringBase+20, uint32(metaOff))
	f.putU32(ringBase+24, uint32(dataOff))

	copy(f.buf[ringBase+schemaOff:], fakeRingSchema)

	for i := uint32(0); i < nThreads; i++ {
		m := ringBase + metaOff + int(i)*metaStride
		f.putU32(m+0, 0)                  // head
		f.putU32(m+4, 1)                  // schema version
		f.putU64(m+8, 0)                  // sequence
		f.putU32(m+16, uint32(schemaOff)) // schema offset
		f.putU32(m+20, uint32(len(fakeRingSchema)+1))
	}
	return f
}

func align8(n int) int { return (n + 7) &^ 7 }

func (f *fakeRing) putU64(off int, v uint64) { *(*uint64)(unsafe.Pointer(&f.buf[off])) = v }
func (f *fakeRing) putU32(off int, v uint32) { *(*uint32)(unsafe.Pointer(&f.buf[off])) = v }

func (f *fakeRing) putDirEntry(dirOff, index int, typ dirType, union uint64, name string) {
	e := (*statSegDirectoryEntryV2)(unsafe.Pointer(&f.buf[dirOff+index*int(unsafe.Sizeof(statSegDirectoryEntryV2{}))]))
	e.directoryType = typ
	e.unionData = union
	copy(e.name[:], name)
	e.name[len(name)] = 0
}

// produce writes n entries as vlib_stats_ring_produce does: entry body first,
// then head, then the sequence that publishes it. Head and sequence are tracked
// separately here, as VPP tracks them, so a slot derived from the wrong one is
// still a visible difference.
func (f *fakeRing) produce(thread uint32, n int) {
	for i := 0; i < n; i++ {
		slot := f.head[thread]
		off := f.ringBase + f.dataOff +
			int(thread)*int(f.ringSize)*int(f.entrySize) +
			int(slot)*int(f.entrySize)
		binary.LittleEndian.PutUint64(f.buf[off:], f.seq[thread])
		f.head[thread] = (slot + 1) % f.ringSize
		f.seq[thread]++
	}
	m := f.ringBase + f.metaOff + int(thread)*f.metaStride
	f.putU32(m+0, f.head[thread])
	f.putU64(m+8, f.seq[thread])
}

// publishHeadAhead advances the published head by one without publishing the
// sequence that explains it. That is the state a producer is in between its
// plain store of head and its release store of the sequence, and it is what a
// consumer sees when it reads the two astride a commit.
func (f *fakeRing) publishHeadAhead(thread uint32) {
	m := f.ringBase + f.metaOff + int(thread)*f.metaStride
	f.putU32(m+0, (f.head[thread]+1)%f.ringSize)
}

// rewindSequence makes the producer's count go backwards, as a VPP restart on
// the same segment would.
func (f *fakeRing) rewindSequence(thread uint32, to uint64) {
	f.seq[thread] = to
	f.putU64(f.ringBase+f.metaOff+int(thread)*f.metaStride+8, to)
}

func (f *fakeRing) client() *StatsClient {
	sc := &StatsClient{statSegment: newStatSegmentV2(f.buf, int64(len(f.buf)))}
	atomic.StoreUint32(&sc.connected, 1)
	return sc
}

// seqsOf decodes the sequence stamp out of each delivered entry.
func seqsOf(t testing.TB, w adapter.RingBufferWindow, entrySize uint32) []uint64 {
	t.Helper()
	if len(w.Entries) != int(w.Count)*int(entrySize) {
		t.Fatalf("Entries is %d bytes, want Count(%d)*entrySize(%d)=%d",
			len(w.Entries), w.Count, entrySize, int(w.Count)*int(entrySize))
	}
	out := make([]uint64, w.Count)
	for i := range out {
		out[i] = binary.LittleEndian.Uint64(w.Entries[i*int(entrySize):])
	}
	return out
}

func prepareRing(t testing.TB, f *fakeRing, maxEntries uint32, skipBacklog bool) (*StatsClient, *adapter.StatDir, *adapter.RingBufferWindowStat) {
	t.Helper()
	sc := f.client()
	dir, err := sc.PrepareRingBuffer("/fake/records", maxEntries, skipBacklog)
	if err != nil {
		t.Fatalf("PrepareRingBuffer: %v", err)
	}
	s, ok := dir.Entries[0].Data.(*adapter.RingBufferWindowStat)
	if !ok {
		t.Fatalf("prepared Data is %T, want *adapter.RingBufferWindowStat", dir.Entries[0].Data)
	}
	return sc, dir, s
}

func refresh(t testing.TB, sc *StatsClient, dir *adapter.StatDir) {
	t.Helper()
	if err := sc.UpdateDir(dir); err != nil {
		t.Fatalf("UpdateDir: %v", err)
	}
}

func wantWindow(t testing.TB, w adapter.RingBufferWindow, entrySize uint32, seqs []uint64, lost, pending uint64) {
	t.Helper()
	got := seqsOf(t, w, entrySize)
	if len(got) != len(seqs) {
		t.Fatalf("got %d entries %v, want %d %v", len(got), got, len(seqs), seqs)
	}
	for i := range got {
		if got[i] != seqs[i] {
			t.Fatalf("entry %d has sequence %d, want %d (got %v, want %v)", i, got[i], seqs[i], got, seqs)
		}
	}
	if w.Lost != lost {
		t.Errorf("Lost = %d, want %d", w.Lost, lost)
	}
	if w.Pending != pending {
		t.Errorf("Pending = %d, want %d", w.Pending, pending)
	}
}

// PrepareRingBuffer must not read the ring's contents, and the first refresh must
// only position the cursor. A first read that also drained the ring would make
// "what has been produced since I last looked" mean something different on the
// first call than on every later one.
func TestRingBufferWindowFirstRefreshOnlyPositions(t *testing.T) {
	f := newFakeRing(t, 16, 8, 1)
	f.produce(0, 3)

	sc, dir, s := prepareRing(t, f, 0, false)
	refresh(t, sc, dir)

	if s.Config.EntrySize != 16 || s.Config.RingSize != 8 || s.Config.NThreads != 1 {
		t.Fatalf("config = %+v, want entry 16 ring 8 threads 1", s.Config)
	}
	if string(s.Schema) != fakeRingSchema+"\x00" {
		t.Errorf("schema = %q, want %q", s.Schema, fakeRingSchema)
	}
	wantWindow(t, s.Windows[0], 16, nil, 0, 0)
	if s.Windows[0].NextSeq != 0 {
		t.Errorf("NextSeq = %d, want 0: the backlog must still be readable", s.Windows[0].NextSeq)
	}

	// The backlog is then delivered by the next read, in order.
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{0, 1, 2}, 0, 0)
}

func TestRingBufferWindowSkipBacklog(t *testing.T) {
	f := newFakeRing(t, 16, 8, 1)
	f.produce(0, 3)

	sc, dir, s := prepareRing(t, f, 0, true)
	refresh(t, sc, dir)
	if s.Windows[0].NextSeq != 3 {
		t.Fatalf("NextSeq = %d, want 3: SkipBacklog must start at the head", s.Windows[0].NextSeq)
	}

	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, nil, 0, 0)

	f.produce(0, 2)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{3, 4}, 0, 0)
}

// Only what was appended since the last refresh, and nothing twice.
func TestRingBufferWindowDeliversNewEntriesOnly(t *testing.T) {
	f := newFakeRing(t, 16, 8, 1)
	sc, dir, s := prepareRing(t, f, 0, false)
	refresh(t, sc, dir)

	f.produce(0, 3)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{0, 1, 2}, 0, 0)

	// A refresh with nothing new must report nothing, not repeat itself.
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, nil, 0, 0)

	f.produce(0, 2)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{3, 4}, 0, 0)
}

// A window that straddles the ring's seam must come back contiguous and in
// order: the consumer is not told where the seam fell and must not have to care.
func TestRingBufferWindowUnwrapsAcrossSeam(t *testing.T) {
	f := newFakeRing(t, 16, 4, 1)
	sc, dir, s := prepareRing(t, f, 0, false)
	refresh(t, sc, dir)

	f.produce(0, 3) // slots 0,1,2
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{0, 1, 2}, 0, 0)

	f.produce(0, 3) // slots 3,0,1 - wraps
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{3, 4, 5}, 0, 0)
}

// Overwritten entries are Lost - gone - and must be counted exactly, not
// estimated and not silently skipped.
func TestRingBufferWindowReportsLostOnLap(t *testing.T) {
	f := newFakeRing(t, 16, 4, 1)
	sc, dir, s := prepareRing(t, f, 0, false)
	refresh(t, sc, dir)

	f.produce(0, 10) // ring holds 4: sequences 6..9 survive, 0..5 are gone
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{6, 7, 8, 9}, 6, 0)

	// And the reader is re-synced, so the next read is clean.
	f.produce(0, 2)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{10, 11}, 0, 0)
}

// MaxEntries bounds one refresh. What it holds back is Pending, not Lost: it is
// still in the ring and the next read delivers it. Conflating the two would tell
// a rate-limited reader it was losing data.
func TestRingBufferWindowCapsAtMaxEntriesAndReportsPending(t *testing.T) {
	f := newFakeRing(t, 16, 16, 1)
	sc, dir, s := prepareRing(t, f, 2, false)
	refresh(t, sc, dir)

	f.produce(0, 5)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{0, 1}, 0, 3)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{2, 3}, 0, 1)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{4}, 0, 0)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, nil, 0, 0)

	// The buffer is sized from MaxEntries, not from the ring: that is what makes
	// a ring sized for burst headroom affordable to read.
	if got := cap(s.Windows[0].Entries); got != 2*16 {
		t.Errorf("buffer capacity = %d, want %d (MaxEntries*EntrySize)", got, 2*16)
	}
}

// A producer sequence that goes backwards means a restart, not loss: nothing was
// overwritten, the count simply is no longer comparable.
func TestRingBufferWindowResyncsOnSequenceRewind(t *testing.T) {
	f := newFakeRing(t, 16, 8, 1)
	sc, dir, s := prepareRing(t, f, 0, true)
	refresh(t, sc, dir)

	f.produce(0, 5)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{0, 1, 2, 3, 4}, 0, 0)

	f.rewindSequence(0, 2)
	refresh(t, sc, dir)
	if s.Windows[0].Lost != 0 {
		t.Errorf("Lost = %d after a sequence rewind, want 0: nothing was overwritten", s.Windows[0].Lost)
	}
	if s.Windows[0].NextSeq != 2 {
		t.Errorf("NextSeq = %d, want 2: the reader must re-sync to what is there now", s.Windows[0].NextSeq)
	}
}

// VPP pads its metadata array to CLIB_CACHE_LINE_BYTES, 128 on aarch64 builds,
// and the ring header does not say which. Thread 0 sits at the start of the
// array either way, so only threads after it can catch a wrong stride: at 128
// with a hard-coded 64, thread 1's metadata is read out of thread 0's padding
// and the thread looks idle.
func TestRingBufferWindowDerivesMetaStride(t *testing.T) {
	for _, stride := range []int{ringBufferMetaSize, ringBufferMetaSizeAlt} {
		t.Run(fmt.Sprint(stride), func(t *testing.T) {
			f := newFakeRingStride(t, 16, 8, 2, stride)
			f.produce(0, 3)
			f.produce(1, 2)

			sc, dir, s := prepareRing(t, f, 0, false)
			refresh(t, sc, dir)

			if got := s.Threads[1].Sequence; got != 2 {
				t.Fatalf("thread 1 sequence = %d, want 2", got)
			}

			f.produce(0, 2)
			f.produce(1, 1)
			refresh(t, sc, dir)

			wantWindow(t, s.Windows[0], 16, []uint64{0, 1, 2, 3, 4}, 0, 0)
			wantWindow(t, s.Windows[1], 16, []uint64{0, 1, 2}, 0, 0)
		})
	}
}

func TestRingBufferWindowPerThread(t *testing.T) {
	f := newFakeRing(t, 16, 8, 3)
	sc, dir, s := prepareRing(t, f, 0, false)
	refresh(t, sc, dir)

	if len(s.Windows) != 3 {
		t.Fatalf("got %d windows, want 3", len(s.Windows))
	}

	f.produce(0, 2)
	f.produce(2, 3)
	refresh(t, sc, dir)

	wantWindow(t, s.Windows[0], 16, []uint64{0, 1}, 0, 0)
	wantWindow(t, s.Windows[1], 16, nil, 0, 0)
	wantWindow(t, s.Windows[2], 16, []uint64{0, 1, 2}, 0, 0)
}

// The steady path must not allocate. A refresh that allocated per poll would put
// the ring read back on the garbage collector's critical path, which is most of
// what this type exists to get off it.
func TestRingBufferWindowRefreshDoesNotAllocate(t *testing.T) {
	f := newFakeRing(t, 128, 4096, 2)
	sc, dir, _ := prepareRing(t, f, 512, true)
	refresh(t, sc, dir)
	f.produce(0, 64)
	refresh(t, sc, dir)

	allocs := testing.AllocsPerRun(50, func() {
		f.produce(0, 64)
		f.produce(1, 64)
		if err := sc.UpdateDir(dir); err != nil {
			t.Fatalf("UpdateDir: %v", err)
		}
	})
	if allocs != 0 {
		t.Errorf("refresh allocated %v objects per run, want 0", allocs)
	}
}

func TestPrepareRingBufferRejectsNonRing(t *testing.T) {
	f := newFakeRing(t, 16, 8, 1)
	sc := f.client()
	if _, err := sc.PrepareRingBuffer("/sys/fake-scalar", 0, false); err == nil {
		t.Fatal("PrepareRingBuffer: no error for a stat that is not a ring buffer")
	}
	if _, err := sc.PrepareRingBuffer("/fake/nope", 0, false); err == nil {
		t.Fatal("PrepareRingBuffer: no error for a missing stat")
	}
}

// The measurement the design rests on: a refresh costs the entries produced,
// where a full-ring read costs the ring. Run with -benchtime=1000x and compare.
func BenchmarkRingBufferWindowRefresh(b *testing.B) {
	f := newFakeRing(b, 128, 8192, 1)
	sc, dir, _ := prepareRing(b, f, 512, true)
	refresh(b, sc, dir)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.produce(0, 256)
		if err := sc.UpdateDir(dir); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.ReportMetric(256, "entries/op")
}

func BenchmarkRingBufferFullCopy(b *testing.B) {
	f := newFakeRing(b, 128, 8192, 1)
	sc := f.client()
	dir, err := sc.PrepareDir("/fake/records")
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.produce(0, 256)
		if err := sc.UpdateDir(dir); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.ReportMetric(256, "entries/op")
}

// A head that has run ahead of its sequence must not move the window. Slots are
// found from the sequence, which is the only field the producer publishes with a
// barrier; positioning off head would deliver the slot being written and skip the
// oldest live entry, and would do it only under load.
func TestRingBufferWindowIgnoresHeadAheadOfSequence(t *testing.T) {
	f := newFakeRing(t, 16, 8, 1)
	sc, dir, s := prepareRing(t, f, 0, false)
	refresh(t, sc, dir)

	f.produce(0, 3)
	f.publishHeadAhead(0)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{0, 1, 2}, 0, 0)

	// And the cursor is still where the sequence says, so the next refresh
	// continues rather than repeating or skipping.
	f.produce(0, 2)
	f.publishHeadAhead(0)
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{3, 4}, 0, 0)
}

// The segment's optimistic lock covers directory changes, not ring data, so a
// worker can overwrite the slots a refresh is copying. What survives must be
// whole, and what did not must be reported as lost: a torn record delivered as
// good is worse than a lost one, because loss is visible and tearing is not.
func TestRingBufferWindowDropsEntriesOverwrittenDuringCopy(t *testing.T) {
	f := newFakeRing(t, 16, 8, 1)
	sc, dir, s := prepareRing(t, f, 0, false)
	refresh(t, sc, dir)

	f.produce(0, 8) // fills the ring: sequences 0..7

	// Three more entries land while the copy is in flight, overwriting the three
	// oldest slots it just read.
	ringWindowCopyHook = func() {
		ringWindowCopyHook = nil
		f.produce(0, 3)
	}
	t.Cleanup(func() { ringWindowCopyHook = nil })

	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{3, 4, 5, 6, 7}, 3, 0)
	if s.Windows[0].FirstSeq != 3 {
		t.Errorf("FirstSeq = %d, want 3: it must name the first entry actually delivered",
			s.Windows[0].FirstSeq)
	}

	// The cursor is past the whole window, overwritten entries included, so the
	// next refresh picks up the three that arrived during the copy.
	refresh(t, sc, dir)
	wantWindow(t, s.Windows[0], 16, []uint64{8, 9, 10}, 0, 0)
}
