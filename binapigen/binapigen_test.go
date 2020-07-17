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

	. "github.com/onsi/gomega"

	"git.fd.io/govpp.git/binapigen/vppapi"
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

			apifiles := []*vppapi.File{test.file}

			gen, err := New(Options{
				ImportPrefix: "test",
			}, apifiles, nil)
			Expect(err).ToNot(HaveOccurred(), "unexpected generator error: %v", err)

			Expect(gen.Files).To(HaveLen(1))
			Expect(gen.Files[0].PackageName).To(BeEquivalentTo(test.expectPackage))
			Expect(gen.Files[0].GoImportPath).To(BeEquivalentTo("test/" + test.expectPackage))
		})
	}
}
