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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
)

// Create folder with files
func createFolderAndFiles(t *testing.T, dir string, files ...string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Logf("created dir %q", dir)
	t.Cleanup(func() {
		t.Helper()
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
		t.Logf("removed dir %q", dir)
	})

	for _, file := range files {
		filename := filepath.Join(dir, file)
		if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filename)
		if err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
		t.Logf("- file created %q", file)
	}
}

func absPath(path string) string {
	p, _ := filepath.Abs(path)
	return p
}

func TestFindFiles(t *testing.T) {

	createFolderAndFiles(t, "testdata/find_files", "A/one.api.json", "A/two.api.json", "B/three.api.json")

	type args struct {
		dir string
	}
	tests := []struct {
		name      string
		cwd       string
		args      args
		wantFiles []string
		wantErr   bool
	}{
		{
			name:      "find files 1",
			args:      args{dir: "testdata/find_files"},
			wantFiles: []string{"A/one.api.json", "A/two.api.json", "B/three.api.json"},
			wantErr:   false,
		},
		{
			name:      "find files 2",
			args:      args{dir: "./testdata/find_files"},
			wantFiles: []string{"A/one.api.json", "A/two.api.json", "B/three.api.json"},
			wantErr:   false,
		},
		{
			name:      "find files 3",
			args:      args{dir: absPath("./testdata/find_files")},
			wantFiles: []string{"A/one.api.json", "A/two.api.json", "B/three.api.json"},
			wantErr:   false,
		},
		{
			name:      "find files 4",
			cwd:       "testdata",
			args:      args{dir: "./find_files"},
			wantFiles: []string{"A/one.api.json", "A/two.api.json", "B/three.api.json"},
			wantErr:   false,
		},
		{
			name:      "find files 5",
			cwd:       "testdata/find_files",
			args:      args{dir: "."},
			wantFiles: []string{"A/one.api.json", "A/two.api.json", "B/three.api.json"},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cwd != "" {
				old := os.Getenv("PWD")
				if err := os.Chdir(tt.cwd); err != nil {
					t.Fatal(err)
				}
				t.Logf("curWorkingDir changed to: %v", tt.cwd)
				defer func() {
					if err := os.Chdir(old); err != nil {
						t.Fatal(err)
					}
				}()
			}
			gotFiles, err := FindFiles(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindFiles(%s) error = %v, wantErr %v", tt.args.dir, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("FindFiles(%s) gotFiles = %v, want %v", tt.args.dir, gotFiles, tt.wantFiles)
			}
		})
	}
}

func TestGetInputFiles(t *testing.T) {
	RegisterTestingT(t)

	result, err := FindFiles("testdata")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).To(HaveLen(6))
	for _, file := range result {
		Expect("testdata/" + file).To(BeAnExistingFile())
	}
}

func TestGetInputFilesError(t *testing.T) {
	RegisterTestingT(t)

	result, err := FindFiles("nonexisting_directory")
	Expect(err).Should(HaveOccurred())
	Expect(result).To(BeNil())
}

func TestReadJson(t *testing.T) {
	RegisterTestingT(t)

	inputData, err := os.ReadFile("testdata/af_packet.api.json")
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

	inputData, err := os.ReadFile("testdata/input-read-json-error.json")
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
