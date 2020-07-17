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
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	//RegisterPlugin("encoding", GenerateEncoding)
}

func generateMessageSize(g *GenFile, name string, fields []*Field) {
	g.P("func (m *", name, ") Size() int {")
	g.P("if m == nil { return 0 }")
	g.P("var size int")

	sizeBaseType := func(typ, name string, length int, sizefrom string) {
		switch typ {
		case STRING:
			if length > 0 {
				g.P("size += ", length, " // ", name)
			} else {
				g.P("size += 4 + len(", name, ")", " // ", name)
			}
		default:
			var size = BaseTypeSizes[typ]
			if sizefrom != "" {
				g.P("size += ", size, " * len(", name, ")", " // ", name)
			} else {
				if length > 0 {
					g.P("size += ", size, " * ", length, " // ", name)
				} else {
					g.P("size += ", size, " // ", name)
				}
			}
		}
	}

	lvl := 0
	var sizeFields func(fields []*Field, parentName string)
	sizeFields = func(fields []*Field, parentName string) {
		lvl++
		defer func() { lvl-- }()

		getFieldName := func(name string) string {
			return fmt.Sprintf("%s.%s", parentName, name)
		}

		for _, field := range fields {
			name := getFieldName(field.GoName)

			var sizeFromName string
			if field.FieldSizeFrom != nil {
				sizeFromName = getFieldName(field.FieldSizeFrom.GoName)
			}

			if _, ok := BaseTypesGo[field.Type]; ok {
				sizeBaseType(field.Type, name, field.Length, sizeFromName)
				continue
			}

			if field.Array {
				char := fmt.Sprintf("s%d", lvl)
				index := fmt.Sprintf("j%d", lvl)
				if field.Length > 0 {
					g.P("for ", index, " := 0; ", index, " < ", field.Length, "; ", index, "++ {")
				} else if field.FieldSizeFrom != nil {
					g.P("for ", index, " := 0; ", index, " < len(", name, "); ", index, "++ {")
				}
				g.P("var ", char, " ", fieldGoType(g, field))
				g.P("_ = ", char)
				g.P("if ", index, " < len(", name, ") { ", char, " = ", name, "[", index, "] }")
				name = char
			}

			switch {
			case field.TypeEnum != nil:
				enum := field.TypeEnum
				if _, ok := BaseTypesGo[enum.Type]; ok {
					sizeBaseType(enum.Type, name, 0, "")
				} else {
					logrus.Panicf("\t// ??? ENUM %s %s\n", name, enum.Type)
				}
			case field.TypeAlias != nil:
				alias := field.TypeAlias
				if typ := alias.TypeStruct; typ != nil {
					sizeFields(typ.Fields, name)
				} else {
					sizeBaseType(alias.Type, name, alias.Length, "")
				}
			case field.TypeStruct != nil:
				typ := field.TypeStruct
				sizeFields(typ.Fields, name)
			case field.TypeUnion != nil:
				union := field.TypeUnion
				maxSize := getUnionSize(union)
				sizeBaseType("u8", name, maxSize, "")
			default:
				logrus.Panicf("\t// ??? buf[pos] = %s (%s)\n", name, field.Type)
			}

			if field.Array {
				g.P("}")
			}
		}
	}
	sizeFields(fields, "m")

	g.P("return size")
	g.P("}")
}

func encodeBaseType(g *GenFile, typ, name string, length int, sizefrom string) {
	isArray := length > 0 || sizefrom != ""
	if isArray {
		switch typ {
		case U8:
			g.P("buf.EncodeBytes(", name, "[:], ", length, ")")
			return
		case I8, I16, U16, I32, U32, I64, U64, F64:
			gotype := BaseTypesGo[typ]
			if length != 0 {
				g.P("for i := 0; i < ", length, "; i++ {")
			} else if sizefrom != "" {
				g.P("for i := 0; i < len(", name, "); i++ {")
			}
			g.P("var x ", gotype)
			g.P("if i < len(", name, ") { x = ", gotype, "(", name, "[i]) }")
			name = "x"
		}
	}
	switch typ {
	case I8, U8, I16, U16, I32, U32, I64, U64:
		typsize := BaseTypeSizes[typ]
		g.P("buf.EncodeUint", typsize*8, "(uint", typsize*8, "(", name, "))")
	case F64:
		g.P("buf.EncodeFloat64(float64(", name, "))")
	case BOOL:
		g.P("buf.EncodeBool(", name, ")")
	case STRING:
		g.P("buf.EncodeString(", name, ", ", length, ")")
	default:
		logrus.Panicf("// ??? %s %s\n", name, typ)
	}
	if isArray {
		switch typ {
		case I8, U8, I16, U16, I32, U32, I64, U64, F64:
			g.P("}")
		}
	}
}

func encodeFields(g *GenFile, fields []*Field, parentName string, lvl int) {
	getFieldName := func(name string) string {
		return fmt.Sprintf("%s.%s", parentName, name)
	}

	for _, field := range fields {
		name := getFieldName(field.GoName)

		encodeField(g, field, name, getFieldName, lvl)
	}
}

func encodeField(g *GenFile, field *Field, name string, getFieldName func(name string) string, lvl int) {
	if f := field.FieldSizeOf; f != nil {
		if _, ok := BaseTypesGo[field.Type]; ok {
			encodeBaseType(g, field.Type, fmt.Sprintf("len(%s)", getFieldName(f.GoName)), field.Length, "")
			return
		} else {
			panic(fmt.Sprintf("failed to encode base type of sizefrom field: %s (%s)", field.Name, field.Type))
		}
	}
	var sizeFromName string
	if field.FieldSizeFrom != nil {
		sizeFromName = getFieldName(field.FieldSizeFrom.GoName)
	}

	if _, ok := BaseTypesGo[field.Type]; ok {
		encodeBaseType(g, field.Type, name, field.Length, sizeFromName)
		return
	}

	if field.Array {
		char := fmt.Sprintf("v%d", lvl)
		index := fmt.Sprintf("j%d", lvl)
		if field.Length > 0 {
			g.P("for ", index, " := 0; ", index, " < ", field.Length, "; ", index, "++ {")
		} else if field.SizeFrom != "" {
			g.P("for ", index, " := 0; ", index, " < len(", name, "); ", index, "++ {")
		}
		g.P("var ", char, " ", fieldGoType(g, field))
		g.P("if ", index, " < len(", name, ") { ", char, " = ", name, "[", index, "] }")
		name = char
	}

	switch {
	case field.TypeEnum != nil:
		encodeBaseType(g, field.TypeEnum.Type, name, 0, "")
	case field.TypeAlias != nil:
		alias := field.TypeAlias
		if typ := alias.TypeStruct; typ != nil {
			encodeFields(g, typ.Fields, name, lvl+1)
		} else {
			encodeBaseType(g, alias.Type, name, alias.Length, "")
		}
	case field.TypeStruct != nil:
		encodeFields(g, field.TypeStruct.Fields, name, lvl+1)
	case field.TypeUnion != nil:
		g.P("buf.EncodeBytes(", name, ".", fieldUnionData, "[:], 0)")
	default:
		logrus.Panicf("\t// ??? buf[pos] = %s (%s)\n", name, field.Type)
	}

	if field.Array {
		g.P("}")
	}
}

func generateMessageMarshal(g *GenFile, name string, fields []*Field) {
	g.P("func (m *", name, ") Marshal(b []byte) ([]byte, error) {")
	g.P("var buf *", govppCodecPkg.Ident("Buffer"))
	g.P("if b == nil {")
	g.P("buf = ", govppCodecPkg.Ident("NewBuffer"), "(make([]byte, m.Size()))")
	g.P("} else {")
	g.P("buf = ", govppCodecPkg.Ident("NewBuffer"), "(b)")
	g.P("}")

	encodeFields(g, fields, "m", 0)

	g.P("return buf.Bytes(), nil")
	g.P("}")
}

func decodeBaseType(g *GenFile, typ, orig, name string, length int, sizefrom string, alloc bool) {
	isArray := length > 0 || sizefrom != ""
	if isArray {
		switch typ {
		case U8:
			g.P("copy(", name, "[:], buf.DecodeBytes(", length, "))")
			return
		case I8, I16, U16, I32, U32, I64, U64, F64:
			if alloc {
				var size string
				switch {
				case length > 0:
					size = strconv.Itoa(length)
				case sizefrom != "":
					size = sizefrom
				}
				if size != "" {
					g.P(name, " = make([]", orig, ", ", size, ")")
				}
			}
			g.P("for i := 0; i < len(", name, "); i++ {")
			name = fmt.Sprintf("%s[i]", name)
		}
	}
	switch typ {
	case I8, U8, I16, U16, I32, U32, I64, U64:
		typsize := BaseTypeSizes[typ]
		if gotype, ok := BaseTypesGo[typ]; !ok || gotype != orig || strings.HasPrefix(orig, "i") {
			g.P(name, " = ", orig, "(buf.DecodeUint", typsize*8, "())")
		} else {
			g.P(name, " = buf.DecodeUint", typsize*8, "()")
		}
	case F64:
		g.P(name, " = ", orig, "(buf.DecodeFloat64())")
	case BOOL:
		g.P(name, " = buf.DecodeBool()")
	case STRING:
		g.P(name, " = buf.DecodeString(", length, ")")
	default:
		logrus.Panicf("\t// ??? %s %s\n", name, typ)
	}
	if isArray {
		switch typ {
		case I8, U8, I16, U16, I32, U32, I64, U64, F64:
			g.P("}")
		}
	}
}

func generateMessageUnmarshal(g *GenFile, name string, fields []*Field) {
	g.P("func (m *", name, ") Unmarshal(b []byte) error {")

	if len(fields) > 0 {
		g.P("buf := ", govppCodecPkg.Ident("NewBuffer"), "(b)")
		decodeFields(g, fields, "m", 0)
	}

	g.P("return nil")
	g.P("}")
}

func decodeFields(g *GenFile, fields []*Field, parentName string, lvl int) {
	getFieldName := func(name string) string {
		return fmt.Sprintf("%s.%s", parentName, name)
	}

	for _, field := range fields {
		name := getFieldName(field.GoName)

		decodeField(g, field, name, getFieldName, lvl)
	}
}

func decodeField(g *GenFile, field *Field, name string, getFieldName func(string) string, lvl int) {
	var sizeFromName string
	if field.FieldSizeFrom != nil {
		sizeFromName = getFieldName(field.FieldSizeFrom.GoName)
	}

	if _, ok := BaseTypesGo[field.Type]; ok {
		decodeBaseType(g, field.Type, fieldGoType(g, field), name, field.Length, sizeFromName, true)
		return
	}

	if field.Array {
		index := fmt.Sprintf("j%d", lvl)
		if field.Length > 0 {
			g.P("for ", index, " := 0; ", index, " < ", field.Length, ";", index, "++ {")
		} else if field.SizeFrom != "" {
			g.P(name, " = make(", getFieldType(g, field), ", int(", sizeFromName, "))")
			g.P("for ", index, " := 0; ", index, " < len(", name, ");", index, "++ {")
		}
		name = fmt.Sprintf("%s[%s]", name, index)
	}

	if enum := field.TypeEnum; enum != nil {
		if _, ok := BaseTypesGo[enum.Type]; ok {
			decodeBaseType(g, enum.Type, fieldGoType(g, field), name, 0, "", false)
		} else {
			logrus.Panicf("\t// ??? ENUM %s %s\n", name, enum.Type)
		}
	} else if alias := field.TypeAlias; alias != nil {
		if typ := alias.TypeStruct; typ != nil {
			decodeFields(g, typ.Fields, name, lvl+1)
		} else {
			if alias.Length > 0 {
				decodeBaseType(g, alias.Type, BaseTypesGo[alias.Type], name, alias.Length, "", false)
			} else {
				decodeBaseType(g, alias.Type, fieldGoType(g, field), name, alias.Length, "", false)
			}
		}
	} else if typ := field.TypeStruct; typ != nil {
		decodeFields(g, typ.Fields, name, lvl+1)
	} else if union := field.TypeUnion; union != nil {
		maxSize := getUnionSize(union)
		g.P("copy(", name, ".", fieldUnionData, "[:], buf.DecodeBytes(", maxSize, "))")
	} else {
		logrus.Panicf("\t// ??? %s (%v)\n", field.GoName, field.Type)
	}

	if field.Array {
		g.P("}")
	}
}
