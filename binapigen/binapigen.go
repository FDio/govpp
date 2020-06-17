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
	"sort"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

type File struct {
	vppapi.File

	Generate bool

	PackageName string
	ImportPath  string

	Enums   []*Enum
	Unions  []*Union
	Structs []*Struct
	Aliases []*Alias

	packageDir string
	refmap     map[string]string //interface{}
}

func newFile(gen *Generator, apifile vppapi.File) (*File, error) {
	file := &File{
		File:        apifile,
		PackageName: normalizePackageName(apifile.Name),
		refmap:      make(map[string]string),
	}

	sortFileObjects(&file.File)

	// load reference map
	for _, enum := range apifile.EnumTypes {
		file.Enums = append(file.Enums, newEnum(gen, file, enum))
	}
	for _, alias := range apifile.AliasTypes {
		file.Aliases = append(file.Aliases, newAlias(gen, file, alias))
	}
	for _, structType := range apifile.StructTypes {
		file.Structs = append(file.Structs, newStruct(gen, file, structType))
	}
	for _, union := range apifile.UnionTypes {
		file.Unions = append(file.Unions, newUnion(gen, file, union))
	}
	/*for _, msg := range apifile.Messages {
		file.Messages = append(file.Messages, newMessage(gen, file, msg))
	}*/

	return file, nil
}

func (file *File) addRef(typ string, name string, ref interface{}) {
	apiName := toApiType(name) // fmt.Sprintf("%s:%s", typ, toApiType(name))
	if _, ok := file.refmap[apiName]; ok {
		logf("%s type %v already in refmap", typ, apiName)
		return
	}
	file.refmap[apiName] = name //ref
}

func sortFileObjects(file *vppapi.File) {
	// sort imports
	sort.SliceStable(file.Imports, func(i, j int) bool {
		return file.Imports[i] < file.Imports[j]
	})
	// sort enum types
	sort.SliceStable(file.EnumTypes, func(i, j int) bool {
		return file.EnumTypes[i].Name < file.EnumTypes[j].Name
	})
	// sort alias types
	sort.Slice(file.AliasTypes, func(i, j int) bool {
		return file.AliasTypes[i].Name < file.AliasTypes[j].Name
	})
	// sort struct types
	sort.SliceStable(file.StructTypes, func(i, j int) bool {
		return file.StructTypes[i].Name < file.StructTypes[j].Name
	})
	// sort union types
	sort.SliceStable(file.UnionTypes, func(i, j int) bool {
		return file.UnionTypes[i].Name < file.UnionTypes[j].Name
	})
	// sort messages
	sort.SliceStable(file.Messages, func(i, j int) bool {
		return file.Messages[i].Name < file.Messages[j].Name
	})
	// sort services
	if file.Service != nil {
		sort.Slice(file.Service.RPCs, func(i, j int) bool {
			// dumps first
			if file.Service.RPCs[i].Stream != file.Service.RPCs[j].Stream {
				return file.Service.RPCs[i].Stream
			}
			return file.Service.RPCs[i].RequestMsg < file.Service.RPCs[j].RequestMsg
		})
	}
}

func normalizePackageName(name string) string {
	switch name {
	case "interface":
		return "interfaces"
	case "map":
		return "maps"
	default:
		return name
	}
}

type Struct struct {
	vppapi.StructType
}

func newStruct(gen *Generator, file *File, apitype vppapi.StructType) *Struct {
	typ := &Struct{
		StructType: apitype,
	}
	file.addRef("struct", typ.Name, typ)
	return typ
}

type Enum struct {
	vppapi.EnumType
}

func newEnum(gen *Generator, file *File, apitype vppapi.EnumType) *Enum {
	typ := &Enum{
		EnumType: apitype,
	}
	file.addRef("enum", typ.Name, typ)
	return typ
}

type Union struct {
	vppapi.UnionType
}

func newUnion(gen *Generator, file *File, apitype vppapi.UnionType) *Union {
	typ := &Union{
		UnionType: apitype,
	}
	file.addRef("union", typ.Name, typ)
	return typ
}

type Alias struct {
	vppapi.AliasType
}

func newAlias(gen *Generator, file *File, apitype vppapi.AliasType) *Alias {
	typ := &Alias{
		AliasType: apitype,
	}
	file.addRef("alias", typ.Name, typ)
	return typ
}

/*type Message struct {
	vppapi.Message
}

func newMessage(gen *Generator, file *File, apitype vppapi.Message) *Message {
	msg := &Message{
		Message: apitype,
	}
	return msg
}*/

type Message = vppapi.Message
type Service = vppapi.Service
type RPC = vppapi.RPC
type Field = vppapi.Field
