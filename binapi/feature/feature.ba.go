// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.8.0
//  VPP:              23.06-release
// source: core/feature.api.json

// Package feature contains generated bindings for API file feature.api.
//
// Contents:
// -  2 messages
package feature

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
	APIFile    = "feature"
	APIVersion = "1.0.2"
	VersionCrc = 0x8a6e6da1
)

// Feature path enable/disable request
//   - sw_if_index - the interface
//   - enable - 1 = on, 0 = off
//
// FeatureEnableDisable defines message 'feature_enable_disable'.
type FeatureEnableDisable struct {
	SwIfIndex   interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	Enable      bool                           `binapi:"bool,name=enable" json:"enable,omitempty"`
	ArcName     string                         `binapi:"string[64],name=arc_name" json:"arc_name,omitempty"`
	FeatureName string                         `binapi:"string[64],name=feature_name" json:"feature_name,omitempty"`
}

func (m *FeatureEnableDisable) Reset()               { *m = FeatureEnableDisable{} }
func (*FeatureEnableDisable) GetMessageName() string { return "feature_enable_disable" }
func (*FeatureEnableDisable) GetCrcString() string   { return "7531c862" }
func (*FeatureEnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *FeatureEnableDisable) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4  // m.SwIfIndex
	size += 1  // m.Enable
	size += 64 // m.ArcName
	size += 64 // m.FeatureName
	return size
}
func (m *FeatureEnableDisable) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeBool(m.Enable)
	buf.EncodeString(m.ArcName, 64)
	buf.EncodeString(m.FeatureName, 64)
	return buf.Bytes(), nil
}
func (m *FeatureEnableDisable) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.Enable = buf.DecodeBool()
	m.ArcName = buf.DecodeString(64)
	m.FeatureName = buf.DecodeString(64)
	return nil
}

// FeatureEnableDisableReply defines message 'feature_enable_disable_reply'.
type FeatureEnableDisableReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *FeatureEnableDisableReply) Reset()               { *m = FeatureEnableDisableReply{} }
func (*FeatureEnableDisableReply) GetMessageName() string { return "feature_enable_disable_reply" }
func (*FeatureEnableDisableReply) GetCrcString() string   { return "e8d4e804" }
func (*FeatureEnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *FeatureEnableDisableReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *FeatureEnableDisableReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *FeatureEnableDisableReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

func init() { file_feature_binapi_init() }
func file_feature_binapi_init() {
	api.RegisterMessage((*FeatureEnableDisable)(nil), "feature_enable_disable_7531c862")
	api.RegisterMessage((*FeatureEnableDisableReply)(nil), "feature_enable_disable_reply_e8d4e804")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*FeatureEnableDisable)(nil),
		(*FeatureEnableDisableReply)(nil),
	}
}
