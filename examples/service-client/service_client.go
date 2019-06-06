// Copyright (c) 2017 Cisco and/or its affiliates.
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

// service-client is an example VPP management application that exercises the
// govpp API using generated service client.
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/examples/binapi/interfaces"
	"git.fd.io/govpp.git/examples/binapi/vpe"
)

func main() {
	fmt.Println("Starting VPP service client...")

	// connect to VPP
	conn, err := govpp.Connect("")
	if err != nil {
		log.Fatalln("failed to connect:", err)
	}
	defer conn.Disconnect()

	// create an API channel
	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("failed to create channel:", err)
	}
	defer ch.Close()

	showVersion(ch)
	interfaceDump(ch)
}

// showVersion shows an example of simple request with services.
func showVersion(ch api.Channel) {
	c := vpe.NewService(ch)

	version, err := c.ShowVersion(context.Background(), &vpe.ShowVersion{})
	if err != nil {
		log.Fatalln("ShowVersion failed:", err)
	}

	fmt.Printf("Version: %v\n", version.Version)
}

// interfaceDump shows an example of multi request with services.
func interfaceDump(ch api.Channel) {
	c := interfaces.NewService(ch)

	ifaces, err := c.DumpSwInterface(context.Background(), &interfaces.SwInterfaceDump{})
	if err != nil {
		log.Fatalln("DumpSwInterface failed:", err)
	}

	fmt.Printf("Listing %d interfaces:\n", len(ifaces))
	for _, d := range ifaces {
		fmt.Printf("- interface: %s\n", bytes.Trim(d.InterfaceName, "\x00"))
	}
}
