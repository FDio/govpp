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

	"git.fd.io/govpp.git/binapigen/vppapi"
)

// generatedCodeVersion indicates a version of the generated code.
// It is incremented whenever an incompatibility between the generated code and
// GoVPP api package is introduced; the generated code references
// a constant, api.GoVppAPIPackageIsVersionN (where N is generatedCodeVersion).
const generatedCodeVersion = 2

// file options
const (
	optFileVersion = "version"
)

type File struct {
	Desc vppapi.File

	Generate       bool
	FilenamePrefix string
	PackageName    GoPackageName
	GoImportPath   GoImportPath

	Version string
	Imports []string

	Enums   []*Enum
	Unions  []*Union
	Structs []*Struct
	Aliases []*Alias

	Messages []*Message
	Service  *Service
}

func newFile(gen *Generator, apifile *vppapi.File, packageName GoPackageName, importPath GoImportPath) (*File, error) {
	file := &File{
		Desc:         *apifile,
		PackageName:  packageName,
		GoImportPath: importPath,
	}
	if apifile.Options != nil {
		file.Version = apifile.Options[optFileVersion]
	}

	file.FilenamePrefix = path.Join(gen.opts.OutputDir, file.Desc.Name)

	for _, imp := range apifile.Imports {
		file.Imports = append(file.Imports, normalizeImport(imp))
	}

	for _, enumType := range apifile.EnumTypes {
		file.Enums = append(file.Enums, newEnum(gen, file, enumType, false))
	}
	for _, enumflagType := range apifile.EnumflagTypes {
		file.Enums = append(file.Enums, newEnum(gen, file, enumflagType, true))
	}
	for _, aliasType := range apifile.AliasTypes {
		file.Aliases = append(file.Aliases, newAlias(gen, file, aliasType))
	}
	for _, structType := range apifile.StructTypes {
		file.Structs = append(file.Structs, newStruct(gen, file, structType))
	}
	for _, unionType := range apifile.UnionTypes {
		file.Unions = append(file.Unions, newUnion(gen, file, unionType))
	}

	for _, msg := range apifile.Messages {
		file.Messages = append(file.Messages, newMessage(gen, file, msg))
	}
	if apifile.Service != nil {
		file.Service = newService(gen, file, *apifile.Service)
	}

	for _, t := range file.Aliases {
		if err := t.resolveDependencies(gen); err != nil {
			return nil, err
		}
	}
	for _, t := range file.Structs {
		if err := t.resolveDependencies(gen); err != nil {
			return nil, err
		}
	}
	for _, t := range file.Unions {
		if err := t.resolveDependencies(gen); err != nil {
			return nil, err
		}
	}
	for _, m := range file.Messages {
		if err := m.resolveDependencies(gen); err != nil {
			return nil, err
		}
	}
	if file.Service != nil {
		for _, rpc := range file.Service.RPCs {
			if err := rpc.resolveMessages(gen); err != nil {
				return nil, err
			}
		}
	}

	return file, nil
}

func (file *File) isTypesFile() bool {
	return strings.HasSuffix(file.Desc.Name, "_types")
}

func (file *File) hasService() bool {
	return file.Service != nil && len(file.Service.RPCs) > 0
}

func (file *File) importedFiles(gen *Generator) []*File {
	var files []*File
	for _, imp := range file.Imports {
		impFile, ok := gen.FilesByName[imp]
		if !ok {
			logf("file %s import %s not found API files", file.Desc.Name, imp)
			continue
		}
		files = append(files, impFile)
	}
	return files
}

func (file *File) dependsOnFile(gen *Generator, dep string) bool {
	for _, imp := range file.Imports {
		if imp == dep {
			return true
		}
		impFile, ok := gen.FilesByName[imp]
		if ok && impFile.dependsOnFile(gen, dep) {
			return true
		}
	}
	return false
}

const (
	enumFlagSuffix = "_flags"
)

func isEnumFlag(enum *Enum) bool {
	return strings.HasSuffix(enum.Name, enumFlagSuffix)
}

type Enum struct {
	vppapi.EnumType

	GoIdent

	IsFlag bool
}

func newEnum(gen *Generator, file *File, apitype vppapi.EnumType, isFlag bool) *Enum {
	typ := &Enum{
		EnumType: apitype,
		GoIdent: GoIdent{
			GoName:       camelCaseName(apitype.Name),
			GoImportPath: file.GoImportPath,
		},
		IsFlag: isFlag,
	}
	gen.enumsByName[typ.Name] = typ
	return typ
}

type Alias struct {
	vppapi.AliasType

	GoIdent

	TypeBasic  *string
	TypeStruct *Struct
	TypeUnion  *Union
}

func newAlias(gen *Generator, file *File, apitype vppapi.AliasType) *Alias {
	typ := &Alias{
		AliasType: apitype,
		GoIdent: GoIdent{
			GoName:       camelCaseName(apitype.Name),
			GoImportPath: file.GoImportPath,
		},
	}
	gen.aliasesByName[typ.Name] = typ
	return typ
}

func (a *Alias) resolveDependencies(gen *Generator) error {
	if err := a.resolveType(gen); err != nil {
		return fmt.Errorf("unable to resolve field: %w", err)
	}
	return nil
}

func (a *Alias) resolveType(gen *Generator) error {
	if _, ok := BaseTypesGo[a.Type]; ok {
		return nil
	}
	typ := fromApiType(a.Type)
	if t, ok := gen.structsByName[typ]; ok {
		a.TypeStruct = t
		return nil
	}
	if t, ok := gen.unionsByName[typ]; ok {
		a.TypeUnion = t
		return nil
	}
	return fmt.Errorf("unknown type: %q", a.Type)
}

type Struct struct {
	vppapi.StructType

	GoIdent

	Fields []*Field
}

func newStruct(gen *Generator, file *File, apitype vppapi.StructType) *Struct {
	typ := &Struct{
		StructType: apitype,
		GoIdent: GoIdent{
			GoName:       camelCaseName(apitype.Name),
			GoImportPath: file.GoImportPath,
		},
	}
	gen.structsByName[typ.Name] = typ
	for i, fieldType := range apitype.Fields {
		field := newField(gen, file, typ, fieldType, i)
		typ.Fields = append(typ.Fields, field)
	}
	return typ
}

func (m *Struct) resolveDependencies(gen *Generator) (err error) {
	for _, field := range m.Fields {
		if err := field.resolveDependencies(gen); err != nil {
			return fmt.Errorf("unable to resolve for struct %s: %w", m.Name, err)
		}
	}
	return nil
}

type Union struct {
	vppapi.UnionType

	GoIdent

	Fields []*Field
}

func newUnion(gen *Generator, file *File, apitype vppapi.UnionType) *Union {
	typ := &Union{
		UnionType: apitype,
		GoIdent: GoIdent{
			GoName:       withSuffix(camelCaseName(apitype.Name), "Union"),
			GoImportPath: file.GoImportPath,
		},
	}
	gen.unionsByName[typ.Name] = typ
	for i, fieldType := range apitype.Fields {
		field := newField(gen, file, typ, fieldType, i)
		typ.Fields = append(typ.Fields, field)
	}
	return typ
}

func (m *Union) resolveDependencies(gen *Generator) (err error) {
	for _, field := range m.Fields {
		if err := field.resolveDependencies(gen); err != nil {
			return err
		}
	}
	return nil
}

// msgType determines message header fields
type msgType int

const (
	msgTypeBase    msgType = iota // msg_id
	msgTypeRequest                // msg_id, client_index, context
	msgTypeReply                  // msg_id, context
	msgTypeEvent                  // msg_id, client_index
)

// common message fields
const (
	fieldMsgID       = "_vl_msg_id"
	fieldClientIndex = "client_index"
	fieldContext     = "context"
	fieldRetval      = "retval"
)

// field options
const (
	optFieldDefault = "default"
)

type Message struct {
	vppapi.Message

	CRC string

	GoIdent

	Fields []*Field

	msgType msgType
}

func newMessage(gen *Generator, file *File, apitype vppapi.Message) *Message {
	msg := &Message{
		Message: apitype,
		CRC:     strings.TrimPrefix(apitype.CRC, "0x"),
		GoIdent: newGoIdent(file, apitype.Name),
	}
	gen.messagesByName[apitype.Name] = msg
	var n int
	for _, fieldType := range apitype.Fields {
		if n == 0 {
			// skip header fields
			switch strings.ToLower(fieldType.Name) {
			case fieldMsgID, fieldClientIndex, fieldContext:
				continue
			}
		}
		n++
		field := newField(gen, file, msg, fieldType, n)
		msg.Fields = append(msg.Fields, field)
	}
	return msg
}

func (m *Message) resolveDependencies(gen *Generator) (err error) {
	if m.msgType, err = getMsgType(m.Message); err != nil {
		return err
	}
	for _, field := range m.Fields {
		if err := field.resolveDependencies(gen); err != nil {
			return err
		}
	}
	return nil
}

func getMsgType(m vppapi.Message) (msgType, error) {
	if len(m.Fields) == 0 {
		return msgType(-1), fmt.Errorf("message %s has no fields", m.Name)
	}
	var typ msgType
	var wasClientIndex bool
	for i, field := range m.Fields {
		switch i {
		case 0:
			if field.Name != fieldMsgID {
				return msgType(-1), fmt.Errorf("message %s is missing ID field", m.Name)
			}
		case 1:
			if field.Name == fieldClientIndex {
				// "client_index" as the second member,
				// this might be an event message or a request
				typ = msgTypeEvent
				wasClientIndex = true
			} else if field.Name == fieldContext {
				// reply needs "context" as the second member
				typ = msgTypeReply
			}
		case 2:
			if field.Name == fieldContext && wasClientIndex {
				// request needs "client_index" as the second member
				// and "context" as the third member
				typ = msgTypeRequest
			}
		}
	}
	return typ, nil
}

func getRetvalField(m *Message) *Field {
	for _, field := range m.Fields {
		if field.Name == fieldRetval {
			return field
		}
	}
	return nil
}

// Field represents a field for message or struct/union types.
type Field struct {
	vppapi.Field

	GoName string

	// Index defines field index in parent.
	Index int

	// DefaultValue is a default value of field or
	// nil if default value is not defined for field.
	DefaultValue interface{}

	// Reference to actual type of this field.
	//
	// For fields with built-in types all of these are nil,
	// otherwise only one is set to non-nil value.
	TypeEnum   *Enum
	TypeAlias  *Alias
	TypeStruct *Struct
	TypeUnion  *Union

	// Parent in which this field is declared.
	ParentMessage *Message
	ParentStruct  *Struct
	ParentUnion   *Union

	// Field reference for fields with variable size.
	FieldSizeOf   *Field
	FieldSizeFrom *Field
}

func newField(gen *Generator, file *File, parent interface{}, apitype vppapi.Field, index int) *Field {
	typ := &Field{
		Field:  apitype,
		GoName: camelCaseName(apitype.Name),
		Index:  index,
	}
	switch p := parent.(type) {
	case *Message:
		typ.ParentMessage = p
	case *Struct:
		typ.ParentStruct = p
	case *Union:
		typ.ParentUnion = p
	default:
		panic(fmt.Sprintf("invalid field parent: %T", parent))
	}
	if apitype.Meta != nil {
		if val, ok := apitype.Meta[optFieldDefault]; ok {
			typ.DefaultValue = val
		}
	}
	return typ
}

func (f *Field) resolveDependencies(gen *Generator) error {
	if err := f.resolveType(gen); err != nil {
		return fmt.Errorf("unable to resolve field type: %w", err)
	}
	if err := f.resolveFields(gen); err != nil {
		return fmt.Errorf("unable to resolve fields: %w", err)
	}
	return nil
}

func (f *Field) resolveType(gen *Generator) error {
	if _, ok := BaseTypesGo[f.Type]; ok {
		return nil
	}
	typ := fromApiType(f.Type)
	if t, ok := gen.structsByName[typ]; ok {
		f.TypeStruct = t
		return nil
	}
	if t, ok := gen.enumsByName[typ]; ok {
		f.TypeEnum = t
		return nil
	}
	if t, ok := gen.aliasesByName[typ]; ok {
		f.TypeAlias = t
		return nil
	}
	if t, ok := gen.unionsByName[typ]; ok {
		f.TypeUnion = t
		return nil
	}
	return fmt.Errorf("unknown type: %q", f.Type)
}

func (f *Field) resolveFields(gen *Generator) error {
	var fields []*Field
	if f.ParentMessage != nil {
		fields = f.ParentMessage.Fields
	} else if f.ParentStruct != nil {
		fields = f.ParentStruct.Fields
	}
	if f.SizeFrom != "" {
		for _, field := range fields {
			if field.Name == f.SizeFrom {
				f.FieldSizeFrom = field
				break
			}
		}
	} else {
		for _, field := range fields {
			if field.SizeFrom == f.Name {
				f.FieldSizeOf = field
				break
			}
		}
	}
	return nil
}

type Service struct {
	vppapi.Service

	RPCs []*RPC
}

func newService(gen *Generator, file *File, apitype vppapi.Service) *Service {
	svc := &Service{
		Service: apitype,
	}
	for _, rpc := range apitype.RPCs {
		svc.RPCs = append(svc.RPCs, newRpc(file, svc, rpc))
	}
	return svc
}

const (
	serviceNoReply = "null"
)

type RPC struct {
	VPP vppapi.RPC

	GoName string

	Service *Service

	MsgRequest *Message
	MsgReply   *Message
	MsgStream  *Message
}

func newRpc(file *File, service *Service, apitype vppapi.RPC) *RPC {
	rpc := &RPC{
		VPP:     apitype,
		GoName:  camelCaseName(apitype.Request),
		Service: service,
	}
	return rpc
}

func (rpc *RPC) resolveMessages(gen *Generator) error {
	msg, ok := gen.messagesByName[rpc.VPP.Request]
	if !ok {
		return fmt.Errorf("rpc %v: no message for request type %v", rpc.GoName, rpc.VPP.Request)
	}
	rpc.MsgRequest = msg

	if rpc.VPP.Reply != "" && rpc.VPP.Reply != serviceNoReply {
		msg, ok := gen.messagesByName[rpc.VPP.Reply]
		if !ok {
			return fmt.Errorf("rpc %v: no message for reply type %v", rpc.GoName, rpc.VPP.Reply)
		}
		rpc.MsgReply = msg
	}
	if rpc.VPP.StreamMsg != "" {
		msg, ok := gen.messagesByName[rpc.VPP.StreamMsg]
		if !ok {
			return fmt.Errorf("rpc %v: no message for stream type %v", rpc.GoName, rpc.VPP.StreamMsg)
		}
		rpc.MsgStream = msg
	}
	return nil
}

// GoIdent is a Go identifier, consisting of a name and import path.
// The name is a single identifier and may not be a dot-qualified selector.
type GoIdent struct {
	GoName       string
	GoImportPath GoImportPath
}

func (id GoIdent) String() string {
	return fmt.Sprintf("%q.%v", id.GoImportPath, id.GoName)
}

func newGoIdent(f *File, fullName string) GoIdent {
	name := strings.TrimPrefix(fullName, string(f.PackageName)+".")
	return GoIdent{
		GoName:       camelCaseName(name),
		GoImportPath: f.GoImportPath,
	}
}

// GoImportPath is a Go import path for a package.
type GoImportPath string

func (p GoImportPath) String() string {
	return strconv.Quote(string(p))
}

func (p GoImportPath) Ident(s string) GoIdent {
	return GoIdent{GoName: s, GoImportPath: p}
}

type GoPackageName string

func cleanPackageName(name string) GoPackageName {
	return GoPackageName(sanitizedName(name))
}

// baseName returns the last path element of the name, with the last dotted suffix removed.
func baseName(name string) string {
	// First, find the last element
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	// Now drop the suffix
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[:i]
	}
	return name
}

// normalizeImport returns the last path element of the import, with all dotted suffixes removed.
func normalizeImport(imp string) string {
	imp = path.Base(imp)
	if idx := strings.Index(imp, "."); idx >= 0 {
		imp = imp[:idx]
	}
	return imp
}
