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
	"sort"
	"strings"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

type File struct {
	vppapi.File

	Generate bool

	PackageName string
	Imports     []string

	Enums    []*Enum
	Unions   []*Union
	Structs  []*Struct
	Aliases  []*Alias
	Messages []*Message

	imports map[string]string
	refmap  map[string]string
}

func newFile(gen *Generator, apifile *vppapi.File) (*File, error) {
	file := &File{
		File:        *apifile,
		PackageName: sanitizedName(apifile.Name),
		imports:     make(map[string]string),
		refmap:      make(map[string]string),
	}

	sortFileObjects(&file.File)

	for _, imp := range apifile.Imports {
		file.Imports = append(file.Imports, normalizeImport(imp))
	}
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
	for _, msg := range apifile.Messages {
		file.Messages = append(file.Messages, newMessage(gen, file, msg))
	}

	return file, nil
}

func (file *File) isTypes() bool {
	return strings.HasSuffix(file.File.Name, "_types")
}

func (file *File) hasService() bool {
	return file.Service != nil && len(file.Service.RPCs) > 0
}

func (file *File) addRef(typ string, name string, ref interface{}) {
	apiName := toApiType(name)
	if _, ok := file.refmap[apiName]; ok {
		logf("%s type %v already in refmap", typ, apiName)
		return
	}
	file.refmap[apiName] = name
}

func (file *File) importedFiles(gen *Generator) []*File {
	var files []*File
	for _, imp := range file.Imports {
		impFile, ok := gen.FilesByName[imp]
		if !ok {
			logf("file %s import %s not found API files", file.Name, imp)
			continue
		}
		files = append(files, impFile)
	}
	return files
}

func (file *File) loadTypeImports(gen *Generator, typeFiles []*File) {
	if len(typeFiles) == 0 {
		return
	}
	for _, t := range file.Structs {
		for _, imp := range typeFiles {
			if _, ok := file.imports[t.Name]; ok {
				break
			}
			for _, at := range imp.File.StructTypes {
				if at.Name != t.Name {
					continue
				}
				if len(at.Fields) != len(t.Fields) {
					continue
				}
				file.imports[t.Name] = imp.PackageName
			}
		}
	}
	for _, t := range file.AliasTypes {
		for _, imp := range typeFiles {
			if _, ok := file.imports[t.Name]; ok {
				break
			}
			for _, at := range imp.File.AliasTypes {
				if at.Name != t.Name {
					continue
				}
				if at.Length != t.Length {
					continue
				}
				if at.Type != t.Type {
					continue
				}
				file.imports[t.Name] = imp.PackageName
			}
		}
	}
	for _, t := range file.EnumTypes {
		for _, imp := range typeFiles {
			if _, ok := file.imports[t.Name]; ok {
				break
			}
			for _, at := range imp.File.EnumTypes {
				if at.Name != t.Name {
					continue
				}
				if at.Type != t.Type {
					continue
				}
				file.imports[t.Name] = imp.PackageName
			}
		}
	}
	for _, t := range file.UnionTypes {
		for _, imp := range typeFiles {
			if _, ok := file.imports[t.Name]; ok {
				break
			}
			for _, at := range imp.File.UnionTypes {
				if at.Name != t.Name {
					continue
				}
				file.imports[t.Name] = imp.PackageName
				/*if gen.ImportTypes {
					imp.Generate = true
				}*/
			}
		}
	}
}

type Enum struct {
	vppapi.EnumType

	GoName string
}

func newEnum(gen *Generator, file *File, apitype vppapi.EnumType) *Enum {
	typ := &Enum{
		EnumType: apitype,
		GoName:   camelCaseName(apitype.Name),
	}
	gen.enumsByName[fmt.Sprintf("%s.%s", file.Name, typ.Name)] = typ
	file.addRef("enum", typ.Name, typ)
	return typ
}

type Alias struct {
	vppapi.AliasType

	GoName string
}

func newAlias(gen *Generator, file *File, apitype vppapi.AliasType) *Alias {
	typ := &Alias{
		AliasType: apitype,
		GoName:    camelCaseName(apitype.Name),
	}
	gen.aliasesByName[fmt.Sprintf("%s.%s", file.Name, typ.Name)] = typ
	file.addRef("alias", typ.Name, typ)
	return typ
}

type Struct struct {
	vppapi.StructType

	GoName string

	Fields []*Field
}

func newStruct(gen *Generator, file *File, apitype vppapi.StructType) *Struct {
	typ := &Struct{
		StructType: apitype,
		GoName:     camelCaseName(apitype.Name),
	}
	for _, fieldType := range apitype.Fields {
		field := newField(gen, file, fieldType)
		field.ParentStruct = typ
		typ.Fields = append(typ.Fields, field)
	}
	gen.structsByName[fmt.Sprintf("%s.%s", file.Name, typ.Name)] = typ
	file.addRef("struct", typ.Name, typ)
	return typ
}

type Union struct {
	vppapi.UnionType

	GoName string

	Fields []*Field
}

func newUnion(gen *Generator, file *File, apitype vppapi.UnionType) *Union {
	typ := &Union{
		UnionType: apitype,
		GoName:    camelCaseName(apitype.Name),
	}
	gen.unionsByName[fmt.Sprintf("%s.%s", file.Name, typ.Name)] = typ
	for _, fieldType := range apitype.Fields {
		field := newField(gen, file, fieldType)
		field.ParentUnion = typ
		typ.Fields = append(typ.Fields, field)
	}
	file.addRef("union", typ.Name, typ)
	return typ
}

type Message struct {
	vppapi.Message

	GoName string

	Fields []*Field
}

func newMessage(gen *Generator, file *File, apitype vppapi.Message) *Message {
	msg := &Message{
		Message: apitype,
		GoName:  camelCaseName(apitype.Name),
	}
	for _, fieldType := range apitype.Fields {
		field := newField(gen, file, fieldType)
		field.ParentMessage = msg
		msg.Fields = append(msg.Fields, field)
	}
	return msg
}

type Field struct {
	vppapi.Field

	GoName string

	// Field parent
	ParentMessage *Message
	ParentStruct  *Struct
	ParentUnion   *Union

	// Type reference
	Enum   *Enum
	Alias  *Alias
	Struct *Struct
	Union  *Union
}

func newField(gen *Generator, file *File, apitype vppapi.Field) *Field {
	typ := &Field{
		Field:  apitype,
		GoName: camelCaseName(apitype.Name),
	}
	return typ
}

type Service = vppapi.Service
type RPC = vppapi.RPC

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

func sanitizedName(name string) string {
	switch name {
	case "interface":
		return "interfaces"
	case "map":
		return "maps"
	default:
		return name
	}
}

func normalizeImport(imp string) string {
	imp = path.Base(imp)
	if idx := strings.Index(imp, "."); idx >= 0 {
		imp = imp[:idx]
	}
	return imp
}
