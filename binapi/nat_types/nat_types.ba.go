// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.10.0
//  VPP:              24.02-release
// source: plugins/nat_types.api.json

// Package nat_types contains generated bindings for API file nat_types.api.
//
// Contents:
// -  2 enums
// -  1 struct
package nat_types

import (
	"strconv"

	api "go.fd.io/govpp/api"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "nat_types"
	APIVersion = "0.0.1"
	VersionCrc = 0x2ca9110f
)

// NatConfigFlags defines enum 'nat_config_flags'.
type NatConfigFlags uint8

const (
	NAT_IS_NONE           NatConfigFlags = 0
	NAT_IS_TWICE_NAT      NatConfigFlags = 1
	NAT_IS_SELF_TWICE_NAT NatConfigFlags = 2
	NAT_IS_OUT2IN_ONLY    NatConfigFlags = 4
	NAT_IS_ADDR_ONLY      NatConfigFlags = 8
	NAT_IS_OUTSIDE        NatConfigFlags = 16
	NAT_IS_INSIDE         NatConfigFlags = 32
	NAT_IS_STATIC         NatConfigFlags = 64
	NAT_IS_EXT_HOST_VALID NatConfigFlags = 128
)

var (
	NatConfigFlags_name = map[uint8]string{
		0:   "NAT_IS_NONE",
		1:   "NAT_IS_TWICE_NAT",
		2:   "NAT_IS_SELF_TWICE_NAT",
		4:   "NAT_IS_OUT2IN_ONLY",
		8:   "NAT_IS_ADDR_ONLY",
		16:  "NAT_IS_OUTSIDE",
		32:  "NAT_IS_INSIDE",
		64:  "NAT_IS_STATIC",
		128: "NAT_IS_EXT_HOST_VALID",
	}
	NatConfigFlags_value = map[string]uint8{
		"NAT_IS_NONE":           0,
		"NAT_IS_TWICE_NAT":      1,
		"NAT_IS_SELF_TWICE_NAT": 2,
		"NAT_IS_OUT2IN_ONLY":    4,
		"NAT_IS_ADDR_ONLY":      8,
		"NAT_IS_OUTSIDE":        16,
		"NAT_IS_INSIDE":         32,
		"NAT_IS_STATIC":         64,
		"NAT_IS_EXT_HOST_VALID": 128,
	}
)

func (x NatConfigFlags) String() string {
	s, ok := NatConfigFlags_name[uint8(x)]
	if ok {
		return s
	}
	str := func(n uint8) string {
		s, ok := NatConfigFlags_name[uint8(n)]
		if ok {
			return s
		}
		return "NatConfigFlags(" + strconv.Itoa(int(n)) + ")"
	}
	for i := uint8(0); i <= 8; i++ {
		val := uint8(x)
		if val&(1<<i) != 0 {
			if s != "" {
				s += "|"
			}
			s += str(1 << i)
		}
	}
	if s == "" {
		return str(uint8(x))
	}
	return s
}

// NatLogLevel defines enum 'nat_log_level'.
type NatLogLevel uint8

const (
	NAT_LOG_NONE    NatLogLevel = 0
	NAT_LOG_ERROR   NatLogLevel = 1
	NAT_LOG_WARNING NatLogLevel = 2
	NAT_LOG_NOTICE  NatLogLevel = 3
	NAT_LOG_INFO    NatLogLevel = 4
	NAT_LOG_DEBUG   NatLogLevel = 5
)

var (
	NatLogLevel_name = map[uint8]string{
		0: "NAT_LOG_NONE",
		1: "NAT_LOG_ERROR",
		2: "NAT_LOG_WARNING",
		3: "NAT_LOG_NOTICE",
		4: "NAT_LOG_INFO",
		5: "NAT_LOG_DEBUG",
	}
	NatLogLevel_value = map[string]uint8{
		"NAT_LOG_NONE":    0,
		"NAT_LOG_ERROR":   1,
		"NAT_LOG_WARNING": 2,
		"NAT_LOG_NOTICE":  3,
		"NAT_LOG_INFO":    4,
		"NAT_LOG_DEBUG":   5,
	}
)

func (x NatLogLevel) String() string {
	s, ok := NatLogLevel_name[uint8(x)]
	if ok {
		return s
	}
	return "NatLogLevel(" + strconv.Itoa(int(x)) + ")"
}

// NatTimeouts defines type 'nat_timeouts'.
type NatTimeouts struct {
	UDP            uint32 `binapi:"u32,name=udp" json:"udp,omitempty"`
	TCPEstablished uint32 `binapi:"u32,name=tcp_established" json:"tcp_established,omitempty"`
	TCPTransitory  uint32 `binapi:"u32,name=tcp_transitory" json:"tcp_transitory,omitempty"`
	ICMP           uint32 `binapi:"u32,name=icmp" json:"icmp,omitempty"`
}
