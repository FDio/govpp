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

func TestGoModule(t *testing.T) {
	const expected = "git.fd.io/govpp.git/binapi"

	impPath := resolveImportPath("../binapi")
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
			size := getSizeOfBinapiTypeLength(test.input, 1)
			if size != test.expsize {
				t.Errorf("expected %d, got %d", test.expsize, size)
			}
		})
	}
}
