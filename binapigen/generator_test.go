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
	"git.fd.io/govpp.git/binapigen/vppapi"
	. "github.com/onsi/gomega"
	"testing"
)

func TestGoModule(t *testing.T) {
	const expected = "git.fd.io/govpp.git/binapi"

	impPath, err := resolveImportPath("../binapi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if impPath != expected {
		t.Fatalf("expected: %q, got: %q", expected, impPath)
	}
}

func TestBinapiTypeSizes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		expsize int
	}{
		{name: "basic1", input: "u8", expsize: 1},
		{name: "basic2", input: "i8", expsize: 1},
		{name: "basic3", input: "u16", expsize: 2},
		{name: "basic4", input: "i32", expsize: 4},
		{name: "string", input: "string", expsize: 1},
		{name: "invalid1", input: "x", expsize: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			size := getSizeOfBinapiBaseType(test.input, 1)
			if size != test.expsize {
				t.Errorf("expected %d, got %d", test.expsize, size)
			}
		})
	}
}

func TestBinapiUnionSizes(t *testing.T) {
	RegisterTestingT(t)
	tests := []struct {
		testName string
		input    *Union
		expsize  int
	}{
		{testName: "union_alias", input: typeTestData{
			typ: "union", fields: []*typeTestData{{typ: "alias", value: U16},
			}}.getUnion("union1"), expsize: 2},
		{testName: "union_enum", input: typeTestData{
			typ: "union", fields: []*typeTestData{{typ: "enum", value: U32},
			}}.getUnion("union2"), expsize: 4},
		{testName: "union_struct", input: typeTestData{
			typ: "union", fields: []*typeTestData{
				{typ: "struct", fields: []*typeTestData{{value: U8}, {value: U16}, {value: U32}}},
			}}.getUnion("union3"), expsize: 7},
		{testName: "union_structs", input: typeTestData{
			typ: "union", fields: []*typeTestData{
				{typ: "struct", fields: []*typeTestData{{value: U8}, {value: BOOL}}},
				{typ: "struct", fields: []*typeTestData{{value: U16}, {value: U32}}},
				{typ: "struct", fields: []*typeTestData{{value: U32}, {value: U64}}},
			}}.getUnion("union4"), expsize: 12},
		{testName: "union_unions", input: typeTestData{
			typ: "union", fields: []*typeTestData{
				{typ: "union", fields: []*typeTestData{
					{typ: "struct", fields: []*typeTestData{{value: STRING}}},
				}},
				{typ: "union", fields: []*typeTestData{
					{typ: "struct", fields: []*typeTestData{{value: U32}}},
				}},
				{typ: "union", fields: []*typeTestData{
					{typ: "struct", fields: []*typeTestData{{value: U64}}},
				}},
			}}.getUnion("union5"), expsize: 8},
		{testName: "union_combined", input: typeTestData{
			typ: "union", fields: []*typeTestData{
				{typ: "alias", value: U8},
				{typ: "enum", value: U16},
				{typ: "struct", fields: []*typeTestData{{value: U8}, {value: U16}, {value: U32}}}, // <-
				{typ: "union", fields: []*typeTestData{
					{typ: "alias", value: U16},
					{typ: "enum", value: U16},
					{typ: "struct", fields: []*typeTestData{{value: U32}}},
				}},
			}}.getUnion("union6"), expsize: 7},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			size := getUnionSize(test.input)
			Expect(size).To(Equal(test.expsize))
		})
	}
}

// Typed data used for union size evaluation testing.
type typeTestData struct {
	typ    string
	value  string
	fields []*typeTestData
}

func (t typeTestData) getUnion(name string) *Union {
	return &Union{
		UnionType: vppapi.UnionType{Name: name},
		Fields:    t.getUnionFields(name),
	}
}

func (t typeTestData) getUnionFields(parentName string) (fields []*Field) {
	for i, field := range t.fields {
		var (
			dataType   string
			aliasType  *Alias
			enumType   *Enum
			structType *Struct
			unionType  *Union
		)
		switch field.typ {
		case "alias":
			aliasType = &Alias{AliasType: vppapi.AliasType{Name: fmt.Sprintf("%s_alias_%d", parentName, i), Type: field.value}}
		case "enum":
			enumType = &Enum{EnumType: vppapi.EnumType{Name: fmt.Sprintf("%s_enum_%d", parentName, i), Type: field.value}}
		case "struct":
			structType = &Struct{Fields: field.getUnionFields(fmt.Sprintf("%s_struct_%d", parentName, i))}
		case "union":
			unionType = field.getUnion(parentName)
		default:
			dataType = field.value
		}
		fields = append(fields, &Field{
			Field:      vppapi.Field{Name: fmt.Sprintf("%s_field_%d", parentName, i), Type: dataType},
			TypeAlias:  aliasType,
			TypeEnum:   enumType,
			TypeStruct: structType,
			TypeUnion:  unionType,
		})
	}
	return fields
}
