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

package proxy

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/core"
)

// Server defines a proxy server that serves client requests to stats and binapi.
type Server struct {
	rpc *rpc.Server

	statsConn  *core.StatsConnection
	binapiConn *core.Connection
}

func NewServer() *Server {
	return &Server{
		rpc: rpc.NewServer(),
	}
}

func (p *Server) ConnectStats(stats adapter.StatsAPI) error {
	var err error
	p.statsConn, err = core.ConnectStats(stats)
	if err != nil {
		return err
	}
	return nil
}

func (p *Server) DisconnectStats() {
	if p.statsConn != nil {
		p.statsConn.Disconnect()
	}
}

func (p *Server) ConnectBinapi(binapi adapter.VppAPI) error {
	var err error
	p.binapiConn, err = core.Connect(binapi)
	if err != nil {
		return err
	}
	return nil
}

func (p *Server) DisconnectBinapi() {
	if p.binapiConn != nil {
		p.binapiConn.Disconnect()
	}
}

func (p *Server) ListenAndServe(addr string) {
	if p.statsConn != nil {
		statsRPC := NewStatsRPC(p.statsConn)
		if err := p.rpc.Register(statsRPC); err != nil {
			panic(err)
		}
	}
	if p.binapiConn != nil {
		ch, err := p.binapiConn.NewAPIChannel()
		if err != nil {
			panic(err)
		}
		binapiRPC := NewBinapiRPC(ch)
		if err := p.rpc.Register(binapiRPC); err != nil {
			panic(err)
		}
	}

	p.rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	l, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	defer l.Close()

	log.Printf("proxy serving on: %v", addr)

	if err := http.Serve(l, nil); err != nil {
		log.Fatalln(err)
	}
}
