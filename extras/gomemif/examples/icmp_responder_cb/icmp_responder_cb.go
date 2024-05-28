/*
 *------------------------------------------------------------------
 * Copyright (c) 2020 Cisco and/or its affiliates.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *------------------------------------------------------------------
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"go.fd.io/govpp/extras/gomemif/memif"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/profile"
)

func Disconnected(i *memif.Interface) error {
	fmt.Println("Disconnected: ", i.GetName())

	data, ok := i.GetPrivateData().(*interfaceData)
	if !ok {
		return fmt.Errorf("Invalid private data")
	}
	close(data.quitChan) // stop polling
	close(data.errChan)
	data.wg.Wait() // wait until polling stops, then continue disconnect

	return nil
}

func Responder(itf *memif.Interface, rx_qid int) error {
	data, ok := itf.GetPrivateData().(*interfaceData)
	if !ok {
		return fmt.Errorf("Invalid private data")
	}
	data.errChan = make(chan error, 1)
	data.quitChan = make(chan struct{}, 1)
	data.wg.Add(1)

	// allocate packet buffers
	pkt := itf.Pkt
	var tx_bufs []memif.MemifPacketBuffer
	for i := range pkt {
		pkt[i].Buf = make([]byte, 2048)
		pkt[i].Buflen = 2048
	}

	// get rx queue
	rxq, err := itf.GetRxQueue(rx_qid)
	if err != nil {
		return err
	}
	// As this is an example, we will use the same queue id for transmit.
	// i.e. if rx_queue id is 1, we will use tx_queue id 1.
	// get tx queue
	txq, err := itf.GetTxQueue(rx_qid)
	if err != nil {
		return err
	}
	_ = txq

	nPackets, err := rxq.Rx_burst(pkt)
	if err != nil {
		return err
	}

	fmt.Println(nPackets)

	rxq.Refill(int(nPackets))
	_ = err

	for i := 0; i < int(nPackets); i++ {
		gopkt := gopacket.NewPacket(pkt[i].Buf[:pkt[i].Buflen], layers.LayerTypeEthernet, gopacket.NoCopy)
		etherLayer := gopkt.Layer(layers.LayerTypeEthernet)

		// received frame src mac address will become trasmit frame dst mac address.
		tx_dstMAC := etherLayer.(*layers.Ethernet).SrcMAC
		if etherLayer.(*layers.Ethernet).EthernetType == layers.EthernetTypeARP {
			rEth := layers.Ethernet{
				SrcMAC: net.HardwareAddr{0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa},
				DstMAC: tx_dstMAC,

				EthernetType: layers.EthernetTypeARP,
			}
			rArp := layers.ARP{
				AddrType:          layers.LinkTypeEthernet,
				Protocol:          layers.EthernetTypeIPv4,
				HwAddressSize:     6,
				ProtAddressSize:   4,
				Operation:         layers.ARPReply,
				SourceHwAddress:   []byte(net.HardwareAddr{0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa}),
				SourceProtAddress: []byte("\xc0\xa8\x01\x01"),
				DstHwAddress:      []byte(tx_dstMAC),
				DstProtAddress:    []byte("\xc0\xa8\x01\x02"),
			}
			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}
			gopacket.SerializeLayers(buf, opts, &rEth, &rArp)
			// write packet to shared memory
			txq.WritePacket(buf.Bytes())
		}
		if etherLayer.(*layers.Ethernet).EthernetType == layers.EthernetTypeIPv4 {
			ipLayer := gopkt.Layer(layers.LayerTypeIPv4)
			if ipLayer == nil {
				fmt.Println("Missing IPv4 layer.")

			}
			ipv4, _ := ipLayer.(*layers.IPv4)
			if ipv4.Protocol != layers.IPProtocolICMPv4 {
				fmt.Println("Not ICMPv4 protocol.")
			}
			icmpLayer := gopkt.Layer(layers.LayerTypeICMPv4)
			if icmpLayer == nil {
				fmt.Println("Missing ICMPv4 layer.")
			}
			icmp, _ := icmpLayer.(*layers.ICMPv4)
			if icmp.TypeCode.Type() != layers.ICMPv4TypeEchoRequest {
				fmt.Println("Not ICMPv4 echo request.")
			}

			// Build packet layers.
			ethResp := layers.Ethernet{
				SrcMAC: net.HardwareAddr{0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa},
				DstMAC: []byte(tx_dstMAC),

				EthernetType: layers.EthernetTypeIPv4,
			}
			ipv4Resp := layers.IPv4{
				Version:    4,
				IHL:        5,
				TOS:        0,
				Id:         0,
				Flags:      0,
				FragOffset: 0,
				TTL:        255,
				Protocol:   layers.IPProtocolICMPv4,
				SrcIP:      []byte("\xc0\xa8\x01\x01"),
				DstIP:      []byte("\xc0\xa8\x01\x02"),
			}
			icmpResp := layers.ICMPv4{
				TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoReply, 0),
				Id:       icmp.Id,
				Seq:      icmp.Seq,
			}

			// Set up buffer and options for serialization.
			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}
			gopacket.SerializeLayers(buf, opts, &ethResp, &ipv4Resp, &icmpResp,
				gopacket.Payload(icmp.Payload))
			// write packet to shared memory
			tx_bufs = append(tx_bufs, memif.MemifPacketBuffer{Buf: buf.Bytes(), Buflen: len(buf.Bytes())})

		}
	}
	txq.Tx_burst(tx_bufs)

	return nil
}

func Connected(i *memif.Interface) error {
	data, ok := i.GetPrivateData().(*interfaceData)
	_ = data
	if !ok {
		return fmt.Errorf("Invalid private data")
	}
	// allocate packet buffer
	i.Pkt = make([]memif.MemifPacketBuffer, 64)

	// get rx queue
	for j := 0; j < int(i.GetMemoryConfig().NumQueuePairs); j++ {
		rxq, _ := i.GetRxQueue(j)
		rxq.Refill(0)
	}

	return nil
}

type interfaceData struct {
	errChan  chan error
	quitChan chan struct{}
	wg       sync.WaitGroup
}

func interractiveHelp() {
	fmt.Println("help - print this help")
	fmt.Println("start - start connecting loop")
	fmt.Println("show - print interface details")
	fmt.Println("exit - exit the application")
}

func main() {
	cpuprof := flag.String("cpuprof", "", "cpu profiling output file")
	memprof := flag.String("memprof", "", "mem profiling output file")
	role := flag.String("role", "slave", "interface role")
	name := flag.String("name", "gomemif", "interface name")
	socketName := flag.String("socket", "/run/vpp/memif.sock", "control socket filename")

	flag.Parse()

	if *cpuprof != "" {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(*cpuprof)).Stop()
	}
	if *memprof != "" {
		defer profile.Start(profile.MemProfile, profile.ProfilePath(*memprof)).Stop()
	}

	memifErrChan := make(chan error)
	exitChan := make(chan struct{})

	var isMaster bool
	switch *role {
	case "slave":
		isMaster = false
	case "master":
		isMaster = true
	default:
		fmt.Println("Invalid role")
		return
	}

	fmt.Println("GoMemif: Responder")
	fmt.Println("-----------------------")

	socket, err := memif.NewSocket("gomemif_example", *socketName)
	if err != nil {
		fmt.Println("Failed to create socket: ", err)
		return
	}

	data := interfaceData{}
	MemoryConfig := memif.MemoryConfig{NumQueuePairs: 2, Log2RingSize: 11}
	args := &memif.Arguments{
		IsMaster:         isMaster,
		ConnectedFunc:    Connected,
		DisconnectedFunc: Disconnected,
		PrivateData:      &data,
		Name:             *name,
		InterruptFunc:    Responder,
		MemoryConfig:     MemoryConfig,
	}

	i, err := socket.NewInterface(args)
	if err != nil {
		fmt.Println("Failed to create interface on socket %s: %s", socket.GetFilename(), err)
		goto exit
	}

	// slave attempts to connect to control socket
	// to handle control communication call socket.StartPolling()
	if !i.IsMaster() {
		fmt.Println(args.Name, ": Connecting to control socket...")
		for !i.IsConnecting() {
			err = i.RequestConnection()
			if err != nil {
				/* TODO: check for ECONNREFUSED errno
				 * if error is ECONNREFUSED it may simply mean that master
				 * interface is not up yet, use i.RequestConnection()
				 */
				fmt.Println("Failed to connect: ", err)
				goto exit
			}
		}
	}

	go func(exitChan chan<- struct{}) {
		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("gomemif# ")
			text, _ := reader.ReadString('\n')
			// convert CRLF to LF
			text = strings.Replace(text, "\n", "", -1)
			switch text {
			case "help":
				interractiveHelp()
			case "start":
				// start polling for events on this socket
				socket.StartPolling(memifErrChan)
			case "show":
				fmt.Print(i.String())
			case "exit":
				err = socket.StopPolling()
				if err != nil {
					fmt.Println("Failed to stop polling: ", err)
				}
				close(exitChan)
				return
			default:
				fmt.Println("Unknown input")
			}
		}
	}(exitChan)

	for {
		select {
		case <-exitChan:
			goto exit
		case err, ok := <-memifErrChan:
			if ok {
				fmt.Println(err)
			}
		case err, ok := <-data.errChan:
			if ok {
				fmt.Println(err)
			}
		default:
			continue
		}
	}

exit:
	socket.Delete()
	close(memifErrChan)
}
