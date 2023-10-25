// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// union-example is an example to show how to use unions in VPP binary API.
package main

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/alkiranet/govpp/examples/binapi/ip"
	"github.com/alkiranet/govpp/examples/binapi/ip_types"

	"github.com/lunixbochs/struc"
)

func main() {
	encodingExample()
	usageExample()
}

func encodingExample() {
	// create union with IPv4 address
	var unionIP4 ip.AddressUnion
	unionIP4.SetIP4(ip.IP4Address{192, 168, 1, 10})

	// use it in the Address type
	addr := &ip.Address{
		Af: ip_types.ADDRESS_IP4,
		Un: ip_types.AddressUnionIP4(ip.IP4Address{192, 168, 1, 10}),
	}
	log.Printf("encoding union IPv4: %v", addr.Un.GetIP4())

	// encode the address with union
	data := encode(addr)
	// decode the address with union
	addr2 := decode(data)

	log.Printf("decoded union IPv4: %v", addr2.Un.GetIP4())
}

func encode(addr *ip.Address) []byte {
	log.Printf("encoding address: %#v", addr)
	buf := new(bytes.Buffer)
	if err := struc.Pack(buf, addr); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func decode(data []byte) *ip.Address {
	addr := new(ip.Address)
	buf := bytes.NewReader(data)
	if err := struc.Unpack(buf, addr); err != nil {
		panic(err)
	}
	log.Printf("decoded address: %#v", addr)
	return addr
}

func usageExample() {
	var convAddr = func(ip string) {
		addr, err := ipToAddress(ip)
		if err != nil {
			log.Printf("converting ip %q failed: %v", ip, err)
		}
		fmt.Printf("% 0X\n", addr)
	}

	convAddr("10.10.10.10")
	convAddr("::1")
	convAddr("")
}

func ipToAddress(ipstr string) (addr ip.Address, err error) {
	netIP := net.ParseIP(ipstr)
	if netIP == nil {
		return ip.Address{}, fmt.Errorf("invalid IP: %q", ipstr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = ip_types.ADDRESS_IP6
		var ip6addr ip.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ip_types.ADDRESS_IP4
		var ip4addr ip.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}
