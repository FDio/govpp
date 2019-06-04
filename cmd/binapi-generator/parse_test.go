package main

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
		{name: "invalid1", input: "x", expsize: -1},
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

func TestSizeOfType(t *testing.T) {
	tests := []struct {
		name    string
		input   Type
		expsize int
	}{
		{
			name: "basic1",
			input: Type{
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
			ctx := &context{
				packageData: &Package{
					Enums: []Enum{
						{Name: "myenum", Type: "u32"},
					},
				},
			}
			size := getSizeOfType(ctx, &test.input)
			if size != test.expsize {
				t.Errorf("expected %d, got %d", test.expsize, size)
			}
		})
	}
}
