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
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"git.fd.io/govpp.git/version"
)

// generatedCodeVersion indicates a version of the generated code.
// It is incremented whenever an incompatibility between the generated code and
// GoVPP api package is introduced; the generated code references
// a constant, api.GoVppAPIPackageIsVersionN (where N is generatedCodeVersion).
const generatedCodeVersion = 2

// message field names
const (
	msgIdField       = "_vl_msg_id"
	clientIndexField = "client_index"
	contextField     = "context"
	//retvalField      = "retval"
)

const (
	outputFileExt = ".ba.go" // file extension of the Go generated files
	rpcFileSuffix = "_rpc"   // file name suffix for the RPC services

	constModuleName = "ModuleName" // module name constant
	constAPIVersion = "APIVersion" // API version constant
	constVersionCrc = "VersionCrc" // version CRC constant

	unionDataField = "XXX_UnionData" // name for the union data field

	serviceApiName    = "RPCService"    // name for the RPC service interface
	serviceImplName   = "serviceClient" // name for the RPC service implementation
	serviceClientName = "ServiceClient" // name for the RPC service client

	//serviceDescType = "ServiceDesc"             // name for service descriptor type
	//serviceDescName = "_ServiceRPC_serviceDesc" // name for service descriptor var
)

// MessageType represents the type of a VPP message
type MessageType int

const (
	requestMessage MessageType = iota // VPP request message
	replyMessage                      // VPP reply message
	eventMessage                      // VPP event message
	otherMessage                      // other VPP message
)

type Options struct {
	//APIFiles   []*vppapi.File
	VPPVersion string // version of VPP that produced API files

	FilesToGenerate []string
	ImportPrefix    string // defines import path prefix for importing types

	IncludeAPIVersion  bool // include constant with API version string
	IncludeComments    bool // include parts of original source in comments
	IncludeBinapiNames bool // include binary API names as struct tag
	IncludeServices    bool // include service interface with client implementation
	IncludeVppVersion  bool // include info about used VPP version
}

// Context holds a context data for generating code.
type Context struct {
	Options

	// contents of the input file
	inputData []byte

	inputFile     string // input file with VPP API in JSON
	outputFile    string // output file with generated binapi package
	outputFileRPC string // output file with generated RPC package

	moduleName  string // name of the source VPP module
	packageName string // name of the Go package being generated

	packageData *File // parsed package data
	RefMap      map[string]string
}

// newContext returns context details of the code generation task
func newContext(inputFile, outputDir string) (*Context, error) {
	ctx := &Context{
		inputFile: inputFile,
		RefMap:    make(map[string]string),
	}

	// package name
	inputFileName := filepath.Base(inputFile)
	ctx.moduleName = inputFileName[:strings.Index(inputFileName, ".")]

	// alter package names for files that are reserved keywords in Go
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
	ctx.outputFile = filepath.Join(packageDir, ctx.packageName+outputFileExt)
	ctx.outputFileRPC = filepath.Join(packageDir, ctx.packageName+rpcFileSuffix+outputFileExt)

	return ctx, nil
}

func (c *Context) generatesRPC() bool {
	return c.IncludeServices && len(c.packageData.Service.RPCs) > 0
}

func generatePackage(ctx *Context, w io.Writer) error {
	logf("----------------------------")
	logf("generating binapi package %q", ctx.packageName)
	logf("----------------------------")

	generateHeader(ctx, w)
	generateImports(ctx, w)

	// generate module desc
	fmt.Fprintln(w, "const (")
	fmt.Fprintf(w, "\t// %s is the name of this module.\n", constModuleName)
	fmt.Fprintf(w, "\t%s = \"%s\"\n", constModuleName, ctx.moduleName)

	if ctx.IncludeAPIVersion {
		fmt.Fprintf(w, "\t// %s is the API version of this module.\n", constAPIVersion)
		fmt.Fprintf(w, "\t%s = \"%s\"\n", constAPIVersion, ctx.packageData.Version)
		fmt.Fprintf(w, "\t// %s is the CRC of this module.\n", constVersionCrc)
		fmt.Fprintf(w, "\t%s = %v\n", constVersionCrc, ctx.packageData.CRC)
	}
	fmt.Fprintln(w, ")")
	fmt.Fprintln(w)

	// generate enums
	if len(ctx.packageData.Enums) > 0 {
		for _, enum := range ctx.packageData.Enums {
			/*if imp, ok := ctx.packageData.Imports[enum.Name]; ok {
				generateImportedAlias(ctx, w, enum.Name, imp)
				continue
			}*/
			generateEnum(ctx, w, enum)
		}
	}

	// generate aliases
	if len(ctx.packageData.Aliases) > 0 {
		for _, alias := range ctx.packageData.Aliases {
			/*if imp, ok := ctx.packageData.Imports[alias.Name]; ok {
				generateImportedAlias(ctx, w, alias.Name, imp)
				continue
			}*/
			generateAlias(ctx, w, alias)
		}
	}

	// generate types
	if len(ctx.packageData.Structs) > 0 {
		for _, typ := range ctx.packageData.Structs {
			/*if imp, ok := ctx.packageData.Imports[typ.Name]; ok {
				generateImportedAlias(ctx, w, typ.Name, imp)
				continue
			}*/
			generateStruct(ctx, w, typ)
		}
	}

	// generate unions
	if len(ctx.packageData.Unions) > 0 {
		for _, union := range ctx.packageData.Unions {
			/*if imp, ok := ctx.packageData.Imports[union.Name]; ok {
				generateImportedAlias(ctx, w, union.Name, imp)
				continue
			}*/
			generateUnion(ctx, w, union)
		}
	}

	// generate messages
	if len(ctx.packageData.Messages) > 0 {
		for _, msg := range ctx.packageData.Messages {
			generateMessage(ctx, w, &msg)
		}

		initFnName := fmt.Sprintf("file_%s_binapi_init", ctx.packageName)

		// generate message registrations
		fmt.Fprintf(w, "func init() { %s() }\n", initFnName)
		fmt.Fprintf(w, "func %s() {\n", initFnName)
		for _, msg := range ctx.packageData.Messages {
			name := camelCaseName(msg.Name)
			fmt.Fprintf(w, "\tapi.RegisterMessage((*%s)(nil), \"%s\")\n", name, ctx.moduleName+"."+name)
		}
		fmt.Fprintln(w, "}")
		fmt.Fprintln(w)

		// generate list of messages
		fmt.Fprintf(w, "// Messages returns list of all messages in this module.\n")
		fmt.Fprintln(w, "func AllMessages() []api.Message {")
		fmt.Fprintln(w, "\treturn []api.Message{")
		for _, msg := range ctx.packageData.Messages {
			name := camelCaseName(msg.Name)
			fmt.Fprintf(w, "\t(*%s)(nil),\n", name)
		}
		fmt.Fprintln(w, "}")
		fmt.Fprintln(w, "}")
	}

	generateFooter(ctx, w)

	return nil
}

func generatePackageRPC(ctx *Context, w io.Writer) error {
	logf("----------------------------")
	logf("generating binapi RPC package %q", ctx.packageName)
	logf("----------------------------")

	fmt.Fprintln(w, "// Code generated by GoVPP's binapi-generator. DO NOT EDIT.")
	fmt.Fprintln(w)

	fmt.Fprintf(w, "package %s\n", ctx.packageName)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "import (")
	fmt.Fprintln(w, `	"context"`)
	fmt.Fprintln(w, `	"io"`)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "\tapi \"%s\"\n", "git.fd.io/govpp.git/api")
	fmt.Fprintln(w, ")")
	fmt.Fprintln(w)

	// generate services
	if len(ctx.packageData.Service.RPCs) > 0 {
		generateServiceMethods(ctx, w, ctx.packageData.Service.RPCs)
	}

	// generate message registrations
	/*fmt.Fprintln(w, "var _RPCService_desc = api.RPCDesc{")

	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)*/

	fmt.Fprintf(w, "// Reference imports to suppress errors if they are not otherwise used.\n")
	fmt.Fprintf(w, "var _ = api.RegisterMessage\n")
	fmt.Fprintf(w, "var _ = context.Background\n")
	fmt.Fprintf(w, "var _ = io.Copy\n")

	return nil
}

func generateHeader(ctx *Context, w io.Writer) {
	fmt.Fprintln(w, "// Code generated by GoVPP's binapi-generator. DO NOT EDIT.")
	fmt.Fprintln(w, "// versions:")
	fmt.Fprintf(w, "//  binapi-generator: %s\n", version.Version())
	if ctx.IncludeVppVersion {
		fmt.Fprintf(w, "//  VPP:              %s\n", ctx.VPPVersion)
	}
	fmt.Fprintf(w, "// source: %s\n", ctx.inputFile)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "/*")
	fmt.Fprintf(w, "Package %s contains generated code for VPP binary API defined by %s.api (version %s).\n", ctx.packageName, ctx.moduleName, ctx.packageData.Version)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "It consists of:")
	printObjNum := func(obj string, num int) {
		if num > 0 {
			if num > 1 {
				if strings.HasSuffix(obj, "s") {

					obj += "es"
				} else {
					obj += "s"
				}
			}
			fmt.Fprintf(w, "\t%3d %s\n", num, obj)
		}
	}
	printObjNum("RPC", len(ctx.packageData.Service.RPCs))
	printObjNum("alias", len(ctx.packageData.AliasTypes))
	printObjNum("enum", len(ctx.packageData.EnumTypes))
	printObjNum("message", len(ctx.packageData.Messages))
	printObjNum("type", len(ctx.packageData.StructTypes))
	printObjNum("union", len(ctx.packageData.UnionTypes))
	fmt.Fprintln(w, "*/")
	fmt.Fprintf(w, "package %s\n", ctx.packageName)
	fmt.Fprintln(w)
}

func generateImports(ctx *Context, w io.Writer) {
	fmt.Fprintln(w, "import (")
	fmt.Fprintln(w, `	"bytes"`)
	fmt.Fprintln(w, `	"context"`)
	fmt.Fprintln(w, `	"encoding/binary"`)
	fmt.Fprintln(w, `	"io"`)
	fmt.Fprintln(w, `	"math"`)
	fmt.Fprintln(w, `	"strconv"`)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "\tapi \"%s\"\n", "git.fd.io/govpp.git/api")
	fmt.Fprintf(w, "\tcodec \"%s\"\n", "git.fd.io/govpp.git/codec")
	fmt.Fprintf(w, "\tstruc \"%s\"\n", "github.com/lunixbochs/struc")
	//fmt.Fprintf(w, "\tstruc \"%s\"\n", "github.com/lunixbochs/struc")
	if len(ctx.packageData.Imports) > 0 {
		fmt.Fprintln(w)
		for _, imp := range getImports(ctx) {
			importPath := path.Join(ctx.ImportPrefix, imp)
			if importPath == "" {
				importPath = getImportPath(filepath.Dir(ctx.outputFile), imp)
			}
			fmt.Fprintf(w, "\t%s \"%s\"\n", imp, strings.TrimSpace(importPath))
		}
	}
	fmt.Fprintln(w, ")")
	fmt.Fprintln(w)

	fmt.Fprintln(w, "// This is a compile-time assertion to ensure that this generated file")
	fmt.Fprintln(w, "// is compatible with the GoVPP api package it is being compiled against.")
	fmt.Fprintln(w, "// A compilation error at this line likely means your copy of the")
	fmt.Fprintln(w, "// GoVPP api package needs to be updated.")
	fmt.Fprintf(w, "const _ = api.GoVppAPIPackageIsVersion%d // please upgrade the GoVPP api package\n", generatedCodeVersion)
	fmt.Fprintln(w)
}

func getImportPath(outputDir string, pkg string) string {
	absPath, _ := filepath.Abs(filepath.Join(outputDir, "..", pkg))
	cmd := exec.Command("go", "list", absPath)
	var errbuf, outbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	if err := cmd.Run(); err != nil {
		fmt.Printf("ERR: %v\n", errbuf.String())
		panic(err)
	}
	return outbuf.String()
}

func getImports(ctx *Context) (imports []string) {
	impmap := map[string]struct{}{}
	for _, imp := range ctx.packageData.Imports {
		if _, ok := impmap[imp]; !ok {
			imports = append(imports, imp)
			impmap[imp] = struct{}{}
		}
	}
	sort.Strings(imports)
	return imports
}

func generateFooter(ctx *Context, w io.Writer) {
	fmt.Fprintf(w, "// Reference imports to suppress errors if they are not otherwise used.\n")
	fmt.Fprintf(w, "var _ = api.RegisterMessage\n")
	fmt.Fprintf(w, "var _ = codec.DecodeString\n")
	fmt.Fprintf(w, "var _ = bytes.NewBuffer\n")
	fmt.Fprintf(w, "var _ = context.Background\n")
	fmt.Fprintf(w, "var _ = io.Copy\n")
	fmt.Fprintf(w, "var _ = strconv.Itoa\n")
	fmt.Fprintf(w, "var _ = struc.Pack\n")
	fmt.Fprintf(w, "var _ = binary.BigEndian\n")
	fmt.Fprintf(w, "var _ = math.Float32bits\n")
}

func generateComment(ctx *Context, w io.Writer, goName string, vppName string, objKind string) {
	if objKind == "service" {
		fmt.Fprintf(w, "// %s represents RPC service API for %s module.\n", goName, ctx.moduleName)
	} else {
		fmt.Fprintf(w, "// %s represents VPP binary API %s '%s'.\n", goName, objKind, vppName)
	}

	if !ctx.IncludeComments {
		return
	}

	var isNotSpace = func(r rune) bool {
		return !unicode.IsSpace(r)
	}

	// print out the source of the generated object
	mapType := false
	objFound := false
	objTitle := fmt.Sprintf(`"%s",`, vppName)
	switch objKind {
	case "alias", "service":
		objTitle = fmt.Sprintf(`"%s": {`, vppName)
		mapType = true
	}

	inputBuff := bytes.NewBuffer(ctx.inputData)
	inputLine := 0

	var trimIndent string
	var indent int
	for {
		line, err := inputBuff.ReadString('\n')
		if err != nil {
			break
		}
		inputLine++

		noSpaceAt := strings.IndexFunc(line, isNotSpace)
		if !objFound {
			indent = strings.Index(line, objTitle)
			if indent == -1 {
				continue
			}
			trimIndent = line[:indent]
			// If no other non-whitespace character then we are at the message header.
			if trimmed := strings.TrimSpace(line); trimmed == objTitle {
				objFound = true
				fmt.Fprintln(w, "//")
			}
		} else if noSpaceAt < indent {
			break // end of the definition in JSON for array types
		} else if objFound && mapType && noSpaceAt <= indent {
			fmt.Fprintf(w, "//\t%s", strings.TrimPrefix(line, trimIndent))
			break // end of the definition in JSON for map types (aliases, services)
		}
		fmt.Fprintf(w, "//\t%s", strings.TrimPrefix(line, trimIndent))
	}

	fmt.Fprintln(w, "//")
}

func generateEnum(ctx *Context, w io.Writer, enum *Enum) {
	name := camelCaseName(enum.Name)
	typ := binapiTypes[enum.Type]

	logf(" writing enum %q (%s) with %d entries", enum.Name, name, len(enum.Entries))

	// generate enum comment
	generateComment(ctx, w, name, enum.Name, "enum")

	// generate enum definition
	fmt.Fprintf(w, "type %s %s\n", name, typ)
	fmt.Fprintln(w)

	// generate enum entries
	fmt.Fprintln(w, "const (")
	for _, entry := range enum.Entries {
		fmt.Fprintf(w, "\t%s %s = %v\n", entry.Name, name, entry.Value)
	}
	fmt.Fprintln(w, ")")
	fmt.Fprintln(w)

	// generate enum conversion maps
	fmt.Fprintln(w, "var (")
	fmt.Fprintf(w, "%s_name = map[%s]string{\n", name, typ)
	for _, entry := range enum.Entries {
		fmt.Fprintf(w, "\t%v: \"%s\",\n", entry.Value, entry.Name)
	}
	fmt.Fprintln(w, "}")
	fmt.Fprintf(w, "%s_value = map[string]%s{\n", name, typ)
	for _, entry := range enum.Entries {
		fmt.Fprintf(w, "\t\"%s\": %v,\n", entry.Name, entry.Value)
	}
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w, ")")
	fmt.Fprintln(w)

	fmt.Fprintf(w, "func (x %s) String() string {\n", name)
	fmt.Fprintf(w, "\ts, ok := %s_name[%s(x)]\n", name, typ)
	fmt.Fprintf(w, "\tif ok { return s }\n")
	fmt.Fprintf(w, "\treturn \"%s(\" + strconv.Itoa(int(x)) + \")\"\n", name)
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)
}

func generateImportedAlias(ctx *Context, w io.Writer, tName string, imp string) {
	name := camelCaseName(tName)

	fmt.Fprintf(w, "type %s = %s.%s\n", name, imp, name)

	fmt.Fprintln(w)
}

func generateAlias(ctx *Context, w io.Writer, alias *Alias) {
	name := camelCaseName(alias.Name)

	logf(" writing type %q (%s), length: %d", alias.Name, name, alias.Length)

	// generate struct comment
	generateComment(ctx, w, name, alias.Name, "alias")

	// generate struct definition
	fmt.Fprintf(w, "type %s ", name)

	if alias.Length > 0 {
		fmt.Fprintf(w, "[%d]", alias.Length)
	}

	dataType := convertToGoType(ctx.packageData, alias.Type)
	fmt.Fprintf(w, "%s\n", dataType)

	fmt.Fprintln(w)
}

func generateUnion(ctx *Context, w io.Writer, union *Union) {
	name := camelCaseName(union.Name)

	logf(" writing union %q (%s) with %d fields", union.Name, name, len(union.Fields))

	// generate struct comment
	generateComment(ctx, w, name, union.Name, "union")

	// generate struct definition
	fmt.Fprintln(w, "type", name, "struct {")

	// maximum size for union
	maxSize := getUnionSize(ctx.packageData, union)

	// generate data field
	fmt.Fprintf(w, "\t%s [%d]byte\n", unionDataField, maxSize)

	// generate end of the struct
	fmt.Fprintln(w, "}")

	// generate name getter
	generateTypeNameGetter(w, name, union.Name)

	// generate getters for fields
	for _, field := range union.Fields {
		fieldName := camelCaseName(field.Name)
		fieldType := convertToGoType(ctx.packageData, field.Type)
		generateUnionGetterSetter(w, name, fieldName, fieldType)
	}

	// generate union methods
	//generateUnionMethods(w, name)

	fmt.Fprintln(w)
}

// generateUnionMethods generates methods that implement struc.Custom
// interface to allow having XXX_uniondata field unexported
// TODO: do more testing when unions are actually used in some messages
/*func generateUnionMethods(w io.Writer, structName string) {
	// generate struc.Custom implementation for union
	fmt.Fprintf(w, `
func (u *%[1]s) Pack(p []byte, opt *struc.Options) (int, error) {
	var b = new(bytes.Buffer)
	if err := struc.PackWithOptions(b, u.union_data, opt); err != nil {
		return 0, err
	}
	copy(p, b.Bytes())
	return b.Len(), nil
}
func (u *%[1]s) Unpack(r io.Reader, length int, opt *struc.Options) error {
	return struc.UnpackWithOptions(r, u.union_data[:], opt)
}
func (u *%[1]s) Size(opt *struc.Options) int {
	return len(u.union_data)
}
func (u *%[1]s) String() string {
	return string(u.union_data[:])
}
`, structName)
}*/

/*func generateUnionGetterSetterNew(w io.Writer, structName string, getterField, getterStruct string) {
	fmt.Fprintf(w, `
func %[1]s%[2]s(a %[3]s) (u %[1]s) {
	u.Set%[2]s(a)
	return
}
func (u *%[1]s) Set%[2]s(a %[3]s) {
	copy(u.%[4]s[:], a[:])
}
func (u *%[1]s) Get%[2]s() (a %[3]s) {
	copy(a[:], u.%[4]s[:])
	return
}
`, structName, getterField, getterStruct, unionDataField)
}*/

func generateUnionGetterSetter(w io.Writer, structName string, getterField, getterStruct string) {
	fmt.Fprintf(w, `
func %[1]s%[2]s(a %[3]s) (u %[1]s) {
	u.Set%[2]s(a)
	return
}
func (u *%[1]s) Set%[2]s(a %[3]s) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.%[4]s[:], b.Bytes())
}
func (u *%[1]s) Get%[2]s() (a %[3]s) {
	var b = bytes.NewReader(u.%[4]s[:])
	struc.Unpack(b, &a)
	return
}
`, structName, getterField, getterStruct, unionDataField)
}

func generateStruct(ctx *Context, w io.Writer, typ *Struct) {
	name := camelCaseName(typ.Name)

	logf(" writing type %q (%s) with %d fields", typ.Name, name, len(typ.Fields))

	// generate struct comment
	generateComment(ctx, w, name, typ.Name, "type")

	// generate struct definition
	fmt.Fprintf(w, "type %s struct {\n", name)

	// generate struct fields
	for i := range typ.Fields {
		// skip internal fields
		/*switch strings.ToLower(field.Name) {
		case crcField, msgIdField:
			continue
		}*/

		generateField(ctx, w, typ.Fields, i)
	}

	// generate end of the struct
	fmt.Fprintln(w, "}")

	// generate name getter
	generateTypeNameGetter(w, name, typ.Name)

	fmt.Fprintln(w)
}

func generateMessage(ctx *Context, w io.Writer, msg *Message) {
	name := camelCaseName(msg.Name)

	logf(" writing message %q (%s) with %d fields", msg.Name, name, len(msg.Fields))

	// generate struct comment
	generateComment(ctx, w, name, msg.Name, "message")

	// generate struct definition
	fmt.Fprintf(w, "type %s struct {", name)

	msgType := otherMessage
	wasClientIndex := false

	// generate struct fields
	n := 0
	for i, field := range msg.Fields {
		if i == 1 {
			if field.Name == clientIndexField {
				// "client_index" as the second member,
				// this might be an event message or a request
				msgType = eventMessage
				wasClientIndex = true
			} else if field.Name == contextField {
				// reply needs "context" as the second member
				msgType = replyMessage
			}
		} else if i == 2 {
			if wasClientIndex && field.Name == contextField {
				// request needs "client_index" as the second member
				// and "context" as the third member
				msgType = requestMessage
			}
		}

		// skip internal fields
		switch strings.ToLower(field.Name) {
		case /*crcField,*/ msgIdField:
			continue
		case clientIndexField, contextField:
			if n == 0 {
				continue
			}
		}
		n++
		if n == 1 {
			fmt.Fprintln(w)
		}

		generateField(ctx, w, msg.Fields, i)
	}

	// generate end of the struct
	fmt.Fprintln(w, "}")

	// generate message methods
	generateMessageResetMethod(w, name)
	generateMessageNameGetter(w, name, msg.Name)
	generateCrcGetter(w, name, msg.CRC)
	generateMessageTypeGetter(w, name, msgType)
	generateMessageSize(ctx, w, name, msg.Fields)
	generateMessageMarshal(ctx, w, name, msg.Fields)
	generateMessageUnmarshal(ctx, w, name, msg.Fields)

	fmt.Fprintln(w)
}

func generateMessageSize(ctx *Context, w io.Writer, name string, fields []Field) {
	fmt.Fprintf(w, "func (m *%[1]s) Size() int {\n", name)

	fmt.Fprintf(w, "\tif m == nil { return 0 }\n")
	fmt.Fprintf(w, "\tvar size int\n")

	encodeBaseType := func(typ, name string, length int, sizefrom string) bool {
		t, ok := BaseTypeNames[typ]
		if !ok {
			return false
		}

		var s = BaseTypeSizes[t]
		switch t {
		case STRING:
			if length > 0 {
				s = length
				fmt.Fprintf(w, "\tsize += %d\n", s)
			} else {
				s = 4
				fmt.Fprintf(w, "\tsize += %d + len(%s)\n", s, name)
			}
		default:
			if sizefrom != "" {
				//fmt.Fprintf(w, "\tsize += %d * int(%s)\n", s, sizefrom)
				fmt.Fprintf(w, "\tsize += %d * len(%s)\n", s, name)
			} else {
				if length > 0 {
					s = BaseTypeSizes[t] * length
				}
				fmt.Fprintf(w, "\tsize += %d\n", s)
			}
		}

		return true
	}

	lvl := 0
	var encodeFields func(fields []Field, parentName string)
	encodeFields = func(fields []Field, parentName string) {
		lvl++
		defer func() { lvl-- }()

		n := 0
		for _, field := range fields {
			// skip internal fields
			switch strings.ToLower(field.Name) {
			case /*crcField,*/ msgIdField:
				continue
			case clientIndexField, contextField:
				if n == 0 {
					continue
				}
			}
			n++

			fieldName := camelCaseName(strings.TrimPrefix(field.Name, "_"))
			name := fmt.Sprintf("%s.%s", parentName, fieldName)
			sizeFrom := camelCaseName(strings.TrimPrefix(field.SizeFrom, "_"))
			var sizeFromName string
			if sizeFrom != "" {
				sizeFromName = fmt.Sprintf("%s.%s", parentName, sizeFrom)
			}

			fmt.Fprintf(w, "\t// field[%d] %s\n", lvl, name)

			if encodeBaseType(field.Type, name, field.Length, sizeFromName) {
				continue
			}

			char := fmt.Sprintf("s%d", lvl)
			index := fmt.Sprintf("j%d", lvl)

			if field.Array {
				if field.Length > 0 {
					fmt.Fprintf(w, "\tfor %[2]s := 0; %[2]s < %[1]d; %[2]s ++ {\n", field.Length, index)
				} else if field.SizeFrom != "" {
					//fmt.Fprintf(w, "\tfor %[1]s := 0; %[1]s < int(%[2]s.%[3]s); %[1]s++ {\n", index, parentName, sizeFrom)
					fmt.Fprintf(w, "\tfor %[1]s := 0; %[1]s < len(%[2]s); %[1]s++ {\n", index, name)
				}

				fmt.Fprintf(w, "\tvar %[1]s %[2]s\n_ = %[1]s\n", char, convertToGoType(ctx.packageData, field.Type))
				fmt.Fprintf(w, "\tif %[1]s < len(%[2]s) { %[3]s = %[2]s[%[1]s] }\n", index, name, char)
				name = char
			}

			if enum := getEnumByRef(ctx.packageData, field.Type); enum != nil {
				if encodeBaseType(enum.Type, name, 0, "") {
				} else {
					fmt.Fprintf(w, "\t// ??? ENUM %s %s\n", name, enum.Type)
				}
			} else if alias := getAliasByRef(ctx.packageData, field.Type); alias != nil {
				if encodeBaseType(alias.Type, name, alias.Length, "") {
				} else if typ := getTypeByRef(ctx.packageData, alias.Type); typ != nil {
					encodeFields(typ.Fields, name)
				} else {
					fmt.Fprintf(w, "\t// ??? ALIAS %s %s\n", name, alias.Type)
				}
			} else if typ := getTypeByRef(ctx.packageData, field.Type); typ != nil {
				encodeFields(typ.Fields, name)
			} else if union := getUnionByRef(ctx.packageData, field.Type); union != nil {
				maxSize := getUnionSize(ctx.packageData, union)
				fmt.Fprintf(w, "\tsize += %d\n", maxSize)
			} else {
				fmt.Fprintf(w, "\t// ??? buf[pos] = (%s)\n", name)
			}

			if field.Array {
				fmt.Fprintf(w, "\t}\n")
			}
		}
	}

	encodeFields(fields, "m")

	fmt.Fprintf(w, "return size\n")

	fmt.Fprintf(w, "}\n")
}

func generateMessageMarshal(ctx *Context, w io.Writer, name string, fields []Field) {
	fmt.Fprintf(w, "func (m *%[1]s) Marshal(b []byte) ([]byte, error) {\n", name)

	fmt.Fprintf(w, "\to := binary.BigEndian\n")
	fmt.Fprintf(w, "\t_ = o\n")
	fmt.Fprintf(w, "\tpos := 0\n")
	fmt.Fprintf(w, "\t_ = pos\n")

	var buf = new(strings.Builder)

	encodeBaseType := func(typ, name string, length int, sizefrom string) bool {
		t, ok := BaseTypeNames[typ]
		if !ok {
			return false
		}

		isArray := length > 0 || sizefrom != ""

		switch t {
		case I8, U8, I16, U16, I32, U32, I64, U64, F64:
			if isArray {
				if length != 0 {
					fmt.Fprintf(buf, "\tfor i := 0; i < %d; i++ {\n", length)
				} else if sizefrom != "" {
					//fmt.Fprintf(buf, "\tfor i := 0; i < int(%s); i++ {\n", sizefrom)
					fmt.Fprintf(buf, "\tfor i := 0; i < len(%s); i++ {\n", name)
				}
			}
		}

		switch t {
		case I8, U8:
			if isArray {
				fmt.Fprintf(buf, "\tvar x uint8\n")
				fmt.Fprintf(buf, "\tif i < len(%s) { x = uint8(%s[i]) }\n", name, name)
				name = "x"
			}
			fmt.Fprintf(buf, "\tbuf[pos] = uint8(%s)\n", name)
			fmt.Fprintf(buf, "\tpos += 1\n")
			if isArray {
				fmt.Fprintf(buf, "\t}\n")
			}
		case I16, U16:
			if isArray {
				fmt.Fprintf(buf, "\tvar x uint16\n")
				fmt.Fprintf(buf, "\tif i < len(%s) { x = uint16(%s[i]) }\n", name, name)
				name = "x"
			}
			fmt.Fprintf(buf, "\to.PutUint16(buf[pos:pos+2], uint16(%s))\n", name)
			fmt.Fprintf(buf, "\tpos += 2\n")
			if isArray {
				fmt.Fprintf(buf, "\t}\n")
			}
		case I32, U32:
			if isArray {
				fmt.Fprintf(buf, "\tvar x uint32\n")
				fmt.Fprintf(buf, "\tif i < len(%s) { x = uint32(%s[i]) }\n", name, name)
				name = "x"
			}
			fmt.Fprintf(buf, "\to.PutUint32(buf[pos:pos+4], uint32(%s))\n", name)
			fmt.Fprintf(buf, "\tpos += 4\n")
			if isArray {
				fmt.Fprintf(buf, "\t}\n")
			}
		case I64, U64:
			if isArray {
				fmt.Fprintf(buf, "\tvar x uint64\n")
				fmt.Fprintf(buf, "\tif i < len(%s) { x = uint64(%s[i]) }\n", name, name)
				name = "x"
			}
			fmt.Fprintf(buf, "\to.PutUint64(buf[pos:pos+8], uint64(%s))\n", name)
			fmt.Fprintf(buf, "\tpos += 8\n")
			if isArray {
				fmt.Fprintf(buf, "\t}\n")
			}
		case F64:
			if isArray {
				fmt.Fprintf(buf, "\tvar x float64\n")
				fmt.Fprintf(buf, "\tif i < len(%s) { x = float64(%s[i]) }\n", name, name)
				name = "x"
			}
			fmt.Fprintf(buf, "\to.PutUint64(buf[pos:pos+8], math.Float64bits(float64(%s)))\n", name)
			fmt.Fprintf(buf, "\tpos += 8\n")
			if isArray {
				fmt.Fprintf(buf, "\t}\n")
			}
		case BOOL:
			fmt.Fprintf(buf, "\tif %s { buf[pos] = 1 }\n", name)
			fmt.Fprintf(buf, "\tpos += 1\n")
		case STRING:
			if length != 0 {
				fmt.Fprintf(buf, "\tcopy(buf[pos:pos+%d], %s)\n", length, name)
				fmt.Fprintf(buf, "\tpos += %d\n", length)
			} else {
				fmt.Fprintf(buf, "\to.PutUint32(buf[pos:pos+4], uint32(len(%s)))\n", name)
				fmt.Fprintf(buf, "\tpos += 4\n")
				fmt.Fprintf(buf, "\tcopy(buf[pos:pos+len(%s)], %s[:])\n", name, name)
				fmt.Fprintf(buf, "\tpos += len(%s)\n", name)
			}
		default:
			fmt.Fprintf(buf, "\t// ??? %s %s\n", name, typ)
			return false
		}
		return true
	}

	lvl := 0
	var encodeFields func(fields []Field, parentName string)
	encodeFields = func(fields []Field, parentName string) {
		lvl++
		defer func() { lvl-- }()

		n := 0
		for _, field := range fields {
			// skip internal fields
			switch strings.ToLower(field.Name) {
			case /*crcField,*/ msgIdField:
				continue
			case clientIndexField, contextField:
				if n == 0 {
					continue
				}
			}
			n++

			getFieldName := func(name string) string {
				fieldName := camelCaseName(strings.TrimPrefix(name, "_"))
				return fmt.Sprintf("%s.%s", parentName, fieldName)
			}

			fieldName := camelCaseName(strings.TrimPrefix(field.Name, "_"))
			name := fmt.Sprintf("%s.%s", parentName, fieldName)
			sizeFrom := camelCaseName(strings.TrimPrefix(field.SizeFrom, "_"))
			var sizeFromName string
			if sizeFrom != "" {
				sizeFromName = fmt.Sprintf("%s.%s", parentName, sizeFrom)
			}

			fmt.Fprintf(buf, "\t// field[%d] %s\n", lvl, name)

			getSizeOfField := func() *Field {
				for _, f := range fields {
					if f.SizeFrom == field.Name {
						return &f
					}
				}
				return nil
			}
			if f := getSizeOfField(); f != nil {
				if encodeBaseType(field.Type, fmt.Sprintf("len(%s)", getFieldName(f.Name)), field.Length, "") {
					continue
				}
				panic(fmt.Sprintf("failed to encode base type of sizefrom field: %s", field.Name))
			}

			if encodeBaseType(field.Type, name, field.Length, sizeFromName) {
				continue
			}

			char := fmt.Sprintf("v%d", lvl)
			index := fmt.Sprintf("j%d", lvl)

			if field.Array {
				if field.Length > 0 {
					fmt.Fprintf(buf, "\tfor %[2]s := 0; %[2]s < %[1]d; %[2]s ++ {\n", field.Length, index)
				} else if field.SizeFrom != "" {
					//fmt.Fprintf(buf, "\tfor %[1]s := 0; %[1]s < int(%[2]s.%[3]s); %[1]s++ {\n", index, parentName, sizeFrom)
					fmt.Fprintf(buf, "\tfor %[1]s := 0; %[1]s < len(%[2]s); %[1]s++ {\n", index, name)
				}

				fmt.Fprintf(buf, "\tvar %s %s\n", char, convertToGoType(ctx.packageData, field.Type))
				fmt.Fprintf(buf, "\tif %[1]s < len(%[2]s) { %[3]s = %[2]s[%[1]s] }\n", index, name, char)
				name = char
			}

			if enum := getEnumByRef(ctx.packageData, field.Type); enum != nil {
				if encodeBaseType(enum.Type, name, 0, "") {
				} else {
					fmt.Fprintf(buf, "\t// ??? ENUM %s %s\n", name, enum.Type)
				}
			} else if alias := getAliasByRef(ctx.packageData, field.Type); alias != nil {
				if encodeBaseType(alias.Type, name, alias.Length, "") {
				} else if typ := getTypeByRef(ctx.packageData, alias.Type); typ != nil {
					encodeFields(typ.Fields, name)
				} else {
					fmt.Fprintf(buf, "\t// ??? ALIAS %s %s\n", name, alias.Type)
				}
			} else if typ := getTypeByRef(ctx.packageData, field.Type); typ != nil {
				encodeFields(typ.Fields, name)
			} else if union := getUnionByRef(ctx.packageData, field.Type); union != nil {
				maxSize := getUnionSize(ctx.packageData, union)
				fmt.Fprintf(buf, "\tcopy(buf[pos:pos+%d], %s.%s[:])\n", maxSize, name, unionDataField)
				fmt.Fprintf(buf, "\tpos += %d\n", maxSize)
			} else {
				fmt.Fprintf(buf, "\t// ??? buf[pos] = (%s)\n", name)
			}

			if field.Array {
				fmt.Fprintf(buf, "\t}\n")
			}
		}
	}

	encodeFields(fields, "m")

	fmt.Fprintf(w, "\tvar buf []byte\n")
	fmt.Fprintf(w, "\tif b == nil {\n")
	fmt.Fprintf(w, "\tbuf = make([]byte, m.Size())\n")
	fmt.Fprintf(w, "\t} else {\n")
	fmt.Fprintf(w, "\tbuf = b\n")
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprint(w, buf.String())

	fmt.Fprintf(w, "return buf, nil\n")

	fmt.Fprintf(w, "}\n")
}

func generateMessageUnmarshal(ctx *Context, w io.Writer, name string, fields []Field) {
	fmt.Fprintf(w, "func (m *%[1]s) Unmarshal(tmp []byte) error {\n", name)

	fmt.Fprintf(w, "\to := binary.BigEndian\n")
	fmt.Fprintf(w, "\t_ = o\n")
	fmt.Fprintf(w, "\tpos := 0\n")
	fmt.Fprintf(w, "\t_ = pos\n")

	decodeBaseType := func(typ, orig, name string, length int, sizefrom string, alloc bool) bool {
		t, ok := BaseTypeNames[typ]
		if !ok {
			return false
		}

		isArray := length > 0 || sizefrom != ""

		switch t {
		case I8, U8, I16, U16, I32, U32, I64, U64, F64:
			if isArray {
				if alloc {
					if length != 0 {
						fmt.Fprintf(w, "\t%s = make([]%s, %d)\n", name, orig, length)
					} else if sizefrom != "" {
						fmt.Fprintf(w, "\t%s = make([]%s, %s)\n", name, orig, sizefrom)
					}
				}
				fmt.Fprintf(w, "\tfor i := 0; i < len(%s); i++ {\n", name)
			}
		}

		switch t {
		case I8, U8:
			if isArray {
				fmt.Fprintf(w, "\t%s[i] = %s(tmp[pos])\n", name, convertToGoType(ctx.packageData, typ))
			} else {
				fmt.Fprintf(w, "\t%s = %s(tmp[pos])\n", name, orig)
			}
			fmt.Fprintf(w, "\tpos += 1\n")
			if isArray {
				fmt.Fprintf(w, "\t}\n")
			}
		case I16, U16:
			if isArray {
				fmt.Fprintf(w, "\t%s[i] = %s(o.Uint16(tmp[pos:pos+2]))\n", name, orig)
			} else {
				fmt.Fprintf(w, "\t%s = %s(o.Uint16(tmp[pos:pos+2]))\n", name, orig)
			}
			fmt.Fprintf(w, "\tpos += 2\n")
			if isArray {
				fmt.Fprintf(w, "\t}\n")
			}
		case I32, U32:
			if isArray {
				fmt.Fprintf(w, "\t%s[i] = %s(o.Uint32(tmp[pos:pos+4]))\n", name, orig)
			} else {
				fmt.Fprintf(w, "\t%s = %s(o.Uint32(tmp[pos:pos+4]))\n", name, orig)
			}
			fmt.Fprintf(w, "\tpos += 4\n")
			if isArray {
				fmt.Fprintf(w, "\t}\n")
			}
		case I64, U64:
			if isArray {
				fmt.Fprintf(w, "\t%s[i] = %s(o.Uint64(tmp[pos:pos+8]))\n", name, orig)
			} else {
				fmt.Fprintf(w, "\t%s = %s(o.Uint64(tmp[pos:pos+8]))\n", name, orig)
			}
			fmt.Fprintf(w, "\tpos += 8\n")
			if isArray {
				fmt.Fprintf(w, "\t}\n")
			}
		case F64:
			if isArray {
				fmt.Fprintf(w, "\t%s[i] = %s(math.Float64frombits(o.Uint64(tmp[pos:pos+8])))\n", name, orig)
			} else {
				fmt.Fprintf(w, "\t%s = %s(math.Float64frombits(o.Uint64(tmp[pos:pos+8])))\n", name, orig)
			}
			fmt.Fprintf(w, "\tpos += 8\n")
			if isArray {
				fmt.Fprintf(w, "\t}\n")
			}
		case BOOL:
			fmt.Fprintf(w, "\t%s = tmp[pos] != 0\n", name)
			fmt.Fprintf(w, "\tpos += 1\n")
		case STRING:
			if length != 0 {
				fmt.Fprintf(w, "\t{\n")
				fmt.Fprintf(w, "\tnul := bytes.Index(tmp[pos:pos+%d], []byte{0x00})\n", length)
				fmt.Fprintf(w, "\t%[1]s = codec.DecodeString(tmp[pos:pos+nul])\n", name)
				fmt.Fprintf(w, "\tpos += %d\n", length)
				fmt.Fprintf(w, "\t}\n")
			} else {
				fmt.Fprintf(w, "\t{\n")
				fmt.Fprintf(w, "\tsiz := o.Uint32(tmp[pos:pos+4])\n")
				fmt.Fprintf(w, "\tpos += 4\n")
				fmt.Fprintf(w, "\t%[1]s = codec.DecodeString(tmp[pos:pos+int(siz)])\n", name)
				fmt.Fprintf(w, "\tpos += len(%s)\n", name)
				fmt.Fprintf(w, "\t}\n")
			}
		default:
			fmt.Fprintf(w, "\t// ??? %s %s\n", name, typ)
			return false
		}
		return true
	}

	lvl := 0
	var decodeFields func(fields []Field, parentName string)
	decodeFields = func(fields []Field, parentName string) {
		lvl++
		defer func() { lvl-- }()

		n := 0
		for _, field := range fields {
			// skip internal fields
			switch strings.ToLower(field.Name) {
			case /*crcField,*/ msgIdField:
				continue
			case clientIndexField, contextField:
				if n == 0 {
					continue
				}
			}
			n++

			fieldName := camelCaseName(strings.TrimPrefix(field.Name, "_"))
			name := fmt.Sprintf("%s.%s", parentName, fieldName)
			sizeFrom := camelCaseName(strings.TrimPrefix(field.SizeFrom, "_"))
			var sizeFromName string
			if sizeFrom != "" {
				sizeFromName = fmt.Sprintf("%s.%s", parentName, sizeFrom)
			}

			fmt.Fprintf(w, "\t// field[%d] %s\n", lvl, name)

			if decodeBaseType(field.Type, convertToGoType(ctx.packageData, field.Type), name, field.Length, sizeFromName, true) {
				continue
			}

			//char := fmt.Sprintf("v%d", lvl)
			index := fmt.Sprintf("j%d", lvl)

			if field.Array {
				if field.Length > 0 {
					fmt.Fprintf(w, "\tfor %[2]s := 0; %[2]s < %[1]d; %[2]s ++ {\n", field.Length, index)
				} else if field.SizeFrom != "" {
					fieldType := getFieldType(ctx, field)
					if strings.HasPrefix(fieldType, "[]") {
						fmt.Fprintf(w, "\t%s = make(%s, int(%s.%s))\n", name, fieldType, parentName, sizeFrom)
					}
					fmt.Fprintf(w, "\tfor %[1]s := 0; %[1]s < int(%[2]s.%[3]s); %[1]s++ {\n", index, parentName, sizeFrom)
				}

				/*fmt.Fprintf(w, "\tvar %s %s\n", char, convertToGoType(ctx, field.Type))
				fmt.Fprintf(w, "\tif %[1]s < len(%[2]s) { %[3]s = %[2]s[%[1]s] }\n", index, name, char)
				name = char*/
				name = fmt.Sprintf("%s[%s]", name, index)
			}

			if enum := getEnumByRef(ctx.packageData, field.Type); enum != nil {
				if decodeBaseType(enum.Type, convertToGoType(ctx.packageData, field.Type), name, 0, "", false) {
				} else {
					fmt.Fprintf(w, "\t// ??? ENUM %s %s\n", name, enum.Type)
				}
			} else if alias := getAliasByRef(ctx.packageData, field.Type); alias != nil {
				if decodeBaseType(alias.Type, convertToGoType(ctx.packageData, field.Type), name, alias.Length, "", false) {
				} else if typ := getTypeByRef(ctx.packageData, alias.Type); typ != nil {
					decodeFields(typ.Fields, name)
				} else {
					fmt.Fprintf(w, "\t// ??? ALIAS %s %s\n", name, alias.Type)
				}
			} else if typ := getTypeByRef(ctx.packageData, field.Type); typ != nil {
				decodeFields(typ.Fields, name)
			} else if union := getUnionByRef(ctx.packageData, field.Type); union != nil {
				maxSize := getUnionSize(ctx.packageData, union)
				fmt.Fprintf(w, "\tcopy(%s.%s[:], tmp[pos:pos+%d])\n", name, unionDataField, maxSize)
				fmt.Fprintf(w, "\tpos += %d\n", maxSize)
			} else {
				fmt.Fprintf(w, "\t// ??? buf[pos] = (%s)\n", name)
			}

			if field.Array {
				fmt.Fprintf(w, "\t}\n")
			}
		}
	}

	decodeFields(fields, "m")

	fmt.Fprintf(w, "return nil\n")

	fmt.Fprintf(w, "}\n")
}

func getFieldType(ctx *Context, field Field) string {
	fieldName := strings.TrimPrefix(field.Name, "_")
	fieldName = camelCaseName(fieldName)

	dataType := convertToGoType(ctx.packageData, field.Type)
	fieldType := dataType

	// check if it is array
	if field.Length > 0 || field.SizeFrom != "" {
		if dataType == "uint8" {
			dataType = "byte"
		}
		if dataType == "string" && field.Array {
			fieldType = "string"
			dataType = "byte"
		} else if _, ok := BaseTypeNames[field.Type]; !ok && field.SizeFrom == "" {
			fieldType = fmt.Sprintf("[%d]%s", field.Length, dataType)
		} else {
			fieldType = "[]" + dataType
		}
	}

	return fieldType
}

func generateField(ctx *Context, w io.Writer, fields []Field, i int) {
	field := fields[i]

	fieldName := strings.TrimPrefix(field.Name, "_")
	fieldName = camelCaseName(fieldName)

	dataType := convertToGoType(ctx.packageData, field.Type)
	fieldType := dataType

	// generate length field for strings
	if field.Type == "string" && field.Length == 0 {
		fmt.Fprintf(w, "\tXXX_%sLen uint32 `struc:\"sizeof=%s\"`\n", fieldName, fieldName)
	}

	// check if it is array
	if field.Length > 0 || field.SizeFrom != "" {
		if dataType == "uint8" {
			dataType = "byte"
		}
		if dataType == "string" && field.Array {
			fieldType = "string"
			dataType = "byte"
		} else if _, ok := BaseTypeNames[field.Type]; !ok && field.SizeFrom == "" {
			fieldType = fmt.Sprintf("[%d]%s", field.Length, dataType)
		} else {
			fieldType = "[]" + dataType
		}
	}
	fmt.Fprintf(w, "\t%s %s", fieldName, fieldType)

	fieldTags := map[string]string{}

	if field.Length > 0 && field.Array {
		// fixed size array
		fieldTags["struc"] = fmt.Sprintf("[%d]%s", field.Length, dataType)
	} else {
		for _, f := range fields {
			if f.SizeFrom == field.Name {
				// variable sized array
				sizeOfName := camelCaseName(f.Name)
				fieldTags["struc"] = fmt.Sprintf("sizeof=%s", sizeOfName)
			}
		}
	}

	if ctx.IncludeBinapiNames {
		typ := fromApiType(field.Type)
		if field.Array {
			if field.Length > 0 {
				fieldTags["binapi"] = fmt.Sprintf("%s[%d],name=%s", typ, field.Length, field.Name)
			} else if field.SizeFrom != "" {
				fieldTags["binapi"] = fmt.Sprintf("%s[%s],name=%s", typ, field.SizeFrom, field.Name)
			}
		} else {
			fieldTags["binapi"] = fmt.Sprintf("%s,name=%s", typ, field.Name)
		}
	}
	if limit, ok := field.Meta["limit"]; ok && limit.(int) > 0 {
		fieldTags["binapi"] = fmt.Sprintf("%s,limit=%d", fieldTags["binapi"], limit)
	}
	if def, ok := field.Meta["default"]; ok && def != nil {
		actual := getActualType(ctx.packageData, fieldType)
		if t, ok := binapiTypes[actual]; ok && t != "float64" {
			defnum := int(def.(float64))
			fieldTags["binapi"] = fmt.Sprintf("%s,default=%d", fieldTags["binapi"], defnum)
		} else {
			fieldTags["binapi"] = fmt.Sprintf("%s,default=%v", fieldTags["binapi"], def)
		}
	}

	fieldTags["json"] = fmt.Sprintf("%s,omitempty", field.Name)

	if len(fieldTags) > 0 {
		fmt.Fprintf(w, "\t`")
		var keys []string
		for k := range fieldTags {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var n int
		for _, tt := range keys {
			t, ok := fieldTags[tt]
			if !ok {
				continue
			}
			if n > 0 {
				fmt.Fprintf(w, " ")
			}
			n++
			fmt.Fprintf(w, `%s:"%s"`, tt, t)
		}
		fmt.Fprintf(w, "`")
	}

	fmt.Fprintln(w)
}

func getUnderlyingType(ctx *Context, typ string) (actual string) {
	for _, enum := range ctx.packageData.EnumTypes {
		if enum.Name == typ {
			return enum.Type
		}
	}
	for _, alias := range ctx.packageData.AliasTypes {
		if alias.Name == typ {
			return alias.Type
		}
	}
	return typ
}

func generateMessageResetMethod(w io.Writer, structName string) {
	fmt.Fprintf(w, "func (m *%[1]s) Reset() { *m = %[1]s{} }\n", structName)
}

func generateMessageNameGetter(w io.Writer, structName, msgName string) {
	fmt.Fprintf(w, "func (*%s) GetMessageName() string {	return %q }\n", structName, msgName)
}

func generateTypeNameGetter(w io.Writer, structName, msgName string) {
	fmt.Fprintf(w, "func (*%s) GetTypeName() string { return %q }\n", structName, msgName)
}

func generateCrcGetter(w io.Writer, structName, crc string) {
	crc = strings.TrimPrefix(crc, "0x")
	fmt.Fprintf(w, "func (*%s) GetCrcString() string { return %q }\n", structName, crc)
}

func generateMessageTypeGetter(w io.Writer, structName string, msgType MessageType) {
	fmt.Fprintf(w, "func (*"+structName+") GetMessageType() api.MessageType {")
	if msgType == requestMessage {
		fmt.Fprintf(w, "\treturn api.RequestMessage")
	} else if msgType == replyMessage {
		fmt.Fprintf(w, "\treturn api.ReplyMessage")
	} else if msgType == eventMessage {
		fmt.Fprintf(w, "\treturn api.EventMessage")
	} else {
		fmt.Fprintf(w, "\treturn api.OtherMessage")
	}
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)
}

func generateServiceMethods(ctx *Context, w io.Writer, methods []RPC) {

	// generate services comment
	generateComment(ctx, w, serviceApiName, "services", "service")

	// generate service api
	fmt.Fprintf(w, "type %s interface {\n", serviceApiName)
	for _, svc := range methods {
		generateServiceMethod(ctx, w, &svc)
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)

	// generate client implementation
	fmt.Fprintf(w, "type %s struct {\n", serviceImplName)
	fmt.Fprintf(w, "\tch api.Channel\n")
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)

	// generate client constructor
	fmt.Fprintf(w, "func New%s(ch api.Channel) %s {\n", serviceClientName, serviceApiName)
	fmt.Fprintf(w, "\treturn &%s{ch}\n", serviceImplName)
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)

	for _, met := range methods {
		method := camelCaseName(met.RequestMsg)
		if m := strings.TrimSuffix(method, "Dump"); method != m {
			method = "Dump" + m
		}

		fmt.Fprintf(w, "func (c *%s) ", serviceImplName)
		generateServiceMethod(ctx, w, &met)
		fmt.Fprintln(w, " {")
		if met.Stream {
			streamImpl := fmt.Sprintf("%s_%sClient", serviceImplName, method)
			fmt.Fprintf(w, "\tstream := c.ch.SendMultiRequest(in)\n")
			fmt.Fprintf(w, "\tx := &%s{stream}\n", streamImpl)
			fmt.Fprintf(w, "\treturn x, nil\n")
		} else if replyTyp := camelCaseName(met.ReplyMsg); replyTyp != "" {
			fmt.Fprintf(w, "\tout := new(%s)\n", replyTyp)
			fmt.Fprintf(w, "\terr:= c.ch.SendRequest(in).ReceiveReply(out)\n")
			fmt.Fprintf(w, "\tif err != nil { return nil, err }\n")
			fmt.Fprintf(w, "\treturn out, nil\n")
		} else {
			fmt.Fprintf(w, "\tc.ch.SendRequest(in)\n")
			fmt.Fprintf(w, "\treturn nil\n")
		}
		fmt.Fprintln(w, "}")
		fmt.Fprintln(w)

		if met.Stream {
			replyTyp := camelCaseName(met.ReplyMsg)
			method := camelCaseName(met.RequestMsg)
			if m := strings.TrimSuffix(method, "Dump"); method != m {
				method = "Dump" + m
			}
			streamApi := fmt.Sprintf("%s_%sClient", serviceApiName, method)

			fmt.Fprintf(w, "type %s interface {\n", streamApi)
			fmt.Fprintf(w, "\tRecv() (*%s, error)\n", replyTyp)
			fmt.Fprintln(w, "}")
			fmt.Fprintln(w)

			streamImpl := fmt.Sprintf("%s_%sClient", serviceImplName, method)
			fmt.Fprintf(w, "type %s struct {\n", streamImpl)
			fmt.Fprintf(w, "\tapi.MultiRequestCtx\n")
			fmt.Fprintln(w, "}")
			fmt.Fprintln(w)

			fmt.Fprintf(w, "func (c *%s) Recv() (*%s, error) {\n", streamImpl, replyTyp)
			fmt.Fprintf(w, "\tm := new(%s)\n", replyTyp)
			fmt.Fprintf(w, "\tstop, err := c.MultiRequestCtx.ReceiveReply(m)\n")
			fmt.Fprintf(w, "\tif err != nil { return nil, err }\n")
			fmt.Fprintf(w, "\tif stop { return nil, io.EOF }\n")
			fmt.Fprintf(w, "\treturn m, nil\n")
			fmt.Fprintln(w, "}")
			fmt.Fprintln(w)
		}
	}

	/*fmt.Fprintf(w, "var %s = api.%s{\n", serviceDescName, serviceDescType)
	fmt.Fprintf(w, "\tServiceName: \"%s\",\n", ctx.moduleName)
	fmt.Fprintf(w, "\tHandlerType: (*%s)(nil),\n", serviceApiName)
	fmt.Fprintf(w, "\tMethods: []api.MethodDesc{\n")
	for _, method := range methods {
		fmt.Fprintf(w, "\t  {\n")
		fmt.Fprintf(w, "\t    MethodName: \"%s\",\n", method.Name)
		fmt.Fprintf(w, "\t  },\n")
	}
	fmt.Fprintf(w, "\t},\n")
	//fmt.Fprintf(w, "\tCompatibility: %s,\n", messageCrcName)
	//fmt.Fprintf(w, "\tMetadata: reflect.TypeOf((*%s)(nil)).Elem().PkgPath(),\n", serviceApiName)
	fmt.Fprintf(w, "\tMetadata: \"%s\",\n", ctx.inputFile)
	fmt.Fprintln(w, "}")*/

	fmt.Fprintln(w)
}

func generateServiceMethod(ctx *Context, w io.Writer, svc *RPC) {
	reqTyp := camelCaseName(svc.RequestMsg)

	// method name is same as parameter type name by default
	method := reqTyp
	if svc.Stream {
		// use Dump as prefix instead of suffix for stream services
		if m := strings.TrimSuffix(method, "Dump"); method != m {
			method = "Dump" + m
		}
	}

	params := fmt.Sprintf("in *%s", reqTyp)
	returns := "error"

	if replyType := camelCaseName(svc.ReplyMsg); replyType != "" {
		var replyTyp string
		if svc.Stream {
			replyTyp = fmt.Sprintf("%s_%sClient", serviceApiName, method)
		} else {
			replyTyp = fmt.Sprintf("*%s", replyType)
		}
		returns = fmt.Sprintf("(%s, error)", replyTyp)
	}

	fmt.Fprintf(w, "\t%s(ctx context.Context, %s) %s", method, params, returns)
}
