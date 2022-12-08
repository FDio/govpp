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
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"go.fd.io/govpp/binapi/ethernet_types"
	"go.fd.io/govpp/binapi/ip_types"
	"go.fd.io/govpp/binapi/vpe_types"
)

func TestGeneratedParseAddress(t *testing.T) {
	RegisterTestingT(t)

	var data = []struct {
		input  string
		result ip_types.Address
	}{
		{"192.168.0.1", ip_types.Address{
			Af: ip_types.ADDRESS_IP4,
			Un: ip_types.AddressUnionIP4(ip_types.IP4Address{192, 168, 0, 1}),
		}},
		{"aac1:0:ab45::", ip_types.Address{
			Af: ip_types.ADDRESS_IP6,
			Un: ip_types.AddressUnionIP6(ip_types.IP6Address{170, 193, 0, 0, 171, 69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}),
		}},
	}

	for _, entry := range data {
		t.Run(entry.input, func(t *testing.T) {
			parsedAddress, err := ip_types.ParseAddress(entry.input)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(parsedAddress).To(Equal(entry.result))

			originAddress := parsedAddress.String()
			Expect(originAddress).To(Equal(entry.input))
		})
	}
}

func TestGeneratedParseAddressError(t *testing.T) {
	RegisterTestingT(t)

	_, err := ip_types.ParseAddress("malformed_ip")
	Expect(err).Should(HaveOccurred())
}

func TestGeneratedParsePrefix(t *testing.T) {
	RegisterTestingT(t)

	var data = []struct {
		input  string
		result ip_types.Prefix
	}{
		{"192.168.0.1/24", ip_types.Prefix{
			Address: ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{192, 168, 0, 1}),
			},
			Len: 24,
		}},
		{"192.168.0.1", ip_types.Prefix{
			Address: ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{192, 168, 0, 1}),
			},
			Len: 32,
		}},
		{"aac1:0:ab45::/96", ip_types.Prefix{
			Address: ip_types.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ip_types.AddressUnionIP6(ip_types.IP6Address{170, 193, 0, 0, 171, 69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}),
			},
			Len: 96,
		}},
		{"aac1:0:ab45::", ip_types.Prefix{
			Address: ip_types.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ip_types.AddressUnionIP6(ip_types.IP6Address{170, 193, 0, 0, 171, 69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}),
			},
			Len: 128,
		}},
	}

	for _, entry := range data {
		t.Run(entry.input, func(t *testing.T) {
			parsedAddress, err := ip_types.ParsePrefix(entry.input)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(parsedAddress).To(Equal(entry.result))

			// Parsed IP without prefix receives a default one
			// so the input data must be adjusted
			if entry.result.Address.Af == ip_types.ADDRESS_IP4 && !strings.Contains(entry.input, "/") {
				entry.input = entry.input + "/32"
			}
			if entry.result.Address.Af == ip_types.ADDRESS_IP6 && !strings.Contains(entry.input, "/") {
				entry.input = entry.input + "/128"
			}
			originAddress := parsedAddress.String()
			Expect(originAddress).To(Equal(entry.input))
		})
	}
}

func TestGeneratedParsePrefixError(t *testing.T) {
	RegisterTestingT(t)

	_, err := ip_types.ParsePrefix("malformed_ip")
	Expect(err).Should(HaveOccurred())
}

func TestGeneratedParseMAC(t *testing.T) {
	RegisterTestingT(t)

	var data = []struct {
		input  string
		result ethernet_types.MacAddress
	}{
		{"b7:b9:bb:a1:5c:af", ethernet_types.MacAddress{183, 185, 187, 161, 92, 175}},
		{"47:4b:c7:3e:06:c8", ethernet_types.MacAddress{71, 75, 199, 62, 6, 200}},
		{"a7:cc:9f:10:18:e3", ethernet_types.MacAddress{167, 204, 159, 16, 24, 227}},
	}

	for _, entry := range data {
		t.Run(entry.input, func(t *testing.T) {
			parsedMac, err := ethernet_types.ParseMacAddress(entry.input)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(parsedMac).To(Equal(entry.result))

			originAddress := parsedMac.String()
			Expect(originAddress).To(Equal(entry.input))
		})
	}
}

func TestGeneratedParseMACError(t *testing.T) {
	RegisterTestingT(t)

	_, err := ethernet_types.ParseMacAddress("malformed_mac")
	Expect(err).Should(HaveOccurred())
}

func TestGeneratedParseTimestamp(t *testing.T) {
	RegisterTestingT(t)

	var data = []struct {
		input  time.Time
		result vpe_types.Timestamp
	}{
		{time.Unix(0, 0), vpe_types.Timestamp(0)},
		{time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			vpe_types.Timestamp(9.466848e+08)},
	}

	for _, entry := range data {
		t.Run(entry.input.String(), func(t *testing.T) {
			ts := vpe_types.NewTimestamp(entry.input)
			Expect(ts).To(Equal(entry.result))

			Expect(entry.input.Equal(ts.ToTime())).To(BeTrue())

			originTime := ts.String()
			Expect(originTime).To(Equal(entry.input.Local().String()))
		})
	}
}
