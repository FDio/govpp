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

	"go.fd.io/govpp/internal/version"
)

// generated names
const (
	apiName    = "APIFile"    // API file name
	apiVersion = "APIVersion" // API version number
	apiCrc     = "VersionCrc" // version checksum

	fieldUnionData = "XXX_UnionData" // name for the union data field
)

// generated status info
const (
	statusInProgressPrefix = "InProgress"
	statusDeprecatedPrefix = "Deprecated"
	statusOtherPrefix      = "Status"

	statusDeprecatedInfoText = "the message will be removed in the future versions"
	statusInProgressInfoText = "the message form may change in the future versions"
)

func GenerateAPI(gen *Generator, file *File) *GenFile {
	logf("----------------------------")
	logf(" Generate API FILE - %s", file.Desc.Name)
	logf("----------------------------")

	filename := path.Join(file.FilenamePrefix, file.Desc.Name+generatedFilenameSuffix)
	g := gen.NewGenFile(filename, file)

	genCodeGeneratedComment(g)
	if !gen.opts.NoVersionInfo {
		g.P("// versions:")
		g.P("//  binapi-generator: ", version.Version())
		g.P("//  VPP:              ", g.gen.vppapiSchema.Version)
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

func genGenericDefinesComment(g *GenFile, goName string, vppName string, objKind string) {
	g.P("// ", goName, " defines ", objKind, " '", vppName, "'.")
}

func genMessageStatusInfoComment(g *GenFile, msg *Message) {
	switch status, text := getMessageStatus(msg); status {
	case msgStatusInProgress:
		// "in progress" status - might be changed anytime
		if text == "" {
			text = statusInProgressInfoText
		}
		g.P("// ", statusInProgressPrefix, ": ", text)
	case msgStatusDeprecated:
		// "deprecated" status - will be removed later
		if text == "" {
			text = statusDeprecatedInfoText
		}
		g.P("// ", statusDeprecatedPrefix, ": ", text)
	case msgStatusOther:
		// custom status - arbitrary info
		g.P("// ", statusOtherPrefix, ": ", text)
	}
}

func genEnum(g *GenFile, enum *Enum) {
	logf("gen ENUM %s (%s) - %d entries", enum.GoName, enum.Name, len(enum.Entries))

	genGenericDefinesComment(g, enum.GoName, enum.Name, "enum")

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

	genGenericDefinesComment(g, alias.GoName, alias.Name, "alias")

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

	genHelperMethods(g, alias.Name, alias.GoName)
}

func genStruct(g *GenFile, typ *Struct) {
	logf("gen STRUCT %s (%s) - %d fields", typ.GoName, typ.Name, len(typ.Fields))

	genGenericDefinesComment(g, typ.GoName, typ.Name, "type")

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

	genHelperMethods(g, typ.Name, typ.GoName)
}

func genUnion(g *GenFile, union *Union) {
	logf("gen UNION %s (%s) - %d fields", union.GoName, union.Name, len(union.Fields))

	genGenericDefinesComment(g, union.GoName, union.Name, "union")

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
			typ += fmt.Sprintf("[%d]", field.Length)
		} else if field.SizeFrom != "" {
			typ += fmt.Sprintf("[%s]", field.SizeFrom)
		} else {
			typ += "[]"
		}
	}
	tag := []string{
		typ,
		fmt.Sprintf("name=%s", field.Name),
	}

	// limit
	if limit, ok := field.Meta[optFieldLimit]; ok && limit.(int) > 0 {
		tag = append(tag, fmt.Sprintf("limit=%s", limit))
	}

	// default value
	if def, ok := field.Meta[optFieldDefault]; ok && def != nil {
		switch fieldActualType(field) {
		case I8, I16, I32, I64:
			def = int64(def.(float64))
		case U8, U16, U32, U64:
			def = uint64(def.(float64))
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

	// generate definitions
	for _, msg := range g.file.Messages {
		genMessage(g, msg)
	}

	// generate initial registration
	initFnName := fmt.Sprintf("file_%s_binapi_init", g.file.PackageName)
	g.P("func init() { ", initFnName, "() }")
	g.P("func ", initFnName, "() {")
	for _, msg := range g.file.Messages {
		id := fmt.Sprintf("%s_%s", msg.Name, msg.CRC)
		g.P(govppApiPkg.Ident("RegisterMessage"), "((*", msg.GoIdent, ")(nil), ", strconv.Quote(id), ")")
	}
	g.P("}")
	g.P()

	// generate message list
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

	genMessageComment(g, msg)
	genGenericDefinesComment(g, msg.GoName, msg.Name, "message")
	genMessageStatusInfoComment(g, msg)

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

	// base methods
	genMessageBaseMethods(g, msg)

	// encoding methods
	genMessageEncodingMethods(g, msg)

	g.P()
}

func genMessageBaseMethods(g *GenFile, msg *Message) {
	// Reset method
	g.P("func (m *", msg.GoName, ") Reset() { *m = ", msg.GoName, "{} }")

	// GetXXX methods
	genMessageMethods(g, msg)

	g.P()
}

func genMessageComment(g *GenFile, msg *Message) {
	if msg.Comment != "" {
		comment := strings.ReplaceAll(msg.Comment, "\n", "\n// ")
		g.P("// ", comment)
	}
}

func genMessageMethods(g *GenFile, msg *Message) {
	// GetMessageName method
	g.P("func (*", msg.GoName, ") GetMessageName() string { return ", strconv.Quote(msg.Name), " }")

	// GetCrcString method
	g.P("func (*", msg.GoName, ") GetCrcString() string { return ", strconv.Quote(msg.CRC), " }")

	// GetMessageType method
	g.P("func (*", msg.GoName, ") GetMessageType() api.MessageType {")
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
