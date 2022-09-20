// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.6.0
//  VPP:              22.06-release
// source: /usr/share/vpp/api/plugins/mss_clamp.api.json

// Package mss_clamp contains generated bindings for API file mss_clamp.api.
//
// Contents:
// -  1 enum
// -  5 messages
package mss_clamp

import (
	"strconv"

	api "go.fd.io/govpp/api"
	interface_types "go.fd.io/govpp/binapi/interface_types"
	codec "go.fd.io/govpp/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "mss_clamp"
	APIVersion = "1.0.0"
	VersionCrc = 0xea8186c0
)

// MssClampDir defines enum 'mss_clamp_dir'.
type MssClampDir uint8

const (
	MSS_CLAMP_DIR_NONE MssClampDir = 0
	MSS_CLAMP_DIR_RX   MssClampDir = 1
	MSS_CLAMP_DIR_TX   MssClampDir = 2
)

var (
	MssClampDir_name = map[uint8]string{
		0: "MSS_CLAMP_DIR_NONE",
		1: "MSS_CLAMP_DIR_RX",
		2: "MSS_CLAMP_DIR_TX",
	}
	MssClampDir_value = map[string]uint8{
		"MSS_CLAMP_DIR_NONE": 0,
		"MSS_CLAMP_DIR_RX":   1,
		"MSS_CLAMP_DIR_TX":   2,
	}
)

func (x MssClampDir) String() string {
	s, ok := MssClampDir_name[uint8(x)]
	if ok {
		return s
	}
	str := func(n uint8) string {
		s, ok := MssClampDir_name[uint8(n)]
		if ok {
			return s
		}
		return "MssClampDir(" + strconv.Itoa(int(n)) + ")"
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

// MssClampDetails defines message 'mss_clamp_details'.
type MssClampDetails struct {
	SwIfIndex     interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IPv4Mss       uint16                         `binapi:"u16,name=ipv4_mss" json:"ipv4_mss,omitempty"`
	IPv6Mss       uint16                         `binapi:"u16,name=ipv6_mss" json:"ipv6_mss,omitempty"`
	IPv4Direction MssClampDir                    `binapi:"mss_clamp_dir,name=ipv4_direction" json:"ipv4_direction,omitempty"`
	IPv6Direction MssClampDir                    `binapi:"mss_clamp_dir,name=ipv6_direction" json:"ipv6_direction,omitempty"`
}

func (m *MssClampDetails) Reset()               { *m = MssClampDetails{} }
func (*MssClampDetails) GetMessageName() string { return "mss_clamp_details" }
func (*MssClampDetails) GetCrcString() string   { return "d3a4de61" }
func (*MssClampDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *MssClampDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	size += 2 // m.IPv4Mss
	size += 2 // m.IPv6Mss
	size += 1 // m.IPv4Direction
	size += 1 // m.IPv6Direction
	return size
}
func (m *MssClampDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint16(m.IPv4Mss)
	buf.EncodeUint16(m.IPv6Mss)
	buf.EncodeUint8(uint8(m.IPv4Direction))
	buf.EncodeUint8(uint8(m.IPv6Direction))
	return buf.Bytes(), nil
}
func (m *MssClampDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IPv4Mss = buf.DecodeUint16()
	m.IPv6Mss = buf.DecodeUint16()
	m.IPv4Direction = MssClampDir(buf.DecodeUint8())
	m.IPv6Direction = MssClampDir(buf.DecodeUint8())
	return nil
}

// MssClampEnableDisable defines message 'mss_clamp_enable_disable'.
type MssClampEnableDisable struct {
	SwIfIndex     interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IPv4Mss       uint16                         `binapi:"u16,name=ipv4_mss" json:"ipv4_mss,omitempty"`
	IPv6Mss       uint16                         `binapi:"u16,name=ipv6_mss" json:"ipv6_mss,omitempty"`
	IPv4Direction MssClampDir                    `binapi:"mss_clamp_dir,name=ipv4_direction" json:"ipv4_direction,omitempty"`
	IPv6Direction MssClampDir                    `binapi:"mss_clamp_dir,name=ipv6_direction" json:"ipv6_direction,omitempty"`
}

func (m *MssClampEnableDisable) Reset()               { *m = MssClampEnableDisable{} }
func (*MssClampEnableDisable) GetMessageName() string { return "mss_clamp_enable_disable" }
func (*MssClampEnableDisable) GetCrcString() string   { return "d31b44e3" }
func (*MssClampEnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *MssClampEnableDisable) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	size += 2 // m.IPv4Mss
	size += 2 // m.IPv6Mss
	size += 1 // m.IPv4Direction
	size += 1 // m.IPv6Direction
	return size
}
func (m *MssClampEnableDisable) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint16(m.IPv4Mss)
	buf.EncodeUint16(m.IPv6Mss)
	buf.EncodeUint8(uint8(m.IPv4Direction))
	buf.EncodeUint8(uint8(m.IPv6Direction))
	return buf.Bytes(), nil
}
func (m *MssClampEnableDisable) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IPv4Mss = buf.DecodeUint16()
	m.IPv6Mss = buf.DecodeUint16()
	m.IPv4Direction = MssClampDir(buf.DecodeUint8())
	m.IPv6Direction = MssClampDir(buf.DecodeUint8())
	return nil
}

// MssClampEnableDisableReply defines message 'mss_clamp_enable_disable_reply'.
type MssClampEnableDisableReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *MssClampEnableDisableReply) Reset()               { *m = MssClampEnableDisableReply{} }
func (*MssClampEnableDisableReply) GetMessageName() string { return "mss_clamp_enable_disable_reply" }
func (*MssClampEnableDisableReply) GetCrcString() string   { return "e8d4e804" }
func (*MssClampEnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *MssClampEnableDisableReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *MssClampEnableDisableReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *MssClampEnableDisableReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// MssClampGet defines message 'mss_clamp_get'.
type MssClampGet struct {
	Cursor    uint32                         `binapi:"u32,name=cursor" json:"cursor,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *MssClampGet) Reset()               { *m = MssClampGet{} }
func (*MssClampGet) GetMessageName() string { return "mss_clamp_get" }
func (*MssClampGet) GetCrcString() string   { return "47250981" }
func (*MssClampGet) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *MssClampGet) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Cursor
	size += 4 // m.SwIfIndex
	return size
}
func (m *MssClampGet) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.Cursor)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *MssClampGet) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Cursor = buf.DecodeUint32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// MssClampGetReply defines message 'mss_clamp_get_reply'.
type MssClampGetReply struct {
	Retval int32  `binapi:"i32,name=retval" json:"retval,omitempty"`
	Cursor uint32 `binapi:"u32,name=cursor" json:"cursor,omitempty"`
}

func (m *MssClampGetReply) Reset()               { *m = MssClampGetReply{} }
func (*MssClampGetReply) GetMessageName() string { return "mss_clamp_get_reply" }
func (*MssClampGetReply) GetCrcString() string   { return "53b48f5d" }
func (*MssClampGetReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *MssClampGetReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.Cursor
	return size
}
func (m *MssClampGetReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(m.Cursor)
	return buf.Bytes(), nil
}
func (m *MssClampGetReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.Cursor = buf.DecodeUint32()
	return nil
}

func init() { file_mss_clamp_binapi_init() }
func file_mss_clamp_binapi_init() {
	api.RegisterMessage((*MssClampDetails)(nil), "mss_clamp_details_d3a4de61")
	api.RegisterMessage((*MssClampEnableDisable)(nil), "mss_clamp_enable_disable_d31b44e3")
	api.RegisterMessage((*MssClampEnableDisableReply)(nil), "mss_clamp_enable_disable_reply_e8d4e804")
	api.RegisterMessage((*MssClampGet)(nil), "mss_clamp_get_47250981")
	api.RegisterMessage((*MssClampGetReply)(nil), "mss_clamp_get_reply_53b48f5d")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*MssClampDetails)(nil),
		(*MssClampEnableDisable)(nil),
		(*MssClampEnableDisableReply)(nil),
		(*MssClampGet)(nil),
		(*MssClampGetReply)(nil),
	}
}
