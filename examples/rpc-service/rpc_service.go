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
	"net"
	"os"
	"sync"

	"go.fd.io/govpp"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/api"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/interface_types"
	"go.fd.io/govpp/binapi/ip_types"
	"go.fd.io/govpp/binapi/vpe"
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

	getVppInfo(conn)
	idx := createLoopback(conn)
	listInterfaces(conn)
	addIPAddress(conn, idx)
	watchInterfaceEvents(conn, idx)
}

func getVppInfo(conn api.Connection) {
	c := vpe.NewServiceClient(conn)

	version, err := c.ShowVersion(context.Background(), &vpe.ShowVersion{})
	if err != nil {
		log.Fatalln("ERROR: getting VPP version failed:", err)
	}
	fmt.Printf("VPP Version: %v\n", version.Version)

	systime, err := c.ShowVpeSystemTime(context.Background(), &vpe.ShowVpeSystemTime{})
	if err != nil {
		log.Fatalln("ERROR: getting system time failed:", err)
	}
	fmt.Printf("System Time: %v\n", systime.VpeSystemTime)
}

func createLoopback(conn api.Connection) interface_types.InterfaceIndex {
	c := interfaces.NewServiceClient(conn)

	reply, err := c.CreateLoopback(context.Background(), &interfaces.CreateLoopback{})
	if err != nil {
		log.Fatalln("ERROR: creating loopback failed:", err)
	}
	fmt.Printf("Loopback interface created: %v\n", reply.SwIfIndex)

	return reply.SwIfIndex
}

func listInterfaces(conn api.Connection) {
	c := interfaces.NewServiceClient(conn)

	stream, err := c.SwInterfaceDump(context.Background(), &interfaces.SwInterfaceDump{
		SwIfIndex: ^interface_types.InterfaceIndex(0),
	})
	if err != nil {
		log.Fatalln("ERROR: listing interfaces failed:", err)
	}
	for {
		iface, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("ERROR: receiving interface list failed:", err)
		}
		fmt.Printf("- interface: %s (index: %v)\n", iface.InterfaceName, iface.SwIfIndex)
	}
}

func addIPAddress(conn api.Connection, ifIdx interface_types.InterfaceIndex) {
	c := interfaces.NewServiceClient(conn)

	addr := ip_types.NewAddress(net.IPv4(10, 10, 0, byte(ifIdx)))

	_, err := c.SwInterfaceAddDelAddress(context.Background(), &interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: ifIdx,
		IsAdd:     true,
		Prefix:    ip_types.AddressWithPrefix{Address: addr, Len: 32},
	})
	if err != nil {
		log.Fatalln("ERROR: adding IP address failed:", err)
	}

	fmt.Printf("IP address %v added\n", addr)
}

func watchInterfaceEvents(conn api.Connection, index interface_types.InterfaceIndex) {
	c := interfaces.NewServiceClient(conn)

	// start watcher for specific event message
	watcher, err := conn.WatchEvent(context.Background(), (*interfaces.SwInterfaceEvent)(nil))
	if err != nil {
		log.Fatalln("ERROR: watching interface events failed:", err)
	}

	// enable interface events in VPP
	_, err = c.WantInterfaceEvents(context.Background(), &interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	})
	if err != nil {
		log.Fatalln("ERROR: enabling interface events failed:", err)
	}

	fmt.Printf("watching interface events for index %d\n", index)

	var wg sync.WaitGroup

	// receive notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		for notif := range watcher.Events() {
			e := notif.(*interfaces.SwInterfaceEvent)
			fmt.Printf("incoming interface event: %+v\n", e)
		}
		fmt.Println("watcher done")
	}()

	// generate some events in VPP
	setInterfaceStatus(conn, index, true)
	setInterfaceStatus(conn, index, false)

	// disable interface events in VPP
	_, err = c.WantInterfaceEvents(context.Background(), &interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 0,
	})
	if err != nil {
		log.Fatalln("ERROR: disabling interface events failed:", err)
	}

	// close watcher to stop receiving notifications
	watcher.Close()

	// generate ignored events in VPP
	setInterfaceStatus(conn, index, true)

	wg.Wait()
}

func setInterfaceStatus(conn api.Connection, ifIdx interface_types.InterfaceIndex, up bool) {
	c := interfaces.NewServiceClient(conn)

	var flags interface_types.IfStatusFlags
	if up {
		flags = interface_types.IF_STATUS_API_FLAG_ADMIN_UP
	} else {
		flags = 0
	}
	_, err := c.SwInterfaceSetFlags(context.Background(), &interfaces.SwInterfaceSetFlags{
		SwIfIndex: ifIdx,
		Flags:     flags,
	})
	if err != nil {
		log.Fatalln("ERROR: setting interface flags failed:", err)
	}
	if up {
		fmt.Printf("interface status set to UP")
	} else {
		fmt.Printf("interface status set to DOWN")
	}
}
