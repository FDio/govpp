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

// Binary stats-client is an example VPP management application that exercises the
// govpp API for interface counters together with asynchronous connection to VPP.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/examples/bin_api/stats"
)

/*

	IMPORTANT NOTICE!

	The binary API module stats used in this example will be deprecated in VPP 19.01.
	VPP's new stats API should be used, you can find basic usage of new stats API in example stats-api.

*/

func main() {
	fmt.Println("Starting stats VPP client..")

	// async connect to VPP
	conn, statCh, err := govpp.AsyncConnect("")
	if err != nil {
		log.Fatalln("Error:", err)
	}
	defer conn.Disconnect()

	// create an API channel that will be used in the examples
	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatalln("Error:", err)
	}
	defer ch.Close()

	// create channel for Interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	var notifChan chan api.Message
	var simpleSub api.SubscriptionCtx
	var combinedSub api.SubscriptionCtx

	// loop until Interrupt signal is received
loop:
	for {
		select {
		case connEvent := <-statCh:
			// VPP connection state change
			switch connEvent.State {
			case core.Connected:
				fmt.Println("VPP connected.")
				if notifChan == nil {
					simpleSub, combinedSub, notifChan = subscribeNotifications(ch)
				}
				requestStatistics(ch)

			case core.Disconnected:
				fmt.Println("VPP disconnected.")
			}

		case msg := <-notifChan:
			switch notif := msg.(type) {
			case *stats.VnetInterfaceSimpleCounters:
				// simple counter notification received
				processSimpleCounters(notif)
			case *stats.VnetInterfaceCombinedCounters:
				// combined counter notification received
				processCombinedCounters(notif)
			default:
				fmt.Println("Ignoring unknown VPP notification")
			}

		case <-sigChan:
			// interrupt received
			fmt.Println("Interrupt received, exiting.")
			break loop
		}
	}

	simpleSub.Unsubscribe()
	combinedSub.Unsubscribe()
}

// subscribeNotifications subscribes for interface counters notifications.
func subscribeNotifications(ch api.Channel) (api.SubscriptionCtx, api.SubscriptionCtx, chan api.Message) {
	notifChan := make(chan api.Message, 100)

	simpleSub, err := ch.SubscribeNotification(notifChan, &stats.VnetInterfaceSimpleCounters{})
	if err != nil {
		panic(err)
	}
	combinedSub, err := ch.SubscribeNotification(notifChan, &stats.VnetInterfaceCombinedCounters{})
	if err != nil {
		panic(err)
	}

	return simpleSub, combinedSub, notifChan
}

// requestStatistics requests interface counters notifications from VPP.
func requestStatistics(ch api.Channel) {
	if err := ch.SendRequest(&stats.WantStats{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(&stats.WantStatsReply{}); err != nil {
		panic(err)
	}
}

// processSimpleCounters processes simple counters received from VPP.
func processSimpleCounters(counters *stats.VnetInterfaceSimpleCounters) {
	fmt.Printf("SimpleCounters: %+v\n", counters)

	counterNames := []string{
		"Drop", "Punt",
		"IPv4", "IPv6",
		"RxNoBuf", "RxMiss",
		"RxError", "TxError",
		"MPLS",
	}

	for i := uint32(0); i < counters.Count; i++ {
		fmt.Printf("Interface '%d': %s = %d\n",
			counters.FirstSwIfIndex+i, counterNames[counters.VnetCounterType], counters.Data[i])
	}
}

// processCombinedCounters processes combined counters received from VPP.
func processCombinedCounters(counters *stats.VnetInterfaceCombinedCounters) {
	fmt.Printf("CombinedCounters: %+v\n", counters)

	counterNames := []string{"Rx", "Tx"}

	for i := uint32(0); i < counters.Count; i++ {
		if len(counterNames) <= int(counters.VnetCounterType) {
			continue
		}
		fmt.Printf("Interface '%d': %s packets = %d, %s bytes = %d\n",
			counters.FirstSwIfIndex+i,
			counterNames[counters.VnetCounterType], counters.Data[i].Packets,
			counterNames[counters.VnetCounterType], counters.Data[i].Bytes)
	}
}
