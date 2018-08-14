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
	"fmt"
	"net"
	"os"
	"strings"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/examples/bin_api/acl"
	"git.fd.io/govpp.git/examples/bin_api/interfaces"
	"git.fd.io/govpp.git/examples/bin_api/ip"
)

func main() {
	fmt.Println("Starting simple VPP client...")

	// connect to VPP
	conn, err := govpp.Connect("")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer conn.Disconnect()

	// create an API channel that will be used in the examples
	ch, err := conn.NewAPIChannel()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer ch.Close()

	// individual examples
	aclVersion(ch)
	aclConfig(ch)
	aclDump(ch)

	interfaceDump(ch)
	ipAddressDump(ch)

	setIpUnnumbered(ch)
	ipUnnumberedDump(ch)

	interfaceNotifications(ch)
}

// aclVersion is the simplest API example - one empty request message and one reply message.
func aclVersion(ch api.Channel) {
	req := &acl.ACLPluginGetVersion{}
	reply := &acl.ACLPluginGetVersionReply{}

	err := ch.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%+v\n", reply)
	}
}

// aclConfig is another simple API example - in this case, the request contains structured data.
func aclConfig(ch api.Channel) {
	req := &acl.ACLAddReplace{
		ACLIndex: ^uint32(0),
		Tag:      []byte("access list 1"),
		R: []acl.ACLRule{
			{
				IsPermit:       1,
				SrcIPAddr:      net.ParseIP("10.0.0.0").To4(),
				SrcIPPrefixLen: 8,
				DstIPAddr:      net.ParseIP("192.168.1.0").To4(),
				DstIPPrefixLen: 24,
				Proto:          6,
			},
			{
				IsPermit:       1,
				SrcIPAddr:      net.ParseIP("8.8.8.8").To4(),
				SrcIPPrefixLen: 32,
				DstIPAddr:      net.ParseIP("172.16.0.0").To4(),
				DstIPPrefixLen: 16,
				Proto:          6,
			},
		},
	}
	reply := &acl.ACLAddReplaceReply{}

	err := ch.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if reply.Retval != 0 {
		fmt.Println("Retval:", reply.Retval)
		return
	}

	fmt.Printf("%+v\n", reply)

}

// aclDump shows an example where SendRequest and ReceiveReply are not chained together.
func aclDump(ch api.Channel) {
	req := &acl.ACLDump{}
	reply := &acl.ACLDetails{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%+v\n", reply)
	}
}

// interfaceDump shows an example of multipart request (multiple replies are expected).
func interfaceDump(ch api.Channel) {
	fmt.Println("Dumping interfaces")

	req := &interfaces.SwInterfaceDump{}
	reqCtx := ch.SendMultiRequest(req)

	for {
		msg := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		ifaceName := strings.TrimFunc(string(msg.InterfaceName), func(r rune) bool {
			return r == 0x00
		})
		fmt.Printf("Interface: %q %+v\n", ifaceName, msg)
	}
}

func ipAddressDump(ch api.Channel) {
	fmt.Println("Dumping IP addresses")

	req := &ip.IPAddressDump{
		SwIfIndex: 1, //^uint32(0),
	}
	reqCtx := ch.SendMultiRequest(req)

	for {
		msg := &ip.IPAddressDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Printf("ip address: %d %+v\n", msg.SwIfIndex, msg)
	}
}

// aclDump shows an example where SendRequest and ReceiveReply are not chained together.
func setIpUnnumbered(ch api.Channel) {
	req := &interfaces.SwInterfaceSetUnnumbered{
		SwIfIndex:           1,
		UnnumberedSwIfIndex: 2,
		IsAdd:               1,
	}
	reply := &interfaces.SwInterfaceSetUnnumberedReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%+v\n", reply)
	}
}

func ipUnnumberedDump(ch api.Channel) {
	fmt.Println("Dumping IP unnumbered")

	req := &ip.IPUnnumberedDump{
		SwIfIndex: ^uint32(0),
	}
	reqCtx := ch.SendMultiRequest(req)

	for {
		msg := &ip.IPUnnumberedDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Printf("ip unnumbered: %+v\n", msg)
	}
}

// interfaceNotifications shows the usage of notification API. Note that for notifications,
// you are supposed to create your own Go channel with your preferred buffer size. If the channel's
// buffer is full, the notifications will not be delivered into it.
func interfaceNotifications(ch api.Channel) {
	// subscribe for specific notification message
	notifChan := make(chan api.Message, 100)
	subs, err := ch.SubscribeNotification(notifChan, interfaces.NewSwInterfaceEvent)
	if err != nil {
		panic(err)
	}

	// enable interface events in VPP
	err = ch.SendRequest(&interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(&interfaces.WantInterfaceEventsReply{})
	if err != nil {
		panic(err)
	}

	// generate some events in VPP
	err = ch.SendRequest(&interfaces.SwInterfaceSetFlags{
		SwIfIndex:   0,
		AdminUpDown: 0,
	}).ReceiveReply(&interfaces.SwInterfaceSetFlagsReply{})
	if err != nil {
		panic(err)
	}
	err = ch.SendRequest(&interfaces.SwInterfaceSetFlags{
		SwIfIndex:   0,
		AdminUpDown: 1,
	}).ReceiveReply(&interfaces.SwInterfaceSetFlagsReply{})
	if err != nil {
		panic(err)
	}

	// receive one notification
	notif := (<-notifChan).(*interfaces.SwInterfaceEvent)
	fmt.Printf("NOTIF: %+v\n", notif)

	// disable interface events in VPP
	err = ch.SendRequest(&interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 0,
	}).ReceiveReply(&interfaces.WantInterfaceEventsReply{})
	if err != nil {
		panic(err)
	}

	// unsubscribe from delivery of the notifications
	err = ch.UnsubscribeNotification(subs)
	if err != nil {
		panic(err)
	}
}
