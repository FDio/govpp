// Copyright (c) 2019 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package statsclient

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"regexp"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/ftrvxmtrx/fd"
	logger "github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/adapter"
)

var (
	// Debug is global variable that determines debug mode
	Debug = os.Getenv("DEBUG_GOVPP_STATS") != ""

	// Log is global logger
	Log = logger.New()
)

// init initializes global logger, which logs debug level messages to stdout.
func init() {
	Log.Out = os.Stdout
	if Debug {
		Log.Level = logger.DebugLevel
		Log.Debug("enabled debug mode")
	}
}

// StatsClient is the pure Go implementation for VPP stats API.
type StatsClient struct {
	sockAddr string

	currentEpoch    int64
	sharedHeader    []byte
	directoryVector uintptr
	memorySize      int
}

// NewStatsClient returns new VPP stats API client.
func NewStatsClient(socketName string) *StatsClient {
	return &StatsClient{
		sockAddr: socketName,
	}
}

func (c *StatsClient) Connect() error {
	var sockName string
	if c.sockAddr == "" {
		sockName = adapter.DefaultStatsSocket
	} else {
		sockName = c.sockAddr
	}

	if _, err := os.Stat(sockName); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("stats socket file %q does not exists, ensure that VPP is running with `statseg { ... }` section in config", sockName)
		}
		return fmt.Errorf("stats socket file error: %v", err)
	}

	if err := c.statSegmentConnect(sockName); err != nil {
		return err
	}

	return nil
}

const statshmFilename = "statshm"

func (c *StatsClient) statSegmentConnect(sockName string) error {
	addr := &net.UnixAddr{
		Net:  "unixpacket",
		Name: sockName,
	}

	Log.Debugf("connecting to: %v", addr)

	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		Log.Warnf("connecting to socket %s failed: %s", addr, err)
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			Log.Warnf("closing socket failed: %v", err)
		}
	}()

	Log.Debugf("connected to socket: %v", addr)

	files, err := fd.Get(conn, 1, []string{statshmFilename})
	if err != nil {
		return fmt.Errorf("getting file descriptor over socket failed: %v", err)
	} else if len(files) == 0 {
		return fmt.Errorf("no files received over socket")
	}
	defer func() {
		for _, f := range files {
			if err := f.Close(); err != nil {
				Log.Warnf("closing file %s failed: %v", f.Name(), err)
			}
		}
	}()

	Log.Debugf("received %d files over socket", len(files))

	f := files[0]

	info, err := f.Stat()
	if err != nil {
		return err
	}

	size := int(info.Size())

	Log.Debugf("fd: name=%v size=%v", info.Name(), size)

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		Log.Warnf("mapping shared memory failed: %v", err)
		return fmt.Errorf("mapping shared memory failed: %v", err)
	}

	Log.Debugf("successfuly mapped shared memory")

	c.sharedHeader = data
	c.memorySize = size

	return nil
}

func (c *StatsClient) Disconnect() error {
	err := syscall.Munmap(c.sharedHeader)
	if err != nil {
		Log.Warnf("unmapping shared memory failed: %v", err)
		return fmt.Errorf("unmapping shared memory failed: %v", err)
	}

	Log.Debugf("successfuly unmapped shared memory")

	return nil
}

func nameMatches(name string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func (c *StatsClient) ListStats(patterns ...string) (statNames []string, err error) {
	sa := c.accessStart()
	if sa == nil {
		return nil, fmt.Errorf("access failed")
	}

	dirOffset, _, _ := c.readOffsets()
	Log.Debugf("dirOffset: %v", dirOffset)

	vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirOffset]))
	Log.Debugf("vecLen: %v", vecLen)
	Log.Debugf("unsafe.Sizeof(statSegDirectoryEntry{}): %v", unsafe.Sizeof(statSegDirectoryEntry{}))

	for i := uint64(0); i < vecLen; i++ {
		offset := uintptr(i) * unsafe.Sizeof(statSegDirectoryEntry{})
		dirEntry := (*statSegDirectoryEntry)(add(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

		nul := bytes.IndexByte(dirEntry.name[:], '\x00')
		if nul < 0 {
			Log.Warnf("no zero byte found for: %q", dirEntry.name[:])
			continue
		}
		name := string(dirEntry.name[:nul])

		Log.Debugf(" %80q (type: %v, data: %d, offset: %d) ", name, dirEntry.directoryType, dirEntry.unionData, dirEntry.offsetVector)

		if nameMatches(name, patterns) {
			statNames = append(statNames, name)
		}

		// TODO: copy the listed entries elsewhere
	}

	if !c.accessEnd(sa) {
		return nil, adapter.ErrStatDirBusy
	}

	c.currentEpoch = sa.epoch

	return statNames, nil
}

func (c *StatsClient) DumpStats(patterns ...string) (entries []*adapter.StatEntry, err error) {
	epoch, _ := c.readEpoch()
	if c.currentEpoch > 0 && c.currentEpoch != epoch { // TODO: do list stats before dump
		return nil, fmt.Errorf("old data")
	}

	sa := c.accessStart()
	if sa == nil {
		return nil, fmt.Errorf("access failed")
	}

	dirOffset, _, _ := c.readOffsets()
	vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirOffset]))

	for i := uint64(0); i < vecLen; i++ {
		offset := uintptr(i) * unsafe.Sizeof(statSegDirectoryEntry{})
		dirEntry := (*statSegDirectoryEntry)(add(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

		entry := c.copyData(dirEntry)
		if nameMatches(entry.Name, patterns) {
			entries = append(entries, &entry)
		}
	}

	if !c.accessEnd(sa) {
		return nil, adapter.ErrStatDumpBusy
	}

	return entries, nil
}

func (c *StatsClient) copyData(dirEntry *statSegDirectoryEntry) (statEntry adapter.StatEntry) {
	name := dirEntry.name[:]
	if nul := bytes.IndexByte(name, '\x00'); nul < 0 {
		Log.Warnf("no zero byte found for: %q", dirEntry.name[:])
	} else {
		name = dirEntry.name[:nul]
	}

	statEntry.Name = string(name)
	statEntry.Type = adapter.StatType(dirEntry.directoryType)

	Log.Debugf(" - %s (type: %v, data: %v, offset: %v) ", statEntry.Name, statEntry.Type, dirEntry.unionData, dirEntry.offsetVector)

	switch statEntry.Type {
	case adapter.ScalarIndex:
		statEntry.Data = adapter.ScalarStat(dirEntry.unionData)

	case adapter.ErrorIndex:
		_, errOffset, _ := c.readOffsets()
		offsetVector := unsafe.Pointer(&c.sharedHeader[errOffset])
		vecLen := vectorLen(offsetVector)

		var errData adapter.Counter
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(add(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			offset := uintptr(cb) + uintptr(dirEntry.unionData)*unsafe.Sizeof(adapter.Counter(0))
			val := *(*adapter.Counter)(add(unsafe.Pointer(&c.sharedHeader[0]), offset))
			errData += val
		}
		statEntry.Data = adapter.ErrorStat(errData)

	case adapter.SimpleCounterVector:
		if dirEntry.unionData == 0 {
			Log.Debugf("\toffset is not valid")
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			Log.Debugf("\toffset out of range")
			break
		}

		simpleCounter := unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]) // offset
		vecLen := vectorLen(simpleCounter)
		offsetVector := add(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.Counter, vecLen)
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(add(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := vectorLen(counterVec)
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(add(counterVec, offset))
				data[i] = append(data[i], val)
			}
		}
		statEntry.Data = adapter.SimpleCounterStat(data)

	case adapter.CombinedCounterVector:
		if dirEntry.unionData == 0 {
			Log.Debugf("\toffset is not valid")
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			Log.Debugf("\toffset out of range")
			break
		}

		combinedCounter := unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]) // offset
		vecLen := vectorLen(combinedCounter)
		offsetVector := add(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.CombinedCounter, vecLen)
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(add(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := vectorLen(counterVec)
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				val := *(*adapter.CombinedCounter)(add(counterVec, offset))
				data[i] = append(data[i], val)
			}
		}
		statEntry.Data = adapter.CombinedCounterStat(data)

	case adapter.NameVector:
		if dirEntry.unionData == 0 {
			Log.Debugf("\toffset is not valid")
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			Log.Debugf("\toffset out of range")
			break
		}

		nameVector := unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]) // offset
		vecLen := vectorLen(nameVector)
		offsetVector := add(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([]adapter.Name, vecLen)
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(add(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			nameVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := vectorLen(nameVec)

			var nameStr []byte
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				val := *(*byte)(add(nameVec, offset))
				if val > 0 {
					nameStr = append(nameStr, val)
				}
			}
			data[i] = adapter.Name(nameStr)
		}
		statEntry.Data = adapter.NameStat(data)

	default:
		Log.Warnf("Unknown type %d for stat entry: %s", statEntry.Type, statEntry.Name)
	}

	Log.Debugf("\tentry data: %#v", statEntry.Data)

	return statEntry
}

type statDirectoryType int32

func (t statDirectoryType) String() string {
	return adapter.StatType(t).String()
}

type statSegDirectoryEntry struct {
	directoryType statDirectoryType
	// unionData can represent: offset, index or value
	unionData    uint64
	offsetVector uint64
	name         [128]byte
}

type statSegSharedHeader struct {
	version         uint64
	epoch           int64
	inProgress      int64
	directoryOffset int64
	errorOffset     int64
	statsOffset     int64
}

func (c *StatsClient) readVersion() uint64 {
	header := *(*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	version := atomic.LoadUint64(&header.version)
	return version
}

func (c *StatsClient) readEpoch() (int64, bool) {
	header := *(*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	epoch := atomic.LoadInt64(&header.epoch)
	inprog := atomic.LoadInt64(&header.inProgress)
	return epoch, inprog != 0
}

func (c *StatsClient) readOffsets() (dir, err, stat int64) {
	header := *(*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	dirOffset := atomic.LoadInt64(&header.directoryOffset)
	errOffset := atomic.LoadInt64(&header.errorOffset)
	statOffset := atomic.LoadInt64(&header.statsOffset)
	return dirOffset, errOffset, statOffset
}

type statSegAccess struct {
	epoch int64
}

var maxWaitInProgress = 1 * time.Second

func (c *StatsClient) accessStart() *statSegAccess {
	epoch, inprog := c.readEpoch()
	t := time.Now()
	for inprog {
		if time.Since(t) > maxWaitInProgress {
			return nil
		}
		epoch, inprog = c.readEpoch()
	}
	return &statSegAccess{
		epoch: epoch,
	}
}

func (c *StatsClient) accessEnd(acc *statSegAccess) bool {
	epoch, inprog := c.readEpoch()
	if acc.epoch != epoch || inprog {
		return false
	}
	return true
}

type vecHeader struct {
	length     uint64
	vectorData [0]uint8
}

func vectorLen(v unsafe.Pointer) uint64 {
	vec := *(*vecHeader)(unsafe.Pointer(uintptr(v) - unsafe.Sizeof(uintptr(0))))
	return vec.length
}

//go:nosplit
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}
