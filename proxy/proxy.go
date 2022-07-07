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
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"

	"go.fd.io/govpp/adapter"
)

// Server defines a proxy server that serves client requests to stats and binapi.
type Server struct {
	rpc *rpc.Server

	statsRPC  *StatsRPC
	binapiRPC *BinapiRPC
}

func NewServer() (*Server, error) {
	srv := &Server{
		rpc:       rpc.NewServer(),
		statsRPC:  &StatsRPC{},
		binapiRPC: &BinapiRPC{},
	}

	if err := srv.rpc.Register(srv.statsRPC); err != nil {
		return nil, err
	}

	if err := srv.rpc.Register(srv.binapiRPC); err != nil {
		return nil, err
	}

	return srv, nil
}

func (p *Server) ConnectStats(stats adapter.StatsAPI) error {
	return p.statsRPC.connect(stats)
}

func (p *Server) DisconnectStats() {
	p.statsRPC.disconnect()
}

func (p *Server) ConnectBinapi(binapi adapter.VppAPI) error {
	return p.binapiRPC.connect(binapi)
}

func (p *Server) DisconnectBinapi() {
	p.binapiRPC.disconnect()
}

func (p *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.rpc.ServeHTTP(w, req)
}

func (p *Server) ServeCodec(codec rpc.ServerCodec) {
	p.rpc.ServeCodec(codec)
}

func (p *Server) ServeConn(conn io.ReadWriteCloser) {
	p.rpc.ServeConn(conn)
}

func (p *Server) ListenAndServe(addr string) error {
	p.rpc.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	l, e := net.Listen("tcp", addr)
	if e != nil {
		return fmt.Errorf("listen failed: %v", e)
	}
	defer l.Close()

	log.Printf("proxy serving on: %v", addr)

	return http.Serve(l, nil)
}
