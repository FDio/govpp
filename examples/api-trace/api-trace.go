// Copyright (c) 2021 Cisco and/or its affiliates.
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

// api-trace is and example how to use the GoVPP API trace tool.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"go.fd.io/govpp"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/api"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/interface_types"
	"go.fd.io/govpp/binapi/ip_types"
	"go.fd.io/govpp/binapi/memclnt"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/core"
)

var (
	sockAddr = flag.String("socket", socketclient.DefaultSocketName, "Path to VPP API socket file")
)

func main() {
	flag.Parse()

	fmt.Printf("Starting api-trace tool example\n\n")

	// make synchronous VPP connection
	conn, err := govpp.Connect(*sockAddr)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	defer conn.Disconnect()

	fmt.Printf("=> Enabling API trace...\n")
	trace := core.NewTrace(conn, 50)

	singleChannel(conn, trace)
	multiChannel(conn, trace)
	stream(conn, trace)

	trace.Close()
	fmt.Printf("Api-trace tool example finished\n\n")
}

func singleChannel(conn *core.Connection, trace api.Trace) {
	// create the new channel and perform simple compatibility check
	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("ERROR: creating channel failed:", err)
	}
	defer ch.Close()

	fmt.Printf("=> Example 1\n\nEnabling API trace...\n")
	if err = ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
		log.Fatalf("compatibility check failed: %v", err)
	}
	if err = ch.CheckCompatiblity(interfaces.AllMessages()...); err != nil {
		log.Printf("compatibility check failed: %v", err)
	}

	// do some API calls
	fmt.Printf("Calling VPP API...\n")
	retrieveVersion(ch)
	idx := createLoopback(ch)
	addIPAddress("10.10.0.1/24", ch, idx)
	interfaceDump(ch)
	deleteLoopback(ch, idx)
	fmt.Println()

	fmt.Printf("API trace (api calls: %d):\n", len(trace.GetRecords()))
	fmt.Printf("--------------------\n")
	for _, item := range trace.GetRecords() {
		printTrace(item)
	}
	fmt.Printf("--------------------\n")

	fmt.Printf("Clearing API trace...\n\n")
	trace.Clear()
}

func multiChannel(conn *core.Connection, trace api.Trace) {
	ch1, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("ERROR: creating channel failed:", err)
	}
	defer ch1.Close()
	ch2, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("ERROR: creating channel failed:", err)
	}
	defer ch2.Close()

	// do API call again
	fmt.Printf("=> Example 2\n\nCalling VPP API (multi-channel)...\n")
	retrieveVersion(ch1)
	idx1 := createLoopback(ch1)
	idx2 := createLoopback(ch2)
	addIPAddress("20.10.0.1/24", ch1, idx1)
	addIPAddress("30.10.0.1/24", ch2, idx2)
	interfaceDump(ch1)
	deleteLoopback(ch2, idx1)
	deleteLoopback(ch1, idx2)
	fmt.Println()

	chan1, ok := ch1.(*core.Channel)
	if !ok {
		log.Fatalln("ERROR: incorrect type of channel 1:", err)
	}
	chan2, ok := ch2.(*core.Channel)
	if !ok {
		log.Fatalln("ERROR: incorrect type of channel 2:", err)
	}

	fmt.Printf("API trace for channel 1 (api calls: %d):\n", len(trace.GetRecordsForChannel(chan1.GetID())))
	fmt.Printf("--------------------\n")
	for _, item := range trace.GetRecordsForChannel(chan1.GetID()) {
		printTrace(item)
	}
	fmt.Printf("--------------------\n")
	fmt.Printf("API trace for channel 2 (api calls: %d):\n", len(trace.GetRecordsForChannel(chan2.GetID())))
	fmt.Printf("--------------------\n")
	for _, item := range trace.GetRecordsForChannel(chan2.GetID()) {
		printTrace(item)
	}
	fmt.Printf("--------------------\n")
	fmt.Printf("cumulative API trace (api calls: %d):\n", len(trace.GetRecords()))
	fmt.Printf("--------------------\n")
	for _, item := range trace.GetRecords() {
		printTrace(item)
	}
	fmt.Printf("--------------------\n")

	fmt.Printf("Clearing API trace...\n\n")
	trace.Clear()
}

func stream(conn *core.Connection, trace api.Trace) {
	// create the new channel and perform simple compatibility check
	s, err := conn.NewStream(context.Background())
	if err != nil {
		log.Fatalln("ERROR: creating channel failed:", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Fatalf("failed to close stream: %v", err)
		}
	}()

	// do some API calls
	fmt.Printf("=> Example 3\n\nCalling VPP API (stream)...\n")
	invokeRetrieveVersion(conn)
	idx := invokeCreateLoopback(conn)
	invokeAddIPAddress("40.10.0.1/24", conn, idx)
	invokeInterfaceDump(conn)
	invokeDeleteLoopback(conn, idx)
	fmt.Println()

	fmt.Printf("stream API trace (api calls: %d):\n", len(trace.GetRecords()))
	fmt.Printf("--------------------\n")
	for _, item := range trace.GetRecords() {
		printTrace(item)
	}
	fmt.Printf("--------------------\n")

	fmt.Printf("Clearing API trace...\n\n")
	trace.GetRecords()
}

func retrieveVersion(ch api.Channel) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf(" - retrieved VPP version: %s\n", reply.Version)
}

func invokeRetrieveVersion(c api.Connection) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := c.Invoke(context.Background(), req, reply); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	fmt.Printf(" - retrieved VPP version: %s\n", reply.Version)
}

func createLoopback(ch api.Channel) interface_types.InterfaceIndex {
	req := &interfaces.CreateLoopback{}
	reply := &interfaces.CreateLoopbackReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return 0
	}
	fmt.Printf(" - created loopback with index: %d\n", reply.SwIfIndex)
	return reply.SwIfIndex
}

func invokeCreateLoopback(c api.Connection) interface_types.InterfaceIndex {
	req := &interfaces.CreateLoopback{}
	reply := &interfaces.CreateLoopbackReply{}

	if err := c.Invoke(context.Background(), req, reply); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return 0
	}
	fmt.Printf(" - created loopback with index: %d\n", reply.SwIfIndex)
	return reply.SwIfIndex
}

func deleteLoopback(ch api.Channel, index interface_types.InterfaceIndex) {
	req := &interfaces.DeleteLoopback{
		SwIfIndex: index,
	}
	reply := &interfaces.DeleteLoopbackReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf(" - deleted loopback with index: %d\n", index)
}

func invokeDeleteLoopback(c api.Connection, index interface_types.InterfaceIndex) {
	req := &interfaces.DeleteLoopback{
		SwIfIndex: index,
	}
	reply := &interfaces.DeleteLoopbackReply{}

	if err := c.Invoke(context.Background(), req, reply); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf(" - deleted loopback with index: %d\n", index)
}

func addIPAddress(addr string, ch api.Channel, index interface_types.InterfaceIndex) {
	ipAddr, err := ip_types.ParsePrefix(addr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	req := &interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: index,
		IsAdd:     true,
		Prefix:    ip_types.AddressWithPrefix(ipAddr),
	}
	reply := &interfaces.SwInterfaceAddDelAddressReply{}

	if err = ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf(" - IP address %s added to interface with index %d\n", addr, index)
}

func invokeAddIPAddress(addr string, c api.Connection, index interface_types.InterfaceIndex) {
	ipAddr, err := ip_types.ParsePrefix(addr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	req := &interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: index,
		IsAdd:     true,
		Prefix:    ip_types.AddressWithPrefix(ipAddr),
	}
	reply := &interfaces.SwInterfaceAddDelAddressReply{}

	if err = c.Invoke(context.Background(), req, reply); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf(" - IP address %s added to interface with index %d\n", addr, index)
}

func interfaceDump(ch api.Channel) {
	reqCtx := ch.SendMultiRequest(&interfaces.SwInterfaceDump{
		SwIfIndex: ^interface_types.InterfaceIndex(0),
	})
	for {
		msg := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return
		}
		fmt.Printf(" - retrieved interface: %v (idx: %d)\n", msg.InterfaceName, msg.SwIfIndex)
	}
}

func invokeInterfaceDump(c api.Connection) {
	s, err := c.NewStream(context.Background())
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	if err = s.SendMsg(&interfaces.SwInterfaceDump{}); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	if err = s.SendMsg(&memclnt.ControlPing{}); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	for {
		reply, err := s.RecvMsg()
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return
		}
		switch msg := reply.(type) {
		case *interfaces.SwInterfaceDetails:
			fmt.Printf(" - retrieved interface: %v (idx: %d)\n", msg.InterfaceName, msg.SwIfIndex)
		case *memclnt.ControlPingReply:
			return
		}
	}
}

func printTrace(item *api.Record) {
	h, m, s := item.Timestamp.Clock()
	reply := ""
	if item.IsReceived {
		reply = "(reply)"
	}
	fmt.Printf("%dh:%dm:%ds:%dns %s sucess: %t %s\n", h, m, s,
		item.Timestamp.Nanosecond(), item.Message.GetMessageName(), item.Succeeded, reply)
}
