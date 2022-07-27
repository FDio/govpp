//  Copyright (c) 2020 Cisco and/or its affiliates.
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
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"go.fd.io/govpp"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/binapi/vlib"
	"go.fd.io/govpp/binapi/vpe"
	"go.fd.io/govpp/binapigen"
	"go.fd.io/govpp/binapigen/vppapi"
)

func main() {
	flag.Parse()

	apifiles, err := vppapi.ParseDir(vppapi.DefaultDir)
	if err != nil {
		log.Fatal(err)
	}

	switch cmd := flag.Arg(0); cmd {
	case "server":
		runServer(apifiles, ":7777")
	case "vppapi":
		showVPPAPI(os.Stdout, apifiles)
	case "vppapijson":
		if flag.NArg() == 1 {
			writeAsJSON(os.Stdout, apifiles)
		} else {
			f := flag.Arg(1)
			var found bool
			for _, apifile := range apifiles {
				if apifile.Name == f {
					writeAsJSON(os.Stdout, apifile)
					found = true
					break
				}
			}
			if !found {
				log.Fatalf("VPP API file %q not found", f)
			}
		}
	case "rpc":
		showRPC(apifiles)
	case "cli":
		args := flag.Args()
		if len(args) == 0 {
			args = []string{"?"}
		}
		sendCLI(args[1:])
	default:
		log.Fatalf("invalid command: %q", cmd)
	}

}

func writeAsJSON(w io.Writer, data interface{}) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
		return
	}
	if _, err := w.Write(b); err != nil {
		panic(err)
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

func showVPPAPI(out io.Writer, apifiles []*vppapi.File) {
	binapigen.SortFilesByImports(apifiles)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "API\tOPTIONS\tCRC\tPATH\tIMPORTED\tTYPES\t\n")

	for _, apifile := range apifiles {
		importedTypes := binapigen.ListImportedTypes(apifiles, apifile)
		var options []string
		for k, v := range apifile.Options {
			options = append(options, fmt.Sprintf("%s=%v", k, v))
		}
		imports := fmt.Sprintf("%d apis, %2d types", len(apifile.Imports), len(importedTypes))
		path := strings.TrimPrefix(apifile.Path, vppapi.DefaultDir+"/")
		types := fmt.Sprintf("%2d enum, %2d enumflag, %2d alias, %2d struct, %2d union, %2d msg",
			len(apifile.EnumTypes), len(apifile.EnumflagTypes), len(apifile.AliasTypes), len(apifile.StructTypes), len(apifile.UnionTypes), len(apifile.Messages))
		fmt.Fprintf(w, " %s\t%s\t%s\t%s\t%v\t%s\t\n",
			apifile.Name, strings.Join(options, " "), apifile.CRC, path, imports, types)
	}

	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(out, buf.String())
}

func sendCLI(args []string) {
	cmd := strings.Join(args, " ")
	fmt.Printf("# %s\n", cmd)

	conn, err := govpp.Connect("/run/vpp/api.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Disconnect()

	ch, err := conn.NewAPIChannel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	if err := ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
		log.Fatal(err)
	}

	client := vlib.NewServiceClient(conn)
	reply, err := client.CliInband(context.Background(), &vlib.CliInband{
		Cmd: cmd,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(reply.Reply)
}

func runServer(apifiles []*vppapi.File, addr string) {
	apiRoutes(apifiles, http.DefaultServeMux)

	conn, err := govpp.Connect(socketclient.DefaultSocketName)
	if err != nil {
		log.Fatal(err)
	}

	vpeRPC := vpe.NewServiceClient(conn)
	c := vpe.HTTPHandler(vpeRPC)

	http.Handle("/", c)

	log.Printf("listening on %v", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func apiRoutes(apifiles []*vppapi.File, mux *http.ServeMux) {
	for _, apifile := range apifiles {
		name := apifile.Name
		mux.HandleFunc("/vppapi/"+name, apiFileHandler(apifile))
		mux.HandleFunc("/raw/"+name, rawHandler(apifile))
		mux.HandleFunc("/rpc/"+name, rpcHandler(apifile))
	}
	mux.HandleFunc("/vppapi", apiHandler(apifiles))
}

func rpcHandler(apifile *vppapi.File) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		msgName := strings.TrimPrefix(req.URL.Path, "/rpc/"+apifile.Name+"/")
		if msgName == "" {
			http.Error(w, "no message name", 500)
			return
		}

		input, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		msgReq := make(map[string]interface{})
		err = json.Unmarshal(input, &msgReq)
		if err != nil {
			http.Error(w, err.Error(), 500)
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
			http.Error(w, "unknown message name: "+msgName, 500)
			return
		}

	}
}

func apiHandler(apifiles []*vppapi.File) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := json.MarshalIndent(apifiles, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(b)
	}
}

func apiFileHandler(apifile *vppapi.File) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := json.MarshalIndent(apifile, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(b)
	}
}

func rawHandler(apifile *vppapi.File) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := ioutil.ReadFile(apifile.Path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(b)
	}
}
