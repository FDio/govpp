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
	"git.fd.io/govpp.git/adapter"
	"sync/atomic"
	"time"
	"unsafe"
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

const (
	statDirIllegal               = 0
	statDirScalarIndex           = 1
	statDirCounterVectorSimple   = 2
	statDirCounterVectorCombined = 3
	statDirErrorIndex            = 4
	statDirNameVector            = 5
	statDirEmpty                 = 6
)

type statDirectoryType int32

type statDirectoryName []byte

// statSegment represents common API for every stats API version
type statSegment interface {
	// GetDirectoryVector returns pointer to memory where the beginning
	// of the data directory is located.
	GetDirectoryVector() unsafe.Pointer

	// GetStatDirOnIndex accepts directory vector and particular index.
	// Returns pointer to the beginning of the segment. Also the directory
	// name as [128]byte and the directory type is returned for easy use
	// without needing to know the exact segment version.
	//
	// Note that if the index is equal to 0, the result pointer points to
	// the same memory address as the argument.
	GetStatDirOnIndex(directory unsafe.Pointer, index uint32) (unsafe.Pointer, statDirectoryName, statDirectoryType)

	// GetEpoch re-loads stats header and returns current epoch
	//and 'inProgress' value
	GetEpoch() (int64, bool)

	// CopyEntryData accepts pointer to a directory segment and returns adapter.Stat
	// based on directory type populated with data
	CopyEntryData(segment unsafe.Pointer, limit uint32) adapter.Stat

	// UpdateEntryData accepts pointer to a directory segment with data, and stat
	// segment to update
	UpdateEntryData(segment unsafe.Pointer, s *adapter.Stat) error
}

const copyLimitNone = ^uint32(0)

// vecHeader represents a vector header
type vecHeader struct {
	length     uint64
	vectorData [0]uint8
}

func (t statDirectoryType) String() string {
	return adapter.StatType(t).String()
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

func vectorLen(v unsafe.Pointer) unsafe.Pointer {
	vec := *(*vecHeader)(unsafe.Pointer(uintptr(v) - unsafe.Sizeof(uint64(0))))
	return unsafe.Pointer(&vec.length)
}

//go:nosplit
func statSegPointer(p unsafe.Pointer, offset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + offset)
}
