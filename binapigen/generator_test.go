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
	"testing"
)

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
			size := getBinapiTypeSize(test.input)
			if size != test.expsize {
				t.Errorf("expected %d, got %d", test.expsize, size)
			}
		})
	}
}

/*func TestSizeOfType(t *testing.T) {
	tests := []struct {
		name    string
		input   StructType
		expsize int
	}{
		{
			name: "basic1",
			input: StructType{
				Fields: []Field{
					{Type: "u8"},
				},
			},
			expsize: 1,
		},
		{
			name: "basic2",
			input: Type{
				Fields: []Field{
					{Type: "u8", Length: 4},
				},
			},
			expsize: 4,
		},
		{
			name: "basic3",
			input: Type{
				Fields: []Field{
					{Type: "u8", Length: 16},
				},
			},
			expsize: 16,
		},
		{
			name: "withEnum",
			input: Type{
				Fields: []Field{
					{Type: "u16"},
					{Type: "vl_api_myenum_t"},
				},
			},
			expsize: 6,
		},
		{
			name: "invalid1",
			input: Type{
				Fields: []Field{
					{Type: "x", Length: 16},
				},
			},
			expsize: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module := &File{
				Enums: []Enum{
					{Name: "myenum", Type: "u32"},
				},
			}
			size := getSizeOfType(module, &test.input)
			if size != test.expsize {
				t.Errorf("expected %d, got %d", test.expsize, size)
			}
		})
	}
}*/
