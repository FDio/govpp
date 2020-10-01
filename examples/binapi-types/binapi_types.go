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
	"fmt"
	"log"

	"git.fd.io/govpp.git/binapi/ethernet_types"
	"git.fd.io/govpp.git/binapi/ip"
	"git.fd.io/govpp.git/binapi/ip_types"
	"git.fd.io/govpp.git/codec"
)

func init() {
	log.SetFlags(0)
}

func main() {
	addressUnionExample()
	ipAddressExample()

	// convert IP from string form into Address type containing union
	convertIP("10.10.1.1")
	convertIP("ff80::1")

	// convert IP from string form into Prefix type
	convertIPPrefix("20.10.1.1/24")
	convertIPPrefix("21.10.1.1")
	convertIPPrefix("ff90::1/64")
	convertIPPrefix("ff91::1")

	// convert MAC address from string into MacAddress
	convertToMacAddress("00:10:ab:4f:00:01")
}

func addressUnionExample() {
	var union ip_types.AddressUnion

	// initialize union using constructors
	union = ip_types.AddressUnionIP4(ip_types.IP4Address{192, 168, 1, 10})
	union = ip_types.AddressUnionIP6(ip_types.IP6Address{0xff, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02})

	// get union value using getters
	ip4 := union.GetIP4()
	ip6 := union.GetIP6()

	// set union value using setters
	union.SetIP4(ip4)
	union.SetIP6(ip6)
}

func ipAddressExample() {
	// parse string into IP address
	addrIP4, err := ip_types.ParseAddress("192.168.1.10")
	if err != nil {
		panic(err)
	}
	/*addrIP6, err := ip_types.ParseAddress("ff:2::2")
	if err != nil {
		panic(err)
	}*/

	var msg = ip.IPPuntRedirect{
		IsAdd: true,
		Punt: ip.PuntRedirect{
			Nh: addrIP4,
		},
	}

	log.Printf("encoding message: %#v", msg)

	var c = codec.DefaultCodec

	b, err := c.EncodeMsg(&msg, 1)
	if err != nil {
		log.Fatal(err)
	}

	// decode into this message
	var msg2 ip.IPPuntRedirect
	if err := c.DecodeMsg(b, &msg2); err != nil {
		log.Fatal(err)
	}
	log.Printf("decoded message: %#v", msg2)

	// compare the messages
	if !msg.Punt.Nh.ToIP().Equal(msg2.Punt.Nh.ToIP()) {
		log.Fatal("messages are not equal")
	}
}

func convertIP(ip string) {
	addr, err := ip_types.ParseAddress(ip)
	if err != nil {
		log.Printf("error converting IP to Address: %v", err)
		return
	}
	fmt.Printf("converted IP %q to: %#v\n", ip, addr)

	fmt.Printf("Address converted back to string IP %#v to: %s\n", addr, addr)
}

func convertIPPrefix(ip string) {
	prefix, err := ip_types.ParsePrefix(ip)
	if err != nil {
		log.Printf("error converting prefix to IP4Prefix: %v", err)
		return
	}
	fmt.Printf("converted prefix %q to: %#v\n", ip, prefix)

	fmt.Printf("IP4Prefix converted back to string prefix %#v to: %s\n", prefix, prefix)
}

func convertToMacAddress(mac string) {
	parsedMac, err := ethernet_types.ParseMacAddress(mac)
	if err != nil {
		log.Printf("error converting MAC to MacAddress: %v", err)
		return
	}
	fmt.Printf("converted mac %q to: %#v\n", mac, parsedMac)

	fmt.Printf("MacAddress converted back to string %#v to: %s\n", parsedMac, parsedMac)
}
