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
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"go.fd.io/govpp/adapter"
)

var (
	// ErrStatDataLenIncorrect is returned when stat data does not match vector
	// length of a respective data directory
	ErrStatDataLenIncorrect = fmt.Errorf("stat data length incorrect")
)

var (
	MaxWaitInProgress    = time.Millisecond * 100
	CheckDelayInProgress = time.Microsecond * 10
)

const (
	minVersion = 1
	maxVersion = 2
)

var dirTypeMapping = map[dirType]adapter.StatType{
	1: adapter.ScalarIndex,
	2: adapter.SimpleCounterVector,
	3: adapter.CombinedCounterVector,
	4: adapter.NameVector,
	5: adapter.Empty,
	6: adapter.Symlink,
	9: adapter.GaugeIndex,
}

var dirTypeMappingLegacy = map[dirType]adapter.StatType{
	1: adapter.ScalarIndex,
	2: adapter.SimpleCounterVector,
	3: adapter.CombinedCounterVector,
	4: adapter.ErrorIndex,
	5: adapter.NameVector,
	6: adapter.Empty,
	7: adapter.Symlink,
}

type (
	dirVector  unsafe.Pointer
	dirSegment unsafe.Pointer
	dirName    []byte
	dirType    int32
)

// statSegment represents common API for every stats API version
type statSegment interface {
	// GetDirectoryVector returns pointer to memory where the beginning
	// of the data directory is located.
	GetDirectoryVector() dirVector

	// GetStatDirOnIndex accepts directory vector and particular index.
	// Returns pointer to the beginning of the segment. Also the directory
	// name as [128]byte and the directory type is returned for easy use
	// without needing to know the exact segment version.
	//
	// Note that if the index is equal to 0, the result pointer points to
	// the same memory address as the argument.
	GetStatDirOnIndex(v dirVector, index uint32) (dirSegment, dirName, adapter.StatType)

	// GetEpoch re-loads stats header and returns current epoch
	//and 'inProgress' value
	GetEpoch() (int64, bool)

	// CopyEntryData accepts pointer to a directory segment and returns adapter.Stat
	// based on directory type populated with data. The index is an optional parameter
	// (used by symlinks) returning stats for item on the given index only.
	// Use ^uint32(0) as an empty index (since 0 is a valid value).
	CopyEntryData(segment dirSegment, index uint32) adapter.Stat

	// UpdateEntryData accepts pointer to a directory segment with data, and stat
	// segment to update
	UpdateEntryData(segment dirSegment, s *adapter.Stat) error
}

// vecHeader represents a vector header
type vecHeader struct {
	length     uint64
	vectorData [0]uint8
}

func getVersion(data []byte) uint64 {
	type apiVersion struct {
		value uint64
	}
	header := (*apiVersion)(unsafe.Pointer(&data[0]))
	version := &apiVersion{
		value: atomic.LoadUint64(&header.value),
	}
	debugf("stats API version loaded: %d", version.value)
	return version.value
}

func vectorLen(v dirVector) dirVector {
	vec := *(*vecHeader)(unsafe.Pointer(uintptr(v) - unsafe.Sizeof(uint64(0))))
	return dirVector(&vec.length)
}

func getStatType(dirTypeNum dirType, useLegacyMapping bool) (dirTyp adapter.StatType) {
	var exists bool
	if useLegacyMapping {
		dirTyp, exists = dirTypeMappingLegacy[dirTypeNum]
	} else {
		dirTyp, exists = dirTypeMapping[dirTypeNum]
	}
	if exists {
		return dirTyp
	}
	return adapter.Unknown
}

//go:nosplit
func statSegPointer(v dirVector, offset uintptr) dirVector {
	return dirVector(uintptr(v) + offset)
}
