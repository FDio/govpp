// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.11.0
//  VPP:              24.06-release
// source: plugins/urpf.api.json

// Package urpf contains generated bindings for API file urpf.api.
//
// Contents:
// -  1 enum
// -  6 messages
package urpf

import (
	"strconv"

	api "go.fd.io/govpp/api"
	_ "go.fd.io/govpp/binapi/fib_types"
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
	APIFile    = "urpf"
	APIVersion = "1.0.0"
	VersionCrc = 0x88759016
)

// UrpfMode defines enum 'urpf_mode'.
type UrpfMode uint8

const (
	URPF_API_MODE_OFF    UrpfMode = 0
	URPF_API_MODE_LOOSE  UrpfMode = 1
	URPF_API_MODE_STRICT UrpfMode = 2
)

var (
	UrpfMode_name = map[uint8]string{
		0: "URPF_API_MODE_OFF",
		1: "URPF_API_MODE_LOOSE",
		2: "URPF_API_MODE_STRICT",
	}
	UrpfMode_value = map[string]uint8{
		"URPF_API_MODE_OFF":    0,
		"URPF_API_MODE_LOOSE":  1,
		"URPF_API_MODE_STRICT": 2,
	}
)

func (x UrpfMode) String() string {
	s, ok := UrpfMode_name[uint8(x)]
	if ok {
		return s
	}
	return "UrpfMode(" + strconv.Itoa(int(x)) + ")"
}

// @brief uRPF enabled interface details
// UrpfInterfaceDetails defines message 'urpf_interface_details'.
type UrpfInterfaceDetails struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IsInput   bool                           `binapi:"bool,name=is_input" json:"is_input,omitempty"`
	Mode      UrpfMode                       `binapi:"urpf_mode,name=mode" json:"mode,omitempty"`
	Af        ip_types.AddressFamily         `binapi:"address_family,name=af" json:"af,omitempty"`
	TableID   uint32                         `binapi:"u32,name=table_id" json:"table_id,omitempty"`
}

func (m *UrpfInterfaceDetails) Reset()               { *m = UrpfInterfaceDetails{} }
func (*UrpfInterfaceDetails) GetMessageName() string { return "urpf_interface_details" }
func (*UrpfInterfaceDetails) GetCrcString() string   { return "f94b5374" }
func (*UrpfInterfaceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *UrpfInterfaceDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	size += 1 // m.IsInput
	size += 1 // m.Mode
	size += 1 // m.Af
	size += 4 // m.TableID
	return size
}
func (m *UrpfInterfaceDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeBool(m.IsInput)
	buf.EncodeUint8(uint8(m.Mode))
	buf.EncodeUint8(uint8(m.Af))
	buf.EncodeUint32(m.TableID)
	return buf.Bytes(), nil
}
func (m *UrpfInterfaceDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IsInput = buf.DecodeBool()
	m.Mode = UrpfMode(buf.DecodeUint8())
	m.Af = ip_types.AddressFamily(buf.DecodeUint8())
	m.TableID = buf.DecodeUint32()
	return nil
}

// @brief Dump uRPF enabled interface(s) in zero or more urpf_interface_details replies
//   - sw_if_index - sw_if_index of a specific interface, or -1 (default)
//     to return all uRPF enabled interfaces
//
// UrpfInterfaceDump defines message 'urpf_interface_dump'.
type UrpfInterfaceDump struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index,default=4294967295" json:"sw_if_index,omitempty"`
}

func (m *UrpfInterfaceDump) Reset()               { *m = UrpfInterfaceDump{} }
func (*UrpfInterfaceDump) GetMessageName() string { return "urpf_interface_dump" }
func (*UrpfInterfaceDump) GetCrcString() string   { return "f9e6675e" }
func (*UrpfInterfaceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *UrpfInterfaceDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *UrpfInterfaceDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *UrpfInterfaceDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// * @brief Enable uRPF on a given interface in a given direction
//   - - mode - Mode
//   - - af - Address Family
//   - - sw_if_index - Interface
//   - - is_input - Direction.
//
// UrpfUpdate defines message 'urpf_update'.
type UrpfUpdate struct {
	IsInput   bool                           `binapi:"bool,name=is_input,default=true" json:"is_input,omitempty"`
	Mode      UrpfMode                       `binapi:"urpf_mode,name=mode" json:"mode,omitempty"`
	Af        ip_types.AddressFamily         `binapi:"address_family,name=af" json:"af,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *UrpfUpdate) Reset()               { *m = UrpfUpdate{} }
func (*UrpfUpdate) GetMessageName() string { return "urpf_update" }
func (*UrpfUpdate) GetCrcString() string   { return "cc274cd1" }
func (*UrpfUpdate) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *UrpfUpdate) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 // m.IsInput
	size += 1 // m.Mode
	size += 1 // m.Af
	size += 4 // m.SwIfIndex
	return size
}
func (m *UrpfUpdate) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsInput)
	buf.EncodeUint8(uint8(m.Mode))
	buf.EncodeUint8(uint8(m.Af))
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *UrpfUpdate) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsInput = buf.DecodeBool()
	m.Mode = UrpfMode(buf.DecodeUint8())
	m.Af = ip_types.AddressFamily(buf.DecodeUint8())
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// UrpfUpdateReply defines message 'urpf_update_reply'.
type UrpfUpdateReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *UrpfUpdateReply) Reset()               { *m = UrpfUpdateReply{} }
func (*UrpfUpdateReply) GetMessageName() string { return "urpf_update_reply" }
func (*UrpfUpdateReply) GetCrcString() string   { return "e8d4e804" }
func (*UrpfUpdateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *UrpfUpdateReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *UrpfUpdateReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *UrpfUpdateReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// * @brief Enable uRPF on a given interface in a given direction
//   - - mode - Mode
//   - - af - Address Family
//   - - sw_if_index - Interface
//   - - is_input - Direction.
//   - - table-id - Table ID
//
// UrpfUpdateV2 defines message 'urpf_update_v2'.
type UrpfUpdateV2 struct {
	IsInput   bool                           `binapi:"bool,name=is_input,default=true" json:"is_input,omitempty"`
	Mode      UrpfMode                       `binapi:"urpf_mode,name=mode" json:"mode,omitempty"`
	Af        ip_types.AddressFamily         `binapi:"address_family,name=af" json:"af,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	TableID   uint32                         `binapi:"u32,name=table_id,default=4294967295" json:"table_id,omitempty"`
}

func (m *UrpfUpdateV2) Reset()               { *m = UrpfUpdateV2{} }
func (*UrpfUpdateV2) GetMessageName() string { return "urpf_update_v2" }
func (*UrpfUpdateV2) GetCrcString() string   { return "b873d028" }
func (*UrpfUpdateV2) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *UrpfUpdateV2) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 // m.IsInput
	size += 1 // m.Mode
	size += 1 // m.Af
	size += 4 // m.SwIfIndex
	size += 4 // m.TableID
	return size
}
func (m *UrpfUpdateV2) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsInput)
	buf.EncodeUint8(uint8(m.Mode))
	buf.EncodeUint8(uint8(m.Af))
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint32(m.TableID)
	return buf.Bytes(), nil
}
func (m *UrpfUpdateV2) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsInput = buf.DecodeBool()
	m.Mode = UrpfMode(buf.DecodeUint8())
	m.Af = ip_types.AddressFamily(buf.DecodeUint8())
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.TableID = buf.DecodeUint32()
	return nil
}

// UrpfUpdateV2Reply defines message 'urpf_update_v2_reply'.
type UrpfUpdateV2Reply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *UrpfUpdateV2Reply) Reset()               { *m = UrpfUpdateV2Reply{} }
func (*UrpfUpdateV2Reply) GetMessageName() string { return "urpf_update_v2_reply" }
func (*UrpfUpdateV2Reply) GetCrcString() string   { return "e8d4e804" }
func (*UrpfUpdateV2Reply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *UrpfUpdateV2Reply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *UrpfUpdateV2Reply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *UrpfUpdateV2Reply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

func init() { file_urpf_binapi_init() }
func file_urpf_binapi_init() {
	api.RegisterMessage((*UrpfInterfaceDetails)(nil), "urpf_interface_details_f94b5374")
	api.RegisterMessage((*UrpfInterfaceDump)(nil), "urpf_interface_dump_f9e6675e")
	api.RegisterMessage((*UrpfUpdate)(nil), "urpf_update_cc274cd1")
	api.RegisterMessage((*UrpfUpdateReply)(nil), "urpf_update_reply_e8d4e804")
	api.RegisterMessage((*UrpfUpdateV2)(nil), "urpf_update_v2_b873d028")
	api.RegisterMessage((*UrpfUpdateV2Reply)(nil), "urpf_update_v2_reply_e8d4e804")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*UrpfInterfaceDetails)(nil),
		(*UrpfInterfaceDump)(nil),
		(*UrpfUpdate)(nil),
		(*UrpfUpdateReply)(nil),
		(*UrpfUpdateV2)(nil),
		(*UrpfUpdateV2Reply)(nil),
	}
}
