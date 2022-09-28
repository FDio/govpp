//  Copyright (c) 2019 Cisco and/or its affiliates.
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
	"unsafe"

	"go.fd.io/govpp/adapter"
)

type statSegmentV1 struct {
	sharedHeader []byte
	memorySize   int64
}

type sharedHeaderV1 struct {
	version         uint64
	epoch           int64
	inProgress      int64
	directoryOffset int64
	errorOffset     int64
	statsOffset     int64
}

type statSegDirectoryEntryV1 struct {
	directoryType dirType
	// unionData can represent:
	// - offset
	// - index
	// - value
	unionData    uint64
	offsetVector uint64
	name         [128]byte
}

func newStatSegmentV1(data []byte, size int64) *statSegmentV1 {
	return &statSegmentV1{
		sharedHeader: data,
		memorySize:   size,
	}
}

func (ss *statSegmentV1) loadSharedHeader(b []byte) (header sharedHeaderV1) {
	h := (*sharedHeaderV1)(unsafe.Pointer(&b[0]))
	return sharedHeaderV1{
		version:         atomic.LoadUint64(&h.version),
		epoch:           atomic.LoadInt64(&h.epoch),
		inProgress:      atomic.LoadInt64(&h.inProgress),
		directoryOffset: atomic.LoadInt64(&h.directoryOffset),
		errorOffset:     atomic.LoadInt64(&h.errorOffset),
		statsOffset:     atomic.LoadInt64(&h.statsOffset),
	}
}

func (ss *statSegmentV1) GetDirectoryVector() dirVector {
	dirOffset, _, _ := ss.getOffsets()
	return dirVector(&ss.sharedHeader[dirOffset])
}

func (ss *statSegmentV1) GetStatDirOnIndex(v dirVector, index uint32) (dirSegment, dirName, adapter.StatType) {
	statSegDir := dirSegment(uintptr(v) + uintptr(index)*unsafe.Sizeof(statSegDirectoryEntryV1{}))
	dir := (*statSegDirectoryEntryV1)(statSegDir)
	var name []byte
	for n := 0; n < len(dir.name); n++ {
		if dir.name[n] == 0 {
			name = dir.name[:n]
			break
		}
	}
	return statSegDir, name, getStatType(dir.directoryType, true)
}

func (ss *statSegmentV1) GetEpoch() (int64, bool) {
	sh := ss.loadSharedHeader(ss.sharedHeader)
	return sh.epoch, sh.inProgress != 0
}

func (ss *statSegmentV1) CopyEntryData(segment dirSegment, _ uint32) adapter.Stat {
	dirEntry := (*statSegDirectoryEntryV1)(segment)
	typ := getStatType(dirEntry.directoryType, true)

	switch typ {
	case adapter.ScalarIndex:
		return adapter.ScalarStat(dirEntry.unionData)

	case adapter.ErrorIndex:
		if dirEntry.unionData >= uint64(len(ss.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		_, errOffset, _ := ss.getOffsets()
		offsetVector := dirVector(&ss.sharedHeader[errOffset])

		var errData []adapter.Counter

		vecLen := *(*uint32)(vectorLen(offsetVector))
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			offset := uintptr(cb) + uintptr(dirEntry.unionData)*unsafe.Sizeof(adapter.Counter(0))
			debugf("error index, cb: %d, offset: %d", cb, offset)
			val := *(*adapter.Counter)(statSegPointer(dirVector(&ss.sharedHeader[0]), offset))
			errData = append(errData, val)
		}
		return adapter.ErrorStat(errData)

	case adapter.SimpleCounterVector:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(ss.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := *(*uint32)(vectorLen(dirVector(&ss.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(dirVector(&ss.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.Counter, vecLen)
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := dirVector(&ss.sharedHeader[uintptr(cb)])
			vecLen2 := *(*uint32)(vectorLen(counterVec))
			data[i] = make([]adapter.Counter, vecLen2)
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(statSegPointer(counterVec, offset))
				data[i][j] = val
			}
		}
		return adapter.SimpleCounterStat(data)

	case adapter.CombinedCounterVector:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(ss.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := *(*uint32)(vectorLen(dirVector(&ss.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(dirVector(&ss.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.CombinedCounter, vecLen)
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := dirVector(&ss.sharedHeader[uintptr(cb)])
			vecLen2 := *(*uint32)(vectorLen(counterVec))
			data[i] = make([]adapter.CombinedCounter, vecLen2)
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				val := *(*adapter.CombinedCounter)(statSegPointer(counterVec, offset))
				data[i][j] = val
			}
		}
		return adapter.CombinedCounterStat(data)

	case adapter.NameVector:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(ss.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := *(*uint32)(vectorLen(dirVector(&ss.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(dirVector(&ss.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([]adapter.Name, vecLen)
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			if cb == 0 {
				debugf("name vector out of range for %s (%v)", dirEntry.name, i)
				continue
			}
			nameVec := dirVector(&ss.sharedHeader[cb])
			vecLen2 := *(*uint32)(vectorLen(nameVec))

			nameStr := make([]byte, 0, vecLen2)
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				val := *(*byte)(statSegPointer(nameVec, offset))
				if val > 0 {
					nameStr = append(nameStr, val)
				}
			}
			data[i] = nameStr
		}
		return adapter.NameStat(data)

	case adapter.Empty:
		// no-op

	case adapter.Symlink:
		debugf("Symlinks are not supported for stats v1")

	default:
		// TODO: monitor occurrences with metrics
		debugf("Unknown type %d for stat entry: %q", dirEntry.directoryType, dirEntry.name)
	}
	return nil
}

func (ss *statSegmentV1) UpdateEntryData(segment dirSegment, stat *adapter.Stat) error {
	dirEntry := (*statSegDirectoryEntryV1)(segment)
	switch (*stat).(type) {
	case adapter.ScalarStat:
		*stat = adapter.ScalarStat(dirEntry.unionData)

	case adapter.ErrorStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(ss.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		_, errOffset, _ := ss.getOffsets()
		offsetVector := dirVector(&ss.sharedHeader[errOffset])

		var errData []adapter.Counter

		vecLen := *(*uint32)(vectorLen(dirVector(&ss.sharedHeader[errOffset])))
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			offset := uintptr(cb) + uintptr(dirEntry.unionData)*unsafe.Sizeof(adapter.Counter(0))
			val := *(*adapter.Counter)(statSegPointer(dirVector(&ss.sharedHeader[0]), offset))
			errData = append(errData, val)
		}
		*stat = adapter.ErrorStat(errData)

	case adapter.SimpleCounterStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(ss.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := *(*uint32)(vectorLen(dirVector(&ss.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(dirVector(&ss.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := (*stat).(adapter.SimpleCounterStat)
		if uint32(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := dirVector(&ss.sharedHeader[uintptr(cb)])
			vecLen2 := *(*uint32)(vectorLen(counterVec))
			simpleData := data[i]
			if uint32(len(simpleData)) != vecLen2 {
				return ErrStatDataLenIncorrect
			}
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(statSegPointer(counterVec, offset))
				simpleData[j] = val
			}
		}

	case adapter.CombinedCounterStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(ss.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := *(*uint32)(vectorLen(dirVector(&ss.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(dirVector(&ss.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := (*stat).(adapter.CombinedCounterStat)
		if uint32(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := dirVector(&ss.sharedHeader[uintptr(cb)])
			vecLen2 := *(*uint32)(vectorLen(counterVec))
			combData := data[i]
			if uint32(len(combData)) != vecLen2 {
				return ErrStatDataLenIncorrect
			}
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				val := *(*adapter.CombinedCounter)(statSegPointer(counterVec, offset))
				combData[j] = val
			}
		}

	case adapter.NameStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(ss.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := *(*uint32)(vectorLen(dirVector(&ss.sharedHeader[dirEntry.unionData])))
		offsetVector := statSegPointer(dirVector(&ss.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := (*stat).(adapter.NameStat)
		if uint32(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint32(0); i < vecLen; i++ {
			cb := *(*uint64)(statSegPointer(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			if cb == 0 {
				continue
			}
			nameVec := dirVector(&ss.sharedHeader[cb])
			vecLen2 := *(*uint32)(vectorLen(nameVec))

			nameData := data[i]
			if uint32(len(nameData))+1 != vecLen2 {
				return ErrStatDataLenIncorrect
			}
			for j := uint32(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				val := *(*byte)(statSegPointer(nameVec, offset))
				if val == 0 {
					break
				}
				nameData[j] = val
			}
		}

	default:
		if Debug {
			Log.Debugf("Unrecognized stat type %T for stat entry: %v", stat, dirEntry.name)
		}
	}
	return nil
}

// Get offsets for various types of data
func (ss *statSegmentV1) getOffsets() (dir, err, stat int64) {
	sh := ss.loadSharedHeader(ss.sharedHeader)
	return sh.directoryOffset, sh.errorOffset, sh.statsOffset
}
