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
	"fmt"
	"log"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/vppapiclient"
)

// This example shows how to work with VPP's new stats API.

func main() {
	fmt.Println("Starting VPP stats API example..")

	client := vppapiclient.NewStatClient(vppapiclient.DefaultStatSocket)

	// connect to stats API
	if err := client.Connect(); err != nil {
		log.Fatalln("connecting client failed:", err)
	}
	defer client.Disconnect()

	// list stats by patterns
	// you can omit parameters to list all stats
	list, err := client.ListStats("/if", "/sys")
	if err != nil {
		log.Fatalln("listing stats failed:", err)
	}

	for _, stat := range list {
		fmt.Printf(" - %v\n", stat)
	}
	fmt.Printf("listed %d stats\n", len(list))

	// dump stats by patterns to retrieve stats with the stats data
	stats, err := client.DumpStats()
	if err != nil {
		log.Fatalln("dumping stats failed:", err)
	}

	for _, stat := range stats {
		switch data := stat.Data.(type) {
		case adapter.ErrorStat:
			if data == 0 {
				// skip printing errors with 0 value
				continue
			}
		}
		fmt.Printf(" - %-25s %25v %+v\n", stat.Name, stat.Type, stat.Data)
	}
}
