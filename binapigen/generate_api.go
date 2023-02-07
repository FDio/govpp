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
	"fmt"
	"path"
	"strconv"
	"strings"

	"go.fd.io/govpp/version"
)

// generated names
const (
	apiName    = "APIFile"    // API file name
	apiVersion = "APIVersion" // API version number
	apiCrc     = "VersionCrc" // version checksum

	fieldUnionData = "XXX_UnionData" // name for the union data field
)

// option keys
const (
	msgStatus     = "status"
	msgDeprecated = "deprecated"
	msgInProgress = "in_progress"
)

// generated option messages
const (
	deprecatedMsg = "the message will be removed in the future versions"
	inProgressMsg = "the message form may change in the future versions"
)

func GenerateAPI(gen *Generator, file *File) *GenFile {
	logf("----------------------------")
	logf(" Generate API - %s", file.Desc.Name)
	logf("----------------------------")

	filename := path.Join(file.FilenamePrefix, file.Desc.Name+generatedFilenameSuffix)
	g := gen.NewGenFile(filename, file)

	genCodeGeneratedComment(g)
	if !gen.opts.NoVersionInfo {
		g.P("// versions:")
		g.P("//  binapi-generator: ", version.Version())
		g.P("//  VPP:              ", g.gen.vppVersion)
		if !gen.opts.NoSourcePathInfo {
			g.P("// source: ", g.file.Desc.Path)
		}
	}
	g.P()

	// package definition
	genPackageComment(g)
	g.P("package ", file.PackageName)
	g.P()

	// package imports
	for _, imp := range g.file.Imports {
		genImport(g, imp)
	}

	// GoVPP API version assertion
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the GoVPP api package it is being compiled against.")
	g.P("// A compilation error at this line likely means your copy of the")
	g.P("// GoVPP api package needs to be updated.")
	g.P("const _ = ", govppApiPkg.Ident("GoVppAPIPackageIsVersion"), generatedCodeVersion)
	g.P()

	// API meta info
	genApiInfo(g)

	// API types
	for _, enum := range g.file.Enums {
		genEnum(g, enum)
	}
	for _, alias := range g.file.Aliases {
		genAlias(g, alias)
	}
	for _, typ := range g.file.Structs {
		genStruct(g, typ)
	}
	for _, union := range g.file.Unions {
		genUnion(g, union)
	}

	// API messages
	genMessages(g)

	return g
}

func genApiInfo(g *GenFile) {
	// generate API info
	g.P("const (")
	g.P(apiName, " = ", strconv.Quote(g.file.Desc.Name))
	g.P(apiVersion, " = ", strconv.Quote(g.file.Version))
	g.P(apiCrc, " = ", g.file.Desc.CRC)
	g.P(")")
	g.P()
}

func genPackageComment(g *GenFile) {
	apifile := g.file.Desc.Name + ".api"
	g.P("// Package ", g.file.PackageName, " contains generated bindings for API file ", apifile, ".")
	g.P("//")
	g.P("// Contents:")
	printObjNum := func(obj string, num int) {
		if num > 0 {
			if num > 1 {
				if strings.HasSuffix(obj, "s") {
					obj += "es"
				} else {
					obj += "s"
				}
			}
			g.P("// - ", fmt.Sprintf("%2d", num), " ", obj)
		}
	}
	printObjNum("alias", len(g.file.Aliases))
	printObjNum("enum", len(g.file.Enums))
	printObjNum("struct", len(g.file.Structs))
	printObjNum("union", len(g.file.Unions))
	printObjNum("message", len(g.file.Messages))
}

func genImport(g *GenFile, imp string) {
	impFile, ok := g.gen.FilesByName[imp]
	if !ok {
		return
	}
	if impFile.GoImportPath == g.file.GoImportPath {
		// Skip generating imports for types in the same package
		return
	}
	// Generate imports for all dependencies, even if not used
	g.Import(impFile.GoImportPath)
}

func genTypeComment(g *GenFile, goName string, vppName string, objKind string) {
	g.P("// ", goName, " defines ", objKind, " '", vppName, "'.")
}

func genTypeOptionComment(g *GenFile, options map[string]string) {
	// all messages for API versions < 1.0.0 are in_progress by default
	if msg, ok := options[msgInProgress]; ok || options[msgStatus] == msgInProgress ||
		len(g.file.Version) > 1 && g.file.Version[0:2] == "0." {
		if msg == "" {
			msg = inProgressMsg
		}
		g.P("// InProgress: ", msg)
	}
	if msg, ok := options[msgDeprecated]; ok || options[msgStatus] == msgDeprecated {
		if msg == "" {
			msg = deprecatedMsg
		}
		g.P("// Deprecated: ", msg)
	}
}

func genEnum(g *GenFile, enum *Enum) {
	logf("gen ENUM %s (%s) - %d entries", enum.GoName, enum.Name, len(enum.Entries))

	genTypeComment(g, enum.GoName, enum.Name, "enum")

	gotype := BaseTypesGo[enum.Type]

	g.P("type ", enum.GoName, " ", gotype)
	g.P()

	// generate enum entries
	g.P("const (")
	for _, entry := range enum.Entries {
		g.P(entry.Name, " ", enum.GoName, " = ", entry.Value)
	}
	g.P(")")
	g.P()

	// generate enum conversion maps
	g.P("var (")
	g.P(enum.GoName, "_name = map[", gotype, "]string{")
	for _, entry := range enum.Entries {
		g.P(entry.Value, ": ", strconv.Quote(entry.Name), ",")
	}
	g.P("}")
	g.P(enum.GoName, "_value = map[string]", gotype, "{")
	for _, entry := range enum.Entries {
		g.P(strconv.Quote(entry.Name), ": ", entry.Value, ",")
	}
	g.P("}")
	g.P(")")
	g.P()

	if enum.IsFlag || isEnumFlag(enum) {
		size := BaseTypeSizes[enum.Type] * 8
		g.P("func (x ", enum.GoName, ") String() string {")
		g.P("	s, ok := ", enum.GoName, "_name[", gotype, "(x)]")
		g.P("	if ok { return s }")
		g.P("	str := func(n ", gotype, ") string {")
		g.P("		s, ok := ", enum.GoName, "_name[", gotype, "(n)]")
		g.P("		if ok {")
		g.P("			return s")
		g.P("		}")
		g.P("		return \"", enum.GoName, "(\" + ", strconvPkg.Ident("Itoa"), "(int(n)) + \")\"")
		g.P("	}")
		g.P("	for i := ", gotype, "(0); i <= ", size, "; i++ {")
		g.P("		val := ", gotype, "(x)")
		g.P("		if val&(1<<i) != 0 {")
		g.P("			if s != \"\" {")
		g.P("				s += \"|\"")
		g.P("			}")
		g.P("			s += str(1<<i)")
		g.P("		}")
		g.P("	}")
		g.P("	if s == \"\" {")
		g.P("		return str(", gotype, "(x))")
		g.P("	}")
		g.P("	return s")
		g.P("}")
		g.P()
	} else {
		g.P("func (x ", enum.GoName, ") String() string {")
		g.P("	s, ok := ", enum.GoName, "_name[", gotype, "(x)]")
		g.P("	if ok { return s }")
		g.P("	return \"", enum.GoName, "(\" + ", strconvPkg.Ident("Itoa"), "(int(x)) + \")\"")
		g.P("}")
		g.P()
	}
}

func genAlias(g *GenFile, alias *Alias) {
	logf("gen ALIAS %s (%s) - type: %s length: %d", alias.GoName, alias.Name, alias.Type, alias.Length)

	genTypeComment(g, alias.GoName, alias.Name, "alias")

	var gotype string
	switch {
	case alias.TypeStruct != nil:
		gotype = g.GoIdent(alias.TypeStruct.GoIdent)
	case alias.TypeUnion != nil:
		gotype = g.GoIdent(alias.TypeUnion.GoIdent)
	default:
		gotype = BaseTypesGo[alias.Type]
	}
	if alias.Length > 0 {
		gotype = fmt.Sprintf("[%d]%s", alias.Length, gotype)
	}

	g.P("type ", alias.GoName, " ", gotype)
	g.P()

	// generate alias-specific methods
	switch alias.Name {
	case "ip4_address":
		genIPXAddressHelpers(g, alias.GoName, 4)
	case "ip6_address":
		genIPXAddressHelpers(g, alias.GoName, 6)
	case "address_with_prefix":
		genAddressWithPrefixHelpers(g, alias.GoName)
	case "mac_address":
		genMacAddressHelpers(g, alias.GoName)
	case "timestamp":
		genTimestampHelpers(g, alias.GoName)
	}
}

func genStruct(g *GenFile, typ *Struct) {
	logf("gen STRUCT %s (%s) - %d fields", typ.GoName, typ.Name, len(typ.Fields))

	genTypeComment(g, typ.GoName, typ.Name, "type")

	if len(typ.Fields) == 0 {
		g.P("type ", typ.GoName, " struct {}")
	} else {
		g.P("type ", typ.GoName, " struct {")
		for i := range typ.Fields {
			genField(g, typ.Fields, i)
		}
		g.P("}")
	}
	g.P()

	// generate type-specific methods
	switch typ.Name {
	case "address":
		genAddressHelpers(g, typ.GoName)
	case "prefix":
		genPrefixHelpers(g, typ.GoName)
	case "ip4_prefix":
		genIPXPrefixHelpers(g, typ.GoName, 4)
	case "ip6_prefix":
		genIPXPrefixHelpers(g, typ.GoName, 6)
	}
}

func genUnion(g *GenFile, union *Union) {
	logf("gen UNION %s (%s) - %d fields", union.GoName, union.Name, len(union.Fields))

	genTypeComment(g, union.GoName, union.Name, "union")

	g.P("type ", union.GoName, " struct {")

	// generate field comments
	g.P("// ", union.GoName, " can be one of:")
	for _, field := range union.Fields {
		g.P("// - ", field.GoName, " *", getFieldType(g, field))
	}

	// generate data field
	maxSize := getUnionSize(union)
	g.P(fieldUnionData, " [", maxSize, "]byte")

	// generate end of the struct
	g.P("}")
	g.P()

	// generate methods for fields
	for _, field := range union.Fields {
		genUnionFieldMethods(g, union, field)
	}
	g.P()
}

func genUnionFieldMethods(g *GenFile, union *Union, field *Field) {
	fieldType := fieldGoType(g, field)
	constructorName := union.GoName + field.GoName

	// Constructor
	g.P("func ", constructorName, "(a ", fieldType, ") (u ", union.GoName, ") {")
	g.P("	u.Set", field.GoName, "(a)")
	g.P("	return")
	g.P("}")

	// Setter
	g.P("func (u *", union.GoName, ") Set", field.GoName, "(a ", fieldType, ") {")
	g.P("	buf := ", govppCodecPkg.Ident("NewBuffer"), "(u.", fieldUnionData, "[:])")
	encodeField(g, field, "a", func(name string) string {
		return "a." + name
	}, 0)
	g.P("}")

	// Getter
	g.P("func (u *", union.GoName, ") Get", field.GoName, "() (a ", fieldType, ") {")
	g.P("	buf := ", govppCodecPkg.Ident("NewBuffer"), "(u.", fieldUnionData, "[:])")
	decodeField(g, field, "a", func(name string) string {
		return "a." + name
	}, 0)
	g.P("	return")
	g.P("}")

	g.P()
}

func withSuffix(s string, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}
	return s + suffix
}

func genField(g *GenFile, fields []*Field, i int) {
	field := fields[i]

	logf(" gen FIELD[%d] %s (%s) - type: %q (array: %v/%v)", i, field.GoName, field.Name, field.Type, field.Array, field.Length)

	gotype := getFieldType(g, field)
	tags := structTags{
		"binapi": fieldTagBinapi(field),
		"json":   fieldTagJson(field),
	}

	g.P(field.GoName, " ", gotype, tags)
}

func fieldTagJson(field *Field) string {
	if field.FieldSizeOf != nil {
		return "-"
	}
	return fmt.Sprintf("%s,omitempty", field.Name)
}

func fieldTagBinapi(field *Field) string {
	typ := fromApiType(field.Type)
	if field.Array {
		if field.Length > 0 {
			typ = fmt.Sprintf("%s[%d]", typ, field.Length)
		} else if field.SizeFrom != "" {
			typ = fmt.Sprintf("%s[%s]", typ, field.SizeFrom)
		} else {
			typ = fmt.Sprintf("%s[]", typ)
		}
	}
	tag := []string{
		typ,
		fmt.Sprintf("name=%s", field.Name),
	}
	if limit, ok := field.Meta["limit"]; ok && limit.(int) > 0 {
		tag = append(tag, fmt.Sprintf("limit=%s", limit))
	}
	if def, ok := field.Meta["default"]; ok && def != nil {
		switch fieldActualType(field) {
		case I8, I16, I32, I64:
			def = int(def.(float64))
		case U8, U16, U32, U64:
			def = uint(def.(float64))
		case F64:
			def = def.(float64)
		}
		tag = append(tag, fmt.Sprintf("default=%v", def))
	}
	return strings.Join(tag, ",")
}

func genMessages(g *GenFile) {
	if len(g.file.Messages) == 0 {
		return
	}

	for _, msg := range g.file.Messages {
		genMessage(g, msg)
	}

	// generate registrations
	initFnName := fmt.Sprintf("file_%s_binapi_init", g.file.PackageName)

	g.P("func init() { ", initFnName, "() }")
	g.P("func ", initFnName, "() {")
	for _, msg := range g.file.Messages {
		id := fmt.Sprintf("%s_%s", msg.Name, msg.CRC)
		g.P(govppApiPkg.Ident("RegisterMessage"), "((*", msg.GoIdent, ")(nil), ", strconv.Quote(id), ")")
	}
	g.P("}")
	g.P()

	// generate list of messages
	g.P("// Messages returns list of all messages in this module.")
	g.P("func AllMessages() []", govppApiPkg.Ident("Message"), " {")
	g.P("return []", govppApiPkg.Ident("Message"), "{")
	for _, msg := range g.file.Messages {
		g.P("(*", msg.GoIdent, ")(nil),")
	}
	g.P("}")
	g.P("}")
}

func genMessage(g *GenFile, msg *Message) {
	logf("gen MESSAGE %s (%s) - %d fields", msg.GoName, msg.Name, len(msg.Fields))

	genTypeComment(g, msg.GoIdent.GoName, msg.Name, "message")
	genTypeOptionComment(g, msg.Options)

	// generate message definition
	if len(msg.Fields) == 0 {
		g.P("type ", msg.GoIdent, " struct {}")
	} else {
		g.P("type ", msg.GoIdent, " struct {")
		for i := range msg.Fields {
			genField(g, msg.Fields, i)
		}
		g.P("}")
	}
	g.P()

	// Reset method
	g.P("func (m *", msg.GoIdent.GoName, ") Reset() { *m = ", msg.GoIdent.GoName, "{} }")

	// GetXXX methods
	genMessageMethods(g, msg)

	// codec methods
	genMessageMethodSize(g, msg.GoIdent.GoName, msg.Fields)
	genMessageMethodMarshal(g, msg.GoIdent.GoName, msg.Fields)
	genMessageMethodUnmarshal(g, msg.GoIdent.GoName, msg.Fields)

	g.P()
}

func genMessageMethods(g *GenFile, msg *Message) {
	// GetMessageName method
	g.P("func (*", msg.GoIdent.GoName, ") GetMessageName() string { return ", strconv.Quote(msg.Name), " }")

	// GetCrcString method
	g.P("func (*", msg.GoIdent.GoName, ") GetCrcString() string { return ", strconv.Quote(msg.CRC), " }")

	// GetMessageType method
	g.P("func (*", msg.GoIdent.GoName, ") GetMessageType() api.MessageType {")
	g.P("	return ", msgType2apiMessageType(msg.msgType))
	g.P("}")

	g.P()
}

func msgType2apiMessageType(t msgType) GoIdent {
	switch t {
	case msgTypeRequest:
		return govppApiPkg.Ident("RequestMessage")
	case msgTypeReply:
		return govppApiPkg.Ident("ReplyMessage")
	case msgTypeEvent:
		return govppApiPkg.Ident("EventMessage")
	default:
		return govppApiPkg.Ident("OtherMessage")
	}
}
