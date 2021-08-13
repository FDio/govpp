// Copyright (C) 2021 Cisco Systems Inc.
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

type IPIPTunnel struct {
	Src       net.IP
	Dst       net.IP
	TableID   uint32
	SwIfIndex uint32
}

func (t *IPIPTunnel) String() string {
	if t.TableID != 0 {
		return fmt.Sprintf("[%d] %s->%s tbl:%d", t.SwIfIndex, t.Src.String(), t.Dst.String(), t.TableID)
	}
	return fmt.Sprintf("[%d] %s->%s", t.SwIfIndex, t.Src.String(), t.Dst.String())
}

type VppIPIP interface {
	ListIPIPTunnels() ([]*IPIPTunnel, error)
	AddIPIPTunnel(tunnel *IPIPTunnel) (uint32, error)
	DelIPIPTunnel(tunnel *IPIPTunnel) (err error)
}
