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

package vppapi

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetInputFiles(t *testing.T) {
	RegisterTestingT(t)

	result, err := FindFiles("testdata", 1)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).To(HaveLen(5))
	for _, file := range result {
		Expect(file).To(BeAnExistingFile())
	}
}

func TestGetInputFilesError(t *testing.T) {
	RegisterTestingT(t)

	result, err := FindFiles("nonexisting_directory", 1)
	Expect(err).Should(HaveOccurred())
	Expect(result).To(BeNil())
}

func TestReadJson(t *testing.T) {
	RegisterTestingT(t)

	inputData, err := ioutil.ReadFile("testdata/af_packet.api.json")
	Expect(err).ShouldNot(HaveOccurred())
	result, err := ParseRaw(inputData)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).ToNot(BeNil())
	Expect(result.EnumTypes).To(HaveLen(0))
	Expect(result.StructTypes).To(HaveLen(0))
	Expect(result.Messages).To(HaveLen(6))
	Expect(result.Service.RPCs).To(HaveLen(3))
}

func TestReadJsonError(t *testing.T) {
	RegisterTestingT(t)

	inputData, err := ioutil.ReadFile("testdata/input-read-json-error.json")
	Expect(err).ShouldNot(HaveOccurred())
	result, err := ParseRaw(inputData)
	Expect(err).Should(HaveOccurred())
	Expect(result).To(BeNil())
}

func TestParseFile(t *testing.T) {
	module, err := ParseFile("testdata/vpe.api.json")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	b, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("parsed module: %s", b)

	if module.Name != "vpe" {
		t.Errorf("expected Name=%s, got %v", "vpe", module.Name)
	}
	if module.Path != "testdata/vpe.api.json" {
		t.Errorf("expected Path=%s, got %v", "testdata/vpe.api.json", module.Path)
	}
	if module.CRC != "0xbd2c94f4" {
		t.Errorf("expected CRC=%s, got %v", "0xbd2c94f4", module.CRC)
	}

	if version := module.Options["version"]; version != "1.6.1" {
		t.Errorf("expected option[version]=%s, got %v", "1.6.1", version)
	}
	if len(module.Imports) == 0 {
		t.Errorf("expected imports, got none")
	}
	if len(module.EnumTypes) == 0 {
		t.Errorf("expected enums, got none")
	}
	if len(module.AliasTypes) == 0 {
		t.Errorf("expected aliases, got none")
	}
	if len(module.StructTypes) == 0 {
		t.Errorf("expected types, got none")
	}
	if len(module.Messages) == 0 {
		t.Errorf("expected messages, got none")
	}
	if len(module.Service.RPCs) == 0 {
		t.Errorf("expected service RPCs, got none")
	}
}

func TestParseFileUnsupported(t *testing.T) {
	_, err := ParseFile("testdata/input.txt")
	if err == nil {
		t.Fatal("expected error")
	}
}
