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
	"strings"

	"github.com/sirupsen/logrus"
)

// define api
const (
	defineApiPrefix = "vl_api_"
	defineApiSuffix = "_t"
)

// BaseType represents base types in VPP binary API.
type BaseType int

const (
	U8 BaseType = iota + 1
	I8
	U16
	I16
	U32
	I32
	U64
	I64
	F64
	BOOL
	STRING
)

var (
	BaseTypes = map[BaseType]string{
		U8:     "u8",
		I8:     "i8",
		U16:    "u16",
		I16:    "i16",
		U32:    "u32",
		I32:    "i32",
		U64:    "u64",
		I64:    "i64",
		F64:    "f64",
		BOOL:   "bool",
		STRING: "string",
	}
	BaseTypeNames = map[string]BaseType{
		"u8":     U8,
		"i8":     I8,
		"u16":    U16,
		"i16":    I16,
		"u32":    U32,
		"i32":    I32,
		"u64":    U64,
		"i64":    I64,
		"f64":    F64,
		"bool":   BOOL,
		"string": STRING,
	}
)

var BaseTypeSizes = map[BaseType]int{
	U8:     1,
	I8:     1,
	U16:    2,
	I16:    2,
	U32:    4,
	I32:    4,
	U64:    8,
	I64:    8,
	F64:    8,
	BOOL:   1,
	STRING: 1,
}

type Kind int

const (
	_ = iota
	Uint8Kind
	Int8Kind
	Uint16Kind
	Int16Kind
	Uint32Kind
	Int32Kind
	Uint64Kind
	Int64Kind
	Float64Kind
	BoolKind
	StringKind
	EnumKind
	AliasKind
	StructKind
	UnionKind
	MessageKind
)

// toApiType returns name that is used as type reference in VPP binary API
func toApiType(name string) string {
	return defineApiPrefix + name + defineApiSuffix
}

func fromApiType(typ string) string {
	name := typ
	name = strings.TrimPrefix(name, defineApiPrefix)
	name = strings.TrimSuffix(name, defineApiSuffix)
	return name
}

func getSizeOfType(module *File, typ *Struct) (size int) {
	for _, field := range typ.Fields {
		enum := getEnumByRef(module, field.Type)
		if enum != nil {
			size += getSizeOfBinapiTypeLength(enum.Type, field.Length)
			continue
		}
		size += getSizeOfBinapiTypeLength(field.Type, field.Length)
	}
	return size
}

func getEnumByRef(file *File, ref string) *Enum {
	for _, typ := range file.Enums {
		if ref == toApiType(typ.Name) {
			return typ
		}
	}
	return nil
}

func getTypeByRef(file *File, ref string) *Struct {
	for _, typ := range file.Structs {
		if ref == toApiType(typ.Name) {
			return typ
		}
	}
	return nil
}

func getAliasByRef(file *File, ref string) *Alias {
	for _, alias := range file.Aliases {
		if ref == toApiType(alias.Name) {
			return alias
		}
	}
	return nil
}

func getUnionByRef(file *File, ref string) *Union {
	for _, union := range file.Unions {
		if ref == toApiType(union.Name) {
			return union
		}
	}
	return nil
}

func getBinapiTypeSize(binapiType string) (size int) {
	typName := BaseTypeNames[binapiType]
	return BaseTypeSizes[typName]
}

// toApiType returns name that is used as type reference in VPP binary API
/*func toApiType(name string) string {
	return fmt.Sprintf("vl_api_%s_t", name)
}

func fromApiType(typ string) string {
	name := typ
	name = strings.TrimPrefix(name, "vl_api_")
	name = strings.TrimSuffix(name, "_t")
	return name
}*/

// binapiTypes is a set of types used VPP binary API for translation to Go types
var binapiTypes = map[string]string{
	"u8":  "uint8",
	"i8":  "int8",
	"u16": "uint16",
	"i16": "int16",
	"u32": "uint32",
	"i32": "int32",
	"u64": "uint64",
	"i64": "int64",
	"f64": "float64",
}
var BaseTypesGo = map[BaseType]string{
	U8:     "uint8",
	I8:     "int8",
	U16:    "uint16",
	I16:    "int16",
	U32:    "uint32",
	I32:    "int32",
	U64:    "uint64",
	I64:    "int64",
	F64:    "float64",
	BOOL:   "bool",
	STRING: "string",
}

/*func getBinapiTypeSize(binapiType string) int {
	if _, ok := binapiTypes[binapiType]; ok {
		b, err := strconv.Atoi(strings.TrimLeft(binapiType, "uif"))
		if err == nil {
			return b / 8
		}
	}
	return -1
}*/

func getActualType(file *File, typ string) (actual string) {
	for _, enum := range file.EnumTypes {
		if camelCaseName(enum.Name) == typ {
			return enum.Type
		}
	}
	for _, alias := range file.AliasTypes {
		if camelCaseName(alias.Name) == typ {
			return alias.Type
		}
	}
	return typ
}

// convertToGoType translates the VPP binary API type into Go type
func convertToGoType(file *File, binapiType string) (typ string) {
	if t, ok := binapiTypes[binapiType]; ok {
		// basic types
		typ = t
	} else if r, ok := file.refmap[binapiType]; ok {
		// specific types (enums/types/unions)
		typ = camelCaseName(r)
	} else {
		switch binapiType {
		case "bool", "string":
			typ = binapiType
		default:
			// fallback type
			logrus.Warnf("found unknown VPP binary API type %q, using byte", binapiType)
			typ = "byte"
		}
	}
	return typ
}

/*func getSizeOfType(ctx *Context, typ *Type) (size int) {
	for _, field := range typ.Fields {
		enum := getEnumByRef(ctx, field.Type)
		if enum != nil {
			size += getSizeOfBinapiTypeLength(enum.Type, field.Length)
			continue
		}
		size += getSizeOfBinapiTypeLength(field.Type, field.Length)
	}
	return size
}*/

func getSizeOfBinapiTypeLength(typ string, length int) (size int) {
	if n := getBinapiTypeSize(typ); n > 0 {
		if length > 0 {
			return n * length
		} else {
			return n
		}
	}

	return
}

/*func getEnumByRef(ctx *Context, ref string) *EnumType {
	for _, typ := range ctx.packageData.EnumTypes {
		if ref == toApiType(typ.Name) {
			return &typ
		}
	}
	return nil
}

func getTypeByRef(ctx *Context, ref string) *Type {
	for _, typ := range ctx.packageData.StructTypes {
		if ref == toApiType(typ.Name) {
			return &typ
		}
	}
	return nil
}

func getAliasByRef(ctx *Context, ref string) *Alias {
	for _, alias := range ctx.packageData.AliasTypes {
		if ref == toApiType(alias.Name) {
			return &alias
		}
	}
	return nil
}

func getUnionByRef(ctx *Context, ref string) *Union {
	for _, union := range ctx.packageData.UnionTypes {
		if ref == toApiType(union.Name) {
			return &union
		}
	}
	return nil
}*/

func getUnionSize(file *File, union *Union) (maxSize int) {
	for _, field := range union.Fields {
		typ := getTypeByRef(file, field.Type)
		if typ != nil {
			if size := getSizeOfType(file, typ); size > maxSize {
				maxSize = size
			}
			continue
		}
		alias := getAliasByRef(file, field.Type)
		if alias != nil {
			if size := getSizeOfBinapiTypeLength(alias.Type, alias.Length); size > maxSize {
				maxSize = size
			}
			continue
		} else {
			logf("no type or alias found for union %s field type %q", union.Name, field.Type)
			continue
		}
	}
	logf("getUnionSize: %s %+v max=%v", union.Name, union.Fields, maxSize)
	return
}
