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
	"os"
	"time"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/api"
	interfaces "git.fd.io/govpp.git/binapi/interface"
	"git.fd.io/govpp.git/binapi/interface_types"
	"git.fd.io/govpp.git/binapi/ip"
	"git.fd.io/govpp.git/binapi/ip_types"
	"git.fd.io/govpp.git/binapi/mactime"
	"git.fd.io/govpp.git/binapi/memclnt"
	"git.fd.io/govpp.git/binapi/vpe"
	"git.fd.io/govpp.git/core"
)

var (
	sockAddr = flag.String("sock", socketclient.DefaultSocketName, "Path to VPP binary API socket file")
)

func main() {
	flag.Parse()

	fmt.Println("Starting stream client example")
	fmt.Println()

	// connect to VPP asynchronously
	conn, connEv, err := govpp.AsyncConnect(*sockAddr, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	defer conn.Disconnect()

	// wait for Connected event
	select {
	case e := <-connEv:
		if e.State != core.Connected {
			log.Fatalln("ERROR: connecting to VPP failed:", e.Error)
		}
	}

	// check compatibility of used messages
	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("ERROR: creating channel failed:", err)
	}
	defer ch.Close()
	if err := ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
		log.Fatal(err)
	}
	if err := ch.CheckCompatiblity(interfaces.AllMessages()...); err != nil {
		log.Fatal(err)
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
	getVppVersionStream(stream)
	idx := createLoopbackStream(stream)
	interfaceDumpStream(stream)
	addIPAddressStream(stream, idx)
	ipAddressDumpStream(stream, idx)
	mactimeDump(stream)
	return
}

func getVppVersionStream(stream api.Stream) {
	fmt.Println("Retrieving version..")

	req := &vpe.ShowVersion{}
	if err := stream.SendMsg(req); err != nil {
		logError(err, "ShowVersion sending message")
		return
	}
	recv, err := stream.RecvMsg()
	if err != nil {
		logError(err, "ShowVersion receive message")
		return
	}
	recvMsg := recv.(*vpe.ShowVersionReply)

	fmt.Printf("Retrieved VPP version: %q\n", recvMsg.Version)
	fmt.Println("OK")
	fmt.Println()
}

func createLoopbackStream(stream api.Stream) (ifIdx interface_types.InterfaceIndex) {
	fmt.Println("Creating the loopback interface..")

	req := &interfaces.CreateLoopback{}
	if err := stream.SendMsg(req); err != nil {
		logError(err, "CreateLoopback sending message")
		return
	}
	recv, err := stream.RecvMsg()
	if err != nil {
		logError(err, "CreateLoopback receive message")
		return
	}
	recvMsg := recv.(*interfaces.CreateLoopbackReply)

	fmt.Printf("Loopback interface index: %v\n", recvMsg.SwIfIndex)
	fmt.Println("OK")
	fmt.Println()

	return recvMsg.SwIfIndex
}

func interfaceDumpStream(stream api.Stream) {
	fmt.Println("Dumping interfaces..")

	if err := stream.SendMsg(&interfaces.SwInterfaceDump{
		SwIfIndex: ^interface_types.InterfaceIndex(0),
	}); err != nil {
		logError(err, "SwInterfaceDump sending message")
		return
	}
	if err := stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		logError(err, "ControlPing sending message")
		return
	}

Loop:
	for {
		msg, err := stream.RecvMsg()
		if err != nil {
			logError(err, "SwInterfaceDump receiving message ")
			return
		}

		switch msg.(type) {
		case *interfaces.SwInterfaceDetails:
			fmt.Printf(" - SwInterfaceDetails: %+v\n", msg)

		case *memclnt.ControlPingReply:
			fmt.Printf(" - ControlPingReply: %+v\n", msg)
			break Loop

		default:
			logError(err, "unexpected message")
			return
		}
	}

	fmt.Println("OK")
	fmt.Println()
}

func addIPAddressStream(stream api.Stream, index interface_types.InterfaceIndex) {
	fmt.Printf("Adding IP address to the interface index %d..\n", index)

	if err := stream.SendMsg(&interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: index,
		IsAdd:     true,
		Prefix: ip_types.AddressWithPrefix{
			Address: ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{10, 10, 0, uint8(index)}),
			},
			Len: 32,
		},
	}); err != nil {
		logError(err, "SwInterfaceAddDelAddress sending message")
		return
	}

	recv, err := stream.RecvMsg()
	if err != nil {
		logError(err, "SwInterfaceAddDelAddressReply receiving message")
		return
	}
	recvMsg := recv.(*interfaces.SwInterfaceAddDelAddressReply)

	fmt.Printf("Added IP address to interface: %v (return value: %d)\n", index, recvMsg.Retval)
	fmt.Println("OK")
	fmt.Println()
}

func ipAddressDumpStream(stream api.Stream, index interface_types.InterfaceIndex) {
	fmt.Printf("Dumping IP addresses for interface index %d..\n", index)

	if err := stream.SendMsg(&ip.IPAddressDump{
		SwIfIndex: index,
	}); err != nil {
		logError(err, "IPAddressDump sending message")
		return
	}
	if err := stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		logError(err, "ControlPing sending sending message")
		return
	}

Loop:
	for {
		msg, err := stream.RecvMsg()
		if err != nil {
			logError(err, "IPAddressDump receiving message ")
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

	fmt.Println("OK")
	fmt.Println()
}

// Mactime dump uses MactimeDumpReply message as an end of the stream
// notification instead of the control ping.
func mactimeDump(stream api.Stream) {
	fmt.Println("Sending mactime dump..")

	if err := stream.SendMsg(&mactime.MactimeDump{}); err != nil {
		logError(err, "sending mactime dump")
		return
	}

Loop:
	for {
		msg, err := stream.RecvMsg()
		if err != nil {
			logError(err, "MactimeDump receiving message")
			return
		}

		switch msg.(type) {
		case *mactime.MactimeDetails:
			fmt.Printf(" - MactimeDetails: %+v\n", msg)

		case *mactime.MactimeDumpReply:
			fmt.Printf(" - MactimeDumpReply: %+v\n", msg)
			break Loop

		default:
			logError(err, "unexpected message")
			return
		}
	}

	fmt.Println("OK")
	fmt.Println()
}

var errors []error

func logError(err error, msg string) {
	fmt.Printf("ERROR: %s: %v\n", msg, err)
	errors = append(errors, err)
}
