// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.10.0
//  VPP:              24.02-release
// source: plugins/ioam_export.api.json

// Package ioam_export contains generated bindings for API file ioam_export.api.
//
// Contents:
// -  2 messages
package ioam_export

import (
	api "go.fd.io/govpp/api"
	ip_types "go.fd.io/govpp/binapi/ip_types"
	codec "go.fd.io/govpp/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "ioam_export"
	APIVersion = "1.0.0"
	VersionCrc = 0x26bebf64
)

// /* Define a simple binary API to control the feature
// IoamExportIP6EnableDisable defines message 'ioam_export_ip6_enable_disable'.
type IoamExportIP6EnableDisable struct {
	IsDisable        bool                `binapi:"bool,name=is_disable" json:"is_disable,omitempty"`
	CollectorAddress ip_types.IP4Address `binapi:"ip4_address,name=collector_address" json:"collector_address,omitempty"`
	SrcAddress       ip_types.IP4Address `binapi:"ip4_address,name=src_address" json:"src_address,omitempty"`
}

func (m *IoamExportIP6EnableDisable) Reset()               { *m = IoamExportIP6EnableDisable{} }
func (*IoamExportIP6EnableDisable) GetMessageName() string { return "ioam_export_ip6_enable_disable" }
func (*IoamExportIP6EnableDisable) GetCrcString() string   { return "d4c76d3a" }
func (*IoamExportIP6EnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *IoamExportIP6EnableDisable) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1     // m.IsDisable
	size += 1 * 4 // m.CollectorAddress
	size += 1 * 4 // m.SrcAddress
	return size
}
func (m *IoamExportIP6EnableDisable) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsDisable)
	buf.EncodeBytes(m.CollectorAddress[:], 4)
	buf.EncodeBytes(m.SrcAddress[:], 4)
	return buf.Bytes(), nil
}
func (m *IoamExportIP6EnableDisable) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsDisable = buf.DecodeBool()
	copy(m.CollectorAddress[:], buf.DecodeBytes(4))
	copy(m.SrcAddress[:], buf.DecodeBytes(4))
	return nil
}

// IoamExportIP6EnableDisableReply defines message 'ioam_export_ip6_enable_disable_reply'.
type IoamExportIP6EnableDisableReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *IoamExportIP6EnableDisableReply) Reset() { *m = IoamExportIP6EnableDisableReply{} }
func (*IoamExportIP6EnableDisableReply) GetMessageName() string {
	return "ioam_export_ip6_enable_disable_reply"
}
func (*IoamExportIP6EnableDisableReply) GetCrcString() string { return "e8d4e804" }
func (*IoamExportIP6EnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *IoamExportIP6EnableDisableReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *IoamExportIP6EnableDisableReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *IoamExportIP6EnableDisableReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

func init() { file_ioam_export_binapi_init() }
func file_ioam_export_binapi_init() {
	api.RegisterMessage((*IoamExportIP6EnableDisable)(nil), "ioam_export_ip6_enable_disable_d4c76d3a")
	api.RegisterMessage((*IoamExportIP6EnableDisableReply)(nil), "ioam_export_ip6_enable_disable_reply_e8d4e804")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*IoamExportIP6EnableDisable)(nil),
		(*IoamExportIP6EnableDisableReply)(nil),
	}
}
