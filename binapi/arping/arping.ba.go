// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.11.0
//  VPP:              25.02-release
// source: plugins/arping.api.json

// Package arping contains generated bindings for API file arping.api.
//
// Contents:
// -  4 messages
package arping

import (
	api "go.fd.io/govpp/api"
	ethernet_types "go.fd.io/govpp/binapi/ethernet_types"
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
	APIFile    = "arping"
	APIVersion = "1.0.0"
	VersionCrc = 0x8b2c8f39
)

// - client_index - opaque cookie to identify the sender
//   - address - address to send arp request or gratuitous arp.
//   - sw_if_index - interface to send
//   - repeat - number of packets to send
//   - interval - if more than 1 packet is sent, the delay between send
//   - is_garp - is garp or arp request
//
// Arping defines message 'arping'.
type Arping struct {
	Address   ip_types.Address               `binapi:"address,name=address" json:"address,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IsGarp    bool                           `binapi:"bool,name=is_garp" json:"is_garp,omitempty"`
	Repeat    uint32                         `binapi:"u32,name=repeat,default=1" json:"repeat,omitempty"`
	Interval  float64                        `binapi:"f64,name=interval,default=1" json:"interval,omitempty"`
}

func (m *Arping) Reset()               { *m = Arping{} }
func (*Arping) GetMessageName() string { return "arping" }
func (*Arping) GetCrcString() string   { return "48817482" }
func (*Arping) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Arping) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1      // m.Address.Af
	size += 1 * 16 // m.Address.Un
	size += 4      // m.SwIfIndex
	size += 1      // m.IsGarp
	size += 4      // m.Repeat
	size += 8      // m.Interval
	return size
}
func (m *Arping) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint8(uint8(m.Address.Af))
	buf.EncodeBytes(m.Address.Un.XXX_UnionData[:], 16)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeBool(m.IsGarp)
	buf.EncodeUint32(m.Repeat)
	buf.EncodeFloat64(m.Interval)
	return buf.Bytes(), nil
}
func (m *Arping) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Address.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.Address.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IsGarp = buf.DecodeBool()
	m.Repeat = buf.DecodeUint32()
	m.Interval = buf.DecodeFloat64()
	return nil
}

// /*
//   - Address Conflict Detection
//
// ArpingAcd defines message 'arping_acd'.
type ArpingAcd struct {
	Address   ip_types.Address               `binapi:"address,name=address" json:"address,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IsGarp    bool                           `binapi:"bool,name=is_garp" json:"is_garp,omitempty"`
	Repeat    uint32                         `binapi:"u32,name=repeat,default=1" json:"repeat,omitempty"`
	Interval  float64                        `binapi:"f64,name=interval,default=1" json:"interval,omitempty"`
}

func (m *ArpingAcd) Reset()               { *m = ArpingAcd{} }
func (*ArpingAcd) GetMessageName() string { return "arping_acd" }
func (*ArpingAcd) GetCrcString() string   { return "48817482" }
func (*ArpingAcd) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *ArpingAcd) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1      // m.Address.Af
	size += 1 * 16 // m.Address.Un
	size += 4      // m.SwIfIndex
	size += 1      // m.IsGarp
	size += 4      // m.Repeat
	size += 8      // m.Interval
	return size
}
func (m *ArpingAcd) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint8(uint8(m.Address.Af))
	buf.EncodeBytes(m.Address.Un.XXX_UnionData[:], 16)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeBool(m.IsGarp)
	buf.EncodeUint32(m.Repeat)
	buf.EncodeFloat64(m.Interval)
	return buf.Bytes(), nil
}
func (m *ArpingAcd) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Address.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.Address.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IsGarp = buf.DecodeBool()
	m.Repeat = buf.DecodeUint32()
	m.Interval = buf.DecodeFloat64()
	return nil
}

// ArpingAcdReply defines message 'arping_acd_reply'.
type ArpingAcdReply struct {
	Retval     int32                     `binapi:"i32,name=retval" json:"retval,omitempty"`
	ReplyCount uint32                    `binapi:"u32,name=reply_count" json:"reply_count,omitempty"`
	MacAddress ethernet_types.MacAddress `binapi:"mac_address,name=mac_address" json:"mac_address,omitempty"`
}

func (m *ArpingAcdReply) Reset()               { *m = ArpingAcdReply{} }
func (*ArpingAcdReply) GetMessageName() string { return "arping_acd_reply" }
func (*ArpingAcdReply) GetCrcString() string   { return "e08c3b05" }
func (*ArpingAcdReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *ArpingAcdReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4     // m.Retval
	size += 4     // m.ReplyCount
	size += 1 * 6 // m.MacAddress
	return size
}
func (m *ArpingAcdReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(m.ReplyCount)
	buf.EncodeBytes(m.MacAddress[:], 6)
	return buf.Bytes(), nil
}
func (m *ArpingAcdReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.ReplyCount = buf.DecodeUint32()
	copy(m.MacAddress[:], buf.DecodeBytes(6))
	return nil
}

// - context - sender context, to match reply w/ request
//   - retval - return value for request
//     @reply_count - return value for reply count
//
// ArpingReply defines message 'arping_reply'.
type ArpingReply struct {
	Retval     int32  `binapi:"i32,name=retval" json:"retval,omitempty"`
	ReplyCount uint32 `binapi:"u32,name=reply_count" json:"reply_count,omitempty"`
}

func (m *ArpingReply) Reset()               { *m = ArpingReply{} }
func (*ArpingReply) GetMessageName() string { return "arping_reply" }
func (*ArpingReply) GetCrcString() string   { return "bb9d1cbd" }
func (*ArpingReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *ArpingReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.ReplyCount
	return size
}
func (m *ArpingReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(m.ReplyCount)
	return buf.Bytes(), nil
}
func (m *ArpingReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.ReplyCount = buf.DecodeUint32()
	return nil
}

func init() { file_arping_binapi_init() }
func file_arping_binapi_init() {
	api.RegisterMessage((*Arping)(nil), "arping_48817482")
	api.RegisterMessage((*ArpingAcd)(nil), "arping_acd_48817482")
	api.RegisterMessage((*ArpingAcdReply)(nil), "arping_acd_reply_e08c3b05")
	api.RegisterMessage((*ArpingReply)(nil), "arping_reply_bb9d1cbd")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*Arping)(nil),
		(*ArpingAcd)(nil),
		(*ArpingAcdReply)(nil),
		(*ArpingReply)(nil),
	}
}
