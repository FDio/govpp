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

// Package integration contains tests against running VPP instance.
// The test cases are only executed if env contains TEST=integration.
package integration

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"

	_ "net/http/pprof"
)

var (
	// IntegrationTestsActive is set to true if integration tests should run.
	IntegrationTestsActive = os.Getenv("TEST") == "integration"
)

func TestMain(m *testing.M) {
	if IntegrationTestsActive {
		go func() {
			fmt.Fprintln(os.Stderr, http.ListenAndServe(":6060", nil))
		}()
		os.Exit(m.Run())
	}
	fmt.Fprintf(os.Stderr, "integration tests are NOT enabled (set TEST='integration' to enable)\n")
	os.Exit(0)
}

func skipTestIfGoNotInstalled(t *testing.T) {
	_, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("`go` command is not available, skipping test")
	}
}
