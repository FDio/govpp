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

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const (
	apiImportPath = "git.fd.io/govpp.git/api" // import path of the govpp API package
	inputFileExt  = ".api.json"               // file extension of the VPP binary API files
	outputFileExt = ".ba.go"                  // file extension of the Go generated files
)

// context is a structure storing data for code generation
type context struct {
	inputFile  string // input file with VPP API in JSON
	outputFile string // output file with generated Go package

	inputData []byte        // contents of the input file
	inputBuff *bytes.Buffer // contents of the input file currently being read
	inputLine int           // currently processed line in the input file

	moduleName  string // name of the source VPP module
	packageName string // name of the Go package being generated

	packageData *Package // parsed package data
}

// getContext returns context details of the code generation task
func getContext(inputFile, outputDir string) (*context, error) {
	if !strings.HasSuffix(inputFile, inputFileExt) {
		return nil, fmt.Errorf("invalid input file name: %q", inputFile)
	}

	ctx := &context{
		inputFile: inputFile,
	}

	// package name
	inputFileName := filepath.Base(inputFile)
	ctx.moduleName = inputFileName[:strings.Index(inputFileName, ".")]

	// alter package names for modules that are reserved keywords in Go
	switch ctx.moduleName {
	case "interface":
		ctx.packageName = "interfaces"
	case "map":
		ctx.packageName = "maps"
	default:
		ctx.packageName = ctx.moduleName
	}

	// output file
	packageDir := filepath.Join(outputDir, ctx.packageName)
	outputFileName := ctx.packageName + outputFileExt
	ctx.outputFile = filepath.Join(packageDir, outputFileName)

	return ctx, nil
}

// generatePackage generates code for the parsed package data and writes it into w
func generatePackage(ctx *context, w *bufio.Writer) error {
	logf("generating package %q", ctx.packageName)

	// generate file header
	generateHeader(ctx, w)

	// generate enums
	ctx.inputBuff = bytes.NewBuffer(ctx.inputData)
	ctx.inputLine = 0
	for _, enum := range ctx.packageData.Enums {
		generateEnum(ctx, w, &enum)
	}

	// generate types
	ctx.inputBuff = bytes.NewBuffer(ctx.inputData)
	ctx.inputLine = 0
	for _, typ := range ctx.packageData.Types {
		generateType(ctx, w, &typ)
	}

	// generate messages
	ctx.inputBuff = bytes.NewBuffer(ctx.inputData)
	ctx.inputLine = 0
	for _, msg := range ctx.packageData.Messages {
		generateMessage(ctx, w, &msg)
	}

	// TODO: generate unions

	// flush the data:
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing data to %s failed: %v", ctx.outputFile, err)
	}

	return nil
}

// generateHeader writes generated package header into w
func generateHeader(ctx *context, w io.Writer) {
	fmt.Fprintln(w, "// Code generated by GoVPP binapi-generator. DO NOT EDIT.")
	fmt.Fprintf(w, "// source: %s\n", ctx.inputFile)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "/*")
	fmt.Fprintf(w, "Package %s is a generated VPP binary API of the '%s' VPP module.\n", ctx.packageName, ctx.moduleName)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "It is generated from this file:")
	fmt.Fprintf(w, "\t%s\n", filepath.Base(ctx.inputFile))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "It contains these VPP binary API objects:")
	var printObjNum = func(obj string, num int) {
		if num > 0 {
			if num > 1 {
				obj += "s"
			}
			fmt.Fprintf(w, "\t%d %s\n", num, obj)
		}
	}
	printObjNum("message", len(ctx.packageData.Messages))
	printObjNum("type", len(ctx.packageData.Types))
	printObjNum("enum", len(ctx.packageData.Enums))
	printObjNum("union", len(ctx.packageData.Unions))
	fmt.Fprintln(w, "*/")
	fmt.Fprintf(w, "package %s\n", ctx.packageName)
	fmt.Fprintln(w)

	fmt.Fprintf(w, "import \"%s\"", apiImportPath)
	fmt.Fprintln(w)

	if *includeAPIVer {
		const APIVerConstName = "VlAPIVersion"
		fmt.Fprintf(w, "// %s represents version of the API.", APIVerConstName)
		fmt.Fprintf(w, "const %s = %v\n", APIVerConstName, ctx.packageData.APIVersion)
		fmt.Fprintln(w)
	}
}

// generateComment writes generated comment for the object into w
func generateComment(ctx *context, w io.Writer, goName string, vppName string, objKind string) {
	fmt.Fprintf(w, "// %s represents the VPP binary API %s '%s'.\n", goName, objKind, vppName)

	// print out the source of the generated object
	objFound := false
	objTitle := fmt.Sprintf(`"%s",`, vppName)
	var indent int
	for {
		line, err := ctx.inputBuff.ReadString('\n')
		if err != nil {
			break
		}
		ctx.inputLine++

		if !objFound {
			indent = strings.Index(line, objTitle)
			if indent == -1 {
				continue
			}
			// If no other non-whitespace character then we are at the message header.
			if trimmed := strings.TrimSpace(line); trimmed == objTitle {
				objFound = true
				fmt.Fprintf(w, "// Generated from '%s', line %d:\n", filepath.Base(ctx.inputFile), ctx.inputLine)
				fmt.Fprintln(w, "//")
			}
		} else {
			if strings.IndexFunc(line, isNotSpace) < indent {
				break // end of the object definition in JSON
			}
		}
		fmt.Fprint(w, "//", line)
	}

	fmt.Fprintln(w, "//")
}

// generateEnum writes generated code for the enum into w
func generateEnum(ctx *context, w io.Writer, enum *Enum) {
	name := camelCaseName(strings.Title(enum.Name))
	typ := binapiTypes[enum.Type]

	// generate enum comment
	generateComment(ctx, w, name, enum.Name, "enum")

	// generate enum definition
	fmt.Fprintf(w, "type %s %s\n", name, typ)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "const (")

	// generate enum entries
	for _, entry := range enum.Entries {
		fmt.Fprintf(w, "\t%s %s = %v\n", entry.Name, name, entry.Value)
	}

	fmt.Fprintln(w, ")")

	fmt.Fprintln(w)
}

// generateType writes generated code for the type into w
func generateType(ctx *context, w io.Writer, typ *Type) {
	name := camelCaseName(strings.Title(typ.Name))

	logf(" writing type %q (%s) with %d fields", typ.Name, name, len(typ.Fields))

	// generate struct comment
	generateComment(ctx, w, name, typ.Name, "type")

	// generate struct definition
	fmt.Fprintln(w, "type", name, "struct {")

	// generate struct fields
	for _, field := range typ.Fields {
		// skip internal fields
		switch strings.ToLower(field.Name) {
		case "crc", "_vl_msg_id":
			continue
		}

		generateField(ctx, w, &field)
	}

	// generate end of the struct
	fmt.Fprintln(w, "}")

	// generate name getter
	generateTypeNameGetter(w, name, typ.Name)

	// generate CRC getter
	generateCrcGetter(w, name, typ.CRC)

	fmt.Fprintln(w)
}

// generateMessage writes generated code for the message into w
func generateMessage(ctx *context, w io.Writer, msg *Message) {
	name := camelCaseName(strings.Title(msg.Name))

	logf(" writing message %q (%s) with %d fields", msg.Name, name, len(msg.Fields))

	// generate struct comment
	generateComment(ctx, w, name, msg.Name, "message")

	// generate struct definition
	fmt.Fprintln(w, "type", name, "struct {")

	msgType := otherMessage
	wasClientIndex := false

	// generate struct fields
	n := 0
	for i, field := range msg.Fields {
		if i == 2 {
			if field.Name == "client_index" {
				// "client_index" as the second member, this might be an event message or a request
				msgType = eventMessage
				wasClientIndex = true
			} else if field.Name == "context" {
				// reply needs "context" as the second member
				msgType = replyMessage
			}
		} else if i == 3 {
			if wasClientIndex && field.Name == "context" {
				// request needs "client_index" as the second member and "context" as the third member
				msgType = requestMessage
			}
		}

		// skip internal fields
		switch strings.ToLower(field.Name) {
		case "crc", "_vl_msg_id":
			continue
		case "client_index", "context":
			if n == 0 {
				continue
			}
		}
		n++

		generateField(ctx, w, &field)
	}

	// generate end of the struct
	fmt.Fprintln(w, "}")

	// generate name getter
	generateMessageNameGetter(w, name, msg.Name)

	// generate CRC getter
	generateCrcGetter(w, name, msg.CRC)

	// generate message type getter method
	generateMessageTypeGetter(w, name, msgType)

	// generate message factory
	generateMessageFactory(w, name)
}

// generateField writes generated code for the field into w
func generateField(ctx *context, w io.Writer, field *Field) {
	fieldName := strings.TrimPrefix(field.Name, "_")
	fieldName = camelCaseName(strings.Title(fieldName))

	isArray := field.Length > 0 || field.SizeFrom != ""
	dataType := convertToGoType(ctx, field.Type, isArray)

	fieldType := dataType
	if isArray {
		fieldType = "[]" + dataType
	}
	fmt.Fprintf(w, "\t%s %s", fieldName, fieldType)

	if field.Length > 0 {
		// fixed size array
		fmt.Fprintf(w, "\t`struc:\"[%d]%s\"`", uint64(field.Length), dataType)
	} else if field.SizeFrom != "" {
		// variable sized array
		sizeFromName := camelCaseName(strings.Title(field.SizeFrom))
		fmt.Fprintf(w, "\t`struc:\"sizefrom=%s\"`", sizeFromName)
	}

	fmt.Fprintln(w)
}

// generateMessageNameGetter generates getter for original VPP message name into the provider writer
func generateMessageNameGetter(w io.Writer, structName string, msgName string) {
	fmt.Fprintln(w, "func (*"+structName+") GetMessageName() string {")
	fmt.Fprintln(w, "\treturn \""+msgName+"\"")
	fmt.Fprintln(w, "}")
}

// generateTypeNameGetter generates getter for original VPP type name into the provider writer
func generateTypeNameGetter(w io.Writer, structName string, msgName string) {
	fmt.Fprintln(w, "func (*"+structName+") GetTypeName() string {")
	fmt.Fprintln(w, "\treturn \""+msgName+"\"")
	fmt.Fprintln(w, "}")
}

// generateMessageTypeGetter generates message factory for the generated message into the provider writer
func generateMessageTypeGetter(w io.Writer, structName string, msgType MessageType) {
	fmt.Fprintln(w, "func (*"+structName+") GetMessageType() api.MessageType {")
	if msgType == requestMessage {
		fmt.Fprintln(w, "\treturn api.RequestMessage")
	} else if msgType == replyMessage {
		fmt.Fprintln(w, "\treturn api.ReplyMessage")
	} else if msgType == eventMessage {
		fmt.Fprintln(w, "\treturn api.EventMessage")
	} else {
		fmt.Fprintln(w, "\treturn api.OtherMessage")
	}
	fmt.Fprintln(w, "}")
}

// generateCrcGetter generates getter for CRC checksum of the message definition into the provider writer
func generateCrcGetter(w io.Writer, structName string, crc string) {
	crc = strings.TrimPrefix(crc, "0x")
	fmt.Fprintln(w, "func (*"+structName+") GetCrcString() string {")
	fmt.Fprintln(w, "\treturn \""+crc+"\"")
	fmt.Fprintln(w, "}")
}

// generateMessageFactory generates message factory for the generated message into the provider writer
func generateMessageFactory(w io.Writer, structName string) {
	fmt.Fprintln(w, "func New"+structName+"() api.Message {")
	fmt.Fprintln(w, "\treturn &"+structName+"{}")
	fmt.Fprintln(w, "}")
}
