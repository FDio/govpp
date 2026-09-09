//  Copyright (c) 2020 Cisco and/or its affiliates.
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
	"bytes"
	"sync/atomic"
	"unsafe"

	"go.fd.io/govpp/adapter"
)

type statSegmentV2 struct {
	sharedHeader []byte
	memorySize   int64
}

type sharedHeaderV2 struct {
	version     uint64
	base        unsafe.Pointer
	epoch       int64
	inProgress  int64
	dirVector   unsafe.Pointer
	errorVector unsafe.Pointer
}

type statSegDirectoryEntryV2 struct {
	directoryType dirType
	// unionData can represent:
	// - symlink indexes
	// - index
	// - value
	// - pointer to data
	unionData uint64
	name      [128]byte
}

func newStatSegmentV2(data []byte, size int64) *statSegmentV2 {
	return &statSegmentV2{
		sharedHeader: data,
		memorySize:   size,
	}
}

func (ss *statSegmentV2) loadSharedHeader(b []byte) (header sharedHeaderV2) {
	h := (*sharedHeaderV2)(unsafe.Pointer(&b[0]))
	return sharedHeaderV2{
		version:     atomic.LoadUint64(&h.version),
		base:        atomic.LoadPointer(&h.base),
		epoch:       atomic.LoadInt64(&h.epoch),
		inProgress:  atomic.LoadInt64(&h.inProgress),
		dirVector:   atomic.LoadPointer(&h.dirVector),
		errorVector: atomic.LoadPointer(&h.errorVector),
	}
}

func (ss *statSegmentV2) GetDirectoryVector() dirVector {
	header := ss.loadSharedHeader(ss.sharedHeader)
	return ss.adjust(dirVector(&header.dirVector))
}

func (ss *statSegmentV2) GetStatDirOnIndex(v dirVector, index uint32) (dirSegment, dirName, adapter.StatType) {
	statSegDir := dirSegment(uintptr(v) + uintptr(index)*unsafe.Sizeof(statSegDirectoryEntryV2{}))
	dir := (*statSegDirectoryEntryV2)(statSegDir)
	var name []byte
	for n := 0; n < len(dir.name); n++ {
		if dir.name[n] == 0 {
			// Copy the name off the statseg shared memory region before returning.
			// The caller must not hold a reference into the mmap'd region after
			// DumpStats returns, because VPP may unmap the segment at any time.
			name = bytes.Clone(dir.name[:n])
			break
		}
	}
	return statSegDir, name, getStatType(dir.directoryType, ss.getErrorVector() != nil)
}

// StatDirOnIndexMatches compares the entry name in place - see the interface.
func (ss *statSegmentV2) StatDirOnIndexMatches(v dirVector, index uint32, want []byte) (dirSegment, adapter.StatType, bool) {
	statSegDir := dirSegment(uintptr(v) + uintptr(index)*unsafe.Sizeof(statSegDirectoryEntryV2{}))
	dir := (*statSegDirectoryEntryV2)(statSegDir)
	n := 0
	for ; n < len(dir.name); n++ {
		if dir.name[n] == 0 {
			break
		}
	}
	if n == 0 || n != len(want) {
		return statSegDir, adapter.Unknown, false
	}
	for i := 0; i < n; i++ {
		if dir.name[i] != want[i] {
			return statSegDir, adapter.Unknown, false
		}
	}
	return statSegDir, getStatType(dir.directoryType, ss.getErrorVector() != nil), true
}

func (ss *statSegmentV2) GetEpoch() (int64, bool) {
	sh := ss.loadSharedHeader(ss.sharedHeader)
	return sh.epoch, sh.inProgress != 0
}

func (ss *statSegmentV2) CopyEntryData(segment dirSegment, index uint32) adapter.Stat {
	dirEntry := (*statSegDirectoryEntryV2)(segment)
	typ := getStatType(dirEntry.directoryType, ss.getErrorVector() != nil)
	// skip zero pointer value
	if typ != adapter.ScalarIndex && typ != adapter.GaugeIndex && typ != adapter.Empty && typ != adapter.ErrorIndex && dirEntry.unionData == 0 {
		debugf("data pointer not defined for %s", dirEntry.name)
		return nil
	}

	switch typ {
	case adapter.ScalarIndex:
		return adapter.ScalarStat(dirEntry.unionData)

	case adapter.ErrorIndex:
		dirVector := ss.getErrorVector()
		if dirVector == nil {
			debugf("error vector pointer is out of range for %s", dirEntry.name)
			return nil
		}
		vecLen := *(*uint32)(vectorLen(dirVector))
		var errData []adapter.Counter
		for i := uint32(0); i < vecLen; i++ {
			cb := statSegPointer(dirVector, uintptr(i+1)*unsafe.Sizeof(uint64(0)))
			cbVal := ss.adjust(vectorLen(cb))
			if cbVal == nil {
				debugf("error counter pointer out of range")
				continue
			}
			offset := uintptr(dirEntry.unionData) * unsafe.Sizeof(adapter.Counter(0))
			val := *(*adapter.Counter)(statSegPointer(cbVal, offset))
			errData = append(errData, val)
		}
		return adapter.ErrorStat(errData)

	case adapter.SimpleCounterVector:
		dirVector := ss.adjust(dirVector(&dirEntry.unionData))
		if dirVector == nil {
			debugf("data vector pointer is out of range for %s", dirEntry.name)
			return nil
		}
		vecLen := *(*uint32)(vectorLen(dirVector))
		data := make([][]adapter.Counter, vecLen)
		for i := uint32(0); i < vecLen; i++ {
			counterVectorOffset := statSegPointer(dirVector, uintptr(i+1)*unsafe.Sizeof(uint64(0)))
			counterVector := ss.adjust(vectorLen(counterVectorOffset))
			if counterVector == nil {
				debugf("counter (vector simple) pointer out of range")
				continue
			}
			counterVectorLength := *(*uint32)(vectorLen(counterVector))
			if index == ^uint32(0) {
				data[i] = make([]adapter.Counter, counterVectorLength)
				for j := uint32(0); j < counterVectorLength; j++ {
					offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
					data[i][j] = *(*adapter.Counter)(statSegPointer(counterVector, offset))
				}
			} else {
				data[i] = make([]adapter.Counter, 1) // expect single value
				for j := uint32(0); j < counterVectorLength; j++ {
					offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
					if index == j {
						data[i][0] = *(*adapter.Counter)(statSegPointer(counterVector, offset))
						break
					}
				}
			}
		}
		return adapter.SimpleCounterStat(data)

	case adapter.CombinedCounterVector:
		dirVector := ss.adjust(dirVector(&dirEntry.unionData))
		if dirVector == nil {
			debugf("data vector pointer is out of range for %s", dirEntry.name)
			return nil
		}
		vecLen := *(*uint32)(vectorLen(dirVector))
		data := make([][]adapter.CombinedCounter, vecLen)
		for i := uint32(0); i < vecLen; i++ {
			counterVectorOffset := statSegPointer(dirVector, uintptr(i+1)*unsafe.Sizeof(uint64(0)))
			counterVector := ss.adjust(vectorLen(counterVectorOffset))
			if counterVector == nil {
				debugf("counter (vector combined) pointer out of range")
				continue
			}
			counterVectorLength := *(*uint32)(vectorLen(counterVector))
			if index == ^uint32(0) {
				data[i] = make([]adapter.CombinedCounter, counterVectorLength)
				for j := uint32(0); j < counterVectorLength; j++ {
					offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
					data[i][j] = *(*adapter.CombinedCounter)(statSegPointer(counterVector, offset))
				}
			} else {
				data[i] = make([]adapter.CombinedCounter, 1) // expect single value pair
				for j := uint32(0); j < counterVectorLength; j++ {
					offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
					if index == j {
						data[i][0] = *(*adapter.CombinedCounter)(statSegPointer(counterVector, offset))
						break
					}
				}
			}
		}
		return adapter.CombinedCounterStat(data)

	case adapter.NameVector:
		dirVector := ss.adjust(dirVector(&dirEntry.unionData))
		if dirVector == nil {
			debugf("data vector pointer is out of range for %s", dirEntry.name)
			return nil
		}
		vecLen := *(*uint32)(vectorLen(dirVector))
		data := make([]adapter.Name, vecLen)
		for i := uint32(0); i < vecLen; i++ {
			nameVectorOffset := statSegPointer(dirVector, uintptr(i+1)*unsafe.Sizeof(uint64(0)))
			if uintptr(nameVectorOffset) == 0 {
				debugf("name vector out of range for %s (%v)", dirEntry.name, i)
				continue
			}
			nameVector := ss.adjust(vectorLen(nameVectorOffset))
			if nameVector == nil {
				debugf("name data pointer out of range")
				continue
			}
			nameVectorLen := *(*uint32)(vectorLen(nameVector))
			name := make([]byte, 0, nameVectorLen)
			for j := uint32(0); j < nameVectorLen; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				value := *(*byte)(statSegPointer(nameVector, offset))
				if value > 0 {
					name = append(name, value)
				}
			}
			data[i] = name
		}
		return adapter.NameStat(data)

	case adapter.Empty:
		return adapter.EmptyStat("<none>")
		// no-op

	case adapter.Symlink:
		// prevent recursion loops
		if index != ^uint32(0) {
			debugf("received symlink with defined item index")
			return nil
		}
		i1, i2 := ss.getSymlinkIndexes(dirEntry)
		// use first index to get the stats directory the symlink points to
		header := ss.loadSharedHeader(ss.sharedHeader)
		dirVector := ss.adjust(dirVector(&header.dirVector))
		statSegDir2 := dirSegment(uintptr(dirVector) + uintptr(i1)*unsafe.Sizeof(statSegDirectoryEntryV2{}))

		// retry with actual stats segment and use second index to get
		// stats for the required item
		return ss.CopyEntryData(statSegDir2, i2)

	case adapter.HistogramLog2:
		return ss.copyHistogramLog2Data(dirEntry)

	case adapter.GaugeIndex:
		return adapter.GaugeStat(dirEntry.unionData)

	case adapter.RingBuffer:
		return ss.copyRingBufferData(dirEntry)

	default:
		// TODO: monitor occurrences with metrics
		debugf("Unknown type %d for stat entry: %q", dirEntry.directoryType, dirEntry.name)
	}
	return nil
}

func (ss *statSegmentV2) UpdateEntryData(segment dirSegment, stat *adapter.Stat) error {
	dirEntry := (*statSegDirectoryEntryV2)(segment)
	switch s := (*stat).(type) {
	case adapter.ScalarStat:
		*stat = adapter.ScalarStat(dirEntry.unionData)

	case adapter.ErrorStat:
		dirVector := ss.getErrorVector()
		if dirVector == nil {
			debugf("error vector pointer is out of range for %s", dirEntry.name)
			return nil
		}
		vecLen := *(*uint32)(vectorLen(dirVector))
		var errData []adapter.Counter
		for i := uint32(0); i < vecLen; i++ {
			cb := statSegPointer(dirVector, uintptr(i+1)*unsafe.Sizeof(uint64(0)))
			cbVal := ss.adjust(vectorLen(cb))
			if cbVal == nil {
				debugf("error counter pointer out of range")
				continue
			}
			offset := uintptr(dirEntry.unionData) * unsafe.Sizeof(adapter.Counter(0))
			val := *(*adapter.Counter)(statSegPointer(cbVal, offset))
			errData = append(errData, val)
		}
		*stat = adapter.ErrorStat(errData)

	case adapter.SimpleCounterStat:
		dirVector := ss.adjust(dirVector(&dirEntry.unionData))
		if dirVector == nil {
			debugf("data vector pointer is out of range for %s", dirEntry.name)
			return nil
		}
		vecLen := *(*uint32)(vectorLen(dirVector))
		data := (*stat).(adapter.SimpleCounterStat)
		if uint32(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint32(0); i < vecLen; i++ {
			counterVectorOffset := statSegPointer(dirVector, uintptr(i+1)*unsafe.Sizeof(uint64(0)))
			counterVector := ss.adjust(vectorLen(counterVectorOffset))
			if counterVector == nil {
				debugf("counter (vector simple) pointer out of range")
				continue
			}
			counterVectorLength := *(*uint32)(vectorLen(counterVector))
			data[i] = make([]adapter.Counter, counterVectorLength)
			for j := uint32(0); j < counterVectorLength; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(statSegPointer(counterVector, offset))
				data[i][j] = val
			}
		}

	case adapter.CombinedCounterStat:
		dirVector := ss.adjust(dirVector(&dirEntry.unionData))
		if dirVector == nil {
			debugf("data vector pointer is out of range for %s", dirEntry.name)
			return nil
		}
		vecLen := *(*uint32)(vectorLen(dirVector))
		data := (*stat).(adapter.CombinedCounterStat)
		if uint32(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint32(0); i < vecLen; i++ {
			counterVectorOffset := statSegPointer(dirVector, uintptr(i+1)*unsafe.Sizeof(uint64(0)))
			counterVector := ss.adjust(vectorLen(counterVectorOffset))
			if counterVector == nil {
				debugf("counter (vector combined) pointer out of range")
				continue
			}
			counterVectorLength := *(*uint32)(vectorLen(counterVector))
			data[i] = make([]adapter.CombinedCounter, counterVectorLength)
			for j := uint32(0); j < counterVectorLength; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				val := *(*adapter.CombinedCounter)(statSegPointer(counterVector, offset))
				data[i][j] = val
			}
		}

	case adapter.NameStat:
		dirVector := ss.adjust(dirVector(&dirEntry.unionData))
		if dirVector == nil {
			debugf("data vector pointer is out of range for %s", dirEntry.name)
			return nil
		}
		vecLen := *(*uint32)(vectorLen(dirVector))
		data := (*stat).(adapter.NameStat)
		if uint32(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint32(0); i < vecLen; i++ {
			nameVectorOffset := statSegPointer(dirVector, uintptr(i+1)*unsafe.Sizeof(uint64(0)))
			if uintptr(nameVectorOffset) == 0 {
				debugf("name vector out of range for %s (%v)", dirEntry.name, i)
				continue
			}
			nameVector := ss.adjust(vectorLen(nameVectorOffset))
			if nameVector == nil {
				debugf("name data pointer out of range")
				continue
			}
			nameVectorLen := *(*uint32)(vectorLen(nameVector))
			nameData := data[i]
			if uint32(len(nameData))+1 != nameVectorLen {
				return ErrStatDataLenIncorrect
			}
			for j := uint32(0); j < nameVectorLen; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				value := *(*byte)(statSegPointer(nameVector, offset))
				if value == 0 {
					break
				}
				nameData[j] = value
			}
		}

	case adapter.HistogramLog2Stat:
		histStat := ss.copyHistogramLog2Data(dirEntry)
		if histStat == nil {
			debugf("failed to read histogram log2 data for %s", dirEntry.name)
			return ErrStatDataLenIncorrect
		}
		*stat = histStat

	case adapter.GaugeStat:
		*stat = adapter.GaugeStat(dirEntry.unionData)

	case adapter.RingBufferStat:
		ringBufferStat := ss.copyRingBufferData(dirEntry)
		if ringBufferStat == nil {
			debugf("failed to read ring buffer data for %s", dirEntry.name)
			return ErrStatDataLenIncorrect
		}
		*stat = ringBufferStat

	case *adapter.RingBufferWindowStat:
		// Refreshed in place: the point of a window stat is that neither the copy
		// nor the allocation scales with ring size, so it must not be replaced by
		// a fresh value the way the other cases are.
		if err := ss.refreshRingBufferWindow(dirEntry, s); err != nil {
			return err
		}

	default:
		if Debug {
			Log.Debugf("Unrecognized stat type %T for stat entry: %v", stat, dirEntry.name)
		}
	}
	return nil
}

func (ss *statSegmentV2) copyHistogramLog2Data(dirEntry *statSegDirectoryEntryV2) adapter.Stat {
	dirVector := ss.adjust(dirVector(&dirEntry.unionData))
	if dirVector == nil {
		debugf("data vector pointer is out of range for %s", dirEntry.name)
		return nil
	}
	vecLen := *(*uint32)(vectorLen(dirVector))
	data := make(adapter.HistogramLog2Stat, vecLen)
	// Iterate over each worker's vector of bins
	for i := uint32(0); i < vecLen; i++ {
		counterVectorOffset := statSegPointer(dirVector, uintptr(i+1)*unsafe.Sizeof(uint64(0)))
		counterVector := ss.adjust(vectorLen(counterVectorOffset))
		if counterVector == nil {
			debugf("histogram log2 pointer out of range for thread %d", i)
			continue
		}
		counterVectorLength := *(*uint32)(vectorLen(counterVector))
		if counterVectorLength < 1 {
			continue
		}

		// Per thread vectors: bins[0] = min_exp, bins[1:] = bin counts.
		data[i].MinExp = *(*uint64)(statSegPointer(counterVector, 0))
		binCount := counterVectorLength - 1
		data[i].Counts = make([]uint64, binCount)
		for j := uint32(0); j < binCount; j++ {
			offset := uintptr(j+1) * unsafe.Sizeof(uint64(0))
			data[i].Counts[j] = *(*uint64)(statSegPointer(counterVector, offset))
		}
	}
	return data
}

type ringBufferHeader struct {
	EntrySize      uint32
	RingSize       uint32
	NThreads       uint32
	SchemaSize     uint32
	SchemaVersion  uint32
	MetadataOffset uint32
	DataOffset     uint32
}

// VPP pads vlib_stats_ring_metadata_t to CLIB_CACHE_LINE_BYTES, which
// src/cmake/cpu.cmake sets to 128 on aarch64 and 64 everywhere else.
const (
	ringBufferMetaSize    = 64
	ringBufferMetaSizeAlt = 128
)

// Only the leading fields are read; the stride between threads is
// ringBufferMetaStride, not this struct's size.
type ringBufferThreadMeta struct {
	Head          uint32
	SchemaVersion uint32
	Sequence      uint64
	SchemaOffset  uint32
	SchemaSize    uint32
	_             [40]byte // padding to cache line size
}

// ringBufferMetaStride returns the byte distance between adjacent threads'
// metadata.
//
// A schema-bearing ring states it: VPP puts the schema directly after the
// metadata array and records that offset in every thread's metadata, including
// thread 0's, which sits at the start of the array whatever the stride is. So
// the array's size, and from it the stride, follows from thread 0 alone.
// Without a schema there is nothing to derive it from and 64 is assumed.
func ringBufferMetaStride(base dirVector, header *ringBufferHeader, segSize uint64) uintptr {
	if header.SchemaSize == 0 || header.NThreads == 0 {
		return ringBufferMetaSize
	}
	if uint64(header.MetadataOffset)+ringBufferMetaSize > segSize {
		return ringBufferMetaSize
	}
	meta := (*ringBufferThreadMeta)(unsafe.Pointer(
		uintptr(unsafe.Pointer(base)) + uintptr(header.MetadataOffset),
	))
	schemaOffset := meta.SchemaOffset
	if schemaOffset <= header.MetadataOffset {
		return ringBufferMetaSize
	}
	metaSize := uint64(schemaOffset - header.MetadataOffset)
	if stride := metaSize / uint64(header.NThreads); stride*uint64(header.NThreads) == metaSize {
		if stride == ringBufferMetaSize || stride == ringBufferMetaSizeAlt {
			return uintptr(stride)
		}
	}
	return ringBufferMetaSize
}

func (ss *statSegmentV2) copyRingBufferData(dirEntry *statSegDirectoryEntryV2) adapter.Stat {
	base, header, stride, ok := ss.ringBufferRegions(dirEntry)
	if !ok {
		return nil
	}

	config := adapter.RingBufferConfig{
		EntrySize:     header.EntrySize,
		RingSize:      header.RingSize,
		NThreads:      header.NThreads,
		SchemaSize:    header.SchemaSize,
		SchemaVersion: header.SchemaVersion,
	}

	threads := make([]adapter.RingBufferThreadMeta, header.NThreads)
	for i := uint32(0); i < header.NThreads; i++ {
		meta := ringBufferThreadMetaAt(base, header, stride, i)
		threads[i] = adapter.RingBufferThreadMeta{
			Head:          meta.Head,
			SchemaVersion: meta.SchemaVersion,
			Sequence:      meta.Sequence,
			SchemaOffset:  meta.SchemaOffset,
			SchemaSize:    meta.SchemaSize,
		}
	}

	// Copy every thread's whole ring. This is what adapter.RingBufferWindowStat
	// exists to avoid: the cost here is the ring's size, not the entries anyone
	// has produced into it.
	data := make([][]byte, header.NThreads)
	for i := uint32(0); i < header.NThreads; i++ {
		src := ringBufferThreadData(base, header, i)
		threadData := make([]byte, len(src))
		copy(threadData, src)
		data[i] = threadData
	}

	return adapter.RingBufferStat{
		Config:  config,
		Threads: threads,
		Schema:  ss.ringBufferSchema(base, threads, dirEntry),
		Data:    data,
	}
}

// Adjust data pointer using shared header and base and return
// the pointer to a data segment
func (ss *statSegmentV2) adjust(data dirVector) dirVector {
	header := ss.loadSharedHeader(ss.sharedHeader)
	adjusted := dirVector(uintptr(unsafe.Pointer(&ss.sharedHeader[0])) +
		uintptr(*(*uint64)(data)) - uintptr(*(*uint64)(unsafe.Pointer(&header.base))))
	if uintptr(unsafe.Pointer(&ss.sharedHeader[len(ss.sharedHeader)-1])) <= uintptr(adjusted) ||
		uintptr(unsafe.Pointer(&ss.sharedHeader[0])) >= uintptr(adjusted) {
		return nil
	}
	return adjusted
}

func (ss *statSegmentV2) getErrorVector() dirVector {
	header := ss.loadSharedHeader(ss.sharedHeader)
	return ss.adjust(dirVector(&header.errorVector))
}

// GetSymlinkIndexes returns the target directory index and item index encoded in a
// symlink directory segment's union data, or ok false if the segment is not a symlink.
func (ss *statSegmentV2) GetSymlinkIndexes(segment dirSegment) (targetIndex, itemIndex uint32, ok bool) {
	dirEntry := (*statSegDirectoryEntryV2)(segment)
	if getStatType(dirEntry.directoryType, ss.getErrorVector() != nil) != adapter.Symlink {
		return 0, 0, false
	}
	targetIndex, itemIndex = ss.getSymlinkIndexes(dirEntry)
	return targetIndex, itemIndex, true
}

func (ss *statSegmentV2) getSymlinkIndexes(dirEntry *statSegDirectoryEntryV2) (index1, index2 uint32) {
	// The union holds the two indexes packed into one uint64, low half first.
	// Serialising it through a bytes.Buffer to take them apart allocated three
	// times per call - which UpdateDir now pays once per prepared symlink per
	// refresh, thousands of times on a real box.
	//
	// unionData is read as a host-order uint64, and the old code wrote it out
	// little-endian and reassembled it little-endian, so it round-tripped to the
	// same numeric value on any host. Shifting does the same, without the buffer.
	return uint32(dirEntry.unionData), uint32(dirEntry.unionData >> 32)
}

// ringBufferRegions resolves and bounds-checks the three regions of a ring
// buffer entry: its header, its per-thread metadata and its data.
//
// Every read of a ring has to do this, and doing it in one place is not only
// tidiness: these checks are what stand between a corrupt or racing header and
// an out-of-bounds read of the mapped segment, so a second copy of them is a
// second chance to get one of them subtly wrong.
func (ss *statSegmentV2) ringBufferRegions(dirEntry *statSegDirectoryEntryV2) (base dirVector, header *ringBufferHeader, stride uintptr, ok bool) {
	base = ss.adjust(dirVector(&dirEntry.unionData))
	if base == nil {
		debugf("ring buffer data pointer is out of range for %s", dirEntry.name)
		return nil, nil, 0, false
	}

	baseAddr := uintptr(unsafe.Pointer(base))
	segEnd := uintptr(unsafe.Pointer(&ss.sharedHeader[len(ss.sharedHeader)-1])) + 1
	if baseAddr >= segEnd {
		debugf("ring buffer base is outside shared memory for %s", dirEntry.name)
		return nil, nil, 0, false
	}
	// Every region is sized against the room left in the segment rather than by
	// forming its end address, because a header claiming absurd geometry makes
	// base+offset+size wrap and compare as if it fit.
	segSize := uint64(segEnd - baseAddr)

	if uint64(unsafe.Sizeof(ringBufferHeader{})) > segSize {
		debugf("ring buffer header extends beyond shared memory for %s", dirEntry.name)
		return nil, nil, 0, false
	}
	header = (*ringBufferHeader)(unsafe.Pointer(base))

	stride = ringBufferMetaStride(base, header, segSize)
	if uint64(header.MetadataOffset) > segSize ||
		uint64(header.NThreads) > (segSize-uint64(header.MetadataOffset))/uint64(stride) {
		debugf("ring buffer metadata extends beyond shared memory for %s (offset=%d, threads=%d, segSize=%d)",
			dirEntry.name, header.MetadataOffset, header.NThreads, segSize)
		return nil, nil, 0, false
	}

	// Both factors are uint32, so the product cannot overflow uint64.
	threadDataSize := uint64(header.RingSize) * uint64(header.EntrySize)
	if uint64(header.DataOffset) > segSize || (threadDataSize > 0 &&
		uint64(header.NThreads) > (segSize-uint64(header.DataOffset))/threadDataSize) {
		debugf("ring buffer data extends beyond shared memory for %s (offset=%d, threads=%d, segSize=%d)",
			dirEntry.name, header.DataOffset, header.NThreads, segSize)
		return nil, nil, 0, false
	}

	return base, header, stride, true
}

// ringBufferThreadMetaAt returns thread t's producer metadata.
func ringBufferThreadMetaAt(base dirVector, header *ringBufferHeader, stride uintptr, t uint32) *ringBufferThreadMeta {
	return (*ringBufferThreadMeta)(unsafe.Pointer(
		uintptr(unsafe.Pointer(base)) + uintptr(header.MetadataOffset) + uintptr(t)*stride,
	))
}

// ringBufferThreadData returns thread t's ring as a slice over the mapped
// segment. Read-only: it aliases shared memory a VPP worker is writing.
func ringBufferThreadData(base dirVector, header *ringBufferHeader, t uint32) []byte {
	threadDataSize := uintptr(header.RingSize) * uintptr(header.EntrySize)
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(base)) + uintptr(header.DataOffset) + uintptr(t)*threadDataSize)
	return unsafe.Slice((*byte)(ptr), threadDataSize)
}

// ringBufferSchema copies the schema blob from the first thread that publishes
// one, or returns nil when none does.
func (ss *statSegmentV2) ringBufferSchema(base dirVector, threads []adapter.RingBufferThreadMeta, dirEntry *statSegDirectoryEntryV2) []byte {
	baseAddr := uintptr(unsafe.Pointer(base))
	segEnd := uintptr(unsafe.Pointer(&ss.sharedHeader[len(ss.sharedHeader)-1])) + 1
	for _, t := range threads {
		if t.SchemaSize == 0 || t.SchemaOffset == 0 {
			continue
		}
		if baseAddr+uintptr(t.SchemaOffset)+uintptr(t.SchemaSize) > segEnd {
			debugf("ring buffer schema extends beyond shared memory for %s", dirEntry.name)
			continue
		}
		// Derived from base in one expression: a uintptr held across statements
		// and converted back loses its provenance, which is undefined and what
		// checkptr rejects under -race.
		src := unsafe.Slice((*byte)(unsafe.Pointer(
			uintptr(unsafe.Pointer(base))+uintptr(t.SchemaOffset))), t.SchemaSize)
		schema := make([]byte, t.SchemaSize)
		copy(schema, src)
		return schema
	}
	return nil
}

// ringWindowCopyHook runs between a window's entries being copied out of the
// segment and their validation against the producer. It is nil in production and
// costs one nil check per thread per refresh; it exists because the case the
// validation is there for - a worker overwriting the slots while the copy runs -
// is otherwise reachable only by racing a real producer, which is not something
// a test can assert on.
var ringWindowCopyHook func()

// refreshRingBufferWindow copies, for each producer thread, only the entries
// appended since the previous refresh - into buffers the stat already owns.
//
// This is the whole reason the type exists, so it is worth being explicit about
// what is not here: no allocation on the steady path, and no read of any slot
// the producer has not written since we last looked. The cost of a refresh is
// the entries produced, not the size of the ring they were produced into.
func (ss *statSegmentV2) refreshRingBufferWindow(dirEntry *statSegDirectoryEntryV2, s *adapter.RingBufferWindowStat) error {
	base, header, stride, ok := ss.ringBufferRegions(dirEntry)
	if !ok {
		return ErrStatDataLenIncorrect
	}
	if header.EntrySize == 0 || header.RingSize == 0 || header.NThreads == 0 {
		debugf("ring buffer has zero geometry for %s (entry_size=%d, ring_size=%d, threads=%d)",
			dirEntry.name, header.EntrySize, header.RingSize, header.NThreads)
		return ErrStatDataLenIncorrect
	}

	cfg := adapter.RingBufferConfig{
		EntrySize:     header.EntrySize,
		RingSize:      header.RingSize,
		NThreads:      header.NThreads,
		SchemaSize:    header.SchemaSize,
		SchemaVersion: header.SchemaVersion,
	}

	// A geometry change invalidates both the buffers, which are sized from it,
	// and the cursors, which are positions within it. Treat it as a first read
	// rather than trying to carry a cursor across a ring that is no longer the
	// same ring.
	reinit := s.Config != cfg || len(s.Windows) != int(header.NThreads)
	if reinit {
		s.Config = cfg
		s.Threads = make([]adapter.RingBufferThreadMeta, header.NThreads)
		s.Windows = make([]adapter.RingBufferWindow, header.NThreads)
	}

	maxEntries := s.MaxEntries
	if maxEntries == 0 || maxEntries > header.RingSize {
		maxEntries = header.RingSize
	}

	for t := uint32(0); t < header.NThreads; t++ {
		meta := ringBufferThreadMetaAt(base, header, stride, t)
		// The sequence is the only field of the pair that carries an ordering
		// guarantee: a producer writes the entry, advances head with a plain
		// store, and then publishes the sequence with a release store. So a
		// consumer that reads both can see a head one slot ahead of the sequence
		// that explains it, and everything below positions off the sequence
		// alone. Head is loaded only to report it.
		head := atomic.LoadUint32(&meta.Head)
		seq := atomic.LoadUint64(&meta.Sequence)

		s.Threads[t] = adapter.RingBufferThreadMeta{
			Head:          head,
			SchemaVersion: meta.SchemaVersion,
			Sequence:      seq,
			SchemaOffset:  meta.SchemaOffset,
			SchemaSize:    meta.SchemaSize,
		}

		w := &s.Windows[t]
		w.Lost, w.Pending, w.Count = 0, 0, 0

		if cap(w.Entries) < int(maxEntries)*int(header.EntrySize) {
			w.Entries = make([]byte, 0, int(maxEntries)*int(header.EntrySize))
		}
		buf := w.Entries[:cap(w.Entries)]

		if reinit {
			// Position the cursor without delivering anything. A first read that
			// also returned a ring of history would make "what happened since I
			// last looked" mean something different on the first call than on
			// every later one.
			//
			// A geometry change mid-stream reaches here too, SchemaVersion
			// included, so with SkipBacklog whatever the ring already held is
			// dropped without appearing in Lost.
			if s.SkipBacklog {
				w.NextSeq = seq
			} else {
				w.NextSeq = seq - min(seq, uint64(header.RingSize))
			}
			w.FirstSeq = w.NextSeq
			w.Entries = buf[:0]
			continue
		}

		if seq < w.NextSeq {
			// The producer's sequence went backwards: VPP restarted, or the entry
			// was reused for a different ring. Re-sync to what is there now and do
			// not report it as loss - nothing was overwritten, the count simply is
			// not comparable to the one we held.
			w.NextSeq = seq - min(seq, uint64(header.RingSize))
		}

		available := seq - w.NextSeq
		if available > uint64(header.RingSize) {
			// Lapped: everything older than the last RingSize entries is gone.
			w.Lost = available - uint64(header.RingSize)
			w.NextSeq = seq - uint64(header.RingSize)
			available = uint64(header.RingSize)
		}

		deliver := available
		if deliver > uint64(maxEntries) {
			deliver = uint64(maxEntries)
			w.Pending = available - deliver
		}

		w.FirstSeq = w.NextSeq
		if deliver == 0 {
			w.Entries = buf[:0]
			continue
		}

		// The entry with sequence x sits at slot x mod RingSize. Head is not used
		// to find it: head and the sequence both start at zero and advance
		// together on every commit, so head is congruent to the sequence and
		// carries no information the sequence does not - but it is published
		// first and without a barrier, so deriving the slot from it delivers the
		// slot a worker is writing right now whenever the two are read astride a
		// commit.
		first := uint32(w.NextSeq % uint64(header.RingSize))

		data := ringBufferThreadData(base, header, t)
		entry := uintptr(header.EntrySize)
		n := uint32(deliver)

		// At most two copies: the run up to the ring's seam, then the rest from
		// slot zero. The consumer sees them unwrapped and never learns where the
		// seam was.
		firstRun := header.RingSize - first
		if firstRun > n {
			firstRun = n
		}
		copy(buf[:uintptr(firstRun)*entry], data[uintptr(first)*entry:uintptr(first+firstRun)*entry])
		if rest := n - firstRun; rest > 0 {
			copy(buf[uintptr(firstRun)*entry:uintptr(n)*entry], data[:uintptr(rest)*entry])
		}

		if ringWindowCopyHook != nil {
			ringWindowCopyHook()
		}

		// Check the copy against the producer, which did not stop while it ran:
		// the segment's optimistic lock covers directory changes, not ring data,
		// so a worker is free to overwrite the slots being copied. Anything older
		// than seqAfter-RingSize was overwritten underneath us, and those are the
		// oldest entries of the window, so dropping them from the front of the
		// buffer leaves exactly the ones that are still whole.
		//
		// Without this a lapped reader delivers entries that are half one record
		// and half another and reports no loss, which is worse than losing them:
		// loss is visible and a torn record is not.
		windowStart := w.NextSeq
		if seqAfter := atomic.LoadUint64(&meta.Sequence); seqAfter > uint64(header.RingSize) {
			if oldest := seqAfter - uint64(header.RingSize); oldest > windowStart {
				overrun := min(oldest-windowStart, uint64(n))
				if overrun < uint64(n) {
					copy(buf, buf[uintptr(overrun)*entry:uintptr(n)*entry])
				}
				w.Lost += overrun
				n -= uint32(overrun)
			}
		}

		w.Count = n
		w.NextSeq += deliver
		// Derived rather than remembered, so it stays right when the check above
		// drops entries from the front: FirstSeq is the sequence of Entries[0],
		// and equals NextSeq when nothing was delivered.
		w.FirstSeq = w.NextSeq - uint64(n)
		w.Entries = buf[:uintptr(n)*entry]
	}

	if s.Schema == nil || reinit {
		s.Schema = ss.ringBufferSchema(base, s.Threads, dirEntry)
	}
	return nil
}
