// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.4.0-dev
//  VPP:              20.01
// source: .vppapi/core/vxlan.api.json

// Package vxlan contains generated bindings for API file vxlan.api.
//
// Contents:
//   8 messages
//
package vxlan

import (
	api "git.fd.io/govpp.git/api"
	codec "git.fd.io/govpp.git/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "vxlan"
	APIVersion = "1.1.0"
	VersionCrc = 0xa95aa271
)

// SwInterfaceSetVxlanBypass defines message 'sw_interface_set_vxlan_bypass'.
type SwInterfaceSetVxlanBypass struct {
	SwIfIndex uint32 `binapi:"u32,name=sw_if_index" json:"sw_if_index,omitempty"`
	IsIPv6    uint8  `binapi:"u8,name=is_ipv6" json:"is_ipv6,omitempty"`
	Enable    uint8  `binapi:"u8,name=enable" json:"enable,omitempty"`
}

func (m *SwInterfaceSetVxlanBypass) Reset()               { *m = SwInterfaceSetVxlanBypass{} }
func (*SwInterfaceSetVxlanBypass) GetMessageName() string { return "sw_interface_set_vxlan_bypass" }
func (*SwInterfaceSetVxlanBypass) GetCrcString() string   { return "e74ca095" }
func (*SwInterfaceSetVxlanBypass) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwInterfaceSetVxlanBypass) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	size += 1 // m.IsIPv6
	size += 1 // m.Enable
	return size
}
func (m *SwInterfaceSetVxlanBypass) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.SwIfIndex)
	buf.EncodeUint8(m.IsIPv6)
	buf.EncodeUint8(m.Enable)
	return buf.Bytes(), nil
}
func (m *SwInterfaceSetVxlanBypass) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = buf.DecodeUint32()
	m.IsIPv6 = buf.DecodeUint8()
	m.Enable = buf.DecodeUint8()
	return nil
}

// SwInterfaceSetVxlanBypassReply defines message 'sw_interface_set_vxlan_bypass_reply'.
type SwInterfaceSetVxlanBypassReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *SwInterfaceSetVxlanBypassReply) Reset() { *m = SwInterfaceSetVxlanBypassReply{} }
func (*SwInterfaceSetVxlanBypassReply) GetMessageName() string {
	return "sw_interface_set_vxlan_bypass_reply"
}
func (*SwInterfaceSetVxlanBypassReply) GetCrcString() string { return "e8d4e804" }
func (*SwInterfaceSetVxlanBypassReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwInterfaceSetVxlanBypassReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *SwInterfaceSetVxlanBypassReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *SwInterfaceSetVxlanBypassReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// VxlanAddDelTunnel defines message 'vxlan_add_del_tunnel'.
type VxlanAddDelTunnel struct {
	IsAdd          uint8  `binapi:"u8,name=is_add" json:"is_add,omitempty"`
	IsIPv6         uint8  `binapi:"u8,name=is_ipv6" json:"is_ipv6,omitempty"`
	Instance       uint32 `binapi:"u32,name=instance" json:"instance,omitempty"`
	SrcAddress     []byte `binapi:"u8[16],name=src_address" json:"src_address,omitempty"`
	DstAddress     []byte `binapi:"u8[16],name=dst_address" json:"dst_address,omitempty"`
	McastSwIfIndex uint32 `binapi:"u32,name=mcast_sw_if_index" json:"mcast_sw_if_index,omitempty"`
	EncapVrfID     uint32 `binapi:"u32,name=encap_vrf_id" json:"encap_vrf_id,omitempty"`
	DecapNextIndex uint32 `binapi:"u32,name=decap_next_index" json:"decap_next_index,omitempty"`
	Vni            uint32 `binapi:"u32,name=vni" json:"vni,omitempty"`
}

func (m *VxlanAddDelTunnel) Reset()               { *m = VxlanAddDelTunnel{} }
func (*VxlanAddDelTunnel) GetMessageName() string { return "vxlan_add_del_tunnel" }
func (*VxlanAddDelTunnel) GetCrcString() string   { return "00f4bdd0" }
func (*VxlanAddDelTunnel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *VxlanAddDelTunnel) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1      // m.IsAdd
	size += 1      // m.IsIPv6
	size += 4      // m.Instance
	size += 1 * 16 // m.SrcAddress
	size += 1 * 16 // m.DstAddress
	size += 4      // m.McastSwIfIndex
	size += 4      // m.EncapVrfID
	size += 4      // m.DecapNextIndex
	size += 4      // m.Vni
	return size
}
func (m *VxlanAddDelTunnel) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint8(m.IsAdd)
	buf.EncodeUint8(m.IsIPv6)
	buf.EncodeUint32(m.Instance)
	buf.EncodeBytes(m.SrcAddress, 16)
	buf.EncodeBytes(m.DstAddress, 16)
	buf.EncodeUint32(m.McastSwIfIndex)
	buf.EncodeUint32(m.EncapVrfID)
	buf.EncodeUint32(m.DecapNextIndex)
	buf.EncodeUint32(m.Vni)
	return buf.Bytes(), nil
}
func (m *VxlanAddDelTunnel) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsAdd = buf.DecodeUint8()
	m.IsIPv6 = buf.DecodeUint8()
	m.Instance = buf.DecodeUint32()
	m.SrcAddress = make([]byte, 16)
	copy(m.SrcAddress, buf.DecodeBytes(len(m.SrcAddress)))
	m.DstAddress = make([]byte, 16)
	copy(m.DstAddress, buf.DecodeBytes(len(m.DstAddress)))
	m.McastSwIfIndex = buf.DecodeUint32()
	m.EncapVrfID = buf.DecodeUint32()
	m.DecapNextIndex = buf.DecodeUint32()
	m.Vni = buf.DecodeUint32()
	return nil
}

// VxlanAddDelTunnelReply defines message 'vxlan_add_del_tunnel_reply'.
type VxlanAddDelTunnelReply struct {
	Retval    int32  `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex uint32 `binapi:"u32,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *VxlanAddDelTunnelReply) Reset()               { *m = VxlanAddDelTunnelReply{} }
func (*VxlanAddDelTunnelReply) GetMessageName() string { return "vxlan_add_del_tunnel_reply" }
func (*VxlanAddDelTunnelReply) GetCrcString() string   { return "fda5941f" }
func (*VxlanAddDelTunnelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *VxlanAddDelTunnelReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *VxlanAddDelTunnelReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(m.SwIfIndex)
	return buf.Bytes(), nil
}
func (m *VxlanAddDelTunnelReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = buf.DecodeUint32()
	return nil
}

// VxlanOffloadRx defines message 'vxlan_offload_rx'.
type VxlanOffloadRx struct {
	HwIfIndex uint32 `binapi:"u32,name=hw_if_index" json:"hw_if_index,omitempty"`
	SwIfIndex uint32 `binapi:"u32,name=sw_if_index" json:"sw_if_index,omitempty"`
	Enable    uint8  `binapi:"u8,name=enable" json:"enable,omitempty"`
}

func (m *VxlanOffloadRx) Reset()               { *m = VxlanOffloadRx{} }
func (*VxlanOffloadRx) GetMessageName() string { return "vxlan_offload_rx" }
func (*VxlanOffloadRx) GetCrcString() string   { return "f0b08786" }
func (*VxlanOffloadRx) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *VxlanOffloadRx) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.HwIfIndex
	size += 4 // m.SwIfIndex
	size += 1 // m.Enable
	return size
}
func (m *VxlanOffloadRx) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.HwIfIndex)
	buf.EncodeUint32(m.SwIfIndex)
	buf.EncodeUint8(m.Enable)
	return buf.Bytes(), nil
}
func (m *VxlanOffloadRx) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.HwIfIndex = buf.DecodeUint32()
	m.SwIfIndex = buf.DecodeUint32()
	m.Enable = buf.DecodeUint8()
	return nil
}

// VxlanOffloadRxReply defines message 'vxlan_offload_rx_reply'.
type VxlanOffloadRxReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *VxlanOffloadRxReply) Reset()               { *m = VxlanOffloadRxReply{} }
func (*VxlanOffloadRxReply) GetMessageName() string { return "vxlan_offload_rx_reply" }
func (*VxlanOffloadRxReply) GetCrcString() string   { return "e8d4e804" }
func (*VxlanOffloadRxReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *VxlanOffloadRxReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *VxlanOffloadRxReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *VxlanOffloadRxReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// VxlanTunnelDetails defines message 'vxlan_tunnel_details'.
type VxlanTunnelDetails struct {
	SwIfIndex      uint32 `binapi:"u32,name=sw_if_index" json:"sw_if_index,omitempty"`
	Instance       uint32 `binapi:"u32,name=instance" json:"instance,omitempty"`
	SrcAddress     []byte `binapi:"u8[16],name=src_address" json:"src_address,omitempty"`
	DstAddress     []byte `binapi:"u8[16],name=dst_address" json:"dst_address,omitempty"`
	McastSwIfIndex uint32 `binapi:"u32,name=mcast_sw_if_index" json:"mcast_sw_if_index,omitempty"`
	EncapVrfID     uint32 `binapi:"u32,name=encap_vrf_id" json:"encap_vrf_id,omitempty"`
	DecapNextIndex uint32 `binapi:"u32,name=decap_next_index" json:"decap_next_index,omitempty"`
	Vni            uint32 `binapi:"u32,name=vni" json:"vni,omitempty"`
	IsIPv6         uint8  `binapi:"u8,name=is_ipv6" json:"is_ipv6,omitempty"`
}

func (m *VxlanTunnelDetails) Reset()               { *m = VxlanTunnelDetails{} }
func (*VxlanTunnelDetails) GetMessageName() string { return "vxlan_tunnel_details" }
func (*VxlanTunnelDetails) GetCrcString() string   { return "ce38e127" }
func (*VxlanTunnelDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *VxlanTunnelDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4      // m.SwIfIndex
	size += 4      // m.Instance
	size += 1 * 16 // m.SrcAddress
	size += 1 * 16 // m.DstAddress
	size += 4      // m.McastSwIfIndex
	size += 4      // m.EncapVrfID
	size += 4      // m.DecapNextIndex
	size += 4      // m.Vni
	size += 1      // m.IsIPv6
	return size
}
func (m *VxlanTunnelDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.SwIfIndex)
	buf.EncodeUint32(m.Instance)
	buf.EncodeBytes(m.SrcAddress, 16)
	buf.EncodeBytes(m.DstAddress, 16)
	buf.EncodeUint32(m.McastSwIfIndex)
	buf.EncodeUint32(m.EncapVrfID)
	buf.EncodeUint32(m.DecapNextIndex)
	buf.EncodeUint32(m.Vni)
	buf.EncodeUint8(m.IsIPv6)
	return buf.Bytes(), nil
}
func (m *VxlanTunnelDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = buf.DecodeUint32()
	m.Instance = buf.DecodeUint32()
	m.SrcAddress = make([]byte, 16)
	copy(m.SrcAddress, buf.DecodeBytes(len(m.SrcAddress)))
	m.DstAddress = make([]byte, 16)
	copy(m.DstAddress, buf.DecodeBytes(len(m.DstAddress)))
	m.McastSwIfIndex = buf.DecodeUint32()
	m.EncapVrfID = buf.DecodeUint32()
	m.DecapNextIndex = buf.DecodeUint32()
	m.Vni = buf.DecodeUint32()
	m.IsIPv6 = buf.DecodeUint8()
	return nil
}

// VxlanTunnelDump defines message 'vxlan_tunnel_dump'.
type VxlanTunnelDump struct {
	SwIfIndex uint32 `binapi:"u32,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *VxlanTunnelDump) Reset()               { *m = VxlanTunnelDump{} }
func (*VxlanTunnelDump) GetMessageName() string { return "vxlan_tunnel_dump" }
func (*VxlanTunnelDump) GetCrcString() string   { return "529cb13f" }
func (*VxlanTunnelDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *VxlanTunnelDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *VxlanTunnelDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.SwIfIndex)
	return buf.Bytes(), nil
}
func (m *VxlanTunnelDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = buf.DecodeUint32()
	return nil
}

func init() { file_vxlan_binapi_init() }
func file_vxlan_binapi_init() {
	api.RegisterMessage((*SwInterfaceSetVxlanBypass)(nil), "sw_interface_set_vxlan_bypass_e74ca095")
	api.RegisterMessage((*SwInterfaceSetVxlanBypassReply)(nil), "sw_interface_set_vxlan_bypass_reply_e8d4e804")
	api.RegisterMessage((*VxlanAddDelTunnel)(nil), "vxlan_add_del_tunnel_00f4bdd0")
	api.RegisterMessage((*VxlanAddDelTunnelReply)(nil), "vxlan_add_del_tunnel_reply_fda5941f")
	api.RegisterMessage((*VxlanOffloadRx)(nil), "vxlan_offload_rx_f0b08786")
	api.RegisterMessage((*VxlanOffloadRxReply)(nil), "vxlan_offload_rx_reply_e8d4e804")
	api.RegisterMessage((*VxlanTunnelDetails)(nil), "vxlan_tunnel_details_ce38e127")
	api.RegisterMessage((*VxlanTunnelDump)(nil), "vxlan_tunnel_dump_529cb13f")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*SwInterfaceSetVxlanBypass)(nil),
		(*SwInterfaceSetVxlanBypassReply)(nil),
		(*VxlanAddDelTunnel)(nil),
		(*VxlanAddDelTunnelReply)(nil),
		(*VxlanOffloadRx)(nil),
		(*VxlanOffloadRxReply)(nil),
		(*VxlanTunnelDetails)(nil),
		(*VxlanTunnelDump)(nil),
	}
}