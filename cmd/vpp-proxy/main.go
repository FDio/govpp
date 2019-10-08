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
	"context"
	"encoding/gob"
	"flag"
	"io"
	"log"

	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/adapter/statsclient"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/examples/binapi/interfaces"
	"git.fd.io/govpp.git/examples/binapi/vpe"
	"git.fd.io/govpp.git/proxy"
)

var (
	binapiSocket = flag.String("binapi-socket", socketclient.DefaultSocketName, "Path to VPP binapi socket")
	statsSocket  = flag.String("stats-socket", statsclient.DefaultSocketName, "Path to VPP stats socket")
	proxyAddr    = flag.String("addr", ":7878", "Address on which proxy serves RPC.")
)

func init() {
	for _, msg := range api.GetRegisteredMessages() {
		gob.Register(msg)
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

	// - using binapi message directly
	req := &vpe.CliInband{Cmd: "show version"}
	reply := new(vpe.CliInbandReply)
	if err := binapiChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		log.Fatalln("binapi request failed:", err)
	}
	log.Printf("VPP version: %+v", reply.Reply)

	// - or using generated rpc service
	svc := interfaces.NewServiceClient(binapiChannel)
	stream, err := svc.DumpSwInterface(context.Background(), &interfaces.SwInterfaceDump{})
	if err != nil {
		log.Fatalln("binapi request failed:", err)
	}
	for {
		iface, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("- interface: %+v", iface)
	}
}

func runServer() {
	p := proxy.NewServer()

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

	p.ListenAndServe(*proxyAddr)
}
