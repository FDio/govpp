// Copyright (c) 2020 Cisco and/or its affiliates.
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

// multi-vpp is an example of managing multiple VPPs in single application.
package main

import (
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
	"git.fd.io/govpp.git/binapi/vpe"
	"git.fd.io/govpp.git/core"
)

var (
	sockAddrVpp1 = flag.String("sock1", socketclient.DefaultSocketName, "Path to binary API socket file of the first VPP instance")
	sockAddrVpp2 = flag.String("sock2", socketclient.DefaultSocketName, "Path to binary API socket file of the second VPP instance")
)

func main() {
	flag.Parse()
	fmt.Println("Starting multi-vpp example")

	// since both of them default to the same value
	if *sockAddrVpp1 == *sockAddrVpp2 {
		log.Fatalln("ERROR: identical VPP sockets defined, set at least one of them to non-default path")
	}

	// connect VPP1
	conn1, err := connectToVPP(*sockAddrVpp1, 1)
	if err != nil {
		log.Fatalf("ERROR: connecting VPP failed (socket %s): %v\n", *sockAddrVpp1, err)
	}
	defer conn1.Disconnect()
	ch1, err := getAPIChannel(conn1)
	if err != nil {
		log.Fatalf("ERROR: creating channel failed (socket: %s): %v\n", *sockAddrVpp1, err)
	}
	defer ch1.Close()

	// connect VPP2
	conn2, err := connectToVPP(*sockAddrVpp2, 2)
	if err != nil {
		log.Fatalf("ERROR: connecting VPP failed (socket %s): %v\n", *sockAddrVpp2, err)
	}
	defer conn2.Disconnect()
	ch2, err := getAPIChannel(conn2)
	if err != nil {
		log.Fatalf("ERROR: creating channel failed (socket: %s): %v\n", *sockAddrVpp2, err)
	}
	defer ch2.Close()

	// configure VPPs
	ifIdx1 := createLoopback(ch1)
	addIPToInterface(ch1, ifIdx1, "10.10.0.1/24")
	ifIdx2 := createLoopback(ch2)
	addIPToInterface(ch2, ifIdx2, "20.10.0.1/24")

	// retrieve configuration from the VPPs
	retrieveIPAddresses(ch1, ifIdx1)
	retrieveIPAddresses(ch2, ifIdx2)

	if len(Errors) > 0 {
		fmt.Printf("finished with %d errors\n", len(Errors))
		os.Exit(1)
	} else {
		fmt.Println("finished successfully")
	}
}

func connectToVPP(socket string, attempts int) (*core.Connection, error) {
	connection, event, err := govpp.AsyncConnect(socket, attempts, core.DefaultReconnectInterval)
	if err != nil {
		return nil, err
	}

	// handle connection event
	select {
	case e := <-event:
		if e.State != core.Connected {
			return nil, err
		}
	}
	return connection, nil
}

func getAPIChannel(conn *core.Connection) (api.Channel, error) {
	ch, err := conn.NewAPIChannel()
	if err != nil {
		return nil, err
	}

	if err := ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
		return nil, err
	}

	getVppVersion(ch)

	if err := ch.CheckCompatiblity(interfaces.AllMessages()...); err != nil {
		return nil, err
	}
	return ch, nil
}

// getVppVersion returns VPP version (simple API usage)
func getVppVersion(ch api.Channel) {
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

var Errors []error

func logError(err error, msg string) {
	fmt.Printf("ERROR: %s: %v\n", msg, err)
	Errors = append(Errors, err)
}

// createLoopback sends request to create a loopback interface
func createLoopback(ch api.Channel) interface_types.InterfaceIndex {
	fmt.Println("Adding loopback interface")

	req := &interfaces.CreateLoopback{}
	reply := &interfaces.CreateLoopbackReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "adding loopback interface")
		return 0
	}
	fmt.Printf("reply: %+v\n", reply)

	fmt.Printf("interface index: %v\n", reply.SwIfIndex)
	fmt.Println("OK")
	fmt.Println()

	return reply.SwIfIndex
}

// addIPToInterface sends request to add an IP address to an interface.
func addIPToInterface(ch api.Channel, index interface_types.InterfaceIndex, ip string) {
	fmt.Printf("Setting up IP address to the interface with index %d\n", index)
	prefix, err := ip_types.ParsePrefix(ip)
	if err != nil {
		logError(err, "attempt to add invalid IP address")
		return
	}

	req := &interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: index,
		IsAdd:     true,
		Prefix:    ip_types.AddressWithPrefix(prefix),
	}
	reply := &interfaces.SwInterfaceAddDelAddressReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "adding IP address to interface")
		return
	}
	fmt.Printf("reply: %+v\n", reply)

	fmt.Println("OK")
	fmt.Println()
}

func retrieveIPAddresses(ch api.Channel, index interface_types.InterfaceIndex) {
	fmt.Printf("Retrieving IP addresses for interface index %d\n", index)

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
		prefix := ip_types.Prefix(msg.Prefix)
		fmt.Printf(" - ip address: %v\n", prefix)
	}

	fmt.Println("OK")
	fmt.Println()
}
