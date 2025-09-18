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

// Package statsclient is pure Go implementation of VPP stats API client.
package statsclient

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ftrvxmtrx/fd"
	logger "github.com/sirupsen/logrus"
	"go.fd.io/govpp/adapter"
)

const (
	// DefaultSocketName is default VPP stats socket file path.
	DefaultSocketName = adapter.DefaultStatsSocket

	// DefaultSocketRetryPeriod is the time period after the socket availability
	// will be re-checked
	DefaultSocketRetryPeriod = 50 * time.Millisecond

	// DefaultSocketRetryTimeout is the maximum time for the stats socket
	DefaultSocketRetryTimeout = 3 * time.Second
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
		Log.Debug("govpp/statsclient: enabled debug mode")
	}
}

func debugf(f string, a ...interface{}) {
	if Debug {
		Log.Debugf(f, a...)
	}
}

// implements StatsAPI
var _ adapter.StatsAPI = (*StatsClient)(nil)

// StatsClient is the pure Go implementation for VPP stats API.
type StatsClient struct {
	socket       string
	retryPeriod  time.Duration
	retryTimeout time.Duration

	headerData []byte

	// defines the adapter connection state
	connected uint32
	// defines the adapter monitoring loop state
	monitored uint32

	// to quit socket monitor
	done chan struct{}

	// to protect statseg access from concurrent reconnect
	accessLock sync.RWMutex

	statSegment
}

// Option is a StatsClient option
type Option func(*StatsClient)

// SetSocketRetryPeriod is and optional parameter to define a custom
// retry period while waiting for the VPP socket
func SetSocketRetryPeriod(t time.Duration) Option {
	return func(c *StatsClient) {
		c.retryPeriod = t
	}
}

// SetSocketRetryTimeout is and optional parameter to define a custom
// timeout while waiting for the VPP socket
func SetSocketRetryTimeout(t time.Duration) Option {
	return func(c *StatsClient) {
		c.retryTimeout = t
	}
}

// NewStatsClient returns a new StatsClient using socket.
// If socket is empty string DefaultSocketName is used.
func NewStatsClient(socket string, options ...Option) *StatsClient {
	if socket == "" {
		socket = DefaultSocketName
	}
	s := &StatsClient{
		socket: socket,
	}
	for _, option := range options {
		option(s)
	}
	if s.retryPeriod == 0 {
		s.retryPeriod = DefaultSocketRetryPeriod
	}
	if s.retryTimeout == 0 {
		s.retryTimeout = DefaultSocketRetryTimeout
	}
	return s
}

// Connect to validated VPP stats socket and start monitoring
// socket file changes
func (sc *StatsClient) Connect() (err error) {
	if err := sc.waitForSocket(); err != nil {
		return err
	}
	sc.done = make(chan struct{})
	sc.accessLock.Lock()
	defer sc.accessLock.Unlock()
	if sc.statSegment, err = sc.connect(); err != nil {
		return err
	}
	sc.monitorSocket()
	return nil
}

// Disconnect from the socket, unmap shared memory and terminate
// socket monitor
func (sc *StatsClient) Disconnect() error {
	if atomic.CompareAndSwapUint32(&sc.monitored, 1, 0) {
		close(sc.done)
	}

	if !sc.isConnected() {
		return nil
	}

	return sc.disconnect()
}

func (sc *StatsClient) ListStats(patterns ...string) (entries []adapter.StatIdentifier, err error) {
	sc.accessLock.RLock()
	defer sc.accessLock.RUnlock()

	if !sc.isConnected() {
		return nil, adapter.ErrStatsDisconnected
	}
	accessEpoch := sc.accessStart()
	if accessEpoch == 0 {
		return nil, adapter.ErrStatsAccessFailed
	}

	entries, err = sc.getIdentifierEntries(patterns...)
	if err != nil {
		return nil, err
	}

	if !sc.accessEnd(accessEpoch) {
		return nil, adapter.ErrStatsDataBusy
	}
	return entries, nil
}

func (sc *StatsClient) DumpStats(patterns ...string) (entries []adapter.StatEntry, err error) {
	sc.accessLock.RLock()
	defer sc.accessLock.RUnlock()

	if !sc.isConnected() {
		return nil, adapter.ErrStatsDisconnected
	}

	accessEpoch := sc.accessStart()
	if accessEpoch == 0 {
		return nil, adapter.ErrStatsAccessFailed
	}

	entries, err = sc.getStatEntries(patterns...)
	if err != nil {
		return nil, err
	}

	if !sc.accessEnd(accessEpoch) {
		return nil, adapter.ErrStatsDataBusy
	}
	return entries, nil
}

func (sc *StatsClient) PrepareDir(patterns ...string) (*adapter.StatDir, error) {
	sc.accessLock.RLock()
	defer sc.accessLock.RUnlock()

	if !sc.isConnected() {
		return nil, adapter.ErrStatsDisconnected
	}

	accessEpoch := sc.accessStart()
	if accessEpoch == 0 {
		return nil, adapter.ErrStatsAccessFailed
	}

	entries, err := sc.getStatEntries(patterns...)
	if err != nil {
		return nil, err
	}

	if !sc.accessEnd(accessEpoch) {
		return nil, adapter.ErrStatsDataBusy
	}

	dir := &adapter.StatDir{
		Epoch:   accessEpoch,
		Entries: entries,
	}

	return dir, nil
}

func (sc *StatsClient) PrepareDirOnIndex(indexes ...uint32) (*adapter.StatDir, error) {
	sc.accessLock.RLock()
	defer sc.accessLock.RUnlock()

	if !sc.isConnected() {
		return nil, adapter.ErrStatsDisconnected
	}

	accessEpoch := sc.accessStart()
	if accessEpoch == 0 {
		return nil, adapter.ErrStatsAccessFailed
	}
	vector := sc.GetDirectoryVector()
	if vector == nil {
		return nil, fmt.Errorf("failed to prepare dir on index: directory vector is nil")
	}
	entries, err := sc.getStatEntriesOnIndex(vector, indexes...)
	if err != nil {
		return nil, err
	}

	if !sc.accessEnd(accessEpoch) {
		return nil, adapter.ErrStatsDataBusy
	}

	dir := &adapter.StatDir{
		Epoch:   accessEpoch,
		Entries: entries,
	}

	return dir, nil
}

// UpdateDir refreshes directory data for all counters
func (sc *StatsClient) UpdateDir(dir *adapter.StatDir) (err error) {
	sc.accessLock.RLock()
	defer sc.accessLock.RUnlock()

	if !sc.isConnected() {
		return adapter.ErrStatsDisconnected
	}

	epoch, _ := sc.GetEpoch()
	if dir.Epoch != epoch {
		return adapter.ErrStatsDirStale
	}

	accessEpoch := sc.accessStart()
	if accessEpoch == 0 {
		return adapter.ErrStatsAccessFailed
	}
	dirVector := sc.GetDirectoryVector()
	if dirVector == nil {
		return err
	}
	for i := 0; i < len(dir.Entries); i++ {
		if err := sc.updateStatOnIndex(&dir.Entries[i], dirVector); err != nil {
			return err
		}
	}
	if !sc.accessEnd(accessEpoch) {
		return adapter.ErrStatsDataBusy
	}
	return nil
}

// checks the socket existence and waits for it for the designated
// time if it is not available immediately
func (sc *StatsClient) waitForSocket() error {
	if _, err := os.Stat(sc.socket); err != nil {
		if os.IsNotExist(err) {
			n := time.Now()
			ticker := time.NewTicker(sc.retryPeriod)
			timeout := time.After(sc.retryTimeout)
			for {
				select {
				case <-ticker.C:
					if _, err := os.Stat(sc.socket); err == nil {
						return nil
					}
				case <-timeout:
					return fmt.Errorf("stats socket file %s is not ready within timeout (after %.2f s) ",
						sc.socket, time.Since(n).Seconds())
				}
			}
		} else {
			return fmt.Errorf("stats socket error: %v", err)
		}
	}
	return nil
}

// connect to the socket and map it into the memory. According to the
// header version info, an appropriate segment handler is returned
func (sc *StatsClient) connect() (ss statSegment, err error) {
	addr := net.UnixAddr{
		Net:  "unixpacket",
		Name: sc.socket,
	}
	Log.Debugf("connecting to: %v", addr)

	conn, err := net.DialUnix(addr.Net, nil, &addr)
	if err != nil {
		Log.Warnf("connecting to socket %s failed: %s", addr, err)
		return nil, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			Log.Warnf("closing socket failed: %v", err)
		}
	}()
	Log.Debugf("connected to socket")

	files, err := fd.Get(conn, 1, nil)
	if err != nil {
		return nil, fmt.Errorf("getting file descriptor over socket failed: %v", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files received over socket")
	}

	file := files[0]
	defer func() {
		if err := file.Close(); err != nil {
			Log.Warnf("closing file failed: %v", err)
		}
	}()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := info.Size()

	sc.headerData, err = syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		Log.Debugf("mapping shared memory failed: %v", err)
		return nil, fmt.Errorf("mapping shared memory failed: %v", err)
	}
	Log.Debugf("successfully mmapped shared memory segment (size: %v) %v", size, len(sc.headerData))

	version := getVersion(sc.headerData)
	switch version {
	case 1:
		ss = newStatSegmentV1(sc.headerData, size)
	case 2:
		ss = newStatSegmentV2(sc.headerData, size)
	default:
		if err = syscall.Munmap(sc.headerData); err != nil {
			Log.Debugf("unmapping shared memory failed: %v", err)
		}
		return nil, fmt.Errorf("stat segment version is not supported: %v (min: %v, max: %v)",
			version, minVersion, maxVersion)
	}

	// set connected
	atomic.CompareAndSwapUint32(&sc.connected, 0, 1)

	return ss, nil
}

// reconnect disconnects from the socket, re-validates it and
// connects again
func (sc *StatsClient) reconnect() (err error) {
	sc.accessLock.Lock()
	defer sc.accessLock.Unlock()
	if err = sc.disconnect(); err != nil {
		return fmt.Errorf("error disconnecting socket: %v", err)
	}
	if err = sc.waitForSocket(); err != nil {
		return fmt.Errorf("error while waiting on socket: %v", err)
	}
	if sc.statSegment, err = sc.connect(); err != nil {
		return fmt.Errorf("error connecting socket: %v", err)
	}
	return nil
}

// disconnect unmaps socket data from the memory and resets the header
func (sc *StatsClient) disconnect() error {

	if !atomic.CompareAndSwapUint32(&sc.connected, 1, 0) {
		return fmt.Errorf("stats client is already disconnected")
	}
	if sc.headerData == nil {
		return nil
	}
	if err := syscall.Munmap(sc.headerData); err != nil {
		Log.Debugf("unmapping shared memory failed: %v", err)
		return fmt.Errorf("unmapping shared memory failed: %v", err)
	}
	sc.headerData = nil

	Log.Debugf("successfully unmapped shared memory")
	return nil
}

func (sc *StatsClient) monitorSocket() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Log.Errorf("error starting socket monitor: %v", err)
		return
	}

	atomic.StoreUint32(&sc.monitored, 1)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op == fsnotify.Remove && event.Name == sc.socket {
					if err := sc.reconnect(); err != nil {
						Log.Errorf("error occurred during socket reconnect: %v", err)
					}
				}
			case err := <-watcher.Errors:
				Log.Errorf("socket monitor delivered error event: %v", err)
			case <-sc.done:
				err := watcher.Close()
				Log.Debugf("socket monitor closed (error: %v)", err)
				return
			}
		}
	}()

	if err := watcher.Add(filepath.Dir(sc.socket)); err != nil {
		Log.Errorf("failed to add socket address to the watcher: %v", err)
	}
}

// Starts monitoring 'inProgress' field. Returns stats segment
// access epoch when completed, or zero value if not finished
// within MaxWaitInProgress
func (sc *StatsClient) accessStart() (epoch int64) {
	t := time.Now()

	epoch, inProg := sc.GetEpoch()
	for inProg {
		if time.Since(t) > MaxWaitInProgress {
			return int64(0)
		}
		time.Sleep(CheckDelayInProgress)
		epoch, inProg = sc.GetEpoch()
	}
	return epoch
}

// AccessEnd returns true if stats data reading was finished, false
// otherwise
func (sc *StatsClient) accessEnd(accessEpoch int64) bool {
	epoch, inProgress := sc.GetEpoch()
	if accessEpoch != epoch || inProgress {
		return false
	}
	return true
}

// getStatEntries retrieves all stats matching desired patterns, or all stats if no pattern is provided.
func (sc *StatsClient) getStatEntries(patterns ...string) (entries []adapter.StatEntry, err error) {
	vector := sc.GetDirectoryVector()
	if vector == nil {
		return nil, fmt.Errorf("failed to get stat entries: directory vector is nil")
	}
	indexes, err := sc.listIndexes(vector, patterns...)
	if err != nil {
		return nil, err
	}
	return sc.getStatEntriesOnIndex(vector, indexes...)
}

// getIdentifierEntries retrieves all identifiers matching desired patterns, or all identifiers
// if no pattern is provided.
func (sc *StatsClient) getIdentifierEntries(patterns ...string) (identifiers []adapter.StatIdentifier, err error) {
	vector := sc.GetDirectoryVector()
	if vector == nil {
		return nil, fmt.Errorf("failed to get identifier entries: directory vector is nil")
	}
	indexes, err := sc.listIndexes(vector, patterns...)
	if err != nil {
		return nil, err
	}
	return sc.getIdentifierEntriesOnIndex(vector, indexes...)
}

// getStatEntriesOnIndex retrieves stats on indexes, or all stats if indexes are not defined.
func (sc *StatsClient) getStatEntriesOnIndex(vector dirVector, indexes ...uint32) (entries []adapter.StatEntry, err error) {
	dirLen := *(*uint32)(vectorLen(vector))
	for _, index := range indexes {
		if index >= dirLen {
			return nil, fmt.Errorf("stat entry index %d out of dir vector length (%d)", index, dirLen)
		}
		dirPtr, dirName, dirType := sc.GetStatDirOnIndex(vector, index)
		if len(dirName) == 0 {
			return
		}
		var t adapter.StatType
		d := sc.CopyEntryData(dirPtr, ^uint32(0))
		if d != nil {
			t = d.Type()
		}
		entries = append(entries, adapter.StatEntry{
			StatIdentifier: adapter.StatIdentifier{
				Index: index,
				Name:  dirName,
			},
			Type:    t,
			Data:    d,
			Symlink: dirType == adapter.Symlink,
		})
	}
	return entries, nil
}

// getIdentifierEntriesOnIndex retrieves identifiers on indexes, or all identifiers if indexes are not defined.
func (sc *StatsClient) getIdentifierEntriesOnIndex(vector dirVector, indexes ...uint32) (identifiers []adapter.StatIdentifier, err error) {
	dirLen := *(*uint32)(vectorLen(vector))
	for _, index := range indexes {
		if index >= dirLen {
			return nil, fmt.Errorf("stat entry index %d out of dir vector length (%d)", index, dirLen)
		}
		_, dirName, _ := sc.GetStatDirOnIndex(vector, index)
		if len(dirName) == 0 {
			return
		}
		identifiers = append(identifiers, adapter.StatIdentifier{
			Index: index,
			Name:  dirName,
		})
	}
	return identifiers, nil
}

// listIndexes lists indexes for all stat entries that match any of the regex patterns.
func (sc *StatsClient) listIndexes(vector dirVector, patterns ...string) (indexes []uint32, err error) {
	if len(patterns) == 0 {
		return sc.listIndexesFunc(vector, nil)
	}
	var regexes = make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling regexp failed: %v", err)
		}
		regexes[i] = r
	}
	nameMatches := func(name dirName) bool {
		for _, r := range regexes {
			if r.Match(name) {
				return true
			}
		}
		return false
	}
	return sc.listIndexesFunc(vector, nameMatches)
}

// listIndexesFunc lists stats indexes. The optional function
// argument filters returned values or returns all if empty
func (sc *StatsClient) listIndexesFunc(vector dirVector, f func(name dirName) bool) (indexes []uint32, err error) {
	if f == nil {
		// there is around ~3157 stats, so to avoid too many allocations
		// we set capacity to 3200 when listing all stats
		indexes = make([]uint32, 0, 3200)
	}
	vecLen := *(*uint32)(vectorLen(vector))
	for i := uint32(0); i < vecLen; i++ {
		_, dirName, _ := sc.GetStatDirOnIndex(vector, i)
		if f != nil {
			if len(dirName) == 0 || !f(dirName) {
				continue
			}
		}
		indexes = append(indexes, i)
	}

	return indexes, nil
}

func (sc *StatsClient) isConnected() bool {
	return atomic.LoadUint32(&sc.connected) == 1
}

// updateStatOnIndex refreshes the entry data.
func (sc *StatsClient) updateStatOnIndex(entry *adapter.StatEntry, vector dirVector) (err error) {
	dirLen := *(*uint32)(vectorLen(vector))
	if entry.Index >= dirLen {
		return fmt.Errorf("stat entry index %d out of dir vector length (%d)", entry.Index, dirLen)
	}
	dirPtr, dirName, dirType := sc.GetStatDirOnIndex(vector, entry.Index)
	if len(dirName) == 0 ||
		!bytes.Equal(dirName, entry.Name) ||
		dirType != entry.Type ||
		entry.Data == nil {
		return nil
	}
	if err := sc.UpdateEntryData(dirPtr, &entry.Data); err != nil {
		return fmt.Errorf("updating stat data for entry %s failed: %v", dirName, err)
	}
	return
}
