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
	"bufio"
	"fmt"
	"git.fd.io/govpp.git/binapigen/vppapi"
	. "github.com/onsi/gomega"
	"os"
	"strings"
	"testing"
)

func TestGoModule(t *testing.T) {
	const expected = "git.fd.io/govpp.git/binapi"

	impPath, err := vppapi.ResolveImportPath("../binapi")
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

	// order of the union sizes in file generated from union.api.json
	var sizes = []int{16, 4, 32, 16, 64, 111}

	// remove directory created during test
	defer func() {
		err := os.RemoveAll(testOutputDir)
		Expect(err).ToNot(HaveOccurred())
	}()

	err := GenerateFromFile("vppapi/testdata/union.api.json")
	Expect(err).ShouldNot(HaveOccurred())

	file, err := os.Open(testOutputDir + "/union/union.ba.go")
	Expect(err).ShouldNot(HaveOccurred())
	defer func() {
		err := file.Close()
		Expect(err).ToNot(HaveOccurred())
	}()

	// the generated line with union size is in format XXX_UnionData [<size>]byte
	// the prefix identifies these lines (the starting tab is important)
	prefix := fmt.Sprintf("\t%s", "XXX_UnionData [")

	index := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), prefix) {
			Expect(scanner.Text()).To(Equal(prefix + fmt.Sprintf("%d]byte", sizes[index])))
			index++
		}
	}
	// ensure all union sizes were found and tested
	Expect(index).To(Equal(len(sizes)))
}
