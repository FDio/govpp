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

	"git.fd.io/govpp.git/adapter/statclient"
)

func main() {
	fmt.Println("Starting example for VPP stats API...")

	client := statclient.NewStatsClient(statclient.DefaultStatSocket)

	// connect to stats API
	if err := client.Connect(); err != nil {
		log.Fatalln("Stats client connect failed:", err)
	}
	defer client.Disconnect()

	list, err := client.ListStats("/if", "/sys")
	if err != nil {
		log.Fatalln("LisStats failed:", err)
	}

	for _, stat := range list {
		fmt.Printf(" - %v\n", stat)
	}
	fmt.Printf("%d stats\n", len(list))

	stats, err := client.DumpStats("/if", "/sys")
	if err != nil {
		log.Fatalln("DumpStats failed:", err)
	}

	for _, stat := range stats {
		fmt.Printf(" - %-25s %25v %+v\n", stat.Name, stat.Type, stat.Data)
	}
}
