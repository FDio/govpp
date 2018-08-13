package main

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
