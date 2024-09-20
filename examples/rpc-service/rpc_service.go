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
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/alkiranet/govpp"
	"github.com/alkiranet/govpp/adapter/socketclient"
	"github.com/alkiranet/govpp/api"
	interfaces "github.com/alkiranet/govpp/binapi/interface"
	"github.com/alkiranet/govpp/binapi/vpe"
)

var (
	sockAddr = flag.String("sock", socketclient.DefaultSocketName, "Path to VPP binary API socket file")
)

func main() {
	flag.Parse()

	fmt.Println("Starting RPC service example")

	// connect to VPP
	conn, err := govpp.Connect(*sockAddr)
	if err != nil {
		log.Fatalln("ERROR: connecting to VPP failed:", err)
	}
	defer conn.Disconnect()

	showVersion(conn)
	interfaceDump(conn)
}

// showVersion shows an example of simple request with services.
func showVersion(conn api.Connection) {
	c := vpe.NewServiceClient(conn)

	version, err := c.ShowVersion(context.Background(), &vpe.ShowVersion{})
	if err != nil {
		log.Fatalln("ERROR: ShowVersion failed:", err)
	}

	fmt.Printf("Version: %v\n", version.Version)
}

// interfaceDump shows an example of multi request with services.
func interfaceDump(conn api.Connection) {
	c := interfaces.NewServiceClient(conn)

	stream, err := c.SwInterfaceDump(context.Background(), &interfaces.SwInterfaceDump{})
	if err != nil {
		log.Fatalln("ERROR: DumpSwInterface failed:", err)
	}

	fmt.Println("Dumping interfaces")
	for {
		iface, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("ERROR: DumpSwInterface failed:", err)
		}
		fmt.Printf("- interface: %s\n", strings.Trim(iface.InterfaceName, "\x00"))
	}
}
