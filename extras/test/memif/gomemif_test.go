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

package extras

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.fd.io/govpp/test/vpptesting"
)

func TestGoMemifPoll(t *testing.T) {
	test := "/icmp_responder_poll/icmp_responder_poll"
	runGoMemif(t, test)
}

func TestGoMemifCb(t *testing.T) {
	test := "/icmp_responder_cb/icmp_responder_cb"
	runGoMemif(t, test)
}

func runGoMemif(t *testing.T, test string) {
	_, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("`go` command is not available, skipping test")
	}

	// Start VPP
	tc := vpptesting.SetupVPP(t)

	// create memif interface, assign ip address and set it up
	tc.RunCli("create interface memif id 0 master")
	tc.RunCli("set int ip addr memif0/0 192.168.1.2/24")
	tc.RunCli("set int state memif0/0 up")

	cmd := exec.Command("./gomemif/examples" + test)
	t.Logf("executing command '%v'", cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Error creating stdin pipe: %v", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Error creating stdout pipe: %v", err)
		return
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Error starting command: %v", err)
		return
	}

	// Create a scanner to read the output from the command
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			// Print each line of output
			fmt.Println(scanner.Text())
		}
	}()

	// Send the "start" command to the gomemif application
	_, err = stdin.Write([]byte("start\n"))
	if err != nil {
		t.Fatalf("Error writing to stdin: %v", err)
		return
	}

	time.Sleep(1 * time.Second)
	// Send the "show" command to the gomemif application
	_, err = stdin.Write([]byte("show\n"))
	if err != nil {
		t.Fatalf("Error writing to stdin: %v", err)
		return
	}

	vppout, err := tc.RunCli("ping 192.168.1.1")
	if err != nil {
		t.Fatalf("Error running ping command: %v", err)
		return
	} else {
		ouput_split := strings.Split(vppout, "received")[0]
		output_field := strings.Fields(ouput_split)
		received, _ := strconv.Atoi(output_field[len(output_field)-1])
		if received < 1 {
			t.Fatalf("No packets received")
			return
		}
	}

	_, err = stdin.Write([]byte("exit\n"))
	if err != nil {
		t.Fatalf("Error writing to stdin: %v", err)
		return
	}

	err = stdin.Close()
	if err != nil {
		t.Fatalf("Error closing stdin: %v", err)
		return
	}

	t.Logf("test %s output: %s", test, stdout)
}
