//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"encoding/gob"
	"flag"
	"log"

	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/adapter/statsclient"
	"go.fd.io/govpp/api"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/vlib"
	"go.fd.io/govpp/binapi/vpe"
	_ "go.fd.io/govpp/core"
	"go.fd.io/govpp/proxy"
)

var (
	binapiSocket = flag.String("binapi-socket", socketclient.DefaultSocketName, "Path to VPP binapi socket")
	statsSocket  = flag.String("stats-socket", statsclient.DefaultSocketName, "Path to VPP stats socket")
	proxyAddr    = flag.String("addr", ":7878", "Address on which proxy serves RPC.")
)

func init() {
	for _, msgList := range api.GetRegisteredMessages() {
		for _, msg := range msgList {
			gob.Register(msg)
		}
	}
}

func main() {
	flag.Parse()

	switch cmd := flag.Arg(0); cmd {
	case "server":
		runServer()
	case "client":
		runClient()
	default:
		log.Printf("invalid command: %q, (available commands: client, server)", cmd)
	}
}

func runClient() {
	// connect to proxy server
	client, err := proxy.Connect(*proxyAddr)
	if err != nil {
		log.Fatalln("connecting to proxy failed:", err)
	}

	// proxy stats
	statsProvider, err := client.NewStatsClient()
	if err != nil {
		log.Fatalln(err)
	}

	var sysStats api.SystemStats
	if err := statsProvider.GetSystemStats(&sysStats); err != nil {
		log.Fatalln("getting stats failed:", err)
	}
	log.Printf("SystemStats: %+v", sysStats)

	var ifaceStats api.InterfaceStats
	if err := statsProvider.GetInterfaceStats(&ifaceStats); err != nil {
		log.Fatalln("getting stats failed:", err)
	}
	log.Printf("InterfaceStats: %+v", ifaceStats)

	// proxy binapi
	binapiChannel, err := client.NewBinapiClient()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("checking compatibility")
	var msgs []api.Message
	msgs = append(msgs, interfaces.AllMessages()...)
	msgs = append(msgs, vpe.AllMessages()...)
	if err := binapiChannel.CheckCompatiblity(msgs...); err != nil {
		panic(err)
	}

	req := &vlib.CliInband{Cmd: "show version"}
	reply := new(vlib.CliInbandReply)
	if err := binapiChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		log.Fatalln("binapi request failed:", err)
	}
	log.Printf("VPP version: %+v", reply.Reply)
}

func runServer() {
	p, err := proxy.NewServer()
	if err != nil {
		log.Fatalln(err)
	}

	statsAdapter := statsclient.NewStatsClient(*statsSocket)
	binapiAdapter := socketclient.NewVppClient(*binapiSocket)

	if err := p.ConnectStats(statsAdapter); err != nil {
		log.Fatalln("connecting to stats failed:", err)
	}
	defer p.DisconnectStats()

	if err := p.ConnectBinapi(binapiAdapter); err != nil {
		log.Fatalln("connecting to binapi failed:", err)
	}
	defer p.DisconnectBinapi()

	if err := p.ListenAndServe(*proxyAddr); err != nil {
		log.Fatalln(err)
	}
}
