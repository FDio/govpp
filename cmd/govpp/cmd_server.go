//  Copyright (c) 2023 Cisco and/or its affiliates.
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.fd.io/govpp"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/binapigen/vppapi"
)

// TODO:
// - add option to allow server to start without VPP running

const (
	DefaultServerCmdAddress = ":7777"
)

type ServerCmdOptions struct {
	Input      string
	ApiSocket  string
	ServerAddr string
}

func newServerCmd() *cobra.Command {
	var (
		opts = ServerCmdOptions{
			ApiSocket:  socketclient.DefaultSocketName,
			ServerAddr: DefaultServerCmdAddress,
		}
	)
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: add option to set list of VPP APIs to serve
			if opts.Input == "" {
				opts.Input = resolveVppApiInput()
			}
			return runServer(opts)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.Input, "input", opts.Input, "Input for VPP API (e.g. path to VPP API directory, local VPP repo)")
	cmd.PersistentFlags().StringVar(&opts.ApiSocket, "apisock", opts.ApiSocket, "Path to VPP API socket")
	cmd.PersistentFlags().StringVar(&opts.ServerAddr, "addr", opts.ServerAddr, "Address for server to listen on")

	return cmd
}

func runServer(opts ServerCmdOptions) error {
	// Input
	vppInput, err := vppapi.ResolveVppInput(opts.Input)
	if err != nil {
		return err
	}

	logrus.Tracef("VPP input:\n - API dir: %s\n - VPP Version: %s\n - Files: %v",
		vppInput.ApiDirectory, vppInput.VppVersion, len(vppInput.ApiFiles))

	apifiles := vppInput.ApiFiles

	addr := opts.ServerAddr
	serveMux := http.NewServeMux()

	setupServerAPIHandlers(apifiles, serveMux)

	conn, err := govpp.Connect(opts.ApiSocket)
	if err != nil {
		return fmt.Errorf("govpp.Connect: %w", err)
	}

	c := vpe.HTTPHandler(vpe.NewServiceClient(conn))
	//c := memclnt.HTTPHandler(memclnt.NewServiceClient(conn))

	serveMux.Handle("/", c)

	logrus.Infof("server listening on: %v", addr)

	if err := http.ListenAndServe(addr, serveMux); err != nil {
		return err
	}

	// TODO: wait for SIGTERM/SIGINT signal and shutdown the server gracefully

	return nil
}

func setupServerAPIHandlers(apifiles []vppapi.File, mux *http.ServeMux) {
	for _, apifile := range apifiles {
		file := apifile
		name := apifile.Name
		mux.HandleFunc("/api/"+name, apiFileHandler(&file))
		mux.HandleFunc("/raw/"+name, rawHandler(&file))
		mux.HandleFunc("/rpc/"+name, rpcHandler(&file))
	}
	mux.HandleFunc("/api", apiHandler(apifiles))
	// TODO: add home page (/index.html) providing references and example links
}

func rpcHandler(apifile *vppapi.File) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		// extract message name
		msgName := strings.TrimPrefix(req.URL.Path, "/rpc/"+apifile.Name+"/")
		if msgName == "" {
			http.Error(w, "no message name", http.StatusNotFound)
			return
		}

		// parse input data
		input, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		msgReq := make(map[string]interface{})
		err = json.Unmarshal(input, &msgReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var msg *vppapi.Message
		for _, m := range apifile.Messages {
			if m.Name == msgName {
				msg = &m
				break
			}
		}
		if msg == nil {
			http.Error(w, "unknown message name: "+msgName, http.StatusInternalServerError)
			return
		}

	}
}

func showRPC(apifiles []*vppapi.File) {
	for _, apifile := range apifiles {
		fmt.Printf("%s.api\n", apifile.Name)
		if apifile.Service == nil {
			continue
		}
		for _, rpc := range apifile.Service.RPCs {
			req := rpc.Request
			reply := rpc.Reply
			if rpc.Stream {
				reply = "stream " + reply
			}
			fmt.Printf(" rpc (%s) --> (%s)\n", req, reply)
		}
	}
}

func apiHandler(apifiles []vppapi.File) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := json.MarshalIndent(apifiles, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func apiFileHandler(apifile *vppapi.File) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := json.MarshalIndent(apifile, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func rawHandler(apifile *vppapi.File) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := os.ReadFile(apifile.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
