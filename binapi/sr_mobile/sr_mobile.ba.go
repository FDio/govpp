// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.11.0
//  VPP:              24.06-release
// source: plugins/sr_mobile.api.json

// Package sr_mobile contains generated bindings for API file sr_mobile.api.
//
// Contents:
// -  4 messages
package sr_mobile

import (
	api "go.fd.io/govpp/api"
	_ "go.fd.io/govpp/binapi/interface_types"
	ip_types "go.fd.io/govpp/binapi/ip_types"
	_ "go.fd.io/govpp/binapi/sr"
	sr_mobile_types "go.fd.io/govpp/binapi/sr_mobile_types"
	_ "go.fd.io/govpp/binapi/sr_types"
	codec "go.fd.io/govpp/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "sr_mobile"
	APIVersion = "0.1.0"
	VersionCrc = 0x2a93fd77
)

// IPv6 SR for Mobile LocalSID add/del request
//   - is_del - Boolean of whether its a delete instruction
//   - localsid_prefix - IPv6 address of the localsid
//   - behavior - the behavior of the SR policy.
//   - fib_table - FIB table in which we should install the localsid entry
//   - local_fib_table - lookup and forward GTP-U packet based on outer IP destination address. optional
//   - drop_in - that reconverts to GTPv1 mode. optional
//   - nhtype - next-header type. optional.
//   - sr_prefix - v6 src ip encoding prefix.optional.
//   - v4src_position - bit position where IPv4 src address embedded. optional.
//
// SrMobileLocalsidAddDel defines message 'sr_mobile_localsid_add_del'.
type SrMobileLocalsidAddDel struct {
	IsDel          bool                           `binapi:"bool,name=is_del,default=false" json:"is_del,omitempty"`
	LocalsidPrefix ip_types.IP6Prefix             `binapi:"ip6_prefix,name=localsid_prefix" json:"localsid_prefix,omitempty"`
	Behavior       string                         `binapi:"string[64],name=behavior" json:"behavior,omitempty"`
	FibTable       uint32                         `binapi:"u32,name=fib_table" json:"fib_table,omitempty"`
	LocalFibTable  uint32                         `binapi:"u32,name=local_fib_table" json:"local_fib_table,omitempty"`
	DropIn         bool                           `binapi:"bool,name=drop_in" json:"drop_in,omitempty"`
	Nhtype         sr_mobile_types.SrMobileNhtype `binapi:"sr_mobile_nhtype,name=nhtype" json:"nhtype,omitempty"`
	SrPrefix       ip_types.IP6Prefix             `binapi:"ip6_prefix,name=sr_prefix" json:"sr_prefix,omitempty"`
	V4srcAddr      ip_types.IP4Address            `binapi:"ip4_address,name=v4src_addr" json:"v4src_addr,omitempty"`
	V4srcPosition  uint32                         `binapi:"u32,name=v4src_position" json:"v4src_position,omitempty"`
}

func (m *SrMobileLocalsidAddDel) Reset()               { *m = SrMobileLocalsidAddDel{} }
func (*SrMobileLocalsidAddDel) GetMessageName() string { return "sr_mobile_localsid_add_del" }
func (*SrMobileLocalsidAddDel) GetCrcString() string   { return "b85a7ed7" }
func (*SrMobileLocalsidAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SrMobileLocalsidAddDel) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1      // m.IsDel
	size += 1 * 16 // m.LocalsidPrefix.Address
	size += 1      // m.LocalsidPrefix.Len
	size += 64     // m.Behavior
	size += 4      // m.FibTable
	size += 4      // m.LocalFibTable
	size += 1      // m.DropIn
	size += 1      // m.Nhtype
	size += 1 * 16 // m.SrPrefix.Address
	size += 1      // m.SrPrefix.Len
	size += 1 * 4  // m.V4srcAddr
	size += 4      // m.V4srcPosition
	return size
}
func (m *SrMobileLocalsidAddDel) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsDel)
	buf.EncodeBytes(m.LocalsidPrefix.Address[:], 16)
	buf.EncodeUint8(m.LocalsidPrefix.Len)
	buf.EncodeString(m.Behavior, 64)
	buf.EncodeUint32(m.FibTable)
	buf.EncodeUint32(m.LocalFibTable)
	buf.EncodeBool(m.DropIn)
	buf.EncodeUint8(uint8(m.Nhtype))
	buf.EncodeBytes(m.SrPrefix.Address[:], 16)
	buf.EncodeUint8(m.SrPrefix.Len)
	buf.EncodeBytes(m.V4srcAddr[:], 4)
	buf.EncodeUint32(m.V4srcPosition)
	return buf.Bytes(), nil
}
func (m *SrMobileLocalsidAddDel) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsDel = buf.DecodeBool()
	copy(m.LocalsidPrefix.Address[:], buf.DecodeBytes(16))
	m.LocalsidPrefix.Len = buf.DecodeUint8()
	m.Behavior = buf.DecodeString(64)
	m.FibTable = buf.DecodeUint32()
	m.LocalFibTable = buf.DecodeUint32()
	m.DropIn = buf.DecodeBool()
	m.Nhtype = sr_mobile_types.SrMobileNhtype(buf.DecodeUint8())
	copy(m.SrPrefix.Address[:], buf.DecodeBytes(16))
	m.SrPrefix.Len = buf.DecodeUint8()
	copy(m.V4srcAddr[:], buf.DecodeBytes(4))
	m.V4srcPosition = buf.DecodeUint32()
	return nil
}

// SrMobileLocalsidAddDelReply defines message 'sr_mobile_localsid_add_del_reply'.
type SrMobileLocalsidAddDelReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *SrMobileLocalsidAddDelReply) Reset() { *m = SrMobileLocalsidAddDelReply{} }
func (*SrMobileLocalsidAddDelReply) GetMessageName() string {
	return "sr_mobile_localsid_add_del_reply"
}
func (*SrMobileLocalsidAddDelReply) GetCrcString() string { return "e8d4e804" }
func (*SrMobileLocalsidAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SrMobileLocalsidAddDelReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *SrMobileLocalsidAddDelReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *SrMobileLocalsidAddDelReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// IPv6 SR for Mobile policy add
//   - bsid - the bindingSID of the SR Policy
//   - sr_prefix - v6 dst ip encoding prefix. optional
//   - v6src_position - v6 src prefix. optional
//   - behavior - the behavior of the SR policy.
//   - fib_table - the VRF where to install the FIB entry for the BSID
//   - encap_src is a encaps IPv6 source addr. optional
//   - local_fib_table - lookup and forward GTP-U packet based on outer IP destination address. optional
//   - drop_in - that reconverts to GTPv1 mode. optional
//   - nhtype - next-header type.
//
// SrMobilePolicyAdd defines message 'sr_mobile_policy_add'.
type SrMobilePolicyAdd struct {
	BsidAddr      ip_types.IP6Address            `binapi:"ip6_address,name=bsid_addr" json:"bsid_addr,omitempty"`
	SrPrefix      ip_types.IP6Prefix             `binapi:"ip6_prefix,name=sr_prefix" json:"sr_prefix,omitempty"`
	V6srcPrefix   ip_types.IP6Prefix             `binapi:"ip6_prefix,name=v6src_prefix" json:"v6src_prefix,omitempty"`
	Behavior      string                         `binapi:"string[64],name=behavior" json:"behavior,omitempty"`
	FibTable      uint32                         `binapi:"u32,name=fib_table" json:"fib_table,omitempty"`
	LocalFibTable uint32                         `binapi:"u32,name=local_fib_table" json:"local_fib_table,omitempty"`
	EncapSrc      ip_types.IP6Address            `binapi:"ip6_address,name=encap_src" json:"encap_src,omitempty"`
	DropIn        bool                           `binapi:"bool,name=drop_in" json:"drop_in,omitempty"`
	Nhtype        sr_mobile_types.SrMobileNhtype `binapi:"sr_mobile_nhtype,name=nhtype" json:"nhtype,omitempty"`
}

func (m *SrMobilePolicyAdd) Reset()               { *m = SrMobilePolicyAdd{} }
func (*SrMobilePolicyAdd) GetMessageName() string { return "sr_mobile_policy_add" }
func (*SrMobilePolicyAdd) GetCrcString() string   { return "8f051658" }
func (*SrMobilePolicyAdd) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SrMobilePolicyAdd) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 * 16 // m.BsidAddr
	size += 1 * 16 // m.SrPrefix.Address
	size += 1      // m.SrPrefix.Len
	size += 1 * 16 // m.V6srcPrefix.Address
	size += 1      // m.V6srcPrefix.Len
	size += 64     // m.Behavior
	size += 4      // m.FibTable
	size += 4      // m.LocalFibTable
	size += 1 * 16 // m.EncapSrc
	size += 1      // m.DropIn
	size += 1      // m.Nhtype
	return size
}
func (m *SrMobilePolicyAdd) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBytes(m.BsidAddr[:], 16)
	buf.EncodeBytes(m.SrPrefix.Address[:], 16)
	buf.EncodeUint8(m.SrPrefix.Len)
	buf.EncodeBytes(m.V6srcPrefix.Address[:], 16)
	buf.EncodeUint8(m.V6srcPrefix.Len)
	buf.EncodeString(m.Behavior, 64)
	buf.EncodeUint32(m.FibTable)
	buf.EncodeUint32(m.LocalFibTable)
	buf.EncodeBytes(m.EncapSrc[:], 16)
	buf.EncodeBool(m.DropIn)
	buf.EncodeUint8(uint8(m.Nhtype))
	return buf.Bytes(), nil
}
func (m *SrMobilePolicyAdd) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	copy(m.BsidAddr[:], buf.DecodeBytes(16))
	copy(m.SrPrefix.Address[:], buf.DecodeBytes(16))
	m.SrPrefix.Len = buf.DecodeUint8()
	copy(m.V6srcPrefix.Address[:], buf.DecodeBytes(16))
	m.V6srcPrefix.Len = buf.DecodeUint8()
	m.Behavior = buf.DecodeString(64)
	m.FibTable = buf.DecodeUint32()
	m.LocalFibTable = buf.DecodeUint32()
	copy(m.EncapSrc[:], buf.DecodeBytes(16))
	m.DropIn = buf.DecodeBool()
	m.Nhtype = sr_mobile_types.SrMobileNhtype(buf.DecodeUint8())
	return nil
}

// SrMobilePolicyAddReply defines message 'sr_mobile_policy_add_reply'.
type SrMobilePolicyAddReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *SrMobilePolicyAddReply) Reset()               { *m = SrMobilePolicyAddReply{} }
func (*SrMobilePolicyAddReply) GetMessageName() string { return "sr_mobile_policy_add_reply" }
func (*SrMobilePolicyAddReply) GetCrcString() string   { return "e8d4e804" }
func (*SrMobilePolicyAddReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SrMobilePolicyAddReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *SrMobilePolicyAddReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *SrMobilePolicyAddReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

func init() { file_sr_mobile_binapi_init() }
func file_sr_mobile_binapi_init() {
	api.RegisterMessage((*SrMobileLocalsidAddDel)(nil), "sr_mobile_localsid_add_del_b85a7ed7")
	api.RegisterMessage((*SrMobileLocalsidAddDelReply)(nil), "sr_mobile_localsid_add_del_reply_e8d4e804")
	api.RegisterMessage((*SrMobilePolicyAdd)(nil), "sr_mobile_policy_add_8f051658")
	api.RegisterMessage((*SrMobilePolicyAddReply)(nil), "sr_mobile_policy_add_reply_e8d4e804")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*SrMobileLocalsidAddDel)(nil),
		(*SrMobileLocalsidAddDelReply)(nil),
		(*SrMobilePolicyAdd)(nil),
		(*SrMobilePolicyAddReply)(nil),
	}
}
