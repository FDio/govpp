// Copyright (c) 2018 Cisco and/or its affiliates.
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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/adapter/statsclient"
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/core"
)

// ------------------------------------------------------------------
// Example - Stats API
// ------------------------------------------------------------------
// The example stats_api demonstrates how to retrieve stats
// from the VPP using the new stats API.
// ------------------------------------------------------------------

var (
	statsSocket = flag.String("socket", statsclient.DefaultSocketName, "Path to VPP stats socket")
	dumpAll     = flag.Bool("all", false, "Dump all stats including ones with zero values")
	pollPeriod  = flag.Duration("period", time.Second*5, "Polling interval period")
	async       = flag.Bool("async", false, "Use asynchronous connection")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s: usage [ls|dump|poll|errors|interfaces|nodes|system|buffers|memory|epoch] <patterns/index>...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	skipZeros := !*dumpAll

	patterns := make([]string, 0)
	indexes := make([]uint32, 0)
	if flag.NArg() > 0 {
		for _, arg := range flag.Args()[1:] {
			if index, err := strconv.Atoi(arg); err == nil {
				indexes = append(indexes, uint32(index))
				continue
			}
			patterns = append(patterns, arg)
		}
	}

	var (
		client *statsclient.StatsClient
		c      *core.StatsConnection
		err    error
	)

	if *async {
		var statsChan chan core.ConnectionEvent
		client = statsclient.NewStatsClient(*statsSocket, statsclient.SetSocketRetryPeriod(1*time.Second),
			statsclient.SetSocketRetryTimeout(10*time.Second))
		c, statsChan, err = core.AsyncConnectStats(client, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
		if err != nil {
			log.Fatalln("Asynchronous connecting failed:", err)
		}
		e := <-statsChan
		if e.State == core.Connected {
			// OK
		} else {
			log.Fatalf("VPP stats asynchronous connection failed: %s\n", e.State.String())
		}
	} else {
		client = statsclient.NewStatsClient(*statsSocket)
		c, err = core.ConnectStats(client)
		if err != nil {
			log.Fatalln("Connecting failed:", err)
		}
	}
	defer c.Disconnect()

	switch cmd := flag.Arg(0); cmd {
	case "system":
		stats := new(api.SystemStats)
		if err := c.GetSystemStats(stats); err != nil {
			log.Fatalln("getting system stats failed:", err)
		}
		fmt.Printf("System stats: %+v\n", stats)

	case "poll-system":
		pollSystem(c)

	case "nodes":
		fmt.Println("Listing node stats..")
		stats := new(api.NodeStats)
		if err := c.GetNodeStats(stats); err != nil {
			log.Fatalln("getting node stats failed:", err)
		}

		for _, node := range stats.Nodes {
			if skipZeros && node.Calls == 0 && node.Suspends == 0 && node.Clocks == 0 && node.Vectors == 0 {
				continue
			}
			fmt.Printf(" - %+v\n", node)
		}
		fmt.Printf("Listed %d node counters\n", len(stats.Nodes))

	case "interfaces":
		fmt.Println("Listing interface stats..")
		stats := new(api.InterfaceStats)
		if err := c.GetInterfaceStats(stats); err != nil {
			log.Fatalln("getting interface stats failed:", err)
		}
		for _, iface := range stats.Interfaces {
			fmt.Printf(" - %+v\n", iface)
		}
		fmt.Printf("Listed %d interface counters\n", len(stats.Interfaces))

	case "poll-interfaces":
		pollInterfaces(c)

	case "errors":
		fmt.Printf("Listing error stats.. %s\n", strings.Join(patterns, " "))
		stats := new(api.ErrorStats)
		if err := c.GetErrorStats(stats); err != nil {
			log.Fatalln("getting error stats failed:", err)
		}
		n := 0
		for _, counter := range stats.Errors {
			var sum uint32
			for _, valuePerWorker := range counter.Values {
				sum += uint32(valuePerWorker)
			}

			if skipZeros && sum == 0 {
				continue
			}
			fmt.Printf(" - %v %d (per worker: %v)\n", counter.CounterName, sum, counter.Values)
			n++
		}
		fmt.Printf("Listed %d (%d) error counters\n", n, len(stats.Errors))

	case "buffers":
		stats := new(api.BufferStats)
		if err := c.GetBufferStats(stats); err != nil {
			log.Fatalln("getting buffer stats failed:", err)
		}
		fmt.Printf("Buffer stats: %+v\n", stats)

	case "memory":
		stats := new(api.MemoryStats)
		if err := c.GetMemoryStats(stats); err != nil {
			log.Fatalln("getting memory stats failed:", err)
		}
		fmt.Printf("Memory stats: %+v\n", stats)

	case "dump":
		fmt.Printf("Dumping stats.. %s\n", strings.Join(patterns, " "))

		dumpStats(client, patterns, indexes, skipZeros)

	case "poll":
		fmt.Printf("Polling stats.. %s\n", strings.Join(patterns, " "))

		pollStats(client, patterns, skipZeros)

	case "list", "ls", "":
		fmt.Printf("Listing stats.. %s\n", strings.Join(patterns, " "))

		listStats(client, patterns, indexes)

	case "epoch", "e":
		fmt.Printf("Getting epoch..\n")

		getEpoch(client)

	default:
		fmt.Printf("invalid command: %q\n", cmd)
	}
}

func listStats(client adapter.StatsAPI, patterns []string, indexes []uint32) {
	var err error
	list := make([]adapter.StatIdentifier, 0)
	if (len(patterns) == 0 && len(indexes) == 0) || len(patterns) != 0 {
		list, err = client.ListStats(patterns...)
		if err != nil {
			log.Fatalln("listing stats failed:", err)
		}
	}
	if len(indexes) != 0 {
		dir, err := client.PrepareDirOnIndex(indexes...)
		if err != nil {
			log.Fatalln("listing stats failed:", err)
		}
		for _, onIndexSi := range dir.Entries {
			list = append(list, onIndexSi.StatIdentifier)
		}
	}
	for _, stat := range list {
		fmt.Printf(" - %d\t %v\n", stat.Index, string(stat.Name))
	}

	fmt.Printf("Listed %d stats\n", len(list))
}

func getEpoch(client adapter.StatsAPI) {
	dir, err := client.PrepareDir()
	if err != nil {
		log.Fatalln("failed to prepare dir in order to read epoch:", err)
	}
	d := *dir
	fmt.Printf("Epoch %d\n", d.Epoch)
}

func dumpStats(client adapter.StatsAPI, patterns []string, indexes []uint32, skipZeros bool) {
	var err error
	stats := make([]adapter.StatEntry, 0)
	if (len(patterns) == 0 && len(indexes) == 0) || len(patterns) != 0 {
		stats, err = client.DumpStats(patterns...)
		if err != nil {
			log.Fatalln("dumping stats failed:", err)
		}
	}
	if len(indexes) != 0 {
		dir, err := client.PrepareDirOnIndex(indexes...)
		if err != nil {
			log.Fatalln("dumping stats failed:", err)
		}
		stats = append(stats, dir.Entries...)
	}

	n := 0
	for _, stat := range stats {
		if skipZeros && (stat.Data == nil || stat.Data.IsZero()) {
			continue
		}
		fmt.Printf(" - %-50s %25v %+v\n", stat.Name, stat.Type, stat.Data)
		n++
	}

	fmt.Printf("Dumped %d (%d) stats\n", n, len(stats))
}

func pollStats(client adapter.StatsAPI, patterns []string, skipZeros bool) {
	dir, err := client.PrepareDir(patterns...)
	if err != nil {
		log.Fatalln("preparing dir failed:", err)
	}

	tick := time.Tick(*pollPeriod)
	for {
		n := 0
		fmt.Println(time.Now().Format(time.Stamp))
		for _, stat := range dir.Entries {
			if skipZeros && (stat.Data == nil || stat.Data.IsZero()) {
				continue
			}
			fmt.Printf("%-50s %+v\n", stat.Name, stat.Data)
			n++
		}
		fmt.Println()

		<-tick
		if err := client.UpdateDir(dir); err != nil {
			if err == adapter.ErrStatsDirStale {
				if dir, err = client.PrepareDir(patterns...); err != nil {
					log.Fatalln("preparing dir failed:", err)
				}
				continue
			}
			log.Fatalln("updating dir failed:", err)
		}
	}
}

func pollSystem(client api.StatsProvider) {
	stats := new(api.SystemStats)

	if err := client.GetSystemStats(stats); err != nil {
		log.Fatalln("updating system stats failed:", err)
	}

	tick := time.Tick(*pollPeriod)
	for {
		fmt.Printf("System stats: %+v\n", stats)
		fmt.Println()

		<-tick
		if err := client.GetSystemStats(stats); err != nil {
			log.Println("updating system stats failed:", err)
		}
	}
}

func pollInterfaces(client api.StatsProvider) {
	stats := new(api.InterfaceStats)

	if err := client.GetInterfaceStats(stats); err != nil {
		log.Fatalln("updating system stats failed:", err)
	}

	tick := time.Tick(*pollPeriod)
	for {
		fmt.Printf("Interface stats (%d interfaces)\n", len(stats.Interfaces))
		for i := range stats.Interfaces {
			fmt.Printf(" - %+v\n", stats.Interfaces[i])
		}
		fmt.Println()

		<-tick
		if err := client.GetInterfaceStats(stats); err != nil {
			log.Println("updating system stats failed:", err)
		}
	}
}
