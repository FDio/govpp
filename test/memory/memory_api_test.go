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
	"runtime"
	runtimeMetrics "runtime/metrics"
	"strings"
	"testing"
	"time"

	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/test/vpptesting"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var apiNum = flag.Uint("api-num", 0, "Custom number API to test")

// metric names
const (
	totalAlloc = "Total alloc"
	memFreed   = "Memory freed"
	heapAlloc  = "Heap alloc"
	objAlloc   = "Objects alloc"
	objFreed   = "Objects freed"
	objRemain  = "Objects remain"
	msgRate    = "Message rate"
	numGo      = "Num goroutines"
	duration   = "Duration"
)

// metric samples
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

	samples := []namedSample{
		{name: totalAlloc, sample: runtimeMetrics.Sample{Name: heapBytesAllocs}},
		{name: memFreed, sample: runtimeMetrics.Sample{Name: heapBytesFrees}},
		{name: objAlloc, sample: runtimeMetrics.Sample{Name: heapObjectAllocs}},
		{name: objFreed, sample: runtimeMetrics.Sample{Name: heapObjectFrees}},
		{name: objRemain, sample: runtimeMetrics.Sample{Name: heapObjects}},
		{name: numGo, sample: runtimeMetrics.Sample{Name: goroutines}},
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
	nameOrder := []string{totalAlloc, memFreed, heapAlloc, objAlloc, objFreed, objRemain, msgRate, numGo, duration}
	if *apiNum != 0 {
		testAPICalls(shVerApiName, *apiNum, &metrics{names: nameOrder}, samples, shVerApiFunc)
		testAPICalls(loopApiName, *apiNum, &metrics{names: nameOrder}, samples, loopApiFunc)
		return
	}

	// threshold values for the 'show version' (m0) or 'loopback create/delete' (m1) API call for 1k, 10k, 100k
	// and 1M number of repeats.
	m0 := []*metrics{
		{names: nameOrder, metricsByName: map[string]metric{
			totalAlloc: {max: 5767168},
			heapAlloc:  {max: 3145728},
			objRemain:  {max: 50000}},
		},
		{names: nameOrder, metricsByName: map[string]metric{
			totalAlloc: {max: 62914560},
			heapAlloc:  {max: 3145728},
			objRemain:  {max: 50000}},
		},
		{names: nameOrder, metricsByName: map[string]metric{
			totalAlloc: {max: 629145600},
			heapAlloc:  {max: 5242880},
			objRemain:  {max: 50000}},
		},
		{names: nameOrder, metricsByName: map[string]metric{
			totalAlloc: {max: 5368709120},
			heapAlloc:  {max: 5242880},
			objRemain:  {max: 50000}},
		},
	}
	m1 := []*metrics{
		{names: nameOrder, metricsByName: map[string]metric{
			totalAlloc: {max: 10485760},
			heapAlloc:  {max: 3670016},
			objRemain:  {max: 50000}}},
		{names: nameOrder, metricsByName: map[string]metric{
			totalAlloc: {max: 94371840},
			heapAlloc:  {max: 3670016},
			objRemain:  {max: 50000}}},
		{names: nameOrder, metricsByName: map[string]metric{
			totalAlloc: {max: 891289600},
			heapAlloc:  {max: 5242880},
			objRemain:  {max: 50000}}},
		{names: nameOrder, metricsByName: map[string]metric{
			totalAlloc: {max: 8589934592},
			heapAlloc:  {max: 5242880},
			objRemain:  {max: 50000}}},
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

func testAPICalls(name string, repeats uint, m *metrics, samples []namedSample, f func()) (pass bool) {
	now := time.Now()

	fmt.Printf("For %d %s calls:\n\t\tMeasured\tThreshold\n%s\n", repeats, name, strings.Repeat("-", 41))

	runtime.GC()
	metricsBefore := &metrics{}
	metricsBefore.readMetrics(samples)
	for i := 0; i < int(repeats); i++ {
		f()
	}
	m.readMetrics(samples)
	d := time.Since(now)
	m.metricsByName[duration] = metric{value: uint64(d)}
	m.metricsByName[msgRate] = metric{value: uint64(int(float64(repeats) / d.Seconds()))}

	defer m.print()
	return m.diff(metricsBefore)
}

type namedSample struct {
	name   string
	sample runtimeMetrics.Sample
}

type metric struct {
	value uint64
	max   uint64
}

// memMetrics is a list of metrics relevant to the memory test. Metrics suffixed with 'Max' can be pre-defined
// to serve as test criteria (these are not filled during 'readMetrics()').
type metrics struct {
	names         []string
	metricsByName map[string]metric
}

// readMetrics populates memory metrics
func (m *metrics) readMetrics(namedSamples []namedSample) {
	if m.metricsByName == nil {
		m.metricsByName = make(map[string]metric, len(namedSamples))
	}
	samples := make([]runtimeMetrics.Sample, len(namedSamples))
	for sIdx, entry := range namedSamples {
		samples[sIdx] = entry.sample
	}
	runtimeMetrics.Read(samples)
	for sIdx := 0; sIdx < len(namedSamples); sIdx++ {
		if metricEntry, ok := m.metricsByName[namedSamples[sIdx].name]; !ok {
			m.metricsByName[namedSamples[sIdx].name] = metric{value: samples[sIdx].Value.Uint64()}
		} else {
			m.metricsByName[namedSamples[sIdx].name] = metric{value: samples[sIdx].Value.Uint64(), max: metricEntry.max}
		}
	}
	// other metrics that were not directly read from samples
	heapAllocVal := m.metricsByName[totalAlloc].value - m.metricsByName[memFreed].value
	m.metricsByName[heapAlloc] = metric{value: heapAllocVal, max: m.metricsByName[heapAlloc].max}
}

// compares metrics with another metric snapshot taken earlier. Calculates difference between those
// and for selected metrics evaluates pass/fail criteria
func (m *metrics) diff(before *metrics) (pass bool) {
	pass = true
	for name, entry := range m.metricsByName {
		if entry.max > 0 && entry.value-before.metricsByName[name].value > entry.max {
			pass = false
		}
	}
	return
}

func (m *metrics) print() {
	p := message.NewPrinter(language.English)
	for _, name := range m.names {
		entry := m.metricsByName[name]
		switch name {
		case totalAlloc, heapAlloc:
			fmt.Printf("%s:\t%s\t%s\n", name, format(entry.value), format(entry.max))
		case memFreed:
			fmt.Printf("%s:\t%s\n", name, format(entry.value))
		case objAlloc, objFreed, numGo:
			fmt.Printf("%s:\t%s\n", name, p.Sprintf("%d", entry.value))
		case objRemain:
			fmt.Printf("%s:\t%s\t\t%s\n", name, p.Sprintf("%d", entry.value), p.Sprintf("%d", entry.max))
		case msgRate:
			fmt.Printf("%s:\t%d m/s\n", name, entry.value)
		case duration:
			fmt.Printf("%s:\t%s\n", name, time.Duration(entry.value).String())
		default:
			fmt.Printf("%s:\t%d\t%d\n", name, entry.value, entry.max)
		}
	}
	fmt.Println()
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
