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
	"log"

	"git.fd.io/govpp.git/examples/bin_api/ip"
	"github.com/lunixbochs/struc"
)

func main() {
	// create union with IPv4 address
	var unionIP4 ip.AddressUnion
	unionIP4.SetIP4(ip.IP4Address{Address: []byte{192, 168, 1, 10}})

	// use it in the Address type
	addr := &ip.Address{
		Af: ip.ADDRESS_IP4,
		Un: unionIP4,
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
