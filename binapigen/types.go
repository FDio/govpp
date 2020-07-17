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
	"strings"

	"github.com/sirupsen/logrus"
)

// define api
const (
	defineApiPrefix = "vl_api_"
	defineApiSuffix = "_t"
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

const (
	U8     = "u8"
	I8     = "i8"
	U16    = "u16"
	I16    = "i16"
	U32    = "u32"
	I32    = "i32"
	U64    = "u64"
	I64    = "i64"
	F64    = "f64"
	BOOL   = "bool"
	STRING = "string"
)

var BaseTypeSizes = map[string]int{
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

var BaseTypesGo = map[string]string{
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

func fieldActualType(field *Field) (actual string) {
	switch {
	case field.TypeAlias != nil:
		actual = field.TypeAlias.Type
	case field.TypeEnum != nil:
		actual = field.TypeEnum.Type
	}
	return field.Type
}

func fieldGoType(g *GenFile, field *Field) string {
	switch {
	case field.TypeAlias != nil:
		return g.GoIdent(field.TypeAlias.GoIdent)
	case field.TypeEnum != nil:
		return g.GoIdent(field.TypeEnum.GoIdent)
	case field.TypeStruct != nil:
		return g.GoIdent(field.TypeStruct.GoIdent)
	case field.TypeUnion != nil:
		return g.GoIdent(field.TypeUnion.GoIdent)
	}
	t, ok := BaseTypesGo[field.Type]
	if !ok {
		logrus.Panicf("type %s is not base type", field.Type)
	}
	return t
}

func getFieldType(g *GenFile, field *Field) string {
	gotype := fieldGoType(g, field)
	if field.Array {
		switch gotype {
		case "uint8":
			return "[]byte"
		case "string":
			return "string"
		}
		if _, ok := BaseTypesGo[field.Type]; !ok && field.Length > 0 {
			return fmt.Sprintf("[%d]%s", field.Length, gotype)
		}
		return "[]" + gotype
	}
	return gotype
}

func getSizeOfBinapiTypeLength(typ string, length int) (size int) {
	if n := BaseTypeSizes[typ]; n > 0 {
		if length > 0 {
			return n * length
		} else {
			return n
		}
	}
	return
}

func getSizeOfType(typ *Struct) (size int) {
	for _, field := range typ.Fields {
		if enum := field.TypeEnum; enum != nil {
			size += getSizeOfBinapiTypeLength(enum.Type, field.Length)
			continue
		}
		size += getSizeOfBinapiTypeLength(field.Type, field.Length)
	}
	return size
}

func getUnionSize(union *Union) (maxSize int) {
	for _, field := range union.Fields {
		if typ := field.TypeStruct; typ != nil {
			if size := getSizeOfType(typ); size > maxSize {
				maxSize = size
			}
			continue
		}
		if alias := field.TypeAlias; alias != nil {
			if size := getSizeOfBinapiTypeLength(alias.Type, alias.Length); size > maxSize {
				maxSize = size
			}
			continue
		} else {
			logrus.Panicf("no type or alias found for union %s field type %q", union.Name, field.Type)
		}
	}
	//logf("getUnionSize: %s %+v max=%v", union.Name, union.Fields, maxSize)
	return
}
