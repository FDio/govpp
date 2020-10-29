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
	"strings"
	"time"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/statsclient"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
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
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s: usage [ls|dump|poll|errors|interfaces|nodes|system|buffers|memory] <patterns>...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	skipZeros := !*dumpAll

	var patterns []string
	if flag.NArg() > 0 {
		patterns = flag.Args()[1:]
	}

	client := statsclient.NewStatsClient(*statsSocket)

	c, err := core.ConnectStats(client)
	if err != nil {
		log.Fatalln("Connecting failed:", err)
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
			if skipZeros && counter.Value == 0 {
				continue
			}
			fmt.Printf(" - %v\n", counter)
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

		dumpStats(client, patterns, skipZeros)

	case "poll":
		fmt.Printf("Polling stats.. %s\n", strings.Join(patterns, " "))

		pollStats(client, patterns, skipZeros)

	case "list", "ls", "":
		fmt.Printf("Listing stats.. %s\n", strings.Join(patterns, " "))

		listStats(client, patterns)

	default:
		fmt.Printf("invalid command: %q\n", cmd)
	}
}

func listStats(client adapter.StatsAPI, patterns []string) {
	list, err := client.ListStats(patterns...)
	if err != nil {
		log.Fatalln("listing stats failed:", err)
	}

	for _, stat := range list {
		fmt.Printf(" - %v\n", stat)
	}

	fmt.Printf("Listed %d stats\n", len(list))
}

func dumpStats(client adapter.StatsAPI, patterns []string, skipZeros bool) {
	stats, err := client.DumpStats(patterns...)
	if err != nil {
		log.Fatalln("dumping stats failed:", err)
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

		select {
		case <-tick:
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

		select {
		case <-tick:
			if err := client.GetSystemStats(stats); err != nil {
				log.Println("updating system stats failed:", err)
			}
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

		select {
		case <-tick:
			if err := client.GetInterfaceStats(stats); err != nil {
				log.Println("updating system stats failed:", err)
			}
		}
	}
}
