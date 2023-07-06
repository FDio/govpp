//  Copyright (c) 2022 Cisco and/or its affiliates.
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

package integration

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"testing"

	"go.fd.io/govpp/test/vpptesting"
)

func TestExamples(t *testing.T) {
	skipTestIfGoNotInstalled(t)

	if err := filepath.WalkDir("./examples", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() || filepath.Base(d.Name()) == "examples" {
			return nil
		}
		example := filepath.Base(d.Name())
		t.Run(example, func(tt *testing.T) {
			testExample(tt, example)
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func testExample(t *testing.T, example string) {
	vpptesting.SetupVPP(t)

	cmd := exec.Command("go", "run", "./examples/"+example)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s command failed: %+v\n%s", cmd, err, out)
	}
	t.Logf("out: %s", out)
}
