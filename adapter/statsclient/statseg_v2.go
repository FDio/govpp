//  Copyright (c) 2020 Cisco and/or its affiliates.
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
	"encoding/binary"
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
			name = dir.name[:n]
			break
		}
	}
	return statSegDir, name, getStatType(dir.directoryType, ss.getErrorVector() != nil)
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
	switch (*stat).(type) {
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

// Respective VPP struct vlib_stats_ring_metadata_t is padded to 64 bytes for cache alignment
const ringBufferMetaSize = 64

type ringBufferThreadMeta struct {
	Head          uint32
	SchemaVersion uint32
	Sequence      uint64
	SchemaOffset  uint32
	SchemaSize    uint32
	_             [40]byte // padding to cache line size
}

func (ss *statSegmentV2) copyRingBufferData(dirEntry *statSegDirectoryEntryV2) adapter.Stat {
	base := ss.adjust(dirVector(&dirEntry.unionData))
	if base == nil {
		debugf("ring buffer data pointer is out of range for %s", dirEntry.name)
		return nil
	}

	// Get start and end of ring buffer entry mapped region.
	baseAddr := uintptr(unsafe.Pointer(base))
	segEnd := uintptr(unsafe.Pointer(&ss.sharedHeader[len(ss.sharedHeader)-1])) + 1

	// Verify the ring buffer header fits within the mapped region.
	headerSize := unsafe.Sizeof(ringBufferHeader{})
	if baseAddr+headerSize > segEnd {
		debugf("ring buffer header extends beyond shared memory for %s", dirEntry.name)
		return nil
	}

	// Cast to header struct instead of reading fields individually.
	header := (*ringBufferHeader)(unsafe.Pointer(base))

	// Verify metadata region fits within the mapped region.
	metaEnd := baseAddr + uintptr(header.MetadataOffset) + uintptr(header.NThreads)*ringBufferMetaSize
	if metaEnd > segEnd {
		debugf("ring buffer metadata extends beyond shared memory for %s (metaEnd=%d, segEnd=%d)",
			dirEntry.name, metaEnd, segEnd)
		return nil
	}

	// Verify data region fits within the mapped region.
	threadDataSize := uintptr(header.RingSize) * uintptr(header.EntrySize)
	dataEnd := baseAddr + uintptr(header.DataOffset) + uintptr(header.NThreads)*threadDataSize
	if dataEnd > segEnd {
		debugf("ring buffer data extends beyond shared memory for %s (dataEnd=%d, segEnd=%d)",
			dirEntry.name, dataEnd, segEnd)
		return nil
	}

	config := adapter.RingBufferConfig{
		EntrySize:     header.EntrySize,
		RingSize:      header.RingSize,
		NThreads:      header.NThreads,
		SchemaSize:    header.SchemaSize,
		SchemaVersion: header.SchemaVersion,
	}

	// Read per-thread metadata by casting to the raw struct.
	threads := make([]adapter.RingBufferThreadMeta, header.NThreads)
	for i := uint32(0); i < header.NThreads; i++ {
		meta := (*ringBufferThreadMeta)(unsafe.Pointer(
			baseAddr + uintptr(header.MetadataOffset) + uintptr(i)*ringBufferMetaSize,
		))
		threads[i] = adapter.RingBufferThreadMeta{
			Head:          meta.Head,
			SchemaVersion: meta.SchemaVersion,
			Sequence:      meta.Sequence,
			SchemaOffset:  meta.SchemaOffset,
			SchemaSize:    meta.SchemaSize,
		}
	}

	// Read schema from the first thread that has one.
	var schema []byte
	for _, t := range threads {
		if t.SchemaSize > 0 && t.SchemaOffset > 0 {
			schemaEnd := baseAddr + uintptr(t.SchemaOffset) + uintptr(t.SchemaSize)
			if schemaEnd > segEnd {
				debugf("ring buffer schema extends beyond shared memory for %s (schemaEnd=%d, segEnd=%d)",
					dirEntry.name, schemaEnd, segEnd)
				continue
			}
			schemaPtr := unsafe.Pointer(baseAddr + uintptr(t.SchemaOffset))
			srcSchema := unsafe.Slice((*byte)(schemaPtr), t.SchemaSize)
			schema = make([]byte, t.SchemaSize)
			copy(schema, srcSchema)
			break
		}
	}

	// Copy per-thread raw ring data.
	data := make([][]byte, header.NThreads)
	for i := uint32(0); i < header.NThreads; i++ {
		srcPtr := unsafe.Pointer(baseAddr + uintptr(header.DataOffset) + uintptr(i)*threadDataSize)
		srcSlice := unsafe.Slice((*byte)(srcPtr), threadDataSize)
		threadData := make([]byte, threadDataSize)
		copy(threadData, srcSlice)
		data[i] = threadData
	}

	return adapter.RingBufferStat{
		Config:  config,
		Threads: threads,
		Schema:  schema,
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

func (ss *statSegmentV2) getSymlinkIndexes(dirEntry *statSegDirectoryEntryV2) (index1, index2 uint32) {
	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, dirEntry.unionData); err != nil {
		debugf("error getting symlink indexes for %s: %v", dirEntry.name, err)
		return
	}
	if len(b.Bytes()) != 8 {
		debugf("incorrect symlink union data length for %s: expected 8, got %d", dirEntry.name, len(b.Bytes()))
		return
	}
	for i := range b.Bytes()[:4] {
		index1 += uint32(b.Bytes()[i]) << (uint32(i) * 8)
	}
	for i := range b.Bytes()[4:] {
		index2 += uint32(b.Bytes()[i+4]) << (uint32(i) * 8)
	}
	return
}
