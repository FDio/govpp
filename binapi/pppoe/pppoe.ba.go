// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.10.0
//  VPP:              24.02-release
// source: plugins/pppoe.api.json

// Package pppoe contains generated bindings for API file pppoe.api.
//
// Contents:
// -  6 messages
package pppoe

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
	APIFile    = "pppoe"
	APIVersion = "2.0.0"
	VersionCrc = 0xec9e86bf
)

// Create PPPOE control plane interface
//   - sw_if_index - software index of the interface
//   - is_add - to create or to delete
//
// PppoeAddDelCp defines message 'pppoe_add_del_cp'.
type PppoeAddDelCp struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IsAdd     uint8                          `binapi:"u8,name=is_add" json:"is_add,omitempty"`
}

func (m *PppoeAddDelCp) Reset()               { *m = PppoeAddDelCp{} }
func (*PppoeAddDelCp) GetMessageName() string { return "pppoe_add_del_cp" }
func (*PppoeAddDelCp) GetCrcString() string   { return "eacd9aaa" }
func (*PppoeAddDelCp) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *PppoeAddDelCp) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	size += 1 // m.IsAdd
	return size
}
func (m *PppoeAddDelCp) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint8(m.IsAdd)
	return buf.Bytes(), nil
}
func (m *PppoeAddDelCp) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IsAdd = buf.DecodeUint8()
	return nil
}

// reply for create PPPOE control plane interface
//   - retval - return code
//
// PppoeAddDelCpReply defines message 'pppoe_add_del_cp_reply'.
type PppoeAddDelCpReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *PppoeAddDelCpReply) Reset()               { *m = PppoeAddDelCpReply{} }
func (*PppoeAddDelCpReply) GetMessageName() string { return "pppoe_add_del_cp_reply" }
func (*PppoeAddDelCpReply) GetCrcString() string   { return "e8d4e804" }
func (*PppoeAddDelCpReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *PppoeAddDelCpReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *PppoeAddDelCpReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *PppoeAddDelCpReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// Set or delete an PPPOE session
//   - is_add - add address if non-zero, else delete
//   - session_id - PPPoE session ID
//   - client_ip - PPPOE session's client address.
//   - decap_vrf_id - the vrf index for pppoe decaped packet
//   - client_mac - the client ethernet address
//
// PppoeAddDelSession defines message 'pppoe_add_del_session'.
type PppoeAddDelSession struct {
	IsAdd      bool                      `binapi:"bool,name=is_add" json:"is_add,omitempty"`
	SessionID  uint16                    `binapi:"u16,name=session_id" json:"session_id,omitempty"`
	ClientIP   ip_types.Address          `binapi:"address,name=client_ip" json:"client_ip,omitempty"`
	DecapVrfID uint32                    `binapi:"u32,name=decap_vrf_id" json:"decap_vrf_id,omitempty"`
	ClientMac  ethernet_types.MacAddress `binapi:"mac_address,name=client_mac" json:"client_mac,omitempty"`
}

func (m *PppoeAddDelSession) Reset()               { *m = PppoeAddDelSession{} }
func (*PppoeAddDelSession) GetMessageName() string { return "pppoe_add_del_session" }
func (*PppoeAddDelSession) GetCrcString() string   { return "f6fd759e" }
func (*PppoeAddDelSession) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *PppoeAddDelSession) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1      // m.IsAdd
	size += 2      // m.SessionID
	size += 1      // m.ClientIP.Af
	size += 1 * 16 // m.ClientIP.Un
	size += 4      // m.DecapVrfID
	size += 1 * 6  // m.ClientMac
	return size
}
func (m *PppoeAddDelSession) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsAdd)
	buf.EncodeUint16(m.SessionID)
	buf.EncodeUint8(uint8(m.ClientIP.Af))
	buf.EncodeBytes(m.ClientIP.Un.XXX_UnionData[:], 16)
	buf.EncodeUint32(m.DecapVrfID)
	buf.EncodeBytes(m.ClientMac[:], 6)
	return buf.Bytes(), nil
}
func (m *PppoeAddDelSession) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsAdd = buf.DecodeBool()
	m.SessionID = buf.DecodeUint16()
	m.ClientIP.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.ClientIP.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.DecapVrfID = buf.DecodeUint32()
	copy(m.ClientMac[:], buf.DecodeBytes(6))
	return nil
}

// reply for set or delete an PPPOE session
//   - retval - return code
//   - sw_if_index - software index of the interface
//
// PppoeAddDelSessionReply defines message 'pppoe_add_del_session_reply'.
type PppoeAddDelSessionReply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *PppoeAddDelSessionReply) Reset()               { *m = PppoeAddDelSessionReply{} }
func (*PppoeAddDelSessionReply) GetMessageName() string { return "pppoe_add_del_session_reply" }
func (*PppoeAddDelSessionReply) GetCrcString() string   { return "5383d31f" }
func (*PppoeAddDelSessionReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *PppoeAddDelSessionReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *PppoeAddDelSessionReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *PppoeAddDelSessionReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// dump details of an PPPOE session
//   - sw_if_index - software index of the interface
//   - session_id - PPPoE session ID
//   - client_ip - PPPOE session's client address.
//   - encap_if_index - the index of tx interface for pppoe encaped packet
//   - decap_vrf_id - the vrf index for pppoe decaped packet
//   - local_mac - the local ethernet address
//   - client_mac - the client ethernet address
//
// PppoeSessionDetails defines message 'pppoe_session_details'.
type PppoeSessionDetails struct {
	SwIfIndex    interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	SessionID    uint16                         `binapi:"u16,name=session_id" json:"session_id,omitempty"`
	ClientIP     ip_types.Address               `binapi:"address,name=client_ip" json:"client_ip,omitempty"`
	EncapIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=encap_if_index" json:"encap_if_index,omitempty"`
	DecapVrfID   uint32                         `binapi:"u32,name=decap_vrf_id" json:"decap_vrf_id,omitempty"`
	LocalMac     ethernet_types.MacAddress      `binapi:"mac_address,name=local_mac" json:"local_mac,omitempty"`
	ClientMac    ethernet_types.MacAddress      `binapi:"mac_address,name=client_mac" json:"client_mac,omitempty"`
}

func (m *PppoeSessionDetails) Reset()               { *m = PppoeSessionDetails{} }
func (*PppoeSessionDetails) GetMessageName() string { return "pppoe_session_details" }
func (*PppoeSessionDetails) GetCrcString() string   { return "4b8e8a4a" }
func (*PppoeSessionDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *PppoeSessionDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4      // m.SwIfIndex
	size += 2      // m.SessionID
	size += 1      // m.ClientIP.Af
	size += 1 * 16 // m.ClientIP.Un
	size += 4      // m.EncapIfIndex
	size += 4      // m.DecapVrfID
	size += 1 * 6  // m.LocalMac
	size += 1 * 6  // m.ClientMac
	return size
}
func (m *PppoeSessionDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint16(m.SessionID)
	buf.EncodeUint8(uint8(m.ClientIP.Af))
	buf.EncodeBytes(m.ClientIP.Un.XXX_UnionData[:], 16)
	buf.EncodeUint32(uint32(m.EncapIfIndex))
	buf.EncodeUint32(m.DecapVrfID)
	buf.EncodeBytes(m.LocalMac[:], 6)
	buf.EncodeBytes(m.ClientMac[:], 6)
	return buf.Bytes(), nil
}
func (m *PppoeSessionDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.SessionID = buf.DecodeUint16()
	m.ClientIP.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.ClientIP.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.EncapIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.DecapVrfID = buf.DecodeUint32()
	copy(m.LocalMac[:], buf.DecodeBytes(6))
	copy(m.ClientMac[:], buf.DecodeBytes(6))
	return nil
}

// Dump PPPOE session
//   - sw_if_index - software index of the interface
//
// PppoeSessionDump defines message 'pppoe_session_dump'.
type PppoeSessionDump struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *PppoeSessionDump) Reset()               { *m = PppoeSessionDump{} }
func (*PppoeSessionDump) GetMessageName() string { return "pppoe_session_dump" }
func (*PppoeSessionDump) GetCrcString() string   { return "f9e6675e" }
func (*PppoeSessionDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *PppoeSessionDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *PppoeSessionDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *PppoeSessionDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

func init() { file_pppoe_binapi_init() }
func file_pppoe_binapi_init() {
	api.RegisterMessage((*PppoeAddDelCp)(nil), "pppoe_add_del_cp_eacd9aaa")
	api.RegisterMessage((*PppoeAddDelCpReply)(nil), "pppoe_add_del_cp_reply_e8d4e804")
	api.RegisterMessage((*PppoeAddDelSession)(nil), "pppoe_add_del_session_f6fd759e")
	api.RegisterMessage((*PppoeAddDelSessionReply)(nil), "pppoe_add_del_session_reply_5383d31f")
	api.RegisterMessage((*PppoeSessionDetails)(nil), "pppoe_session_details_4b8e8a4a")
	api.RegisterMessage((*PppoeSessionDump)(nil), "pppoe_session_dump_f9e6675e")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*PppoeAddDelCp)(nil),
		(*PppoeAddDelCpReply)(nil),
		(*PppoeAddDelSession)(nil),
		(*PppoeAddDelSessionReply)(nil),
		(*PppoeSessionDetails)(nil),
		(*PppoeSessionDump)(nil),
	}
}
