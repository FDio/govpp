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

	"git.fd.io/govpp.git/adapter"
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

func (ss *statSegmentV2) GetStatDirOnIndex(v dirVector, index uint32) (dirSegment, dirName, dirType) {
	statSegDir := dirSegment(uintptr(v) + uintptr(index)*unsafe.Sizeof(statSegDirectoryEntryV2{}))
	dir := (*statSegDirectoryEntryV2)(statSegDir)
	var name []byte
	for n := 0; n < len(dir.name); n++ {
		if dir.name[n] == 0 {
			name = dir.name[:n]
			break
		}
	}
	return statSegDir, name, dir.directoryType
}

func (ss *statSegmentV2) GetEpoch() (int64, bool) {
	sh := ss.loadSharedHeader(ss.sharedHeader)
	return sh.epoch, sh.inProgress != 0
}

func (ss *statSegmentV2) CopyEntryData(segment dirSegment, imin uint32, imax uint32) adapter.Stat {
	return ss.copyEntryData(segment, imin, imax, false)
}

func (ss *statSegmentV2) copyEntryData(segment dirSegment, imin uint32, imax uint32, nofollow bool) adapter.Stat {
	dirEntry := (*statSegDirectoryEntryV2)(segment)
	typ := adapter.StatType(dirEntry.directoryType)
	// skip zero pointer value
	if typ != statDirScalarIndex && typ != statDirEmpty && dirEntry.unionData == 0 {
		debugf("data pointer not defined for %s", dirEntry.name)
		return nil
	}

	switch typ {
	case statDirScalarIndex:
		return adapter.ScalarStat(dirEntry.unionData)

	case statDirErrorIndex:
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

	case statDirCounterVectorSimple:
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
			imax = applyLimit(imax, *(*uint32)(vectorLen(counterVector)))
			imin = applyLimit(imin, imax)
			data[i] = make([]adapter.Counter, imax-imin)
			for j := imin; j < imax; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				data[i][j-imin] = *(*adapter.Counter)(statSegPointer(counterVector, offset))
			}
		}
		return adapter.SimpleCounterStat(data)

	case statDirCounterVectorCombined:
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
			imax = applyLimit(imax, *(*uint32)(vectorLen(counterVector)))
			imin = applyLimit(imin, imax)
			data[i] = make([]adapter.CombinedCounter, imax-imin)
			for j := imin; j < imax; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				data[i][j-imin] = *(*adapter.CombinedCounter)(statSegPointer(counterVector, offset))
			}
		}
		return adapter.CombinedCounterStat(data)

	case statDirNameVector:
		dirVector := ss.adjust(dirVector(&dirEntry.unionData))
		if dirVector == nil {
			debugf("data vector pointer is out of range for %s", dirEntry.name)
			return nil
		}
		imax = applyLimit(imax, *(*uint32)(vectorLen(dirVector)))
		imin = applyLimit(imin, imax)
		data := make([]adapter.Name, imax-imin)
		for i := imin; i < imax; i++ {
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
			data[i-imin] = name
		}
		return adapter.NameStat(data)

	case statDirEmpty:
		return adapter.EmptyStat("<none>")
		// no-op

	case statDirSymlink:
		// prevent recursion loops
		if nofollow {
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
		return ss.copyEntryData(statSegDir2, i2, i2+1 /* nofollow */, true)

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

	default:
		if Debug {
			Log.Debugf("Unrecognized stat type %T for stat entry: %v", stat, dirEntry.name)
		}
	}
	return nil
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

func applyLimit(val, limit uint32) uint32 {
	if val > limit {
		return limit
	}
	return val
}
