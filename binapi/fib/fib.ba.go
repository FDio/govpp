// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.8.0
//  VPP:              23.06-release
// source: core/fib.api.json

// Package fib contains generated bindings for API file fib.api.
//
// Contents:
// -  1 struct
// -  4 messages
package fib

import (
	api "go.fd.io/govpp/api"
	_ "go.fd.io/govpp/binapi/fib_types"
	_ "go.fd.io/govpp/binapi/ip_types"
	codec "go.fd.io/govpp/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "fib"
	APIVersion = "1.0.0"
	VersionCrc = 0x4ef4abc1
)

// FibSource defines type 'fib_source'.
type FibSource struct {
	Priority uint8  `binapi:"u8,name=priority" json:"priority,omitempty"`
	ID       uint8  `binapi:"u8,name=id" json:"id,omitempty"`
	Name     string `binapi:"string[64],name=name" json:"name,omitempty"`
}

// /*
//   - Copyright (c) 2018 Cisco and/or its affiliates.
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
// FibSourceAdd defines message 'fib_source_add'.
type FibSourceAdd struct {
	Src FibSource `binapi:"fib_source,name=src" json:"src,omitempty"`
}

func (m *FibSourceAdd) Reset()               { *m = FibSourceAdd{} }
func (*FibSourceAdd) GetMessageName() string { return "fib_source_add" }
func (*FibSourceAdd) GetCrcString() string   { return "b3ac2aec" }
func (*FibSourceAdd) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *FibSourceAdd) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1  // m.Src.Priority
	size += 1  // m.Src.ID
	size += 64 // m.Src.Name
	return size
}
func (m *FibSourceAdd) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint8(m.Src.Priority)
	buf.EncodeUint8(m.Src.ID)
	buf.EncodeString(m.Src.Name, 64)
	return buf.Bytes(), nil
}
func (m *FibSourceAdd) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Src.Priority = buf.DecodeUint8()
	m.Src.ID = buf.DecodeUint8()
	m.Src.Name = buf.DecodeString(64)
	return nil
}

// FibSourceAddReply defines message 'fib_source_add_reply'.
type FibSourceAddReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
	ID     uint8 `binapi:"u8,name=id" json:"id,omitempty"`
}

func (m *FibSourceAddReply) Reset()               { *m = FibSourceAddReply{} }
func (*FibSourceAddReply) GetMessageName() string { return "fib_source_add_reply" }
func (*FibSourceAddReply) GetCrcString() string   { return "604fd6f1" }
func (*FibSourceAddReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *FibSourceAddReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 1 // m.ID
	return size
}
func (m *FibSourceAddReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint8(m.ID)
	return buf.Bytes(), nil
}
func (m *FibSourceAddReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.ID = buf.DecodeUint8()
	return nil
}

// FibSourceDetails defines message 'fib_source_details'.
type FibSourceDetails struct {
	Src FibSource `binapi:"fib_source,name=src" json:"src,omitempty"`
}

func (m *FibSourceDetails) Reset()               { *m = FibSourceDetails{} }
func (*FibSourceDetails) GetMessageName() string { return "fib_source_details" }
func (*FibSourceDetails) GetCrcString() string   { return "8668acdb" }
func (*FibSourceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *FibSourceDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1  // m.Src.Priority
	size += 1  // m.Src.ID
	size += 64 // m.Src.Name
	return size
}
func (m *FibSourceDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint8(m.Src.Priority)
	buf.EncodeUint8(m.Src.ID)
	buf.EncodeString(m.Src.Name, 64)
	return buf.Bytes(), nil
}
func (m *FibSourceDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Src.Priority = buf.DecodeUint8()
	m.Src.ID = buf.DecodeUint8()
	m.Src.Name = buf.DecodeString(64)
	return nil
}

// FibSourceDump defines message 'fib_source_dump'.
type FibSourceDump struct{}

func (m *FibSourceDump) Reset()               { *m = FibSourceDump{} }
func (*FibSourceDump) GetMessageName() string { return "fib_source_dump" }
func (*FibSourceDump) GetCrcString() string   { return "51077d14" }
func (*FibSourceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *FibSourceDump) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *FibSourceDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *FibSourceDump) Unmarshal(b []byte) error {
	return nil
}

func init() { file_fib_binapi_init() }
func file_fib_binapi_init() {
	api.RegisterMessage((*FibSourceAdd)(nil), "fib_source_add_b3ac2aec")
	api.RegisterMessage((*FibSourceAddReply)(nil), "fib_source_add_reply_604fd6f1")
	api.RegisterMessage((*FibSourceDetails)(nil), "fib_source_details_8668acdb")
	api.RegisterMessage((*FibSourceDump)(nil), "fib_source_dump_51077d14")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*FibSourceAdd)(nil),
		(*FibSourceAddReply)(nil),
		(*FibSourceDetails)(nil),
		(*FibSourceDump)(nil),
	}
}
