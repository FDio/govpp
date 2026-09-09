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

package adapter

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// DefaultStatsSocket defines a default socket file path for VPP stats API.
	DefaultStatsSocket = "/run/vpp/stats.sock"
)

var (
	ErrStatsDataBusy     = errors.New("stats data busy")
	ErrStatsDirStale     = errors.New("stats dir stale")
	ErrStatsDisconnected = errors.New("stats disconnected")
	ErrStatsAccessFailed = errors.New("stats access failed")
)

// StatsAPI provides connection to VPP stats API.
type StatsAPI interface {
	// Connect establishes client connection to the stats API.
	Connect() error
	// Disconnect terminates client connection.
	Disconnect() error

	// ListStats lists indexed names for stats matching patterns.
	ListStats(patterns ...string) (indexes []StatIdentifier, err error)
	// DumpStats dumps all stat entries.
	DumpStats(patterns ...string) (entries []StatEntry, err error)

	// PrepareDir prepares new stat dir for entries that match any of prefixes.
	PrepareDir(patterns ...string) (*StatDir, error)
	// PrepareDirOnIndex prepares new stat dir for entries that match any of indexes.
	PrepareDirOnIndex(indexes ...uint32) (*StatDir, error)
	// UpdateDir updates stat dir and all of their entries.
	UpdateDir(dir *StatDir) error
}

// StatType represents type of stat directory and simply
// defines what type of stat data is stored in the stat entry.
type StatType string

const (
	Unknown               StatType = "UnknownStatType"
	ScalarIndex           StatType = "ScalarIndex"
	SimpleCounterVector   StatType = "SimpleCounterVector"
	CombinedCounterVector StatType = "CombinedCounterVector"
	ErrorIndex            StatType = "ErrorIndex"
	NameVector            StatType = "NameVector"
	Empty                 StatType = "Empty"
	Symlink               StatType = "Symlink"
	GaugeIndex            StatType = "GaugeIndex"
	HistogramLog2         StatType = "HistogramLog2"
	RingBuffer            StatType = "RingBuffer"
)

// StatDir defines directory of stats entries created by PrepareDir.
type StatDir struct {
	Epoch   int64
	Entries []StatEntry
}

// StatIdentifier holds a stat entry name and index
type StatIdentifier struct {
	Index uint32
	Name  []byte
}

// StatEntry represents single stat entry. The type of stat stored in Data
// is defined by Type.
type StatEntry struct {
	StatIdentifier
	Type    StatType
	Data    Stat
	Symlink bool
}

// Counter represents simple counter with single value, which is usually packet count.
type Counter uint64

// CombinedCounter represents counter with two values, for packet count and bytes count.
type CombinedCounter [2]uint64

func (s CombinedCounter) Packets() uint64 {
	return s[0]
}

func (s CombinedCounter) Bytes() uint64 {
	return s[1]
}

// Name represents string value stored under name vector.
type Name []byte

func (n Name) String() string {
	return string(n)
}

// Stat represents some type of stat which is usually defined by StatType.
type Stat interface {
	// IsZero returns true if all of its values equal to zero.
	IsZero() bool

	// Type returns underlying type of a stat
	Type() StatType

	// isStat is intentionally  unexported to limit implementations of interface to this package,
	isStat()
}

// ScalarStat represents stat for ScalarIndex.
type ScalarStat float64

// ScalarStat represents stat for GaugeIndex.
type GaugeStat float64

// ErrorStat represents stat for ErrorIndex. The array represents workers.
type ErrorStat []Counter

// SimpleCounterStat represents indexed stat for SimpleCounterVector.
// The outer array represents workers and the inner array represents interface/node/.. indexes.
// Values should be aggregated per interface/node for every worker.
// ReduceSimpleCounterStatIndex can be used to reduce specific index.
type SimpleCounterStat [][]Counter

// CombinedCounterStat represents indexed stat for CombinedCounterVector.
// The outer array represents workers and the inner array represents interface/node/.. indexes.
// Values should be aggregated per interface/node for every worker.
// ReduceCombinedCounterStatIndex can be used to reduce specific index.
type CombinedCounterStat [][]CombinedCounter

// NameStat represents stat for NameVector.
type NameStat []Name

// EmptyStat represents removed counter directory
type EmptyStat string

// HistogramLog2Bin represents histogram data.
// Counts[j] represents the histogram bucket with values in range [2^(MinExp + j), 2^(MinExp + j + 1) - 1].
type HistogramLog2Bin struct {
	MinExp uint64
	Counts []uint64
}

// HistogramLog2Stat represents stat for a log2 histogram. The array represents workers.
type HistogramLog2Stat []HistogramLog2Bin

// RingBufferConfig holds the configuration for a ring buffer.
type RingBufferConfig struct {
	EntrySize     uint32
	RingSize      uint32
	NThreads      uint32
	SchemaSize    uint32
	SchemaVersion uint32
}

// RingBufferThreadMeta holds per-thread metadata for a ring buffer.
type RingBufferThreadMeta struct {
	Head          uint32
	SchemaVersion uint32
	Sequence      uint64
	SchemaOffset  uint32
	SchemaSize    uint32
}

// RingBufferStat represents a snapshot of a ring buffer's configuration, per-thread metadata,
// and raw ring data.
type RingBufferStat struct {
	Config  RingBufferConfig
	Threads []RingBufferThreadMeta
	Schema  []byte
	Data    [][]byte // per-thread raw ring data
}

// RingBufferAPI is implemented by adapters that can read a ring-buffer stat
// incrementally, and is separate from StatsAPI because not every adapter can:
// serving a window means carrying a read cursor between refreshes. That cursor
// is consumer-side state - the segment itself is mapped read-only - and a mock
// or a v1 segment has nowhere to keep one. Callers type-assert for it.
//
// core.StatsConnection holds its StatsAPI unexported and offers no accessor, so
// this is reached through a *statsclient.StatsClient the caller owns.
type RingBufferAPI interface {
	// PrepareRingBuffer resolves one ring-buffer stat by name and returns a
	// StatDir holding an incremental reader for it, to be refreshed with
	// UpdateDir. maxEntries bounds entries delivered per thread per refresh, zero
	// meaning the ring size; skipBacklog starts at the producer's head rather
	// than at the oldest entry the ring still holds.
	PrepareRingBuffer(name string, maxEntries uint32, skipBacklog bool) (*StatDir, error)
}

// RingBufferWindow is the run of entries one producer thread appended since a
// consumer last read it.
type RingBufferWindow struct {
	// Entries holds Count entries of RingBufferConfig.EntrySize bytes each,
	// oldest first and already unwrapped, so a consumer indexes it without
	// knowing where the ring's seam fell.
	//
	// It aliases a buffer the stat owns and reuses, so it is only valid until
	// the next refresh. Copy anything that must outlive that.
	//
	// The slice always starts at the buffer's origin and is capped to the entries
	// this refresh delivered, so cap(Entries) is the room a refresh has to fill
	// and there is no second buffer field to keep in step with it.
	Entries []byte
	Count   uint32

	// FirstSeq is the producer sequence of Entries[0], and NextSeq the sequence
	// the following read will start from. Both are absolute counts of entries
	// this thread has ever written, so they stay meaningful across wraps.
	FirstSeq uint64
	NextSeq  uint64

	// Lost counts entries the producer overwrote before this read reached them:
	// data that is gone. Pending counts entries still in the ring that this read
	// did not return because MaxEntries capped it: data that the next read will
	// deliver.
	//
	// They are separate because they call for opposite responses. Pending means
	// read again immediately; Lost means the reader is not keeping up and the
	// gap is unrecoverable. A single "missed" figure would conflate a reader
	// that is behind with one that is merely rate-limited.
	Lost    uint64
	Pending uint64
}

// RingBufferWindowStat reads a ring buffer incrementally: every refresh copies
// only what the producers appended since the previous one, into buffers the stat
// already owns.
//
// This is the difference between a cost proportional to entries produced and one
// proportional to ring size. RingBufferStat copies the whole ring, for every
// thread, into a fresh allocation on every read - so a ring sized for burst
// headroom rather than for poll latency becomes unreadable long before it
// becomes useful. A 16M-entry ring of 128-byte records is 2 GiB per thread per
// read as a RingBufferStat, and the entries actually produced as this.
//
// Put one in a prepared StatDir entry's Data and refresh it with
// StatsClient.UpdateDir, or let StatsClient.PrepareRingBuffer build both.
// CopyEntryData never produces one: windowing needs a read cursor, and only the
// consumer has it.
//
// The zero value is valid and self-initialising. The first refresh reads the
// geometry, allocates the per-thread buffers, and positions the cursor - at the
// oldest entry the ring still holds, or at the producer's head if SkipBacklog is
// set - and returns no entries. SkipBacklog only decides where that first
// refresh starts and is ignored afterwards; MaxEntries is honoured on every
// refresh.
type RingBufferWindowStat struct {
	// MaxEntries bounds how many entries one refresh delivers per thread, and so
	// bounds both the buffer this stat allocates and the work one refresh does.
	// Zero means the ring size, which is the largest window that can ever be
	// available. A consumer draining a fast producer wants this small enough to
	// bound a single read and to loop while Pending is non-zero.
	//
	// It is read on every refresh, so raising or lowering it between refreshes
	// takes effect on the next one; the buffers grow to match and are not shrunk.
	MaxEntries uint32

	// SkipBacklog starts the first read at the producer's head rather than at the
	// oldest entry still in the ring, so a consumer that wants only what happens
	// from now on does not first have to read and discard a ring of history.
	SkipBacklog bool

	Config  RingBufferConfig
	Threads []RingBufferThreadMeta
	Schema  []byte

	// Windows holds one window per producer thread, in thread order.
	Windows []RingBufferWindow
}

func (ScalarStat) isStat()          {}
func (ErrorStat) isStat()           {}
func (SimpleCounterStat) isStat()   {}
func (CombinedCounterStat) isStat() {}
func (NameStat) isStat()            {}
func (EmptyStat) isStat()           {}
func (GaugeStat) isStat()           {}
func (HistogramLog2Stat) isStat()   {}
func (RingBufferStat) isStat()      {}

// Pointer receiver: refreshed in place, so a value must not satisfy Stat.
func (*RingBufferWindowStat) isStat() {}

func (s ScalarStat) IsZero() bool {
	return s == 0
}

func (s ScalarStat) Type() StatType {
	return ScalarIndex
}

func (s ErrorStat) IsZero() bool {
	if s == nil {
		return true
	}
	for _, ss := range s {
		if ss != 0 {
			return false
		}
	}
	return true
}

func (s ErrorStat) Type() StatType {
	return ErrorIndex
}

func (s SimpleCounterStat) IsZero() bool {
	if s == nil {
		return true
	}
	for _, ss := range s {
		for _, sss := range ss {
			if sss != 0 {
				return false
			}
		}
	}
	return true
}

func (s SimpleCounterStat) Type() StatType {
	return SimpleCounterVector
}

func (s CombinedCounterStat) IsZero() bool {
	if s == nil {
		return true
	}
	for _, ss := range s {
		if ss == nil {
			return true
		}
		for _, sss := range ss {
			if sss[0] != 0 || sss[1] != 0 {
				return false
			}
		}
	}
	return true
}

func (s CombinedCounterStat) Type() StatType {
	return CombinedCounterVector
}

func (s NameStat) IsZero() bool {
	if s == nil {
		return true
	}
	for _, ss := range s {
		if len(ss) > 0 {
			return false
		}
	}
	return true
}

func (s NameStat) Type() StatType {
	return NameVector
}

func (s EmptyStat) IsZero() bool {
	return true
}

func (s EmptyStat) Type() StatType {
	return Empty
}

// ReduceSimpleCounterStatIndex returns reduced SimpleCounterStat s for index i.
func ReduceSimpleCounterStatIndex(s SimpleCounterStat, i int) uint64 {
	var val uint64
	for _, w := range s {
		val += uint64(w[i])
	}
	return val
}

// ReduceCombinedCounterStatIndex returns reduced CombinedCounterStat s for index i.
func ReduceCombinedCounterStatIndex(s CombinedCounterStat, i int) [2]uint64 {
	var val [2]uint64
	for _, w := range s {
		val[0] += w[i][0]
		val[1] += w[i][1]
	}
	return val
}

func (s GaugeStat) IsZero() bool {
	return s == 0
}

func (s GaugeStat) Type() StatType {
	return GaugeIndex
}

func (s HistogramLog2Stat) IsZero() bool {
	if s == nil {
		return true
	}
	for _, bins := range s {
		for _, c := range bins.Counts {
			if c != 0 {
				return false
			}
		}
	}
	return true
}

func (s HistogramLog2Stat) Type() StatType {
	return HistogramLog2
}

func (s HistogramLog2Stat) String() string {
	var b strings.Builder
	for i, bin := range s {
		fmt.Fprintf(&b, "\n  [thread %d]: min_exp=%d", i, bin.MinExp)
		var cumulative uint64
		for j, count := range bin.Counts {
			cumulative += count
			fmt.Fprintf(&b, "\n    <= %d: %d (cumulative: %d)",
				uint64(1)<<(bin.MinExp+uint64(j)), count, cumulative)
		}
	}
	return b.String()
}

func (s RingBufferStat) IsZero() bool {
	return s.Config.NThreads == 0 || s.Config.EntrySize == 0
}

func (s RingBufferStat) Type() StatType {
	return RingBuffer
}

func (s *RingBufferWindowStat) IsZero() bool {
	return s.Config.NThreads == 0 || s.Config.EntrySize == 0
}

func (s *RingBufferWindowStat) Type() StatType {
	return RingBuffer
}

func (s *RingBufferWindowStat) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n  config: entry_size=%d, ring_size=%d, threads=%d, schema_version=%d, schema_size=%d",
		s.Config.EntrySize, s.Config.RingSize, s.Config.NThreads, s.Config.SchemaVersion, s.Config.SchemaSize)
	for i, w := range s.Windows {
		fmt.Fprintf(&b, "\n  thread[%d]: entries=%d first_seq=%d next_seq=%d lost=%d pending=%d",
			i, w.Count, w.FirstSeq, w.NextSeq, w.Lost, w.Pending)
	}
	return b.String()
}

func (s RingBufferStat) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n  config: entry_size=%d, ring_size=%d, threads=%d, schema_version=%d, schema_size=%d",
		s.Config.EntrySize, s.Config.RingSize, s.Config.NThreads, s.Config.SchemaVersion, s.Config.SchemaSize)

	for i, t := range s.Threads {
		fmt.Fprintf(&b, "\n  thread[%d]: head=%d seq=%d schema_version=%d",
			i, t.Head, t.Sequence, t.SchemaVersion)
	}

	return b.String()
}
