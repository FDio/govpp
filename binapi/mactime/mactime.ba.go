// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.4.0
//  VPP:              21.06-release
// source: /usr/share/vpp/api/plugins/mactime.api.json

// Package mactime contains generated bindings for API file mactime.api.
//
// Contents:
//   2 structs
//   7 messages
//
package mactime

import (
	api "git.fd.io/govpp.git/api"
	ethernet_types "git.fd.io/govpp.git/binapi/ethernet_types"
	interface_types "git.fd.io/govpp.git/binapi/interface_types"
	codec "git.fd.io/govpp.git/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "mactime"
	APIVersion = "2.0.0"
	VersionCrc = 0xc72e296e
)

// MactimeTimeRange defines type 'mactime_time_range'.
type MactimeTimeRange struct {
	Start float64 `binapi:"f64,name=start" json:"start,omitempty"`
	End   float64 `binapi:"f64,name=end" json:"end,omitempty"`
}

// TimeRange defines type 'time_range'.
type TimeRange struct {
	Start float64 `binapi:"f64,name=start" json:"start,omitempty"`
	End   float64 `binapi:"f64,name=end" json:"end,omitempty"`
}

// MactimeAddDelRange defines message 'mactime_add_del_range'.
type MactimeAddDelRange struct {
	IsAdd      bool                      `binapi:"bool,name=is_add" json:"is_add,omitempty"`
	Drop       bool                      `binapi:"bool,name=drop" json:"drop,omitempty"`
	Allow      bool                      `binapi:"bool,name=allow" json:"allow,omitempty"`
	AllowQuota uint8                     `binapi:"u8,name=allow_quota" json:"allow_quota,omitempty"`
	NoUDP10001 bool                      `binapi:"bool,name=no_udp_10001" json:"no_udp_10001,omitempty"`
	DataQuota  uint64                    `binapi:"u64,name=data_quota" json:"data_quota,omitempty"`
	MacAddress ethernet_types.MacAddress `binapi:"mac_address,name=mac_address" json:"mac_address,omitempty"`
	DeviceName string                    `binapi:"string[64],name=device_name" json:"device_name,omitempty"`
	Count      uint32                    `binapi:"u32,name=count" json:"-"`
	Ranges     []TimeRange               `binapi:"time_range[count],name=ranges" json:"ranges,omitempty"`
}

func (m *MactimeAddDelRange) Reset()               { *m = MactimeAddDelRange{} }
func (*MactimeAddDelRange) GetMessageName() string { return "mactime_add_del_range" }
func (*MactimeAddDelRange) GetCrcString() string   { return "cb56e877" }
func (*MactimeAddDelRange) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *MactimeAddDelRange) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1     // m.IsAdd
	size += 1     // m.Drop
	size += 1     // m.Allow
	size += 1     // m.AllowQuota
	size += 1     // m.NoUDP10001
	size += 8     // m.DataQuota
	size += 1 * 6 // m.MacAddress
	size += 64    // m.DeviceName
	size += 4     // m.Count
	for j1 := 0; j1 < len(m.Ranges); j1++ {
		var s1 TimeRange
		_ = s1
		if j1 < len(m.Ranges) {
			s1 = m.Ranges[j1]
		}
		size += 8 // s1.Start
		size += 8 // s1.End
	}
	return size
}
func (m *MactimeAddDelRange) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.IsAdd)
	buf.EncodeBool(m.Drop)
	buf.EncodeBool(m.Allow)
	buf.EncodeUint8(m.AllowQuota)
	buf.EncodeBool(m.NoUDP10001)
	buf.EncodeUint64(m.DataQuota)
	buf.EncodeBytes(m.MacAddress[:], 6)
	buf.EncodeString(m.DeviceName, 64)
	buf.EncodeUint32(uint32(len(m.Ranges)))
	for j0 := 0; j0 < len(m.Ranges); j0++ {
		var v0 TimeRange // Ranges
		if j0 < len(m.Ranges) {
			v0 = m.Ranges[j0]
		}
		buf.EncodeFloat64(v0.Start)
		buf.EncodeFloat64(v0.End)
	}
	return buf.Bytes(), nil
}
func (m *MactimeAddDelRange) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsAdd = buf.DecodeBool()
	m.Drop = buf.DecodeBool()
	m.Allow = buf.DecodeBool()
	m.AllowQuota = buf.DecodeUint8()
	m.NoUDP10001 = buf.DecodeBool()
	m.DataQuota = buf.DecodeUint64()
	copy(m.MacAddress[:], buf.DecodeBytes(6))
	m.DeviceName = buf.DecodeString(64)
	m.Count = buf.DecodeUint32()
	m.Ranges = make([]TimeRange, m.Count)
	for j0 := 0; j0 < len(m.Ranges); j0++ {
		m.Ranges[j0].Start = buf.DecodeFloat64()
		m.Ranges[j0].End = buf.DecodeFloat64()
	}
	return nil
}

// MactimeAddDelRangeReply defines message 'mactime_add_del_range_reply'.
type MactimeAddDelRangeReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *MactimeAddDelRangeReply) Reset()               { *m = MactimeAddDelRangeReply{} }
func (*MactimeAddDelRangeReply) GetMessageName() string { return "mactime_add_del_range_reply" }
func (*MactimeAddDelRangeReply) GetCrcString() string   { return "e8d4e804" }
func (*MactimeAddDelRangeReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *MactimeAddDelRangeReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *MactimeAddDelRangeReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *MactimeAddDelRangeReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// MactimeDetails defines message 'mactime_details'.
type MactimeDetails struct {
	PoolIndex       uint32                    `binapi:"u32,name=pool_index" json:"pool_index,omitempty"`
	MacAddress      ethernet_types.MacAddress `binapi:"mac_address,name=mac_address" json:"mac_address,omitempty"`
	DataQuota       uint64                    `binapi:"u64,name=data_quota" json:"data_quota,omitempty"`
	DataUsedInRange uint64                    `binapi:"u64,name=data_used_in_range" json:"data_used_in_range,omitempty"`
	Flags           uint32                    `binapi:"u32,name=flags" json:"flags,omitempty"`
	DeviceName      string                    `binapi:"string[64],name=device_name" json:"device_name,omitempty"`
	Nranges         uint32                    `binapi:"u32,name=nranges" json:"-"`
	Ranges          []MactimeTimeRange        `binapi:"mactime_time_range[nranges],name=ranges" json:"ranges,omitempty"`
}

func (m *MactimeDetails) Reset()               { *m = MactimeDetails{} }
func (*MactimeDetails) GetMessageName() string { return "mactime_details" }
func (*MactimeDetails) GetCrcString() string   { return "da25b13a" }
func (*MactimeDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *MactimeDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4     // m.PoolIndex
	size += 1 * 6 // m.MacAddress
	size += 8     // m.DataQuota
	size += 8     // m.DataUsedInRange
	size += 4     // m.Flags
	size += 64    // m.DeviceName
	size += 4     // m.Nranges
	for j1 := 0; j1 < len(m.Ranges); j1++ {
		var s1 MactimeTimeRange
		_ = s1
		if j1 < len(m.Ranges) {
			s1 = m.Ranges[j1]
		}
		size += 8 // s1.Start
		size += 8 // s1.End
	}
	return size
}
func (m *MactimeDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.PoolIndex)
	buf.EncodeBytes(m.MacAddress[:], 6)
	buf.EncodeUint64(m.DataQuota)
	buf.EncodeUint64(m.DataUsedInRange)
	buf.EncodeUint32(m.Flags)
	buf.EncodeString(m.DeviceName, 64)
	buf.EncodeUint32(uint32(len(m.Ranges)))
	for j0 := 0; j0 < len(m.Ranges); j0++ {
		var v0 MactimeTimeRange // Ranges
		if j0 < len(m.Ranges) {
			v0 = m.Ranges[j0]
		}
		buf.EncodeFloat64(v0.Start)
		buf.EncodeFloat64(v0.End)
	}
	return buf.Bytes(), nil
}
func (m *MactimeDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.PoolIndex = buf.DecodeUint32()
	copy(m.MacAddress[:], buf.DecodeBytes(6))
	m.DataQuota = buf.DecodeUint64()
	m.DataUsedInRange = buf.DecodeUint64()
	m.Flags = buf.DecodeUint32()
	m.DeviceName = buf.DecodeString(64)
	m.Nranges = buf.DecodeUint32()
	m.Ranges = make([]MactimeTimeRange, m.Nranges)
	for j0 := 0; j0 < len(m.Ranges); j0++ {
		m.Ranges[j0].Start = buf.DecodeFloat64()
		m.Ranges[j0].End = buf.DecodeFloat64()
	}
	return nil
}

// MactimeDump defines message 'mactime_dump'.
type MactimeDump struct {
	MyTableEpoch uint32 `binapi:"u32,name=my_table_epoch" json:"my_table_epoch,omitempty"`
}

func (m *MactimeDump) Reset()               { *m = MactimeDump{} }
func (*MactimeDump) GetMessageName() string { return "mactime_dump" }
func (*MactimeDump) GetCrcString() string   { return "8f454e23" }
func (*MactimeDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *MactimeDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.MyTableEpoch
	return size
}
func (m *MactimeDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.MyTableEpoch)
	return buf.Bytes(), nil
}
func (m *MactimeDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.MyTableEpoch = buf.DecodeUint32()
	return nil
}

// MactimeDumpReply defines message 'mactime_dump_reply'.
type MactimeDumpReply struct {
	Retval     int32  `binapi:"i32,name=retval" json:"retval,omitempty"`
	TableEpoch uint32 `binapi:"u32,name=table_epoch" json:"table_epoch,omitempty"`
}

func (m *MactimeDumpReply) Reset()               { *m = MactimeDumpReply{} }
func (*MactimeDumpReply) GetMessageName() string { return "mactime_dump_reply" }
func (*MactimeDumpReply) GetCrcString() string   { return "49bcc753" }
func (*MactimeDumpReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *MactimeDumpReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.TableEpoch
	return size
}
func (m *MactimeDumpReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(m.TableEpoch)
	return buf.Bytes(), nil
}
func (m *MactimeDumpReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.TableEpoch = buf.DecodeUint32()
	return nil
}

// MactimeEnableDisable defines message 'mactime_enable_disable'.
type MactimeEnableDisable struct {
	EnableDisable bool                           `binapi:"bool,name=enable_disable" json:"enable_disable,omitempty"`
	SwIfIndex     interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *MactimeEnableDisable) Reset()               { *m = MactimeEnableDisable{} }
func (*MactimeEnableDisable) GetMessageName() string { return "mactime_enable_disable" }
func (*MactimeEnableDisable) GetCrcString() string   { return "3865946c" }
func (*MactimeEnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *MactimeEnableDisable) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 // m.EnableDisable
	size += 4 // m.SwIfIndex
	return size
}
func (m *MactimeEnableDisable) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBool(m.EnableDisable)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *MactimeEnableDisable) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.EnableDisable = buf.DecodeBool()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// MactimeEnableDisableReply defines message 'mactime_enable_disable_reply'.
type MactimeEnableDisableReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *MactimeEnableDisableReply) Reset()               { *m = MactimeEnableDisableReply{} }
func (*MactimeEnableDisableReply) GetMessageName() string { return "mactime_enable_disable_reply" }
func (*MactimeEnableDisableReply) GetCrcString() string   { return "e8d4e804" }
func (*MactimeEnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *MactimeEnableDisableReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *MactimeEnableDisableReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *MactimeEnableDisableReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

func init() { file_mactime_binapi_init() }
func file_mactime_binapi_init() {
	api.RegisterMessage((*MactimeAddDelRange)(nil), "mactime_add_del_range_cb56e877")
	api.RegisterMessage((*MactimeAddDelRangeReply)(nil), "mactime_add_del_range_reply_e8d4e804")
	api.RegisterMessage((*MactimeDetails)(nil), "mactime_details_da25b13a")
	api.RegisterMessage((*MactimeDump)(nil), "mactime_dump_8f454e23")
	api.RegisterMessage((*MactimeDumpReply)(nil), "mactime_dump_reply_49bcc753")
	api.RegisterMessage((*MactimeEnableDisable)(nil), "mactime_enable_disable_3865946c")
	api.RegisterMessage((*MactimeEnableDisableReply)(nil), "mactime_enable_disable_reply_e8d4e804")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*MactimeAddDelRange)(nil),
		(*MactimeAddDelRangeReply)(nil),
		(*MactimeDetails)(nil),
		(*MactimeDump)(nil),
		(*MactimeDumpReply)(nil),
		(*MactimeEnableDisable)(nil),
		(*MactimeEnableDisableReply)(nil),
	}
}
