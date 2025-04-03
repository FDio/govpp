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
//  - add option to allow server to start without VPP running
//  - add option to set list of VPP APIs to serve
//  - wait for SIGTERM/SIGINT signal and shutdown the server gracefully
//  - add home page (/index.html) providing references and example links

const (
	DefaultHttpServiceAddress = ":8000"
)

type HttpCmdOptions struct {
	Input     string
	ApiSocket string
	Address   string
}

func newHttpCmd(Cli) *cobra.Command {
	var opts = HttpCmdOptions{
		ApiSocket: socketclient.DefaultSocketName,
		Address:   DefaultHttpServiceAddress,
	}
	cmd := &cobra.Command{
		Use:   "http",
		Short: "VPP API as HTTP service",
		Long:  "Serves VPP API via HTTP service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHttpCmd(opts)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.Input, "input", opts.Input, "Input for VPP API (e.g. path to VPP API directory, local VPP repo)")
	cmd.PersistentFlags().StringVar(&opts.ApiSocket, "apisock", opts.ApiSocket, "Path to VPP API socket")
	cmd.PersistentFlags().StringVar(&opts.Address, "addr", opts.Address, "HTTP service address")

	return cmd
}

func runHttpCmd(opts HttpCmdOptions) error {
	vppInput, err := resolveVppInput(opts.Input)
	if err != nil {
		return err
	}

	logrus.Debugf("connecting to VPP socket %s", opts.ApiSocket)

	conn, err := govpp.Connect(opts.ApiSocket)
	if err != nil {
		return fmt.Errorf("govpp.Connect: %w", err)
	}

	serveMux := http.NewServeMux()

	setupHttpAPIHandlers(vppInput.Schema.Files, serveMux)

	// TODO: register all api files automatically (requires some regisry for handlers or apifiles
	c := vpe.HTTPHandler(vpe.NewServiceClient(conn))
	//c := memclnt.HTTPHandler(memclnt.NewServiceClient(conn))

	serveMux.Handle("/", c)

	logrus.Infof("HTTP server listening on: %v", opts.Address)

	if err := http.ListenAndServe(opts.Address, serveMux); err != nil {
		return err
	}

	return nil
}

func setupHttpAPIHandlers(apifiles []vppapi.File, mux *http.ServeMux) {
	for _, apifile := range apifiles {
		file := apifile
		name := file.Name
		mux.HandleFunc("/api/"+name, apiFileHandler(&file))
		mux.HandleFunc("/raw/"+name, apiRawHandler(&file))
		mux.HandleFunc("/vpp/"+name, reqHandler(&file))
	}
	mux.HandleFunc("/api", apiFilesHandler(apifiles))
}

func reqHandler(apifile *vppapi.File) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		// extract message name
		msgName := strings.TrimPrefix(req.URL.Path, "/vpp/"+apifile.Name+"/")
		if msgName == "" {
			http.Error(w, "missing message name in URL", http.StatusBadRequest)
			return
		}

		// find message
		var msg *vppapi.Message
		for _, m := range apifile.Messages {
			if m.Name == msgName {
				msg = &m
				break
			}
		}
		if msg == nil {
			http.Error(w, "message not found : "+msgName, http.StatusNotFound)
			return
		}

		switch req.Method {
		case http.MethodPost:
			reqData := make(map[string]interface{})

			// parse body data
			body, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = json.Unmarshal(body, &reqData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// TODO: unmarshal body data into message, send request to VPP, marshal it and send as response
			http.Error(w, "Sending requests is not implemented yet", http.StatusNotImplemented)
		case http.MethodGet:
			b, err := json.MarshalIndent(msg, "", "  ")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = w.Write(b)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "GET or POST are allowed only", http.StatusMethodNotAllowed)
		}
	}
}

func apiFilesHandler(apifiles []vppapi.File) func(http.ResponseWriter, *http.Request) {
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

func apiRawHandler(apifile *vppapi.File) func(http.ResponseWriter, *http.Request) {
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
