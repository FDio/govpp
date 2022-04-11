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
	"fmt"
	"net"
)

type SaFlags uint32
type IPsecTunnelProtection struct {
	SwIfIndex   uint32
	NextHop     net.IP
	OutSAIndex  uint32
	InSAIndices []uint32
}
type Tunnel struct {
	Src     net.IP
	Dst     net.IP
	TableID uint32
}
type IPSecSA struct {
	SAId         uint32
	Spi          uint32
	Salt         uint32
	CryptoKey    []byte
	IntegrityKey []byte
	SrcPort      int
	DstPort      int
	Tunnel       *Tunnel
	Flags        SaFlags
}

func (t *Tunnel) String() string {
	if t.TableID != 0 {
		return fmt.Sprintf("%s->%s tbl:%d", t.Src.String(), t.Dst.String(), t.TableID)
	}
	return fmt.Sprintf("%s->%s", t.Src.String(), t.Dst.String())
}

type VppIpsec interface {
	GetIPsecTunnelProtection(tunnelInterface uint32) (protections []IPsecTunnelProtection, err error)
	AddIpsecSA(sa *IPSecSA) error
	DelIpsecSA(sa *IPSecSA) error
	AddIpsecSAProtect(swIfIndex, saIn, saOut uint32) error
	DelIpsecSAProtect(swIfIndex uint32) error
	AddIpsecInterface() (uint32, error)
	DelIpsecInterface(swIfIndex uint32) error
	GetSaFlagNone() SaFlags
	GetSaFlagUseEsn() SaFlags
	GetSaFlagAntiReplay() SaFlags
	GetSaFlagIsTunnel() SaFlags
	GetSaFlagIsTunnelV6() SaFlags
	GetSaFlagUdpEncap() SaFlags
	GetSaFlagIsInbound() SaFlags
	GetSaFlagAsync() SaFlags
}
