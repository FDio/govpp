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

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/statsclient"
	"git.fd.io/govpp.git/adapter/vppapiclient"
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
	oldclient   = flag.Bool("oldclient", false, "Use old client for stats API (vppapiclient)")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s: usage [ls|dump|errors|interfaces|nodes|system|buffers] <patterns>...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	skipZeros := !*dumpAll

	cmd := flag.Arg(0)
	switch cmd {
	case "", "ls", "dump", "errors", "interfaces", "nodes", "system", "buffers":
	default:
		flag.Usage()
	}

	var patterns []string
	if flag.NArg() > 0 {
		patterns = flag.Args()[1:]
	}

	var client adapter.StatsAPI
	if *oldclient {
		client = vppapiclient.NewStatClient(*statsSocket)
	} else {
		client = statsclient.NewStatsClient(*statsSocket)
	}

	fmt.Printf("Connecting to stats socket: %s\n", *statsSocket)

	c, err := core.ConnectStats(client)
	if err != nil {
		log.Fatalln("Connecting failed:", err)
	}
	defer c.Disconnect()

	switch cmd {
	case "system":
		stats, err := c.GetSystemStats()
		if err != nil {
			log.Fatalln("getting system stats failed:", err)
		}
		fmt.Printf("System stats: %+v\n", stats)

	case "nodes":
		fmt.Println("Listing node stats..")
		stats, err := c.GetNodeStats()
		if err != nil {
			log.Fatalln("getting node stats failed:", err)
		}
		for _, node := range stats.Nodes {
			if node.Calls == 0 && node.Suspends == 0 && node.Clocks == 0 && node.Vectors == 0 && skipZeros {
				continue
			}
			fmt.Printf(" - %+v\n", node)
		}
		fmt.Printf("Listed %d node counters\n", len(stats.Nodes))

	case "interfaces":
		fmt.Println("Listing interface stats..")
		stats, err := c.GetInterfaceStats()
		if err != nil {
			log.Fatalln("getting interface stats failed:", err)
		}
		for _, iface := range stats.Interfaces {
			fmt.Printf(" - %+v\n", iface)
		}
		fmt.Printf("Listed %d interface counters\n", len(stats.Interfaces))

	case "errors":
		fmt.Printf("Listing error stats.. %s\n", strings.Join(patterns, " "))
		stats, err := c.GetErrorStats(patterns...)
		if err != nil {
			log.Fatalln("getting error stats failed:", err)
		}
		n := 0
		for _, counter := range stats.Errors {
			if counter.Value == 0 && skipZeros {
				continue
			}
			fmt.Printf(" - %v\n", counter)
			n++
		}
		fmt.Printf("Listed %d (%d) error counters\n", n, len(stats.Errors))

	case "buffers":
		stats, err := c.GetBufferStats()
		if err != nil {
			log.Fatalln("getting buffer stats failed:", err)
		}
		fmt.Printf("Buffer stats: %+v\n", stats)

	case "dump":
		dumpStats(client, patterns, skipZeros)

	default:
		listStats(client, patterns)
	}
}

func listStats(client adapter.StatsAPI, patterns []string) {
	fmt.Printf("Listing stats.. %s\n", strings.Join(patterns, " "))

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
	fmt.Printf("Dumping stats.. %s\n", strings.Join(patterns, " "))

	stats, err := client.DumpStats(patterns...)
	if err != nil {
		log.Fatalln("dumping stats failed:", err)
	}

	n := 0
	for _, stat := range stats {
		if isZero(stat.Data) && skipZeros {
			continue
		}
		fmt.Printf(" - %-25s %25v %+v\n", stat.Name, stat.Type, stat.Data)
		n++
	}

	fmt.Printf("Dumped %d (%d) stats\n", n, len(stats))
}

func isZero(stat adapter.Stat) bool {
	switch s := stat.(type) {
	case adapter.ScalarStat:
		return s == 0
	case adapter.ErrorStat:
		return s == 0
	case adapter.SimpleCounterStat:
		for _, ss := range s {
			for _, sss := range ss {
				if sss != 0 {
					return false
				}
			}
		}
		return true
	case adapter.CombinedCounterStat:
		for _, ss := range s {
			for _, sss := range ss {
				if sss.Bytes != 0 || sss.Packets != 0 {
					return false
				}
			}
		}
		return true
	}
	return false
}
