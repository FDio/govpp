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
	"strings"

	"go.fd.io/govpp"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/adapter/statsclient"
	"go.fd.io/govpp/api"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/interface_types"
	"go.fd.io/govpp/binapi/ip"
	"go.fd.io/govpp/binapi/ip_types"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/core"
)

var (
	binapiSockAddrVpp1 = flag.String("api-sock-1", socketclient.DefaultSocketName, "Path to binary API socket file of the VPP1")
	statsSockAddrVpp1  = flag.String("stats-sock-1", statsclient.DefaultSocketName, "Path to stats socket file of the VPP1")
	binapiSockAddrVpp2 = flag.String("api-sock-2", socketclient.DefaultSocketName, "Path to binary API socket file of the VPP2")
	statsSockAddrVpp2  = flag.String("stats-sock-2", statsclient.DefaultSocketName, "Path to stats socket file of the VPP2")
)

var errors []error

func main() {
	flag.Parse()
	fmt.Println("Starting multi-vpp example")

	defer func() {
		if len(errors) > 0 {
			logInfo("Finished with %d errors\n", len(errors))
			os.Exit(1)
		} else {
			logInfo("Finished successfully\n")
		}
	}()

	// since sockets default to the same value
	if *binapiSockAddrVpp1 == *binapiSockAddrVpp2 {
		log.Println("ERROR: identical VPP binapi sockets defined, set at least one of them to a non-default path")
	}
	if *statsSockAddrVpp1 == *statsSockAddrVpp2 {
		log.Println("ERROR: identical VPP stats sockets defined, set at least one of them to a non-default path")
	}
	var name1, name2 = "vpp1", "vpp2"
	ch1, statsConn1, disconnect1 := connectVPP(name1, *binapiSockAddrVpp1, *statsSockAddrVpp1)
	defer disconnect1()

	ch2, statsConn2, disconnect2 := connectVPP(name2, *binapiSockAddrVpp2, *statsSockAddrVpp2)
	defer disconnect2()

	fmt.Println()

	// retrieve VPP1 version
	logHeader("Retrieving %s version", name1)
	getVppVersion(ch1, name1)

	// retrieve VPP2 version
	logHeader("Retrieving %s version", name2)
	getVppVersion(ch1, name2)

	// configure VPP1
	logHeader("Configuring %s", name1)
	ifIdx1 := createLoopback(ch1, name1)
	addIPsToInterface(ch1, ifIdx1, []string{"10.10.0.1/24", "15.10.0.1/24"})

	// configure VPP2
	logHeader("Configuring %s", name2)
	ifIdx2 := createLoopback(ch2, name2)
	addIPsToInterface(ch2, ifIdx2, []string{"20.10.0.1/24", "25.10.0.1/24"})

	// retrieve configuration from VPPs
	retrieveIPAddresses(ch1, name1, ifIdx1)
	retrieveIPAddresses(ch2, name2, ifIdx2)

	// retrieve stats from VPPs
	retrieveStats(statsConn1, name1)
	retrieveStats(statsConn2, name2)

	// cleanup
	logHeader("Cleaning up %s", name1)
	deleteIPsToInterface(ch1, ifIdx1, []string{"10.10.0.1/24", "15.10.0.1/24"})
	deleteLoopback(ch1, ifIdx1)
	logHeader("Cleaning up %s", name2)
	deleteIPsToInterface(ch2, ifIdx2, []string{"20.10.0.1/24", "25.10.0.1/24"})
	deleteLoopback(ch2, ifIdx2)
}

func connectVPP(name, binapiSocket, statsSocket string) (api.Channel, api.StatsProvider, func()) {
	fmt.Println()
	logHeader("Connecting to %s", name)

	// connect VPP1 to the binapi socket
	ch, disconnectBinapi, err := connectBinapi(binapiSocket, 1)
	if err != nil {
		log.Fatalf("ERROR: connecting VPP binapi failed (socket %s): %v\n", binapiSocket, err)
	}

	// connect VPP1 to the stats socket
	statsConn, disconnectStats, err := connectStats(name, statsSocket)
	if err != nil {
		disconnectBinapi()
		log.Fatalf("ERROR: connecting VPP stats failed (socket %s): %v\n", statsSocket, err)
	}

	logInfo("OK\n")

	return ch, statsConn, func() {
		disconnectStats()
		disconnectBinapi()
		logInfo("VPP %s disconnected\n", name)
	}
}

// connectBinapi connects to the binary API socket and returns a communication channel
func connectBinapi(socket string, attempts int) (api.Channel, func(), error) {
	logInfo("Attaching to the binapi socket %s\n", socket)
	conn, event, err := govpp.AsyncConnect(socket, attempts, core.DefaultReconnectInterval)
	if err != nil {
		return nil, nil, err
	}
	e := <-event
	if e.State != core.Connected {
		return nil, nil, err
	}
	ch, err := getAPIChannel(conn)
	if err != nil {
		return nil, nil, err
	}
	disconnect := func() {
		if ch != nil {
			ch.Close()
		}
		if conn != nil {
			conn.Disconnect()
		}
	}
	return ch, disconnect, nil
}

// connectStats connects to the stats socket and returns a stats provider
func connectStats(name, socket string) (api.StatsProvider, func(), error) {
	logInfo("Attaching to the stats socket %s\n", socket)
	sc := statsclient.NewStatsClient(socket)
	conn, err := core.ConnectStats(sc)
	if err != nil {
		return nil, nil, err
	}
	disconnect := func() {
		if err := sc.Disconnect(); err != nil {
			logError(err, "failed to disconnect "+name+" stats socket")
		}
	}
	return conn, disconnect, nil
}

// getAPIChannel creates new API channel and verifies its compatibility
func getAPIChannel(c api.ChannelProvider) (api.Channel, error) {
	ch, err := c.NewAPIChannel()
	if err != nil {
		return nil, err
	}
	if err := ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
		return nil, fmt.Errorf("compatibility check failed: %w", err)
	}
	if err := ch.CheckCompatiblity(interfaces.AllMessages()...); err != nil {
		logInfo("compatibility check failed: %v", err)
	}
	return ch, nil
}

// getVppVersion returns VPP version (simple API usage)
func getVppVersion(ch api.Channel, name string) {
	logInfo("Retrieving version of %s ..\n", name)

	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "retrieving version")
		return
	}
	logInfo("Retrieved version is %q\n", reply.Version)
	fmt.Println()
}

// createLoopback sends request to create a loopback interface
func createLoopback(ch api.Channel, name string) interface_types.InterfaceIndex {
	logInfo("Adding loopback interface ..\n")

	req := &interfaces.CreateLoopback{}
	reply := &interfaces.CreateLoopbackReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "adding loopback interface")
		return 0
	}
	logInfo("Interface index %d added to %s\n", reply.SwIfIndex, name)

	return reply.SwIfIndex
}

// deleteLoopback removes created loopback interface
func deleteLoopback(ch api.Channel, ifIdx interface_types.InterfaceIndex) {
	logInfo("Removing loopback interface ..\n")
	req := &interfaces.DeleteLoopback{
		SwIfIndex: ifIdx,
	}
	reply := &interfaces.DeleteLoopbackReply{}

	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
		logError(err, "removing loopback interface")
	}
	logInfo("OK\n")
	fmt.Println()
}

// addIPsToInterface sends request to add IP addresses to an interface.
func addIPsToInterface(ch api.Channel, index interface_types.InterfaceIndex, ips []string) {
	for _, ipAddr := range ips {
		logInfo("Adding IP address %s\n", ipAddr)
		prefix, err := ip_types.ParsePrefix(ipAddr)
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
	}

	logInfo("OK\n")
	fmt.Println()
}

// deleteIPsToInterface sends request to remove IP addresses from an interface.
func deleteIPsToInterface(ch api.Channel, index interface_types.InterfaceIndex, ips []string) {
	for _, ipAddr := range ips {
		logInfo("Removing IP address %s\n", ipAddr)
		prefix, err := ip_types.ParsePrefix(ipAddr)
		if err != nil {
			logError(err, "attempt to remove invalid IP address")
			return
		}

		req := &interfaces.SwInterfaceAddDelAddress{
			SwIfIndex: index,
			Prefix:    ip_types.AddressWithPrefix(prefix),
		}
		reply := &interfaces.SwInterfaceAddDelAddressReply{}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			logError(err, "removing IP address to interface")
			return
		}
	}
}

// retrieveIPAddresses reads IP address from the interface
func retrieveIPAddresses(ch api.Channel, name string, index interface_types.InterfaceIndex) {
	logHeader("Retrieving interface data from %s", name)
	req := &ip.IPAddressDump{
		SwIfIndex: index,
	}
	reqCtx := ch.SendMultiRequest(req)

	logInfo("Dump IP addresses for interface index %d ..\n", index)
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
		logInfo(" - ip address: %v\n", prefix)
	}

	logInfo("OK\n")
	fmt.Println()
}

// retrieveStats reads interface stats
func retrieveStats(s api.StatsProvider, name string) {
	logHeader("Retrieving interface stats from %s", name)
	ifStats := &api.InterfaceStats{}
	err := s.GetInterfaceStats(ifStats)
	if err != nil {
		logError(err, "dumping interface stats")
		return
	}
	logInfo("Dump interface stats ..\n")
	for _, ifStats := range ifStats.Interfaces {
		logInfo(" - %+v\n", ifStats)
	}

	logInfo("OK\n")
	fmt.Println()
}

// logHeader prints underlined message (for better output segmentation)
func logHeader(format string, a ...interface{}) {
	n, _ := fmt.Printf(format+"\n", a...)
	fmt.Println(strings.Repeat("-", n-1))
}

// logInfo prints info message
func logInfo(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// logError prints error message
func logError(err error, msg string) {
	fmt.Printf("[ERROR]: %s: %v\n", msg, err)
	errors = append(errors, err)
}
