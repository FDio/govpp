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

// stream-client is an example VPP management application that exercises the
// govpp API on real-world use-cases.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"go.fd.io/govpp"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/api"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/interface_types"
	"go.fd.io/govpp/binapi/ip"
	"go.fd.io/govpp/binapi/ip_types"
	"go.fd.io/govpp/binapi/mactime"
	"go.fd.io/govpp/binapi/memclnt"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/core"
)

var (
	sockAddr = flag.String("sock", socketclient.DefaultSocketName, "Path to VPP binary API socket file")
)

func main() {
	flag.Parse()

	fmt.Println("Starting stream client example")

	// connect to VPP asynchronously
	conn, connEv, err := govpp.AsyncConnect(*sockAddr, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	defer conn.Disconnect()

	// wait for Connected event
	e := <-connEv
	if e.State != core.Connected {
		log.Fatalln("ERROR: connecting to VPP failed:", e.Error)
	}

	// check compatibility of used messages
	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("ERROR: creating channel failed:", err)
	}
	defer ch.Close()
	if err := ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
		log.Fatalf("compatibility check failed: %v", err)
	}
	if err := ch.CheckCompatiblity(interfaces.AllMessages()...); err != nil {
		log.Printf("compatibility check failed: %v", err)
	}

	// process errors encountered during the example
	defer func() {
		if len(errors) > 0 {
			fmt.Printf("finished with %d errors\n", len(errors))
			os.Exit(1)
		} else {
			fmt.Println("finished successfully")
		}
	}()

	// send and receive messages using stream (low-low level API)
	stream, err := conn.NewStream(context.Background(),
		core.WithRequestSize(50),
		core.WithReplySize(50),
		core.WithReplyTimeout(2*time.Second))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := stream.Close(); err != nil {
			logError(err, "closing the stream")
		}
	}()

	getVppVersion(stream)
	idx := createLoopback(stream)
	interfaceDump(stream)
	addIPAddress(stream, idx)
	ipAddressDump(stream, idx)
	mactimeDump(stream)
	interfaceNotifications(conn, idx)
}

func getVppVersion(stream api.Stream) {
	fmt.Println("Retrieving version")

	if err := stream.SendMsg(&vpe.ShowVersion{}); err != nil {
		logError(err, "get version request")
		return
	}
	recvMsg, err := stream.RecvMsg()
	if err != nil {
		logError(err, "get version reply")
		return
	}
	reply := recvMsg.(*vpe.ShowVersionReply)
	if api.RetvalToVPPApiError(reply.Retval) != nil {
		logError(err, "get version reply retval")
		return
	}

	fmt.Printf("VPP version: %v\n", reply.Version)
}

func createLoopback(stream api.Stream) (ifIdx interface_types.InterfaceIndex) {
	fmt.Println("Creating loopback interface..")

	if err := stream.SendMsg(&interfaces.CreateLoopback{}); err != nil {
		logError(err, "create loopback request")
		return
	}
	recv, err := stream.RecvMsg()
	if err != nil {
		logError(err, "create loopback reply")
		return
	}
	reply := recv.(*interfaces.CreateLoopbackReply)
	if api.RetvalToVPPApiError(reply.Retval) != nil {
		logError(err, "create loopback reply retval")
		return
	}

	fmt.Printf("Loopback interface created: %v\n", reply.SwIfIndex)

	return reply.SwIfIndex
}

func interfaceDump(stream api.Stream) {
	fmt.Println("Listing interfaces")

	if err := stream.SendMsg(&interfaces.SwInterfaceDump{
		SwIfIndex: ^interface_types.InterfaceIndex(0),
	}); err != nil {
		logError(err, "list interfaces request")
		return
	}
	if err := stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		logError(err, "ControlPing request")
		return
	}

Loop:
	for {
		msg, err := stream.RecvMsg()
		if err != nil {
			logError(err, "receiving interface list")
			return
		}

		switch m := msg.(type) {
		case *interfaces.SwInterfaceDetails:
			fmt.Printf("- interface: %s (index: %v)\n", m.InterfaceName, m.SwIfIndex)

		case *memclnt.ControlPingReply:
			fmt.Printf(" - ControlPingReply: %+v\n", m)
			break Loop

		default:
			logError(err, "unexpected message")
			return
		}
	}
}

func addIPAddress(stream api.Stream, index interface_types.InterfaceIndex) {
	addr := ip_types.NewAddress(net.IPv4(10, 10, 0, byte(index)))

	fmt.Printf("Adding IP address %v to interface (index %d)\n", addr, index)

	if err := stream.SendMsg(&interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: index,
		IsAdd:     true,
		Prefix:    ip_types.AddressWithPrefix{Address: addr, Len: 32},
	}); err != nil {
		logError(err, "add IP address request")
		return
	}

	recv, err := stream.RecvMsg()
	if err != nil {
		logError(err, "add IP address reply")
		return
	}
	reply := recv.(*interfaces.SwInterfaceAddDelAddressReply)
	if api.RetvalToVPPApiError(reply.Retval) != nil {
		logError(err, "add IP address reply retval")
		return
	}

	fmt.Printf("IP address %v added\n", addr)
}

func ipAddressDump(stream api.Stream, index interface_types.InterfaceIndex) {
	fmt.Printf("Listing IP addresses for interface (index %d)\n", index)

	if err := stream.SendMsg(&ip.IPAddressDump{
		SwIfIndex: index,
	}); err != nil {
		logError(err, "dump IP address request")
		return
	}
	if err := stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		logError(err, "sending ControlPing")
		return
	}

Loop:
	for {
		msg, err := stream.RecvMsg()
		if err != nil {
			logError(err, "receiving IP addresses")
			return
		}

		switch msg.(type) {
		case *ip.IPAddressDetails:
			fmt.Printf(" - IPAddressDetails: %+v\n", msg)

		case *memclnt.ControlPingReply:
			fmt.Printf(" - ControlPingReply: %+v\n", msg)
			break Loop

		default:
			logError(err, "unexpected message")
			return
		}
	}
}

// Mactime dump uses MactimeDumpReply message as an end of the stream
// notification instead of the control ping.
func mactimeDump(stream api.Stream) {
	fmt.Println("Sending mactime dump")

	if err := stream.SendMsg(&mactime.MactimeDump{}); err != nil {
		logError(err, "mactime dump request")
		return
	}

Loop:
	for {
		msg, err := stream.RecvMsg()
		if err != nil {
			logError(err, "receiving mactime dump")
			return
		}

		switch m := msg.(type) {
		case *mactime.MactimeDetails:
			fmt.Printf(" - MactimeDetails: %+v\n", m)

		case *mactime.MactimeDumpReply:
			if err := api.RetvalToVPPApiError(m.Retval); err != nil && err != api.NO_CHANGE {
				logError(err, "mactime dump reply retval")
				return
			}
			fmt.Printf(" - MactimeDumpReply: %+v\n", m)
			break Loop

		default:
			logError(err, "unexpected message")
			return
		}
	}
}

// interfaceNotifications demonstrates how to watch for interface events.
func interfaceNotifications(conn api.Connection, index interface_types.InterfaceIndex) {

	// start watcher for specific event message
	watcher, err := conn.WatchEvent(context.Background(), (*interfaces.SwInterfaceEvent)(nil))
	if err != nil {
		logError(err, "watching interface events")
		return
	}

	// enable interface events in VPP
	var reply interfaces.WantInterfaceEventsReply
	err = conn.Invoke(context.Background(), &interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}, &reply)
	if err != nil || api.RetvalToVPPApiError(reply.Retval) != nil {
		logError(err, "enabling interface events")
		return
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
	reply.Reset()
	if err := conn.Invoke(context.Background(), &interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 0,
	}, &reply); err != nil || api.RetvalToVPPApiError(reply.Retval) != nil {
		logError(err, "disabling interface events")
		return
	}

	// unsubscribe from delivery of the notifications
	watcher.Close()

	// generate ignored events in VPP
	setInterfaceStatus(conn, index, true)

	wg.Wait()
}

func setInterfaceStatus(conn api.Connection, ifIdx interface_types.InterfaceIndex, up bool) {
	var flags interface_types.IfStatusFlags
	if up {
		flags = interface_types.IF_STATUS_API_FLAG_ADMIN_UP
	} else {
		flags = 0
	}
	var reply interfaces.SwInterfaceSetFlagsReply
	if err := conn.Invoke(context.Background(), &interfaces.SwInterfaceSetFlags{
		SwIfIndex: ifIdx,
		Flags:     flags,
	}, &reply); err != nil {
		logError(err, "setting interface flags")
		return
	} else if err = api.RetvalToVPPApiError(reply.Retval); err != nil {
		logError(err, "setting interface flags retval")
		return
	}
}

var errors []error

func logError(err error, msg string) {
	fmt.Printf("ERROR: %s: %v\n", msg, err)
	errors = append(errors, err)
}
