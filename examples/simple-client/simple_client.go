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

// simple-client is an example VPP management application that exercises the
// govpp API on real-world use-cases.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/api"
	interfaces "git.fd.io/govpp.git/binapi/interface"
	"git.fd.io/govpp.git/binapi/interface_types"
	"git.fd.io/govpp.git/binapi/ip"
	"git.fd.io/govpp.git/binapi/ip_types"
	"git.fd.io/govpp.git/binapi/mactime"
	"git.fd.io/govpp.git/binapi/vpe"
	"git.fd.io/govpp.git/core"
)

var (
	sockAddr = flag.String("sock", socketclient.DefaultSocketName, "Path to VPP binary API socket file")
)

func main() {
	flag.Parse()

	fmt.Println("Starting simple client example")

	// connect to VPP asynchronously
	conn, conev, err := govpp.AsyncConnect(*sockAddr, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	defer conn.Disconnect()

	// wait for Connected event
	select {
	case e := <-conev:
		if e.State != core.Connected {
			log.Fatalln("ERROR: connecting to VPP failed:", e.Error)
		}
	}

	// create an API channel that will be used in the examples
	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("ERROR: creating channel failed:", err)
	}
	defer ch.Close()

	if err := ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
		log.Fatal(err)
	}

	vppVersion(ch)

	if err := ch.CheckCompatiblity(interfaces.AllMessages()...); err != nil {
		log.Fatal(err)
	}

	idx := createLoopback(ch)
	interfaceDump(ch)

	addIPAddress(ch, idx)
	ipAddressDump(ch, idx)
	interfaceNotifications(ch, idx)

	mactimeDump(conn)

	if len(Errors) > 0 {
		fmt.Printf("finished with %d errors\n", len(Errors))
		os.Exit(1)
	} else {
		fmt.Println("finished successfully")
	}
}

var Errors []error

func logError(err error, msg string) {
	fmt.Printf("ERROR: %s: %v\n", msg, err)
	Errors = append(Errors, err)
}

// vppVersion is the simplest API example - it retrieves VPP version.
func vppVersion(ch api.Channel) {
	fmt.Println("Retrieving version")

	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "retrieving version")
		return
	}
	fmt.Printf("reply: %+v\n", reply)

	fmt.Printf("VPP version: %q\n", reply.Version)
	fmt.Println("OK")
	fmt.Println()
}

// createLoopback sends request to create loopback interface.
func createLoopback(ch api.Channel) interface_types.InterfaceIndex {
	fmt.Println("Creating loopback interface")

	req := &interfaces.CreateLoopback{}
	reply := &interfaces.CreateLoopbackReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "creating loopback interface")
		return 0
	}
	fmt.Printf("reply: %+v\n", reply)

	fmt.Printf("interface index: %v\n", reply.SwIfIndex)
	fmt.Println("OK")
	fmt.Println()

	return reply.SwIfIndex
}

// interfaceDump shows an example of multipart request (multiple replies are expected).
func interfaceDump(ch api.Channel) {
	fmt.Println("Dumping interfaces")

	n := 0
	reqCtx := ch.SendMultiRequest(&interfaces.SwInterfaceDump{})
	for {
		msg := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			logError(err, "dumping interfaces")
			return
		}
		n++
		fmt.Printf(" - interface #%d: %+v\n", n, msg)
		marshal(msg)
	}

	fmt.Println("OK")
	fmt.Println()
}

// addIPAddress sends request to add IP address to interface.
func addIPAddress(ch api.Channel, index interface_types.InterfaceIndex) {
	fmt.Printf("Adding IP address to interface to interface index %d\n", index)

	req := &interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: index,
		IsAdd:     true,
		Prefix: ip_types.AddressWithPrefix{
			Address: ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{10, 10, 0, uint8(index)}),
			},
			Len: 32,
		},
	}
	marshal(req)
	reply := &interfaces.SwInterfaceAddDelAddressReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "adding IP address to interface")
		return
	}
	fmt.Printf("reply: %+v\n", reply)

	fmt.Println("OK")
	fmt.Println()
}

func ipAddressDump(ch api.Channel, index interface_types.InterfaceIndex) {
	fmt.Printf("Dumping IP addresses for interface index %d\n", index)

	req := &ip.IPAddressDump{
		SwIfIndex: index,
	}
	reqCtx := ch.SendMultiRequest(req)

	for {
		msg := &ip.IPAddressDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if err != nil {
			logError(err, "dumping IP addresses")
			return
		}
		if stop {
			break
		}
		fmt.Printf(" - ip address: %+v\n", msg)
		marshal(msg)
	}

	fmt.Println("OK")
	fmt.Println()
}

// interfaceNotifications shows the usage of notification API. Note that for notifications,
// you are supposed to create your own Go channel with your preferred buffer size. If the channel's
// buffer is full, the notifications will not be delivered into it.
func interfaceNotifications(ch api.Channel, index interface_types.InterfaceIndex) {
	fmt.Printf("Subscribing to notificaiton events for interface index %d\n", index)

	notifChan := make(chan api.Message, 100)

	// subscribe for specific notification message
	sub, err := ch.SubscribeNotification(notifChan, &interfaces.SwInterfaceEvent{})
	if err != nil {
		logError(err, "subscribing to interface events")
		return
	}

	// enable interface events in VPP
	err = ch.SendRequest(&interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(&interfaces.WantInterfaceEventsReply{})
	if err != nil {
		logError(err, "enabling interface events")
		return
	}

	// receive notifications
	go func() {
		for notif := range notifChan {
			e := notif.(*interfaces.SwInterfaceEvent)
			fmt.Printf("incoming event: %+v\n", e)
			marshal(e)
		}
	}()

	// generate some events in VPP
	err = ch.SendRequest(&interfaces.SwInterfaceSetFlags{
		SwIfIndex: index,
		Flags:     interface_types.IF_STATUS_API_FLAG_ADMIN_UP,
	}).ReceiveReply(&interfaces.SwInterfaceSetFlagsReply{})
	if err != nil {
		logError(err, "setting interface flags")
		return
	}
	err = ch.SendRequest(&interfaces.SwInterfaceSetFlags{
		SwIfIndex: index,
		Flags:     0,
	}).ReceiveReply(&interfaces.SwInterfaceSetFlagsReply{})
	if err != nil {
		logError(err, "setting interface flags")
		return
	}

	// disable interface events in VPP
	err = ch.SendRequest(&interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 0,
	}).ReceiveReply(&interfaces.WantInterfaceEventsReply{})
	if err != nil {
		logError(err, "setting interface flags")
		return
	}

	// unsubscribe from delivery of the notifications
	err = sub.Unsubscribe()
	if err != nil {
		logError(err, "unsubscribing from interface events")
		return
	}

	fmt.Println("OK")
	fmt.Println()
}

func mactimeDump(conn api.Connection) {
	fmt.Println("Sending mactime dump")

	ctx := context.Background()

	stream, err := conn.NewStream(ctx)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	if err := stream.SendMsg(&mactime.MactimeDump{}); err != nil {
		logError(err, "sending mactime dump")
		return
	}

Loop:
	for {
		msg, err := stream.RecvMsg()
		if err != nil {
			logError(err, "dumping mactime")
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

func marshal(v interface{}) {
	fmt.Printf("GO: %#v\n", v)
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("JSON: %s\n", b)
}
