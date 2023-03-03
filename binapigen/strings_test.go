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

func TestInitialism(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expOutput string
	}{
		{name: "id", input: "id", expOutput: "ID"},
		{name: "ipv6", input: "is_ipv6", expOutput: "IsIPv6"},
		{name: "ip6", input: "is_ip6", expOutput: "IsIP6"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := camelCaseName(test.input)
			if output != test.expOutput {
				t.Errorf("expected %q, got %q", test.expOutput, output)
			}
		})
	}
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
