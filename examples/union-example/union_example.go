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
	"git.fd.io/govpp.git/codec"
	"git.fd.io/govpp.git/examples/binapi/interfaces"
	"git.fd.io/govpp.git/examples/binapi/ip"
	"git.fd.io/govpp.git/examples/binapi/ip_types"
	"log"
	"reflect"
)

func init() {
	log.SetFlags(0)
}

func main() {
	constructExample()

	encodingExampleIP()

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

func constructExample() {
	var union ip_types.AddressUnion

	// create AddressUnion with AdressUnionXXX constructors
	union = ip_types.AddressUnionIP4(ip_types.IP4Address{192, 168, 1, 10})
	union = ip_types.AddressUnionIP6(ip_types.IP6Address{0xff, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02})

	// set AddressUnion with SetXXX methods
	union.SetIP4(ip_types.IP4Address{192, 168, 1, 10})
	union.SetIP6(ip_types.IP6Address{0xff, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02})
}

func encodingExampleIP() {
	var c = codec.DefaultCodec

	// encode this message
	var msg = ip.IPPuntRedirect{
		Punt: ip.PuntRedirect{
			Nh: ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{192, 168, 1, 10}),
			},
		},
		IsAdd: true,
	}
	log.Printf("encoding message: %+v", msg)

	b, err := c.EncodeMsg(&msg, 1)
	if err != nil {
		log.Fatal(err)
	}

	// decode into this message
	var msg2 ip.IPPuntRedirect
	if err := c.DecodeMsg(b, &msg2); err != nil {
		log.Fatal(err)
	}
	log.Printf("decoded message: %+v", msg2)

	// compare the messages
	if !reflect.DeepEqual(msg, msg2) {
		log.Fatal("messages are not equal")
	}
}

func convertIP(ip string) {
	addr, err := ip_types.ParseAddress(ip)
	if err != nil {
		log.Printf("error converting IP to Address: %v", err)
		return
	}
	fmt.Printf("converted IP %q to: %+v\n", ip, addr)

	ipStr := addr.ToString()
	fmt.Printf("Address converted back to string IP %+v to: %q\n", addr, ipStr)
}

func convertIPPrefix(ip string) {
	prefix, err := ip_types.ParsePrefix(ip)
	if err != nil {
		log.Printf("error converting prefix to IP4Prefix: %v", err)
		return
	}
	fmt.Printf("converted prefix %q to: %+v\n", ip, prefix)

	ipStr := prefix.ToString()
	fmt.Printf("IP4Prefix converted back to string prefix %+v to: %q\n", prefix, ipStr)
}

func convertToMacAddress(mac string) {
	parsedMac, err := interfaces.ParseMAC(mac)
	if err != nil {
		log.Printf("error converting MAC to MacAddress: %v", err)
		return
	}
	fmt.Printf("converted prefix %q to: %+v\n", mac, parsedMac)

	macStr := parsedMac.ToString()
	fmt.Printf("MacAddress converted back to string %+v to: %q\n", parsedMac, macStr)
}