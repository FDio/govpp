// gopacket is a simple example showing how to answer APR and ICMP echo
// requests through a memif interface. This example is mostly identical
// to icmp-responder example, but it is using MemifPacketHandle API to
// read and write packets using gopacket API.
//
// The appropriate VPP configuration for the opposite memif is:
//
//	vpp$ create memif socket id 1 filename /tmp/gopacket-example
//	vpp$ create interface memif id 1 socket-id 1 slave secret secret no-zero-copy
//	vpp$ set int state memif1/1 up
//	vpp$ set int ip address memif1/1 192.168.1.2/24
//
// To start the example, simply type:
//
//	root$ ./gopacket
//
// gopacket needs to be run as root so that it can access the socket
// created by VPP.
//
// Normally, the memif interface is in the master mode. Pass CLI flag "--slave"
// to create memif in the slave mode:
//
//	root$ ./gopacket --slave
//
// Don't forget to put the opposite memif into the master mode in that case.
//
// To verify the connection, run:
//
//	vpp$ ping 192.168.1.1
//	64 bytes from 192.168.1.1: icmp_seq=2 ttl=255 time=.6974 ms
//	64 bytes from 192.168.1.1: icmp_seq=3 ttl=255 time=.6310 ms
//	64 bytes from 192.168.1.1: icmp_seq=4 ttl=255 time=1.0350 ms
//	64 bytes from 192.168.1.1: icmp_seq=5 ttl=255 time=.5359 ms
//
//	Statistics: 5 sent, 4 received, 20% packet loss
//	vpp$ sh ip arp
//	Time           IP4       Flags      Ethernet              Interface
//	68.5648   192.168.1.1     D    aa:aa:aa:aa:aa:aa memif0/1
//
// Note: it is expected that the first ping is shown as lost. It was actually
// converted to an ARP request. This is a VPP feature common to all interface
// types.
//
// Stop the example with an interrupt signal.
package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"go.fd.io/govpp/extras/libmemif"
)

var (
	// Used to signalize interrupt goroutines to stop
	stopCh chan struct{}

	// MAC address assigned to the memif interface.
	hwAddr net.HardwareAddr

	// IPAddress assigned to the memif interface.
	ipAddr net.IP

	// ErrUnhandledPacket is thrown and printed when an unexpected packet is received.
	ErrUnhandledPacket = errors.New("received an unhandled packet")
)

// OnConnect is called when a memif connection gets established.
func OnConnect(memif *libmemif.Memif) (err error) {
	// Use Memif.GetDetails to get the number of queues.
	details, err := memif.GetDetails()
	if err != nil {
		fmt.Printf("libmemif.GetDetails() error: %v\n", err)
		return
	}

	fmt.Printf("memif %s has been connected: %+v\n", memif.IfName, details)
	stopCh = make(chan struct{})

	// Start a separate go routine for each RX queue.
	// (memif queue is a unit of parallelism for Rx/Tx).
	// Beware: the number of queues created may be lower than what was requested
	// in MemifConfiguration (the master makes the final decision).
	for _, queue := range details.RxQueues {
		ch, err := memif.GetQueueInterruptChan(queue.QueueID)
		if err != nil {
			fmt.Printf("libmemif.Memif.GetQueueInterruptChan() error %v\n", err)
			continue
		}

		go CreateInterruptCallback(memif.NewPacketHandle(queue.QueueID, 10), ch, OnInterrupt)
	}

	return
}

// OnDisconnect is called when a memif connection is lost.
func OnDisconnect(memif *libmemif.Memif) (err error) {
	fmt.Printf("memif %s has been disconnected\n", memif.IfName)
	// Stop all packet producers and consumers.
	close(stopCh)
	return nil
}

// OnInterrupt is called when interrupted
func OnInterrupt(handle *libmemif.MemifPacketHandle) {
	source := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	var responses []gopacket.Packet

	// Process ICMP pings
	for packet := range source.Packets() {
		fmt.Println("Received new packet:")
		fmt.Println(packet.Dump())

		response, err := GeneratePacketResponse(packet)
		if err != nil {
			fmt.Printf("Failed to generate response: %v\n", err)
			continue
		}

		fmt.Println("Sending response:")
		fmt.Println(response.Dump())
		responses = append(responses, response)
	}

	// Answer with ICMP pongs
	for i, response := range responses {
		err := handle.WritePacketData(response.Data())

		switch err {
		case io.EOF:
			return
		case nil:
			fmt.Printf("Succesfully sent packet #%v %v\n", i, len(response.Data()))
		default:
			fmt.Printf("Got error while sending packet #%v %v\n", i, err)
		}
	}
}

// Creates user-friendly memif interrupt callback
func CreateInterruptCallback(handle *libmemif.MemifPacketHandle, interruptCh <-chan struct{}, callback func(handle *libmemif.MemifPacketHandle)) {
	for {
		select {
		case <-interruptCh:
			callback(handle)
		case <-stopCh:
			handle.Close()
			return
		}
	}
}

// GeneratePacketResponse returns an appropriate answer to an ARP request
// or an ICMP echo request.
func GeneratePacketResponse(packet gopacket.Packet) (response gopacket.Packet, err error) {
	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	eth, ok := ethLayer.(*layers.Ethernet)
	if !ok {
		fmt.Println("Missing ETH layer.")
		return nil, ErrUnhandledPacket
	}

	// Set up buffer and options for serialization.
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	switch eth.EthernetType {
	case layers.EthernetTypeARP:
		// Handle ARP request.
		arpLayer := packet.Layer(layers.LayerTypeARP)
		arp, ok := arpLayer.(*layers.ARP)
		if !ok {
			fmt.Println("Missing ARP layer.")
			return nil, ErrUnhandledPacket
		}

		if arp.Operation != layers.ARPRequest {
			fmt.Println("Not ARP request.")
			return nil, ErrUnhandledPacket
		}

		fmt.Println("Received an ARP request.")

		// Build packet layers.
		ethResp := layers.Ethernet{
			SrcMAC:       hwAddr,
			DstMAC:       eth.SrcMAC,
			EthernetType: layers.EthernetTypeARP,
		}

		arpResp := layers.ARP{
			AddrType:          layers.LinkTypeEthernet,
			Protocol:          layers.EthernetTypeIPv4,
			HwAddressSize:     6,
			ProtAddressSize:   4,
			Operation:         layers.ARPReply,
			SourceHwAddress:   []byte(hwAddr),
			SourceProtAddress: []byte(ipAddr),
			DstHwAddress:      arp.SourceHwAddress,
			DstProtAddress:    arp.SourceProtAddress,
		}

		if err := gopacket.SerializeLayers(buf, opts, &ethResp, &arpResp); err != nil {
			fmt.Println("SerializeLayers error: ", err)
			return nil, ErrUnhandledPacket
		}
	case layers.EthernetTypeIPv4:
		// Respond to ICMP request.
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		ipv4, ok := ipLayer.(*layers.IPv4)
		if !ok {
			fmt.Println("Missing IPv4 layer.")
			return nil, ErrUnhandledPacket
		}

		if ipv4.Protocol != layers.IPProtocolICMPv4 {
			fmt.Println("Not ICMPv4 protocol.")
			return nil, ErrUnhandledPacket
		}

		icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
		icmp, ok := icmpLayer.(*layers.ICMPv4)
		if !ok {
			fmt.Println("Missing ICMPv4 layer.")
			return nil, ErrUnhandledPacket
		}

		if icmp.TypeCode.Type() != layers.ICMPv4TypeEchoRequest {
			fmt.Println("Not ICMPv4 echo request.")
			return nil, ErrUnhandledPacket
		}

		fmt.Println("Received an ICMPv4 echo request.")

		// Build packet layers.
		ethResp := layers.Ethernet{
			SrcMAC:       hwAddr,
			DstMAC:       eth.SrcMAC,
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
			SrcIP:      ipAddr,
			DstIP:      ipv4.SrcIP,
		}

		icmpResp := layers.ICMPv4{
			TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoReply, 0),
			Id:       icmp.Id,
			Seq:      icmp.Seq,
		}

		if err := gopacket.SerializeLayers(buf, opts, &ethResp, &ipv4Resp, &icmpResp, gopacket.Payload(icmp.Payload)); err != nil {
			fmt.Println("SerializeLayers error: ", err)
			return nil, ErrUnhandledPacket
		}
	default:
		return nil, ErrUnhandledPacket
	}

	return gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default), nil
}

func main() {
	fmt.Println("Starting 'gopacket' example...")
	var err error

	// Parse MAC address associated with memif interface
	hwAddr, err = net.ParseMAC("aa:aa:aa:aa:aa:aa")
	if err != nil {
		fmt.Printf("Failed to parse the MAC address: %v/n", err)
		return
	}

	// Parse IP address associated with memif interface
	ip := net.ParseIP("192.168.1.1")
	if ip != nil {
		ipAddr = ip.To4()
	}
	if ipAddr == nil {
		fmt.Println("Failed to parse the IP address")
		return
	}

	// If run with the "--slave" option, create memif in the slave mode.
	var isMaster = true
	var appSuffix string
	if len(os.Args) > 1 && (os.Args[1] == "--slave" || os.Args[1] == "-slave") {
		isMaster = false
		appSuffix = "-slave"
	}

	// Initialize libmemif first.
	appName := "gopacket" + appSuffix
	fmt.Println("Initializing libmemif as ", appName)
	err = libmemif.Init(appName)
	if err != nil {
		fmt.Printf("libmemif.Init() error: %v\n", err)
		return
	}

	// Schedule automatic cleanup.
	defer libmemif.Cleanup()

	// Prepare callbacks to use with the memif.
	// The same callbacks could be used with multiple memifs.
	// The first input argument (*libmemif.Memif) can be used to tell which
	// memif the callback was triggered for.
	memifCallbacks := &libmemif.MemifCallbacks{
		OnConnect:    OnConnect,
		OnDisconnect: OnDisconnect,
	}

	// Prepare memif1 configuration.
	memifConfig := &libmemif.MemifConfig{
		MemifMeta: libmemif.MemifMeta{
			IfName:         "memif1",
			ConnID:         1,                       // ConnectionID is an identifier used to match opposite memifs.
			SocketFilename: "/tmp/gopacket-example", // Socket through which the opposite memifs will establish the connection.
			Secret:         "secret",                // Secret used to authenticate the memif connection.
			IsMaster:       isMaster,
			Mode:           libmemif.IfModeEthernet,
			AppName:        appName,
		},
		MemifShmSpecs: libmemif.MemifShmSpecs{
			NumRxQueues:  3, // NumQueues is the (configured!) number of queues for both Rx & Tx.
			NumTxQueues:  3, // The actual number agreed during connection establishment may be smaller!
			BufferSize:   2048,
			Log2RingSize: 10,
		},
	}

	fmt.Printf("Callbacks: %+v\n", memifCallbacks)
	fmt.Printf("Config: %+v\n", memifConfig)

	// Create memif1 interface.
	memif, err := libmemif.CreateInterface(memifConfig, memifCallbacks)
	if err != nil {
		fmt.Printf("libmemif.CreateInterface() error: %v\n", err)
		return
	}

	// Schedule automatic cleanup of the interface.
	defer memif.Close()

	// Wait until an interrupt signal is received.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
