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
	"git.fd.io/govpp.git/adapter/vppapiclient"
)

// ------------------------------------------------------------------
// Example - Stats API
// ------------------------------------------------------------------
// The example stats_api demonstrates how to retrieve stats
// from the VPP using the new stats API.
// ------------------------------------------------------------------

var (
	statsSocket = flag.String("socket", vppapiclient.DefaultStatSocket, "VPP stats segment socket")
	skipZeros   = flag.Bool("skipzero", true, "Skip stats with zero values")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s: usage [ls|dump] <patterns>...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	cmd := flag.Arg(0)

	switch cmd {
	case "", "ls", "dump":
	default:
		flag.Usage()
	}

	var patterns []string
	if flag.NArg() > 0 {
		patterns = flag.Args()[1:]
	}

	client := vppapiclient.NewStatClient(*statsSocket)

	fmt.Printf("Connecting to stats socket: %s\n", *statsSocket)

	if err := client.Connect(); err != nil {
		log.Fatalln("Connecting failed:", err)
	}
	defer client.Disconnect()

	switch cmd {
	case "dump":
		dumpStats(client, patterns)
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

func dumpStats(client adapter.StatsAPI, patterns []string) {
	fmt.Printf("Dumping stats.. %s\n", strings.Join(patterns, " "))

	stats, err := client.DumpStats(patterns...)
	if err != nil {
		log.Fatalln("dumping stats failed:", err)
	}

	n := 0
	for _, stat := range stats {
		if isZero(stat.Data) && *skipZeros {
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
	case adapter.CombinedCounterStat:
		for _, ss := range s {
			for _, sss := range ss {
				if sss.Bytes != 0 || sss.Packets != 0 {
					return false
				}
			}
		}
	}
	return true
}
