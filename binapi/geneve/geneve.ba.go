// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.10.0
//  VPP:              24.02-release
// source: plugins/geneve.api.json

// Package geneve contains generated bindings for API file geneve.api.
//
// Contents:
// -  8 messages
package geneve

import (
	api "go.fd.io/govpp/api"
	_ "go.fd.io/govpp/binapi/ethernet_types"
	interface_types "go.fd.io/govpp/binapi/interface_types"
	ip_types "go.fd.io/govpp/binapi/ip_types"
	codec "go.fd.io/govpp/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "geneve"
	APIVersion = "2.1.0"
	VersionCrc = 0xe3dbb8a3
)

// /*
//   - Copyright (c) 2017 SUSE LLC.
//   - Licensed under the Apache License, Version 2.0 (the "License");
//   - you may not use this file except in compliance with the License.
//   - You may obtain a copy of the License at:
//     *
//   - http://www.apache.org/licenses/LICENSE-2.0
//     *
//   - Unless required by applicable law or agreed to in writing, software
//   - distributed under the License is distributed on an "AS IS" BASIS,
//   - WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   - See the License for the specific language governing permissions and
//   - limitations under the License.
//
// GeneveAddDelTunnel defines message 'geneve_add_del_tunnel'.
// Deprecated: the message will be removed in the future versions
type GeneveAddDelTunnel struct {
	IsAdd          bool                           `binapi:"bool,name=is_add" json:"is_add,omitempty"`
	LocalAddress   ip_types.Address               `binapi:"address,name=local_address" json:"local_address,omitempty"`
	RemoteAddress  ip_types.Address               `binapi:"address,name=remote_address" json:"remote_address,omitempty"`
	McastSwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=mcast_sw_if_index" json:"mcast_sw_if_index,omitempty"`
	EncapVrfID     uint32                         `binapi:"u32,name=encap_vrf_id" json:"encap_vrf_id,omitempty"`
	DecapNextIndex uint32                         `binapi:"u32,name=decap_next_index" json:"decap_next_index,omitempty"`
	Vni            uint32                         `binapi:"u32,name=vni" json:"vni,omitempty"`
}

func (m *GeneveAddDelTunnel) Reset()               { *m = GeneveAddDelTunnel{} }
func (*GeneveAddDelTunnel) GetMessageName() string { return "geneve_add_del_tunnel" }
func (*GeneveAddDelTunnel) GetCrcString() string   { return "99445831" }
func (*GeneveAddDelTunnel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *GeneveAddDelTunnel) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1      // m.IsAdd
	size += 1      // m.LocalAddress.Af
	size += 1 * 16 // m.LocalAddress.Un
	size += 1      // m.RemoteAddress.Af
	size += 1 * 16 // m.RemoteAddress.Un
	size += 4      // m.McastSwIfIndex
	size += 4      // m.EncapVrfID
	size += 4      // m.DecapNextIndex
	size += 4      // m.Vni
	return size
}
func (m *GeneveAddDelTunnel) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsAdd)
	buf.EncodeUint8(uint8(m.LocalAddress.Af))
	buf.EncodeBytes(m.LocalAddress.Un.XXX_UnionData[:], 16)
	buf.EncodeUint8(uint8(m.RemoteAddress.Af))
	buf.EncodeBytes(m.RemoteAddress.Un.XXX_UnionData[:], 16)
	buf.EncodeUint32(uint32(m.McastSwIfIndex))
	buf.EncodeUint32(m.EncapVrfID)
	buf.EncodeUint32(m.DecapNextIndex)
	buf.EncodeUint32(m.Vni)
	return buf.Bytes(), nil
}
func (m *GeneveAddDelTunnel) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsAdd = buf.DecodeBool()
	m.LocalAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.LocalAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.RemoteAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.RemoteAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.McastSwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.EncapVrfID = buf.DecodeUint32()
	m.DecapNextIndex = buf.DecodeUint32()
	m.Vni = buf.DecodeUint32()
	return nil
}

// GeneveAddDelTunnel2 defines message 'geneve_add_del_tunnel2'.
type GeneveAddDelTunnel2 struct {
	IsAdd          bool                           `binapi:"bool,name=is_add" json:"is_add,omitempty"`
	LocalAddress   ip_types.Address               `binapi:"address,name=local_address" json:"local_address,omitempty"`
	RemoteAddress  ip_types.Address               `binapi:"address,name=remote_address" json:"remote_address,omitempty"`
	McastSwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=mcast_sw_if_index" json:"mcast_sw_if_index,omitempty"`
	EncapVrfID     uint32                         `binapi:"u32,name=encap_vrf_id" json:"encap_vrf_id,omitempty"`
	DecapNextIndex uint32                         `binapi:"u32,name=decap_next_index" json:"decap_next_index,omitempty"`
	Vni            uint32                         `binapi:"u32,name=vni" json:"vni,omitempty"`
	L3Mode         bool                           `binapi:"bool,name=l3_mode" json:"l3_mode,omitempty"`
}

func (m *GeneveAddDelTunnel2) Reset()               { *m = GeneveAddDelTunnel2{} }
func (*GeneveAddDelTunnel2) GetMessageName() string { return "geneve_add_del_tunnel2" }
func (*GeneveAddDelTunnel2) GetCrcString() string   { return "8c2a9999" }
func (*GeneveAddDelTunnel2) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *GeneveAddDelTunnel2) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1      // m.IsAdd
	size += 1      // m.LocalAddress.Af
	size += 1 * 16 // m.LocalAddress.Un
	size += 1      // m.RemoteAddress.Af
	size += 1 * 16 // m.RemoteAddress.Un
	size += 4      // m.McastSwIfIndex
	size += 4      // m.EncapVrfID
	size += 4      // m.DecapNextIndex
	size += 4      // m.Vni
	size += 1      // m.L3Mode
	return size
}
func (m *GeneveAddDelTunnel2) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsAdd)
	buf.EncodeUint8(uint8(m.LocalAddress.Af))
	buf.EncodeBytes(m.LocalAddress.Un.XXX_UnionData[:], 16)
	buf.EncodeUint8(uint8(m.RemoteAddress.Af))
	buf.EncodeBytes(m.RemoteAddress.Un.XXX_UnionData[:], 16)
	buf.EncodeUint32(uint32(m.McastSwIfIndex))
	buf.EncodeUint32(m.EncapVrfID)
	buf.EncodeUint32(m.DecapNextIndex)
	buf.EncodeUint32(m.Vni)
	buf.EncodeBool(m.L3Mode)
	return buf.Bytes(), nil
}
func (m *GeneveAddDelTunnel2) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsAdd = buf.DecodeBool()
	m.LocalAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.LocalAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.RemoteAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.RemoteAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.McastSwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.EncapVrfID = buf.DecodeUint32()
	m.DecapNextIndex = buf.DecodeUint32()
	m.Vni = buf.DecodeUint32()
	m.L3Mode = buf.DecodeBool()
	return nil
}

// GeneveAddDelTunnel2Reply defines message 'geneve_add_del_tunnel2_reply'.
type GeneveAddDelTunnel2Reply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *GeneveAddDelTunnel2Reply) Reset()               { *m = GeneveAddDelTunnel2Reply{} }
func (*GeneveAddDelTunnel2Reply) GetMessageName() string { return "geneve_add_del_tunnel2_reply" }
func (*GeneveAddDelTunnel2Reply) GetCrcString() string   { return "5383d31f" }
func (*GeneveAddDelTunnel2Reply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *GeneveAddDelTunnel2Reply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *GeneveAddDelTunnel2Reply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *GeneveAddDelTunnel2Reply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// GeneveAddDelTunnelReply defines message 'geneve_add_del_tunnel_reply'.
type GeneveAddDelTunnelReply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *GeneveAddDelTunnelReply) Reset()               { *m = GeneveAddDelTunnelReply{} }
func (*GeneveAddDelTunnelReply) GetMessageName() string { return "geneve_add_del_tunnel_reply" }
func (*GeneveAddDelTunnelReply) GetCrcString() string   { return "5383d31f" }
func (*GeneveAddDelTunnelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *GeneveAddDelTunnelReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *GeneveAddDelTunnelReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *GeneveAddDelTunnelReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// GeneveTunnelDetails defines message 'geneve_tunnel_details'.
type GeneveTunnelDetails struct {
	SwIfIndex      interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	SrcAddress     ip_types.Address               `binapi:"address,name=src_address" json:"src_address,omitempty"`
	DstAddress     ip_types.Address               `binapi:"address,name=dst_address" json:"dst_address,omitempty"`
	McastSwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=mcast_sw_if_index" json:"mcast_sw_if_index,omitempty"`
	EncapVrfID     uint32                         `binapi:"u32,name=encap_vrf_id" json:"encap_vrf_id,omitempty"`
	DecapNextIndex uint32                         `binapi:"u32,name=decap_next_index" json:"decap_next_index,omitempty"`
	Vni            uint32                         `binapi:"u32,name=vni" json:"vni,omitempty"`
}

func (m *GeneveTunnelDetails) Reset()               { *m = GeneveTunnelDetails{} }
func (*GeneveTunnelDetails) GetMessageName() string { return "geneve_tunnel_details" }
func (*GeneveTunnelDetails) GetCrcString() string   { return "6b16eb24" }
func (*GeneveTunnelDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *GeneveTunnelDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4      // m.SwIfIndex
	size += 1      // m.SrcAddress.Af
	size += 1 * 16 // m.SrcAddress.Un
	size += 1      // m.DstAddress.Af
	size += 1 * 16 // m.DstAddress.Un
	size += 4      // m.McastSwIfIndex
	size += 4      // m.EncapVrfID
	size += 4      // m.DecapNextIndex
	size += 4      // m.Vni
	return size
}
func (m *GeneveTunnelDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint8(uint8(m.SrcAddress.Af))
	buf.EncodeBytes(m.SrcAddress.Un.XXX_UnionData[:], 16)
	buf.EncodeUint8(uint8(m.DstAddress.Af))
	buf.EncodeBytes(m.DstAddress.Un.XXX_UnionData[:], 16)
	buf.EncodeUint32(uint32(m.McastSwIfIndex))
	buf.EncodeUint32(m.EncapVrfID)
	buf.EncodeUint32(m.DecapNextIndex)
	buf.EncodeUint32(m.Vni)
	return buf.Bytes(), nil
}
func (m *GeneveTunnelDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.SrcAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.SrcAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.DstAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.DstAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.McastSwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.EncapVrfID = buf.DecodeUint32()
	m.DecapNextIndex = buf.DecodeUint32()
	m.Vni = buf.DecodeUint32()
	return nil
}

// GeneveTunnelDump defines message 'geneve_tunnel_dump'.
type GeneveTunnelDump struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *GeneveTunnelDump) Reset()               { *m = GeneveTunnelDump{} }
func (*GeneveTunnelDump) GetMessageName() string { return "geneve_tunnel_dump" }
func (*GeneveTunnelDump) GetCrcString() string   { return "f9e6675e" }
func (*GeneveTunnelDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *GeneveTunnelDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *GeneveTunnelDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *GeneveTunnelDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Interface set geneve-bypass request
//   - sw_if_index - interface used to reach neighbor
//   - is_ipv6 - if non-zero, enable ipv6-geneve-bypass, else ipv4-geneve-bypass
//   - enable - if non-zero enable, else disable
//
// SwInterfaceSetGeneveBypass defines message 'sw_interface_set_geneve_bypass'.
type SwInterfaceSetGeneveBypass struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IsIPv6    bool                           `binapi:"bool,name=is_ipv6" json:"is_ipv6,omitempty"`
	Enable    bool                           `binapi:"bool,name=enable" json:"enable,omitempty"`
}

func (m *SwInterfaceSetGeneveBypass) Reset()               { *m = SwInterfaceSetGeneveBypass{} }
func (*SwInterfaceSetGeneveBypass) GetMessageName() string { return "sw_interface_set_geneve_bypass" }
func (*SwInterfaceSetGeneveBypass) GetCrcString() string   { return "65247409" }
func (*SwInterfaceSetGeneveBypass) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwInterfaceSetGeneveBypass) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	size += 1 // m.IsIPv6
	size += 1 // m.Enable
	return size
}
func (m *SwInterfaceSetGeneveBypass) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeBool(m.IsIPv6)
	buf.EncodeBool(m.Enable)
	return buf.Bytes(), nil
}
func (m *SwInterfaceSetGeneveBypass) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IsIPv6 = buf.DecodeBool()
	m.Enable = buf.DecodeBool()
	return nil
}

// SwInterfaceSetGeneveBypassReply defines message 'sw_interface_set_geneve_bypass_reply'.
type SwInterfaceSetGeneveBypassReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *SwInterfaceSetGeneveBypassReply) Reset() { *m = SwInterfaceSetGeneveBypassReply{} }
func (*SwInterfaceSetGeneveBypassReply) GetMessageName() string {
	return "sw_interface_set_geneve_bypass_reply"
}
func (*SwInterfaceSetGeneveBypassReply) GetCrcString() string { return "e8d4e804" }
func (*SwInterfaceSetGeneveBypassReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwInterfaceSetGeneveBypassReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *SwInterfaceSetGeneveBypassReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *SwInterfaceSetGeneveBypassReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

func init() { file_geneve_binapi_init() }
func file_geneve_binapi_init() {
	api.RegisterMessage((*GeneveAddDelTunnel)(nil), "geneve_add_del_tunnel_99445831")
	api.RegisterMessage((*GeneveAddDelTunnel2)(nil), "geneve_add_del_tunnel2_8c2a9999")
	api.RegisterMessage((*GeneveAddDelTunnel2Reply)(nil), "geneve_add_del_tunnel2_reply_5383d31f")
	api.RegisterMessage((*GeneveAddDelTunnelReply)(nil), "geneve_add_del_tunnel_reply_5383d31f")
	api.RegisterMessage((*GeneveTunnelDetails)(nil), "geneve_tunnel_details_6b16eb24")
	api.RegisterMessage((*GeneveTunnelDump)(nil), "geneve_tunnel_dump_f9e6675e")
	api.RegisterMessage((*SwInterfaceSetGeneveBypass)(nil), "sw_interface_set_geneve_bypass_65247409")
	api.RegisterMessage((*SwInterfaceSetGeneveBypassReply)(nil), "sw_interface_set_geneve_bypass_reply_e8d4e804")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*GeneveAddDelTunnel)(nil),
		(*GeneveAddDelTunnel2)(nil),
		(*GeneveAddDelTunnel2Reply)(nil),
		(*GeneveAddDelTunnelReply)(nil),
		(*GeneveTunnelDetails)(nil),
		(*GeneveTunnelDump)(nil),
		(*SwInterfaceSetGeneveBypass)(nil),
		(*SwInterfaceSetGeneveBypassReply)(nil),
	}
}
