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
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"go.fd.io/govpp/binapigen/vppapi"
)

var sampleJson = `{
    "types": [],
    "messages": [],
    "unions": [],
    "enums": [],
    "enumflags": [],
    "services": {},
    "options": {
        "version": "1.7.0"
    },
    "aliases": {},
    "vl_api_version": "0x12345678",
    "imports": [],
    "counters": [],
    "paths": []
}`

func TestGenerator(t *testing.T) {
	RegisterTestingT(t)

	dir, err := os.MkdirTemp("", "govpp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "vpe.api.json")
	if err := os.WriteFile(file, []byte(sampleJson), 0666); err != nil {
		t.Fatal(err)
	}

	os.Setenv(vppapi.VPPVersionEnvVar, "test-version")
	gen, err := New(Options{
		ApiDir:       dir,
		FileFilter:   []string{file},
		ImportPrefix: "test",
	})
	Expect(err).ToNot(HaveOccurred(), "unexpected generator error: %v", err)

	Expect(gen.Files).To(HaveLen(1))
	Expect(gen.Files[0].PackageName).To(BeEquivalentTo("vpe"))
	Expect(gen.Files[0].GoImportPath).To(BeEquivalentTo("test/vpe"))
	Expect(gen.Files[0].Desc.CRC).To(BeEquivalentTo("0x12345678"))
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"interface", "interfaces"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := sanitizedName(test.name)
			if s != test.expected {
				t.Fatalf("expected: %q, got: %q", test.expected, s)
			}
		})
	}
}
