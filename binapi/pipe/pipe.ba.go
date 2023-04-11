// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.7.0
//  VPP:              22.10-release
// source: core/pipe.api.json

// Package pipe contains generated bindings for API file pipe.api.
//
// Contents:
// -  6 messages
package pipe

import (
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
	APIFile    = "pipe"
	APIVersion = "1.0.1"
	VersionCrc = 0xc159134a
)

// PipeCreate defines message 'pipe_create'.
type PipeCreate struct {
	IsSpecified  bool   `binapi:"bool,name=is_specified" json:"is_specified,omitempty"`
	UserInstance uint32 `binapi:"u32,name=user_instance" json:"user_instance,omitempty"`
}

func (m *PipeCreate) Reset()               { *m = PipeCreate{} }
func (*PipeCreate) GetMessageName() string { return "pipe_create" }
func (*PipeCreate) GetCrcString() string   { return "bb263bd3" }
func (*PipeCreate) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *PipeCreate) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 // m.IsSpecified
	size += 4 // m.UserInstance
	return size
}
func (m *PipeCreate) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsSpecified)
	buf.EncodeUint32(m.UserInstance)
	return buf.Bytes(), nil
}
func (m *PipeCreate) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsSpecified = buf.DecodeBool()
	m.UserInstance = buf.DecodeUint32()
	return nil
}

// PipeCreateReply defines message 'pipe_create_reply'.
type PipeCreateReply struct {
	Retval        int32                             `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex     interface_types.InterfaceIndex    `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	PipeSwIfIndex [2]interface_types.InterfaceIndex `binapi:"interface_index[2],name=pipe_sw_if_index" json:"pipe_sw_if_index,omitempty"`
}

func (m *PipeCreateReply) Reset()               { *m = PipeCreateReply{} }
func (*PipeCreateReply) GetMessageName() string { return "pipe_create_reply" }
func (*PipeCreateReply) GetCrcString() string   { return "b7ce310c" }
func (*PipeCreateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *PipeCreateReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	for j1 := 0; j1 < 2; j1++ {
		size += 4 // m.PipeSwIfIndex[j1]
	}
	return size
}
func (m *PipeCreateReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	for j0 := 0; j0 < 2; j0++ {
		buf.EncodeUint32(uint32(m.PipeSwIfIndex[j0]))
	}
	return buf.Bytes(), nil
}
func (m *PipeCreateReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	for j0 := 0; j0 < 2; j0++ {
		m.PipeSwIfIndex[j0] = interface_types.InterfaceIndex(buf.DecodeUint32())
	}
	return nil
}

// PipeDelete defines message 'pipe_delete'.
type PipeDelete struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *PipeDelete) Reset()               { *m = PipeDelete{} }
func (*PipeDelete) GetMessageName() string { return "pipe_delete" }
func (*PipeDelete) GetCrcString() string   { return "f9e6675e" }
func (*PipeDelete) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *PipeDelete) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *PipeDelete) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *PipeDelete) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// PipeDeleteReply defines message 'pipe_delete_reply'.
type PipeDeleteReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *PipeDeleteReply) Reset()               { *m = PipeDeleteReply{} }
func (*PipeDeleteReply) GetMessageName() string { return "pipe_delete_reply" }
func (*PipeDeleteReply) GetCrcString() string   { return "e8d4e804" }
func (*PipeDeleteReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *PipeDeleteReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *PipeDeleteReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *PipeDeleteReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// PipeDetails defines message 'pipe_details'.
type PipeDetails struct {
	SwIfIndex     interface_types.InterfaceIndex    `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	PipeSwIfIndex [2]interface_types.InterfaceIndex `binapi:"interface_index[2],name=pipe_sw_if_index" json:"pipe_sw_if_index,omitempty"`
	Instance      uint32                            `binapi:"u32,name=instance" json:"instance,omitempty"`
}

func (m *PipeDetails) Reset()               { *m = PipeDetails{} }
func (*PipeDetails) GetMessageName() string { return "pipe_details" }
func (*PipeDetails) GetCrcString() string   { return "c52b799d" }
func (*PipeDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *PipeDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	for j1 := 0; j1 < 2; j1++ {
		size += 4 // m.PipeSwIfIndex[j1]
	}
	size += 4 // m.Instance
	return size
}
func (m *PipeDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	for j0 := 0; j0 < 2; j0++ {
		buf.EncodeUint32(uint32(m.PipeSwIfIndex[j0]))
	}
	buf.EncodeUint32(m.Instance)
	return buf.Bytes(), nil
}
func (m *PipeDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	for j0 := 0; j0 < 2; j0++ {
		m.PipeSwIfIndex[j0] = interface_types.InterfaceIndex(buf.DecodeUint32())
	}
	m.Instance = buf.DecodeUint32()
	return nil
}

// PipeDump defines message 'pipe_dump'.
type PipeDump struct{}

func (m *PipeDump) Reset()               { *m = PipeDump{} }
func (*PipeDump) GetMessageName() string { return "pipe_dump" }
func (*PipeDump) GetCrcString() string   { return "51077d14" }
func (*PipeDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *PipeDump) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *PipeDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *PipeDump) Unmarshal(b []byte) error {
	return nil
}

func init() { file_pipe_binapi_init() }
func file_pipe_binapi_init() {
	api.RegisterMessage((*PipeCreate)(nil), "pipe_create_bb263bd3")
	api.RegisterMessage((*PipeCreateReply)(nil), "pipe_create_reply_b7ce310c")
	api.RegisterMessage((*PipeDelete)(nil), "pipe_delete_f9e6675e")
	api.RegisterMessage((*PipeDeleteReply)(nil), "pipe_delete_reply_e8d4e804")
	api.RegisterMessage((*PipeDetails)(nil), "pipe_details_c52b799d")
	api.RegisterMessage((*PipeDump)(nil), "pipe_dump_51077d14")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*PipeCreate)(nil),
		(*PipeCreateReply)(nil),
		(*PipeDelete)(nil),
		(*PipeDeleteReply)(nil),
		(*PipeDetails)(nil),
		(*PipeDump)(nil),
	}
}
