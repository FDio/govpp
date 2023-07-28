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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"go.fd.io/govpp"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/api"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/interface_types"
	"go.fd.io/govpp/binapi/ip"
	"go.fd.io/govpp/binapi/ip_types"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/core"
)

var (
	sockAddr = flag.String("sock", socketclient.DefaultSocketName, "Path to VPP binary API socket file")
)

func main() {
	flag.Parse()

	fmt.Println("Starting simple client example")

	// connect to VPP
	conn, connEv, err := govpp.AsyncConnect(*sockAddr, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	defer conn.Disconnect()

	e := <-connEv
	if e.State != core.Connected {
		log.Fatalln("ERROR: connecting to VPP failed:", e.Error)
	}

	// check message compatibility
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
			log.Fatalf("finished with %d errors", len(errors))
		}
	}()

	// use Channel request/reply (channel API)
	getVppVersion(ch)
	getSystemTime(ch)
	idx := createLoopback(ch)
	listInterfaces(ch)
	addIPAddress(ch, idx)
	listIPaddresses(ch, idx)
	watchInterfaceEvents(ch, idx)
}

func getVppVersion(ch api.Channel) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "retrieving version")
		return
	}

	fmt.Printf("VPP version: %q\n", reply.Version)
}

func getSystemTime(ch api.Channel) {
	req := &vpe.ShowVpeSystemTime{}
	reply := &vpe.ShowVpeSystemTimeReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "retrieving system time")
		return
	}

	fmt.Printf("system time: %v\n", reply.VpeSystemTime)
}

func createLoopback(ch api.Channel) interface_types.InterfaceIndex {
	req := &interfaces.CreateLoopback{}
	reply := &interfaces.CreateLoopbackReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "creating loopback")
		return 0
	}

	fmt.Printf("loopback created: %v\n", reply.SwIfIndex)

	return reply.SwIfIndex
}

func listInterfaces(ch api.Channel) {
	reqCtx := ch.SendMultiRequest(&interfaces.SwInterfaceDump{
		SwIfIndex: ^interface_types.InterfaceIndex(0),
	})
	for {
		iface := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(iface)
		if stop {
			break
		}
		if err != nil {
			logError(err, "listing interfaces")
			return
		}
		fmt.Printf(" - interface: %+v (ifIndex: %v)\n", iface.InterfaceName, iface.SwIfIndex)
		marshal(iface)
	}

	fmt.Println("OK")
	fmt.Println()
}

func addIPAddress(ch api.Channel, ifIdx interface_types.InterfaceIndex) {
	addr := ip_types.NewAddress(net.IPv4(10, 10, 0, byte(ifIdx)))

	req := &interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: ifIdx,
		IsAdd:     true,
		Prefix:    ip_types.AddressWithPrefix{Address: addr, Len: 32},
	}
	marshal(req)
	reply := &interfaces.SwInterfaceAddDelAddressReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "adding IP address")
		return
	}
}

func listIPaddresses(ch api.Channel, index interface_types.InterfaceIndex) {
	reqCtx := ch.SendMultiRequest(&ip.IPAddressDump{
		SwIfIndex: index,
	})
	for {
		ipAddr := &ip.IPAddressDetails{}
		stop, err := reqCtx.ReceiveReply(ipAddr)
		if err != nil {
			logError(err, "listing IP addresses")
			return
		}
		if stop {
			break
		}
		fmt.Printf(" - IP address: %+v\n", ipAddr)
		marshal(ipAddr)
	}
}

// watchInterfaceEvents shows the usage of notification API. Note that for notifications,
// you are supposed to create your own Go channel with your preferred buffer size. If the channel's
// buffer is full, the notifications will not be delivered into it.
func watchInterfaceEvents(ch api.Channel, index interface_types.InterfaceIndex) {
	notifChan := make(chan api.Message, 100)

	// subscribe for specific event message
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

	fmt.Printf("subscribed to interface events for index %d\n", index)

	var wg sync.WaitGroup

	// receive notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		for notif := range notifChan {
			e := notif.(*interfaces.SwInterfaceEvent)
			fmt.Printf("incoming interface event: %+v\n", e)
			marshal(e)
		}
		fmt.Println("watcher done")
	}()

	// generate some events in VPP
	setInterfaceStatus(ch, index, true)
	setInterfaceStatus(ch, index, false)

	// disable interface events in VPP
	err = ch.SendRequest(&interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 0,
	}).ReceiveReply(&interfaces.WantInterfaceEventsReply{})
	if err != nil {
		logError(err, "disabling interface events")
		return
	}

	// unsubscribe from receiving events
	err = sub.Unsubscribe()
	if err != nil {
		logError(err, "unsubscribing from interface events")
		return
	}

	// generate ignored events in VPP
	setInterfaceStatus(ch, index, true)

	wg.Wait()
}

func setInterfaceStatus(ch api.Channel, ifIdx interface_types.InterfaceIndex, up bool) {
	var flags interface_types.IfStatusFlags
	if up {
		flags = interface_types.IF_STATUS_API_FLAG_ADMIN_UP
	} else {
		flags = 0
	}
	if err := ch.SendRequest(&interfaces.SwInterfaceSetFlags{
		SwIfIndex: ifIdx,
		Flags:     flags,
	}).ReceiveReply(&interfaces.SwInterfaceSetFlagsReply{}); err != nil {
		log.Fatalln("ERROR:  setting interface flags failed:", err)
	}
	if up {
		fmt.Printf("interface status set to UP")
	} else {
		fmt.Printf("interface status set to DOWN")
	}
}

func marshal(v interface{}) {
	fmt.Printf("GO: %#v\n", v)
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("JSON: %s\n", b)
}

var errors []error

func logError(err error, msg string) {
	fmt.Printf("ERROR: %s: %v\n", msg, err)
	errors = append(errors, err)
}
