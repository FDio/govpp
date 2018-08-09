package main

import (
	"bytes"
	"log"

	"git.fd.io/govpp.git/examples/bin_api/ip"
	"github.com/lunixbochs/struc"
)

func main() {
	addr := &ip.Address{
		Af: ip.ADDRESS_IP4,
	}
	addr.Un.SetIP4(ip.IP4Address{[]byte{192, 168, 1, 10}})

	log.Printf("addr: %#v", addr)
	buf := new(bytes.Buffer)
	if err := struc.Pack(buf, addr); err != nil {
		panic(err)
	}
	data := buf.Bytes()

	addr2 := new(ip.Address)
	buf2 := bytes.NewReader(data)
	if err := struc.Unpack(buf2, addr2); err != nil {
		panic(err)
	}
	log.Printf("addr2: %#v", addr2)
	log.Printf("addr2: %v", addr2.Un.GetIP4().Address)
}
