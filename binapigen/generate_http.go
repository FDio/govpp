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

package binapigen

import (
	"path"
	"strconv"
)

func init() {
	RegisterPlugin("http", GenerateHTTP)
}

// library dependencies
const (
	httpPkg = GoImportPath("net/http")
	OsPkg   = GoImportPath("os")
	jsonPkg = GoImportPath("encoding/json")
)

func GenerateHTTP(gen *Generator, file *File) *GenFile {
	if file.Service == nil {
		return nil
	}

	logf("----------------------------")
	logf(" Generate HTTP - %s", file.Desc.Name)
	logf("----------------------------")

	filename := path.Join(file.FilenamePrefix, file.Desc.Name+"_http"+generatedFilenameSuffix)
	g := gen.NewGenFile(filename, file)

	// file header
	genCodeGeneratedComment(g)
	g.P()
	g.P("package ", file.PackageName)
	g.P()

	// service HTTP handlers
	if len(file.Service.RPCs) > 0 {
		genHTTPHandler(g, file.Service)
	}

	return g
}

func genHTTPHandler(g *GenFile, svc *Service) {
	// constructor
	g.P("func HTTPHandler(rpc ", serviceApiName, ") ", httpPkg.Ident("Handler"), " {")
	g.P("	mux := ", httpPkg.Ident("NewServeMux"), "()")

	// http handlers for rpc
	for _, rpc := range svc.RPCs {
		if rpc.MsgReply == nil {
			continue
		}
		if rpc.VPP.Stream {
			continue // TODO: implement handler for streaming messages
		}
		g.P("mux.HandleFunc(", strconv.Quote("/"+rpc.VPP.Request), ", func(w ", httpPkg.Ident("ResponseWriter"), ", req *", httpPkg.Ident("Request"), ") {")
		g.P("var request = new(", rpc.MsgRequest.GoName, ")")
		if len(rpc.MsgRequest.Fields) > 0 {
			g.P("b, err := ", OsPkg.Ident("ReadAll"), "(req.Body)")
			g.P("if err != nil {")
			g.P("	", httpPkg.Ident("Error"), "(w, \"read body failed\", ", httpPkg.Ident("StatusBadRequest"), ")")
			g.P("	return")
			g.P("}")
			g.P("if err := ", jsonPkg.Ident("Unmarshal"), "(b, request); err != nil {")
			g.P("	", httpPkg.Ident("Error"), "(w, \"unmarshal data failed\", ", httpPkg.Ident("StatusBadRequest"), ")")
			g.P("	return")
			g.P("}")
		}
		g.P("reply, err := rpc.", rpc.GoName, "(req.Context(), request)")
		g.P("if err != nil {")
		g.P("	", httpPkg.Ident("Error"), "(w, \"request failed: \"+err.Error(), ", httpPkg.Ident("StatusInternalServerError"), ")")
		g.P("	return")
		g.P("}")
		g.P("rep, err := ", jsonPkg.Ident("MarshalIndent"), "(reply, \"\", \"  \")")
		g.P("if err != nil {")
		g.P("	", httpPkg.Ident("Error"), "(w, \"marshal failed: \"+err.Error(), ", httpPkg.Ident("StatusInternalServerError"), ")")
		g.P("	return")
		g.P("}")
		g.P("w.Write(rep)")
		g.P("})")
	}

	g.P("return ", httpPkg.Ident("HandlerFunc"), "(mux.ServeHTTP)")
	g.P("}")
	g.P()
}
