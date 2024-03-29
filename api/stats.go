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

package api

// StatsProvider provides methods for retrieving statistics.
type StatsProvider interface {
	GetSystemStats(*SystemStats) error
	GetNodeStats(*NodeStats) error
	GetInterfaceStats(*InterfaceStats) error
	GetErrorStats(*ErrorStats) error
	GetBufferStats(*BufferStats) error
	GetMemoryStats(*MemoryStats) error
}

// SystemStats represents global system statistics.
type SystemStats struct {
	VectorRate          uint64
	NumWorkerThreads    uint64
	VectorRatePerWorker []uint64
	InputRate           uint64
	LastUpdate          uint64
	LastStatsClear      uint64
	Heartbeat           uint64
}

// NodeStats represents per node statistics.
type NodeStats struct {
	Nodes []NodeCounters
}

// NodeCounters represents node counters.
type NodeCounters struct {
	NodeIndex uint32
	NodeName  string // requires VPP 19.04+

	Clocks   uint64
	Vectors  uint64
	Calls    uint64
	Suspends uint64
}

// InterfaceStats represents per interface statistics.
type InterfaceStats struct {
	Interfaces []InterfaceCounters
}

// InterfaceCounters represents interface counters.
type InterfaceCounters struct {
	InterfaceIndex uint32
	InterfaceName  string // requires VPP 19.04+

	Rx InterfaceCounterCombined
	Tx InterfaceCounterCombined

	RxErrors uint64
	TxErrors uint64

	RxUnicast   InterfaceCounterCombined
	RxMulticast InterfaceCounterCombined
	RxBroadcast InterfaceCounterCombined
	TxUnicast   InterfaceCounterCombined
	TxMulticast InterfaceCounterCombined
	TxBroadcast InterfaceCounterCombined

	Drops   uint64
	Punts   uint64
	IP4     uint64
	IP6     uint64
	RxNoBuf uint64
	RxMiss  uint64
	Mpls    uint64
}

// InterfaceCounterCombined defines combined counters for interfaces.
type InterfaceCounterCombined struct {
	Packets uint64
	Bytes   uint64
}

// ErrorStats represents statistics per error counter.
type ErrorStats struct {
	Errors []ErrorCounter
}

// ErrorCounter represents error counter.
type ErrorCounter struct {
	CounterName string

	Values []uint64
}

// BufferStats represents statistics per buffer pool.
type BufferStats struct {
	Buffer map[string]BufferPool
}

// BufferPool represents buffer pool.
type BufferPool struct {
	PoolName string

	Cached    float64
	Used      float64
	Available float64
}

// MemoryStats represents memory stats segment counters.
type MemoryStats struct {
	// Deprecated: /mem/statseg total memory directory
	Total float64
	// Deprecated: /mem/statseg used memory directory
	Used float64

	// stat/main memory usage per-heap
	Stat map[int]MemoryCounters
	Main map[int]MemoryCounters
}

// MemoryCounters represents values of various memory usage
type MemoryCounters struct {
	Total      uint64
	Used       uint64
	Free       uint64
	UsedMMap   uint64
	TotalAlloc uint64
	FreeChunks uint64
	Releasable uint64
}
