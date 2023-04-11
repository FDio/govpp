// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.7.0
//  VPP:              22.10-release
// source: plugins/nat64.api.json

// Package nat64 contains generated bindings for API file nat64.api.
//
// Contents:
// - 26 messages
package nat64

import (
	api "go.fd.io/govpp/api"
	interface_types "go.fd.io/govpp/binapi/interface_types"
	ip_types "go.fd.io/govpp/binapi/ip_types"
	nat_types "go.fd.io/govpp/binapi/nat_types"
	codec "go.fd.io/govpp/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "nat64"
	APIVersion = "1.0.0"
	VersionCrc = 0xfbd06e33
)

// Nat64AddDelInterface defines message 'nat64_add_del_interface'.
type Nat64AddDelInterface struct {
	IsAdd     bool                           `binapi:"bool,name=is_add" json:"is_add,omitempty"`
	Flags     nat_types.NatConfigFlags       `binapi:"nat_config_flags,name=flags" json:"flags,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *Nat64AddDelInterface) Reset()               { *m = Nat64AddDelInterface{} }
func (*Nat64AddDelInterface) GetMessageName() string { return "nat64_add_del_interface" }
func (*Nat64AddDelInterface) GetCrcString() string   { return "f3699b83" }
func (*Nat64AddDelInterface) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64AddDelInterface) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 // m.IsAdd
	size += 1 // m.Flags
	size += 4 // m.SwIfIndex
	return size
}
func (m *Nat64AddDelInterface) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsAdd)
	buf.EncodeUint8(uint8(m.Flags))
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *Nat64AddDelInterface) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsAdd = buf.DecodeBool()
	m.Flags = nat_types.NatConfigFlags(buf.DecodeUint8())
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Nat64AddDelInterfaceAddr defines message 'nat64_add_del_interface_addr'.
type Nat64AddDelInterfaceAddr struct {
	IsAdd     bool                           `binapi:"bool,name=is_add" json:"is_add,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *Nat64AddDelInterfaceAddr) Reset()               { *m = Nat64AddDelInterfaceAddr{} }
func (*Nat64AddDelInterfaceAddr) GetMessageName() string { return "nat64_add_del_interface_addr" }
func (*Nat64AddDelInterfaceAddr) GetCrcString() string   { return "47d6e753" }
func (*Nat64AddDelInterfaceAddr) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64AddDelInterfaceAddr) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 // m.IsAdd
	size += 4 // m.SwIfIndex
	return size
}
func (m *Nat64AddDelInterfaceAddr) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsAdd)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *Nat64AddDelInterfaceAddr) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsAdd = buf.DecodeBool()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Nat64AddDelInterfaceAddrReply defines message 'nat64_add_del_interface_addr_reply'.
type Nat64AddDelInterfaceAddrReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *Nat64AddDelInterfaceAddrReply) Reset() { *m = Nat64AddDelInterfaceAddrReply{} }
func (*Nat64AddDelInterfaceAddrReply) GetMessageName() string {
	return "nat64_add_del_interface_addr_reply"
}
func (*Nat64AddDelInterfaceAddrReply) GetCrcString() string { return "e8d4e804" }
func (*Nat64AddDelInterfaceAddrReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64AddDelInterfaceAddrReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *Nat64AddDelInterfaceAddrReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *Nat64AddDelInterfaceAddrReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// Nat64AddDelInterfaceReply defines message 'nat64_add_del_interface_reply'.
type Nat64AddDelInterfaceReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *Nat64AddDelInterfaceReply) Reset()               { *m = Nat64AddDelInterfaceReply{} }
func (*Nat64AddDelInterfaceReply) GetMessageName() string { return "nat64_add_del_interface_reply" }
func (*Nat64AddDelInterfaceReply) GetCrcString() string   { return "e8d4e804" }
func (*Nat64AddDelInterfaceReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64AddDelInterfaceReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *Nat64AddDelInterfaceReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *Nat64AddDelInterfaceReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// Nat64AddDelPoolAddrRange defines message 'nat64_add_del_pool_addr_range'.
type Nat64AddDelPoolAddrRange struct {
	StartAddr ip_types.IP4Address `binapi:"ip4_address,name=start_addr" json:"start_addr,omitempty"`
	EndAddr   ip_types.IP4Address `binapi:"ip4_address,name=end_addr" json:"end_addr,omitempty"`
	VrfID     uint32              `binapi:"u32,name=vrf_id" json:"vrf_id,omitempty"`
	IsAdd     bool                `binapi:"bool,name=is_add" json:"is_add,omitempty"`
}

func (m *Nat64AddDelPoolAddrRange) Reset()               { *m = Nat64AddDelPoolAddrRange{} }
func (*Nat64AddDelPoolAddrRange) GetMessageName() string { return "nat64_add_del_pool_addr_range" }
func (*Nat64AddDelPoolAddrRange) GetCrcString() string   { return "a3b944e3" }
func (*Nat64AddDelPoolAddrRange) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64AddDelPoolAddrRange) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 * 4 // m.StartAddr
	size += 1 * 4 // m.EndAddr
	size += 4     // m.VrfID
	size += 1     // m.IsAdd
	return size
}
func (m *Nat64AddDelPoolAddrRange) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBytes(m.StartAddr[:], 4)
	buf.EncodeBytes(m.EndAddr[:], 4)
	buf.EncodeUint32(m.VrfID)
	buf.EncodeBool(m.IsAdd)
	return buf.Bytes(), nil
}
func (m *Nat64AddDelPoolAddrRange) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	copy(m.StartAddr[:], buf.DecodeBytes(4))
	copy(m.EndAddr[:], buf.DecodeBytes(4))
	m.VrfID = buf.DecodeUint32()
	m.IsAdd = buf.DecodeBool()
	return nil
}

// Nat64AddDelPoolAddrRangeReply defines message 'nat64_add_del_pool_addr_range_reply'.
type Nat64AddDelPoolAddrRangeReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *Nat64AddDelPoolAddrRangeReply) Reset() { *m = Nat64AddDelPoolAddrRangeReply{} }
func (*Nat64AddDelPoolAddrRangeReply) GetMessageName() string {
	return "nat64_add_del_pool_addr_range_reply"
}
func (*Nat64AddDelPoolAddrRangeReply) GetCrcString() string { return "e8d4e804" }
func (*Nat64AddDelPoolAddrRangeReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64AddDelPoolAddrRangeReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *Nat64AddDelPoolAddrRangeReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *Nat64AddDelPoolAddrRangeReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// Nat64AddDelPrefix defines message 'nat64_add_del_prefix'.
type Nat64AddDelPrefix struct {
	Prefix ip_types.IP6Prefix `binapi:"ip6_prefix,name=prefix" json:"prefix,omitempty"`
	VrfID  uint32             `binapi:"u32,name=vrf_id" json:"vrf_id,omitempty"`
	IsAdd  bool               `binapi:"bool,name=is_add" json:"is_add,omitempty"`
}

func (m *Nat64AddDelPrefix) Reset()               { *m = Nat64AddDelPrefix{} }
func (*Nat64AddDelPrefix) GetMessageName() string { return "nat64_add_del_prefix" }
func (*Nat64AddDelPrefix) GetCrcString() string   { return "727b2f4c" }
func (*Nat64AddDelPrefix) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64AddDelPrefix) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 * 16 // m.Prefix.Address
	size += 1      // m.Prefix.Len
	size += 4      // m.VrfID
	size += 1      // m.IsAdd
	return size
}
func (m *Nat64AddDelPrefix) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBytes(m.Prefix.Address[:], 16)
	buf.EncodeUint8(m.Prefix.Len)
	buf.EncodeUint32(m.VrfID)
	buf.EncodeBool(m.IsAdd)
	return buf.Bytes(), nil
}
func (m *Nat64AddDelPrefix) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	copy(m.Prefix.Address[:], buf.DecodeBytes(16))
	m.Prefix.Len = buf.DecodeUint8()
	m.VrfID = buf.DecodeUint32()
	m.IsAdd = buf.DecodeBool()
	return nil
}

// Nat64AddDelPrefixReply defines message 'nat64_add_del_prefix_reply'.
type Nat64AddDelPrefixReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *Nat64AddDelPrefixReply) Reset()               { *m = Nat64AddDelPrefixReply{} }
func (*Nat64AddDelPrefixReply) GetMessageName() string { return "nat64_add_del_prefix_reply" }
func (*Nat64AddDelPrefixReply) GetCrcString() string   { return "e8d4e804" }
func (*Nat64AddDelPrefixReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64AddDelPrefixReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *Nat64AddDelPrefixReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *Nat64AddDelPrefixReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// Nat64AddDelStaticBib defines message 'nat64_add_del_static_bib'.
type Nat64AddDelStaticBib struct {
	IAddr ip_types.IP6Address `binapi:"ip6_address,name=i_addr" json:"i_addr,omitempty"`
	OAddr ip_types.IP4Address `binapi:"ip4_address,name=o_addr" json:"o_addr,omitempty"`
	IPort uint16              `binapi:"u16,name=i_port" json:"i_port,omitempty"`
	OPort uint16              `binapi:"u16,name=o_port" json:"o_port,omitempty"`
	VrfID uint32              `binapi:"u32,name=vrf_id" json:"vrf_id,omitempty"`
	Proto uint8               `binapi:"u8,name=proto" json:"proto,omitempty"`
	IsAdd bool                `binapi:"bool,name=is_add" json:"is_add,omitempty"`
}

func (m *Nat64AddDelStaticBib) Reset()               { *m = Nat64AddDelStaticBib{} }
func (*Nat64AddDelStaticBib) GetMessageName() string { return "nat64_add_del_static_bib" }
func (*Nat64AddDelStaticBib) GetCrcString() string   { return "1c404de5" }
func (*Nat64AddDelStaticBib) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64AddDelStaticBib) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 * 16 // m.IAddr
	size += 1 * 4  // m.OAddr
	size += 2      // m.IPort
	size += 2      // m.OPort
	size += 4      // m.VrfID
	size += 1      // m.Proto
	size += 1      // m.IsAdd
	return size
}
func (m *Nat64AddDelStaticBib) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBytes(m.IAddr[:], 16)
	buf.EncodeBytes(m.OAddr[:], 4)
	buf.EncodeUint16(m.IPort)
	buf.EncodeUint16(m.OPort)
	buf.EncodeUint32(m.VrfID)
	buf.EncodeUint8(m.Proto)
	buf.EncodeBool(m.IsAdd)
	return buf.Bytes(), nil
}
func (m *Nat64AddDelStaticBib) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	copy(m.IAddr[:], buf.DecodeBytes(16))
	copy(m.OAddr[:], buf.DecodeBytes(4))
	m.IPort = buf.DecodeUint16()
	m.OPort = buf.DecodeUint16()
	m.VrfID = buf.DecodeUint32()
	m.Proto = buf.DecodeUint8()
	m.IsAdd = buf.DecodeBool()
	return nil
}

// Nat64AddDelStaticBibReply defines message 'nat64_add_del_static_bib_reply'.
type Nat64AddDelStaticBibReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *Nat64AddDelStaticBibReply) Reset()               { *m = Nat64AddDelStaticBibReply{} }
func (*Nat64AddDelStaticBibReply) GetMessageName() string { return "nat64_add_del_static_bib_reply" }
func (*Nat64AddDelStaticBibReply) GetCrcString() string   { return "e8d4e804" }
func (*Nat64AddDelStaticBibReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64AddDelStaticBibReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *Nat64AddDelStaticBibReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *Nat64AddDelStaticBibReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// Nat64BibDetails defines message 'nat64_bib_details'.
type Nat64BibDetails struct {
	IAddr  ip_types.IP6Address      `binapi:"ip6_address,name=i_addr" json:"i_addr,omitempty"`
	OAddr  ip_types.IP4Address      `binapi:"ip4_address,name=o_addr" json:"o_addr,omitempty"`
	IPort  uint16                   `binapi:"u16,name=i_port" json:"i_port,omitempty"`
	OPort  uint16                   `binapi:"u16,name=o_port" json:"o_port,omitempty"`
	VrfID  uint32                   `binapi:"u32,name=vrf_id" json:"vrf_id,omitempty"`
	Proto  uint8                    `binapi:"u8,name=proto" json:"proto,omitempty"`
	Flags  nat_types.NatConfigFlags `binapi:"nat_config_flags,name=flags" json:"flags,omitempty"`
	SesNum uint32                   `binapi:"u32,name=ses_num" json:"ses_num,omitempty"`
}

func (m *Nat64BibDetails) Reset()               { *m = Nat64BibDetails{} }
func (*Nat64BibDetails) GetMessageName() string { return "nat64_bib_details" }
func (*Nat64BibDetails) GetCrcString() string   { return "43bc3ddf" }
func (*Nat64BibDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64BibDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 * 16 // m.IAddr
	size += 1 * 4  // m.OAddr
	size += 2      // m.IPort
	size += 2      // m.OPort
	size += 4      // m.VrfID
	size += 1      // m.Proto
	size += 1      // m.Flags
	size += 4      // m.SesNum
	return size
}
func (m *Nat64BibDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBytes(m.IAddr[:], 16)
	buf.EncodeBytes(m.OAddr[:], 4)
	buf.EncodeUint16(m.IPort)
	buf.EncodeUint16(m.OPort)
	buf.EncodeUint32(m.VrfID)
	buf.EncodeUint8(m.Proto)
	buf.EncodeUint8(uint8(m.Flags))
	buf.EncodeUint32(m.SesNum)
	return buf.Bytes(), nil
}
func (m *Nat64BibDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	copy(m.IAddr[:], buf.DecodeBytes(16))
	copy(m.OAddr[:], buf.DecodeBytes(4))
	m.IPort = buf.DecodeUint16()
	m.OPort = buf.DecodeUint16()
	m.VrfID = buf.DecodeUint32()
	m.Proto = buf.DecodeUint8()
	m.Flags = nat_types.NatConfigFlags(buf.DecodeUint8())
	m.SesNum = buf.DecodeUint32()
	return nil
}

// Nat64BibDump defines message 'nat64_bib_dump'.
type Nat64BibDump struct {
	Proto uint8 `binapi:"u8,name=proto" json:"proto,omitempty"`
}

func (m *Nat64BibDump) Reset()               { *m = Nat64BibDump{} }
func (*Nat64BibDump) GetMessageName() string { return "nat64_bib_dump" }
func (*Nat64BibDump) GetCrcString() string   { return "cfcb6b75" }
func (*Nat64BibDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64BibDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 // m.Proto
	return size
}
func (m *Nat64BibDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint8(m.Proto)
	return buf.Bytes(), nil
}
func (m *Nat64BibDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Proto = buf.DecodeUint8()
	return nil
}

// Nat64GetTimeouts defines message 'nat64_get_timeouts'.
type Nat64GetTimeouts struct{}

func (m *Nat64GetTimeouts) Reset()               { *m = Nat64GetTimeouts{} }
func (*Nat64GetTimeouts) GetMessageName() string { return "nat64_get_timeouts" }
func (*Nat64GetTimeouts) GetCrcString() string   { return "51077d14" }
func (*Nat64GetTimeouts) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64GetTimeouts) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *Nat64GetTimeouts) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *Nat64GetTimeouts) Unmarshal(b []byte) error {
	return nil
}

// Nat64GetTimeoutsReply defines message 'nat64_get_timeouts_reply'.
type Nat64GetTimeoutsReply struct {
	Retval         int32  `binapi:"i32,name=retval" json:"retval,omitempty"`
	UDP            uint32 `binapi:"u32,name=udp" json:"udp,omitempty"`
	TCPEstablished uint32 `binapi:"u32,name=tcp_established" json:"tcp_established,omitempty"`
	TCPTransitory  uint32 `binapi:"u32,name=tcp_transitory" json:"tcp_transitory,omitempty"`
	ICMP           uint32 `binapi:"u32,name=icmp" json:"icmp,omitempty"`
}

func (m *Nat64GetTimeoutsReply) Reset()               { *m = Nat64GetTimeoutsReply{} }
func (*Nat64GetTimeoutsReply) GetMessageName() string { return "nat64_get_timeouts_reply" }
func (*Nat64GetTimeoutsReply) GetCrcString() string   { return "3c4df4e1" }
func (*Nat64GetTimeoutsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64GetTimeoutsReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.UDP
	size += 4 // m.TCPEstablished
	size += 4 // m.TCPTransitory
	size += 4 // m.ICMP
	return size
}
func (m *Nat64GetTimeoutsReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(m.UDP)
	buf.EncodeUint32(m.TCPEstablished)
	buf.EncodeUint32(m.TCPTransitory)
	buf.EncodeUint32(m.ICMP)
	return buf.Bytes(), nil
}
func (m *Nat64GetTimeoutsReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.UDP = buf.DecodeUint32()
	m.TCPEstablished = buf.DecodeUint32()
	m.TCPTransitory = buf.DecodeUint32()
	m.ICMP = buf.DecodeUint32()
	return nil
}

// Nat64InterfaceDetails defines message 'nat64_interface_details'.
type Nat64InterfaceDetails struct {
	Flags     nat_types.NatConfigFlags       `binapi:"nat_config_flags,name=flags" json:"flags,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *Nat64InterfaceDetails) Reset()               { *m = Nat64InterfaceDetails{} }
func (*Nat64InterfaceDetails) GetMessageName() string { return "nat64_interface_details" }
func (*Nat64InterfaceDetails) GetCrcString() string   { return "5d286289" }
func (*Nat64InterfaceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64InterfaceDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 // m.Flags
	size += 4 // m.SwIfIndex
	return size
}
func (m *Nat64InterfaceDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint8(uint8(m.Flags))
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *Nat64InterfaceDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Flags = nat_types.NatConfigFlags(buf.DecodeUint8())
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Nat64InterfaceDump defines message 'nat64_interface_dump'.
type Nat64InterfaceDump struct{}

func (m *Nat64InterfaceDump) Reset()               { *m = Nat64InterfaceDump{} }
func (*Nat64InterfaceDump) GetMessageName() string { return "nat64_interface_dump" }
func (*Nat64InterfaceDump) GetCrcString() string   { return "51077d14" }
func (*Nat64InterfaceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64InterfaceDump) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *Nat64InterfaceDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *Nat64InterfaceDump) Unmarshal(b []byte) error {
	return nil
}

// Nat64PluginEnableDisable defines message 'nat64_plugin_enable_disable'.
// InProgress: the message form may change in the future versions
type Nat64PluginEnableDisable struct {
	BibBuckets    uint32 `binapi:"u32,name=bib_buckets" json:"bib_buckets,omitempty"`
	BibMemorySize uint32 `binapi:"u32,name=bib_memory_size" json:"bib_memory_size,omitempty"`
	StBuckets     uint32 `binapi:"u32,name=st_buckets" json:"st_buckets,omitempty"`
	StMemorySize  uint32 `binapi:"u32,name=st_memory_size" json:"st_memory_size,omitempty"`
	Enable        bool   `binapi:"bool,name=enable" json:"enable,omitempty"`
}

func (m *Nat64PluginEnableDisable) Reset()               { *m = Nat64PluginEnableDisable{} }
func (*Nat64PluginEnableDisable) GetMessageName() string { return "nat64_plugin_enable_disable" }
func (*Nat64PluginEnableDisable) GetCrcString() string   { return "45948b90" }
func (*Nat64PluginEnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64PluginEnableDisable) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.BibBuckets
	size += 4 // m.BibMemorySize
	size += 4 // m.StBuckets
	size += 4 // m.StMemorySize
	size += 1 // m.Enable
	return size
}
func (m *Nat64PluginEnableDisable) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.BibBuckets)
	buf.EncodeUint32(m.BibMemorySize)
	buf.EncodeUint32(m.StBuckets)
	buf.EncodeUint32(m.StMemorySize)
	buf.EncodeBool(m.Enable)
	return buf.Bytes(), nil
}
func (m *Nat64PluginEnableDisable) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.BibBuckets = buf.DecodeUint32()
	m.BibMemorySize = buf.DecodeUint32()
	m.StBuckets = buf.DecodeUint32()
	m.StMemorySize = buf.DecodeUint32()
	m.Enable = buf.DecodeBool()
	return nil
}

// Nat64PluginEnableDisableReply defines message 'nat64_plugin_enable_disable_reply'.
// InProgress: the message form may change in the future versions
type Nat64PluginEnableDisableReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *Nat64PluginEnableDisableReply) Reset() { *m = Nat64PluginEnableDisableReply{} }
func (*Nat64PluginEnableDisableReply) GetMessageName() string {
	return "nat64_plugin_enable_disable_reply"
}
func (*Nat64PluginEnableDisableReply) GetCrcString() string { return "e8d4e804" }
func (*Nat64PluginEnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64PluginEnableDisableReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *Nat64PluginEnableDisableReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *Nat64PluginEnableDisableReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// Nat64PoolAddrDetails defines message 'nat64_pool_addr_details'.
type Nat64PoolAddrDetails struct {
	Address ip_types.IP4Address `binapi:"ip4_address,name=address" json:"address,omitempty"`
	VrfID   uint32              `binapi:"u32,name=vrf_id" json:"vrf_id,omitempty"`
}

func (m *Nat64PoolAddrDetails) Reset()               { *m = Nat64PoolAddrDetails{} }
func (*Nat64PoolAddrDetails) GetMessageName() string { return "nat64_pool_addr_details" }
func (*Nat64PoolAddrDetails) GetCrcString() string   { return "9bb99cdb" }
func (*Nat64PoolAddrDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64PoolAddrDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 * 4 // m.Address
	size += 4     // m.VrfID
	return size
}
func (m *Nat64PoolAddrDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBytes(m.Address[:], 4)
	buf.EncodeUint32(m.VrfID)
	return buf.Bytes(), nil
}
func (m *Nat64PoolAddrDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	copy(m.Address[:], buf.DecodeBytes(4))
	m.VrfID = buf.DecodeUint32()
	return nil
}

// Nat64PoolAddrDump defines message 'nat64_pool_addr_dump'.
type Nat64PoolAddrDump struct{}

func (m *Nat64PoolAddrDump) Reset()               { *m = Nat64PoolAddrDump{} }
func (*Nat64PoolAddrDump) GetMessageName() string { return "nat64_pool_addr_dump" }
func (*Nat64PoolAddrDump) GetCrcString() string   { return "51077d14" }
func (*Nat64PoolAddrDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64PoolAddrDump) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *Nat64PoolAddrDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *Nat64PoolAddrDump) Unmarshal(b []byte) error {
	return nil
}

// Nat64PrefixDetails defines message 'nat64_prefix_details'.
type Nat64PrefixDetails struct {
	Prefix ip_types.IP6Prefix `binapi:"ip6_prefix,name=prefix" json:"prefix,omitempty"`
	VrfID  uint32             `binapi:"u32,name=vrf_id" json:"vrf_id,omitempty"`
}

func (m *Nat64PrefixDetails) Reset()               { *m = Nat64PrefixDetails{} }
func (*Nat64PrefixDetails) GetMessageName() string { return "nat64_prefix_details" }
func (*Nat64PrefixDetails) GetCrcString() string   { return "20568de3" }
func (*Nat64PrefixDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64PrefixDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 * 16 // m.Prefix.Address
	size += 1      // m.Prefix.Len
	size += 4      // m.VrfID
	return size
}
func (m *Nat64PrefixDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBytes(m.Prefix.Address[:], 16)
	buf.EncodeUint8(m.Prefix.Len)
	buf.EncodeUint32(m.VrfID)
	return buf.Bytes(), nil
}
func (m *Nat64PrefixDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	copy(m.Prefix.Address[:], buf.DecodeBytes(16))
	m.Prefix.Len = buf.DecodeUint8()
	m.VrfID = buf.DecodeUint32()
	return nil
}

// Nat64PrefixDump defines message 'nat64_prefix_dump'.
type Nat64PrefixDump struct{}

func (m *Nat64PrefixDump) Reset()               { *m = Nat64PrefixDump{} }
func (*Nat64PrefixDump) GetMessageName() string { return "nat64_prefix_dump" }
func (*Nat64PrefixDump) GetCrcString() string   { return "51077d14" }
func (*Nat64PrefixDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64PrefixDump) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *Nat64PrefixDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *Nat64PrefixDump) Unmarshal(b []byte) error {
	return nil
}

// Nat64SetTimeouts defines message 'nat64_set_timeouts'.
type Nat64SetTimeouts struct {
	UDP            uint32 `binapi:"u32,name=udp" json:"udp,omitempty"`
	TCPEstablished uint32 `binapi:"u32,name=tcp_established" json:"tcp_established,omitempty"`
	TCPTransitory  uint32 `binapi:"u32,name=tcp_transitory" json:"tcp_transitory,omitempty"`
	ICMP           uint32 `binapi:"u32,name=icmp" json:"icmp,omitempty"`
}

func (m *Nat64SetTimeouts) Reset()               { *m = Nat64SetTimeouts{} }
func (*Nat64SetTimeouts) GetMessageName() string { return "nat64_set_timeouts" }
func (*Nat64SetTimeouts) GetCrcString() string   { return "d4746b16" }
func (*Nat64SetTimeouts) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64SetTimeouts) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.UDP
	size += 4 // m.TCPEstablished
	size += 4 // m.TCPTransitory
	size += 4 // m.ICMP
	return size
}
func (m *Nat64SetTimeouts) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.UDP)
	buf.EncodeUint32(m.TCPEstablished)
	buf.EncodeUint32(m.TCPTransitory)
	buf.EncodeUint32(m.ICMP)
	return buf.Bytes(), nil
}
func (m *Nat64SetTimeouts) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.UDP = buf.DecodeUint32()
	m.TCPEstablished = buf.DecodeUint32()
	m.TCPTransitory = buf.DecodeUint32()
	m.ICMP = buf.DecodeUint32()
	return nil
}

// Nat64SetTimeoutsReply defines message 'nat64_set_timeouts_reply'.
type Nat64SetTimeoutsReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *Nat64SetTimeoutsReply) Reset()               { *m = Nat64SetTimeoutsReply{} }
func (*Nat64SetTimeoutsReply) GetMessageName() string { return "nat64_set_timeouts_reply" }
func (*Nat64SetTimeoutsReply) GetCrcString() string   { return "e8d4e804" }
func (*Nat64SetTimeoutsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64SetTimeoutsReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *Nat64SetTimeoutsReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *Nat64SetTimeoutsReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// Nat64StDetails defines message 'nat64_st_details'.
type Nat64StDetails struct {
	IlAddr ip_types.IP6Address `binapi:"ip6_address,name=il_addr" json:"il_addr,omitempty"`
	OlAddr ip_types.IP4Address `binapi:"ip4_address,name=ol_addr" json:"ol_addr,omitempty"`
	IlPort uint16              `binapi:"u16,name=il_port" json:"il_port,omitempty"`
	OlPort uint16              `binapi:"u16,name=ol_port" json:"ol_port,omitempty"`
	IrAddr ip_types.IP6Address `binapi:"ip6_address,name=ir_addr" json:"ir_addr,omitempty"`
	OrAddr ip_types.IP4Address `binapi:"ip4_address,name=or_addr" json:"or_addr,omitempty"`
	RPort  uint16              `binapi:"u16,name=r_port" json:"r_port,omitempty"`
	VrfID  uint32              `binapi:"u32,name=vrf_id" json:"vrf_id,omitempty"`
	Proto  uint8               `binapi:"u8,name=proto" json:"proto,omitempty"`
}

func (m *Nat64StDetails) Reset()               { *m = Nat64StDetails{} }
func (*Nat64StDetails) GetMessageName() string { return "nat64_st_details" }
func (*Nat64StDetails) GetCrcString() string   { return "dd3361ed" }
func (*Nat64StDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Nat64StDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 * 16 // m.IlAddr
	size += 1 * 4  // m.OlAddr
	size += 2      // m.IlPort
	size += 2      // m.OlPort
	size += 1 * 16 // m.IrAddr
	size += 1 * 4  // m.OrAddr
	size += 2      // m.RPort
	size += 4      // m.VrfID
	size += 1      // m.Proto
	return size
}
func (m *Nat64StDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBytes(m.IlAddr[:], 16)
	buf.EncodeBytes(m.OlAddr[:], 4)
	buf.EncodeUint16(m.IlPort)
	buf.EncodeUint16(m.OlPort)
	buf.EncodeBytes(m.IrAddr[:], 16)
	buf.EncodeBytes(m.OrAddr[:], 4)
	buf.EncodeUint16(m.RPort)
	buf.EncodeUint32(m.VrfID)
	buf.EncodeUint8(m.Proto)
	return buf.Bytes(), nil
}
func (m *Nat64StDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	copy(m.IlAddr[:], buf.DecodeBytes(16))
	copy(m.OlAddr[:], buf.DecodeBytes(4))
	m.IlPort = buf.DecodeUint16()
	m.OlPort = buf.DecodeUint16()
	copy(m.IrAddr[:], buf.DecodeBytes(16))
	copy(m.OrAddr[:], buf.DecodeBytes(4))
	m.RPort = buf.DecodeUint16()
	m.VrfID = buf.DecodeUint32()
	m.Proto = buf.DecodeUint8()
	return nil
}

// Nat64StDump defines message 'nat64_st_dump'.
type Nat64StDump struct {
	Proto uint8 `binapi:"u8,name=proto" json:"proto,omitempty"`
}

func (m *Nat64StDump) Reset()               { *m = Nat64StDump{} }
func (*Nat64StDump) GetMessageName() string { return "nat64_st_dump" }
func (*Nat64StDump) GetCrcString() string   { return "cfcb6b75" }
func (*Nat64StDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Nat64StDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 // m.Proto
	return size
}
func (m *Nat64StDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint8(m.Proto)
	return buf.Bytes(), nil
}
func (m *Nat64StDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Proto = buf.DecodeUint8()
	return nil
}

func init() { file_nat64_binapi_init() }
func file_nat64_binapi_init() {
	api.RegisterMessage((*Nat64AddDelInterface)(nil), "nat64_add_del_interface_f3699b83")
	api.RegisterMessage((*Nat64AddDelInterfaceAddr)(nil), "nat64_add_del_interface_addr_47d6e753")
	api.RegisterMessage((*Nat64AddDelInterfaceAddrReply)(nil), "nat64_add_del_interface_addr_reply_e8d4e804")
	api.RegisterMessage((*Nat64AddDelInterfaceReply)(nil), "nat64_add_del_interface_reply_e8d4e804")
	api.RegisterMessage((*Nat64AddDelPoolAddrRange)(nil), "nat64_add_del_pool_addr_range_a3b944e3")
	api.RegisterMessage((*Nat64AddDelPoolAddrRangeReply)(nil), "nat64_add_del_pool_addr_range_reply_e8d4e804")
	api.RegisterMessage((*Nat64AddDelPrefix)(nil), "nat64_add_del_prefix_727b2f4c")
	api.RegisterMessage((*Nat64AddDelPrefixReply)(nil), "nat64_add_del_prefix_reply_e8d4e804")
	api.RegisterMessage((*Nat64AddDelStaticBib)(nil), "nat64_add_del_static_bib_1c404de5")
	api.RegisterMessage((*Nat64AddDelStaticBibReply)(nil), "nat64_add_del_static_bib_reply_e8d4e804")
	api.RegisterMessage((*Nat64BibDetails)(nil), "nat64_bib_details_43bc3ddf")
	api.RegisterMessage((*Nat64BibDump)(nil), "nat64_bib_dump_cfcb6b75")
	api.RegisterMessage((*Nat64GetTimeouts)(nil), "nat64_get_timeouts_51077d14")
	api.RegisterMessage((*Nat64GetTimeoutsReply)(nil), "nat64_get_timeouts_reply_3c4df4e1")
	api.RegisterMessage((*Nat64InterfaceDetails)(nil), "nat64_interface_details_5d286289")
	api.RegisterMessage((*Nat64InterfaceDump)(nil), "nat64_interface_dump_51077d14")
	api.RegisterMessage((*Nat64PluginEnableDisable)(nil), "nat64_plugin_enable_disable_45948b90")
	api.RegisterMessage((*Nat64PluginEnableDisableReply)(nil), "nat64_plugin_enable_disable_reply_e8d4e804")
	api.RegisterMessage((*Nat64PoolAddrDetails)(nil), "nat64_pool_addr_details_9bb99cdb")
	api.RegisterMessage((*Nat64PoolAddrDump)(nil), "nat64_pool_addr_dump_51077d14")
	api.RegisterMessage((*Nat64PrefixDetails)(nil), "nat64_prefix_details_20568de3")
	api.RegisterMessage((*Nat64PrefixDump)(nil), "nat64_prefix_dump_51077d14")
	api.RegisterMessage((*Nat64SetTimeouts)(nil), "nat64_set_timeouts_d4746b16")
	api.RegisterMessage((*Nat64SetTimeoutsReply)(nil), "nat64_set_timeouts_reply_e8d4e804")
	api.RegisterMessage((*Nat64StDetails)(nil), "nat64_st_details_dd3361ed")
	api.RegisterMessage((*Nat64StDump)(nil), "nat64_st_dump_cfcb6b75")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*Nat64AddDelInterface)(nil),
		(*Nat64AddDelInterfaceAddr)(nil),
		(*Nat64AddDelInterfaceAddrReply)(nil),
		(*Nat64AddDelInterfaceReply)(nil),
		(*Nat64AddDelPoolAddrRange)(nil),
		(*Nat64AddDelPoolAddrRangeReply)(nil),
		(*Nat64AddDelPrefix)(nil),
		(*Nat64AddDelPrefixReply)(nil),
		(*Nat64AddDelStaticBib)(nil),
		(*Nat64AddDelStaticBibReply)(nil),
		(*Nat64BibDetails)(nil),
		(*Nat64BibDump)(nil),
		(*Nat64GetTimeouts)(nil),
		(*Nat64GetTimeoutsReply)(nil),
		(*Nat64InterfaceDetails)(nil),
		(*Nat64InterfaceDump)(nil),
		(*Nat64PluginEnableDisable)(nil),
		(*Nat64PluginEnableDisableReply)(nil),
		(*Nat64PoolAddrDetails)(nil),
		(*Nat64PoolAddrDump)(nil),
		(*Nat64PrefixDetails)(nil),
		(*Nat64PrefixDump)(nil),
		(*Nat64SetTimeouts)(nil),
		(*Nat64SetTimeoutsReply)(nil),
		(*Nat64StDetails)(nil),
		(*Nat64StDump)(nil),
	}
}
