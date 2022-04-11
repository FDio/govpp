// Copyright (C) 2019 Cisco Systems Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package types

import (
	"net"

	"golang.org/x/sys/unix"
)

type IPProto uint8

const (
	TCP IPProto = iota
	UDP
	SCTP
	ICMP
	ICMP6
	INVALID
)
const (
	// Family type definitions
	FAMILY_ALL = unix.AF_UNSPEC
	FAMILY_V4  = unix.AF_INET
	FAMILY_V6  = unix.AF_INET6
)

type IfAddress struct {
	IPNet     net.IPNet
	SwIfIndex uint32
}

func GetIPFamily(ip net.IP) int {
	if len(ip) <= net.IPv4len {
		return FAMILY_V4
	}
	if ip.To4() != nil {
		return FAMILY_V4
	}
	return FAMILY_V6
}
func IsIP4(ip net.IP) bool {
	return GetIPFamily(ip) == FAMILY_V4
}
func IsIP6(ip net.IP) bool {
	return GetIPFamily(ip) == FAMILY_V6
}
func formatProto(proto IPProto) string {
	switch proto {
	case UDP:
		return "UDP"
	case TCP:
		return "TCP"
	case SCTP:
		return "SCTP"
	case ICMP:
		return "ICMP"
	case ICMP6:
		return "ICMP6"
	default:
		return "???"
	}
}
