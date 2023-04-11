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
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"go.fd.io/govpp/binapigen/vppapi"
)

func TestGenerator(t *testing.T) {
	tests := []struct {
		name          string
		file          *vppapi.File
		expectPackage string
	}{
		{name: "vpe", file: &vppapi.File{
			Name: "vpe",
			Path: "/usr/share/vpp/api/core/vpe.api.json",
			CRC:  "0x12345678",
		},
			expectPackage: "vpe",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RegisterTestingT(t)

			apiFiles := []vppapi.File{*test.file}

			input := &vppapi.VppInput{Schema: vppapi.Schema{
				Files: apiFiles,
			}}

			gen, err := New(Options{
				ImportPrefix: "test",
			}, input)
			Expect(err).ToNot(HaveOccurred(), "unexpected generator error: %v", err)

			Expect(gen.Files).To(HaveLen(1))
			Expect(gen.Files[0].PackageName).To(BeEquivalentTo(test.expectPackage))
			Expect(gen.Files[0].GoImportPath).To(BeEquivalentTo("test/" + test.expectPackage))
		})
	}
}

func TestGoModule(t *testing.T) {
	const expected = "go.fd.io/govpp/binapi"

	impPath, err := ResolveImportPath("../binapi")
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

	err := GenerateFromFile("vppapi/testdata/union.api.json", Options{OutputDir: testOutputDir})
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
