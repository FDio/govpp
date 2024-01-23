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
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/test/vpptesting"
	"runtime"
	"runtime/metrics"
	"strconv"
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
	goroutines       = "/sched/goroutines:goroutines"
)

// TestAPIMemory tests the GoVPP memory consumption for various numbers of API calls
func BenchmarkAPIMemory(b *testing.B) {
	flag.Parse()

	fmt.Printf("Running GoVPP API calls memory test\n\n")
	test := vpptesting.SetupVPP(b)
	vpeRPC := vpe.NewServiceClient(test.Conn)

	samples := []metrics.Sample{
		{Name: heapBytesAllocs},
		{Name: heapBytesFrees},
		{Name: heapObjectAllocs},
		{Name: heapObjectFrees},
		{Name: goroutines},
	}
	pass := true
	testAPICalls := func(n uint, thresholds [3]uint64) {
		tHolds0, tHolds1, tHolds2 := "N/A", "N/A", "N/A"

		now := time.Now()

		fmt.Printf("For %d:\tMeasured\tThreshold\n%s\n", n, strings.Repeat("-", 41))

		runtime.GC()
		metricsBefore := readMetrics(samples)
		for i := 0; i < int(n); i++ {
			if _, err := vpeRPC.ShowVersion(context.Background(), &vpe.ShowVersion{}); err != nil {
				b.Fatal("calling show version failed:", err)
			}
		}
		metricsAfter := readMetrics(samples)
		d := time.Since(now)

		totalAlloc := metricsAfter[heapBytesAllocs] - metricsBefore[heapBytesAllocs]
		if uint64(totalAlloc) > thresholds[0] {
			pass = false
		}
		heapFrees := metricsAfter[heapBytesFrees] - metricsBefore[heapBytesFrees]
		heapAlloc := totalAlloc - heapFrees
		if uint64(heapAlloc) > thresholds[1] {
			pass = false
		}
		objectAlloc := metricsAfter[heapObjectAllocs] - metricsBefore[heapObjectAllocs]
		objectFreed := metricsAfter[heapObjectFrees] - metricsBefore[heapObjectFrees]
		objectRemain := objectAlloc - objectFreed
		if uint64(objectRemain) > thresholds[2] {
			pass = false
		}

		// if thresholds are set, use them in the output
		if !(thresholds[0] == 0 && thresholds[1] == 0 && thresholds[2] == 0) {
			tHolds0 = goMetric(thresholds[0]).format()
			tHolds1 = goMetric(thresholds[1]).format()
			tHolds2 = goMetric(thresholds[2]).addSep()
		}
		fmt.Printf("Total alloc:\t%s\t%s\n", totalAlloc.format(), tHolds0)
		fmt.Printf("Memory Freed:\t%s\n", heapFrees.format())
		fmt.Printf("Heap alloc:\t%s\t%s\n", heapAlloc.format(), tHolds1)
		fmt.Printf("Objects alloc:\t%s\nObj freed:\t%s\n", objectAlloc.addSep(), objectFreed.addSep())
		fmt.Printf("Objects remain:\t%s\t%s\n", objectRemain.addSep(), tHolds2)
		fmt.Printf("Num goroutines:\t%d\nDuration:\t%s\n\n", metricsAfter[goroutines], d.String())
	}

	// run the custom soak test and skip the rest
	if *apiNum != 0 {
		testAPICalls(*apiNum, [3]uint64{})
		return
	}
	testAPICalls(1000, [3]uint64{2621440, 3145728, 50000})
	testAPICalls(10000, [3]uint64{26214400, 3145728, 50000})
	testAPICalls(100000, [3]uint64{262144000, 5242880, 50000})
	//testAPICalls(1000000, [3]uint64{2684364560, 5242880, 50000})
	//testAPICalls(10000000, [3]uint64{26843645600, 5242880, 50000})

	if !pass {
		b.Fatal("one or more memory thresholds was exceeded")
	}
}

type goMetric uint64

func readMetrics(s []metrics.Sample) map[string]goMetric {
	metrics.Read(s)
	goMetrics := make(map[string]goMetric)
	for _, sample := range s {
		goMetrics[sample.Name] = goMetric(sample.Value.Uint64())
	}
	return goMetrics
}

// shortens the number and adds unit
func (m goMetric) format() string {
	const (
		_  = iota
		KB = 1 << (10 * iota)
		MB
		GB
		TB
		PB
	)

	var unit string
	value := float64(m)

	switch {
	case m < KB:
		unit = "B"
	case m < MB:
		unit = "KB"
		value /= KB
	case m < GB:
		unit = "MB"
		value /= MB
	case m < TB:
		unit = "GB"
		value /= GB
	case m < PB:
		unit = "TB"
		value /= TB
	default:
		unit = "PB"
		value /= PB
	}
	return align(fmt.Sprintf("%.2f %s", value, unit))
}

// add number separators for better readability, like 1000 => 1,000
func (m goMetric) addSep() string {
	numStr := strconv.Itoa(int(m))
	numColumns := (len(numStr) - 1) / 3
	result := make([]rune, 0, len(numStr)+numColumns)
	for i := len(numStr) - 1; i >= 0; i-- {
		result = append(result, rune(numStr[i]))
		if i > 0 && (len(numStr)-i)%3 == 0 {
			result = append(result, ',')
		}
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return align(string(result))
}

// align the table
func align(s string) string {
	if len(s) < 8 {
		return s + strings.Repeat(" ", 8-len(s))
	}
	return s
}
