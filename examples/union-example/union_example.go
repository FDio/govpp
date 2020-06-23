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
	"net"
	"reflect"

	"git.fd.io/govpp.git/codec"
	"git.fd.io/govpp.git/examples/binapi/ip"
	"git.fd.io/govpp.git/examples/binapi/ip_types"
)

func init() {
	log.SetFlags(0)
}

func main() {
	constructExample()

	encodingExample()

	// convert IP from string form into Address type containing union
	convertIP("10.10.1.1")
	convertIP("ff80::1")
}

func constructExample() {
	var union ip_types.AddressUnion

	// create AddressUnion with AdressUnionXXX constructors
	union = ip_types.AddressUnionIP4(ip.IP4Address{192, 168, 1, 10})
	union = ip_types.AddressUnionIP6(ip.IP6Address{0xff, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02})

	// set AddressUnion with SetXXX methods
	union.SetIP4(ip.IP4Address{192, 168, 1, 10})
	union.SetIP6(ip.IP6Address{0xff, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02})
}

func encodingExample() {
	var c = codec.DefaultCodec

	// encode this message
	var msg = ip.IPPuntRedirect{
		Punt: ip.PuntRedirect{
			Nh: ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip.IP4Address{192, 168, 1, 10}),
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
	addr, err := ipToAddress(ip)
	if err != nil {
		log.Printf("error converting IP: %v", err)
		return
	}
	fmt.Printf("converted IP %q to: %+v\n", ip, addr)
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
		copy(ip4addr[:], ip4.To4())
		addr.Un.SetIP4(ip4addr)
	}
	return
}
