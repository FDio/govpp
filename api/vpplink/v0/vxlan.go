// Copyright (C) 2020 Cisco Systems Inc.
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

type VXLanTunnel struct {
	SrcAddress     net.IP
	DstAddress     net.IP
	SrcPort        uint16
	DstPort        uint16
	Vni            uint32
	DecapNextIndex uint32
	SwIfIndex      uint32
}

func (t *VXLanTunnel) String() string {
	return fmt.Sprintf("[%d]vni=%d %s:%d->%s:%d", t.SwIfIndex, t.Vni, t.SrcAddress, t.SrcPort, t.DstAddress, t.DstPort)
}
