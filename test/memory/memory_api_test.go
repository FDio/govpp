//  Copyright (c) 2024 Cisco and/or its affiliates.
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

package memory

import (
	"context"
	"flag"
	"fmt"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/test/vpptesting"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"runtime"
	"runtime/metrics"
	"strings"
	"testing"
	"time"
)

var apiNum = flag.Uint("api-num", 0, "Custom number API to test")

// required go runtime metrics
const (
	heapBytesAllocs  = "/gc/heap/allocs:bytes"
	heapBytesFrees   = "/gc/heap/frees:bytes"
	heapObjectAllocs = "/gc/heap/allocs:objects"
	heapObjectFrees  = "/gc/heap/frees:objects"
	heapObjects      = "/gc/heap/objects:objects"
	goroutines       = "/sched/goroutines:goroutines"
)

// TestAPIMemory tests the GoVPP memory consumption for various numbers of API calls
func BenchmarkAPIMemory(b *testing.B) {
	flag.Parse()

	fmt.Printf("Running GoVPP API calls memory test\n\n")
	test := vpptesting.SetupVPP(b)
	vpeRPC := vpe.NewServiceClient(test.Conn)
	ifRPC := interfaces.NewServiceClient(test.Conn)

	samples := []metrics.Sample{
		{Name: heapBytesAllocs},
		{Name: heapBytesFrees},
		{Name: heapObjectAllocs},
		{Name: heapObjectFrees},
		{Name: heapObjects},
		{Name: goroutines},
	}

	shVerApiFunc := func() {
		// called twice to keep the number of API calls per func the same
		if _, err := vpeRPC.ShowVersion(context.Background(), &vpe.ShowVersion{}); err != nil {
			b.Fatal("calling show version failed:", err)
		}
		if _, err := vpeRPC.ShowVersion(context.Background(), &vpe.ShowVersion{}); err != nil {
			b.Fatal("calling show version failed:", err)
		}
	}
	loopApiFunc := func() {
		if reply, err := ifRPC.CreateLoopback(context.Background(), &interfaces.CreateLoopback{}); err != nil {
			b.Fatal("calling create loopback failed:", err)
		} else if _, err = ifRPC.DeleteLoopback(context.Background(), &interfaces.DeleteLoopback{
			SwIfIndex: reply.SwIfIndex,
		}); err != nil {
			b.Fatal("calling delete loopback failed:", err)
		}
	}
	shVerApiName, loopApiName := "show-version", "create/delete loopback"

	// run the custom soak only
	if *apiNum != 0 {
		testAPICalls(shVerApiName, *apiNum, &memMetrics{}, samples, shVerApiFunc)
		testAPICalls(loopApiName, *apiNum, &memMetrics{}, samples, loopApiFunc)
		return
	}

	// threshold values for the 'show version' (m0) or 'loopback create/delete' (m1) API call for 1k, 10k, 100k
	// and 1M number of repeats.
	m0 := []*memMetrics{
		{totalAllocMax: 2621440, heapAllocMax: 3145728, objectRemainMax: 50000},
		{totalAllocMax: 26214400, heapAllocMax: 3145728, objectRemainMax: 50000},
		{totalAllocMax: 262144000, heapAllocMax: 5242880, objectRemainMax: 50000},
		{totalAllocMax: 2684364560, heapAllocMax: 5242880, objectRemainMax: 50000},
	}
	m1 := []*memMetrics{
		{totalAllocMax: 3145728, heapAllocMax: 3670016, objectRemainMax: 50000},
		{totalAllocMax: 36700160, heapAllocMax: 3670016, objectRemainMax: 50000},
		{totalAllocMax: 367001600, heapAllocMax: 5242880, objectRemainMax: 50000},
		{totalAllocMax: 3758110384, heapAllocMax: 5242880, objectRemainMax: 50000},
	}

	pass := true
	for i, repeats := range []uint{1000, 10000, 100000, 1000000} {
		if passed := testAPICalls(shVerApiName, repeats, m0[i], samples, shVerApiFunc); !passed {
			pass = false
		}
		if passed := testAPICalls(loopApiName, repeats, m1[i], samples, loopApiFunc); !passed {
			pass = false
		}
	}
	if !pass {
		b.Fatal("one or more memory thresholds was exceeded")
	}
}

func testAPICalls(name string, repeats uint, m *memMetrics, samples []metrics.Sample, f func()) (pass bool) {
	now := time.Now()

	fmt.Printf("For %d %s calls:\n\t\tMeasured\tThreshold\n%s\n", repeats, name, strings.Repeat("-", 41))

	runtime.GC()
	metricsBefore := &memMetrics{}
	metricsBefore.readMetrics(samples)
	for i := 0; i < int(repeats); i++ {
		f()
	}
	m.readMetrics(samples)
	m.d = time.Since(now)
	m.rps = int(float64(repeats) / m.d.Seconds())

	defer m.print()
	return m.diff(metricsBefore)
}

// memMetrics is a list of metrics relevant to the memory test. Metrics suffixed with 'Max' can be pre-defined
// to serve as test criteria (these are not filled during 'readMetrics()').
type memMetrics struct {
	totalAlloc      uint64
	totalAllocMax   uint64
	memoryFreed     uint64
	heapAlloc       uint64
	heapAllocMax    uint64
	objectAlloc     uint64
	objectFreed     uint64
	objectRemain    uint64
	objectRemainMax uint64
	goroutines      int
	rps             int
	d               time.Duration
}

// readMetrics populates memory metrics
func (m *memMetrics) readMetrics(s []metrics.Sample) {
	metrics.Read(s)
	for _, sample := range s {
		switch sample.Name {
		case heapBytesAllocs:
			m.totalAlloc = sample.Value.Uint64()
		case heapBytesFrees:
			m.memoryFreed = sample.Value.Uint64()
		case heapObjectAllocs:
			m.objectAlloc = sample.Value.Uint64()
		case heapObjectFrees:
			m.objectFreed = sample.Value.Uint64()
		case heapObjects:
			m.objectRemain = sample.Value.Uint64()
		case goroutines:
			m.goroutines = int(sample.Value.Uint64())
		}
	}
	m.heapAlloc = m.totalAlloc - m.memoryFreed
}

// compares metrics with another metric snapshot taken earlier. Calculates difference between those
// and for selected metrics evaluates pass/fail criteria
func (m *memMetrics) diff(before *memMetrics) (pass bool) {
	pass = true
	m.totalAlloc -= before.totalAlloc
	if m.totalAlloc > m.totalAllocMax {
		pass = false
	}
	m.memoryFreed -= before.memoryFreed
	m.heapAlloc -= before.heapAlloc
	if m.heapAlloc > m.heapAllocMax {
		pass = false
	}
	m.objectAlloc -= before.objectAlloc
	m.objectFreed -= before.objectFreed
	m.objectRemain -= before.objectRemain
	if m.objectRemain > m.objectRemainMax {
		pass = false
	}
	return
}

func (m *memMetrics) print() {
	p := message.NewPrinter(language.English)
	var totalAllocMax, heapAllocMax, objectReaminMax string
	// if thresholds are set, use them in the output
	if !(m.totalAllocMax == 0 && m.heapAllocMax == 0 && m.objectRemainMax == 0) {
		totalAllocMax = format(m.totalAllocMax)
		heapAllocMax = format(m.heapAllocMax)
		objectReaminMax = p.Sprintf("%d", m.objectRemainMax)
	}
	fmt.Printf("Total alloc:\t%s\t%s\n", format(m.totalAlloc), totalAllocMax)
	fmt.Printf("Memory Freed:\t%s\n", format(m.objectFreed))
	fmt.Printf("Heap alloc:\t%s\t%s\n", format(m.heapAlloc), heapAllocMax)
	fmt.Printf("Objects alloc:\t%s\nObj freed:\t%s\n", p.Sprintf("%d", m.objectAlloc),
		p.Sprintf("%d", m.objectFreed))
	fmt.Printf("Objects remain:\t%s\t\t%s\n", p.Sprintf("%d", m.objectRemain), objectReaminMax)
	fmt.Printf("Message rate:\t%d m/s\n", m.rps)
	fmt.Printf("Num goroutines:\t%d\nDuration:\t%s\n\n", m.goroutines, m.d.String())
}

// shortens the number and adds unit
func format(v uint64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota)
		MB
		GB
		TB
		PB
	)

	var unit string
	value := float64(v)

	switch {
	case value < KB:
		unit = "B"
	case value < MB:
		unit = "KB"
		value /= KB
	case value < GB:
		unit = "MB"
		value /= MB
	case value < TB:
		unit = "GB"
		value /= GB
	case value < PB:
		unit = "TB"
		value /= TB
	default:
		unit = "PB"
		value /= PB
	}
	return align(fmt.Sprintf("%.2f %s", value, unit))
}

// align the table
func align(s string) string {
	if len(s) < 8 {
		return s + strings.Repeat(" ", 8-len(s))
	}
	return s
}
