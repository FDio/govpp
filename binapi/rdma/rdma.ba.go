// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.10.0
//  VPP:              24.02-release
// source: plugins/rdma.api.json

// Package rdma contains generated bindings for API file rdma.api.
//
// Contents:
// -  3 enums
// - 10 messages
package rdma

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
	APIFile    = "rdma"
	APIVersion = "3.0.0"
	VersionCrc = 0x351383c2
)

// RdmaMode defines enum 'rdma_mode'.
type RdmaMode uint32

const (
	RDMA_API_MODE_AUTO RdmaMode = 0
	RDMA_API_MODE_IBV  RdmaMode = 1
	RDMA_API_MODE_DV   RdmaMode = 2
)

var (
	RdmaMode_name = map[uint32]string{
		0: "RDMA_API_MODE_AUTO",
		1: "RDMA_API_MODE_IBV",
		2: "RDMA_API_MODE_DV",
	}
	RdmaMode_value = map[string]uint32{
		"RDMA_API_MODE_AUTO": 0,
		"RDMA_API_MODE_IBV":  1,
		"RDMA_API_MODE_DV":   2,
	}
)

func (x RdmaMode) String() string {
	s, ok := RdmaMode_name[uint32(x)]
	if ok {
		return s
	}
	return "RdmaMode(" + strconv.Itoa(int(x)) + ")"
}

// RdmaRss4 defines enum 'rdma_rss4'.
type RdmaRss4 uint32

const (
	RDMA_API_RSS4_AUTO   RdmaRss4 = 0
	RDMA_API_RSS4_IP     RdmaRss4 = 1
	RDMA_API_RSS4_IP_UDP RdmaRss4 = 2
	RDMA_API_RSS4_IP_TCP RdmaRss4 = 3
)

var (
	RdmaRss4_name = map[uint32]string{
		0: "RDMA_API_RSS4_AUTO",
		1: "RDMA_API_RSS4_IP",
		2: "RDMA_API_RSS4_IP_UDP",
		3: "RDMA_API_RSS4_IP_TCP",
	}
	RdmaRss4_value = map[string]uint32{
		"RDMA_API_RSS4_AUTO":   0,
		"RDMA_API_RSS4_IP":     1,
		"RDMA_API_RSS4_IP_UDP": 2,
		"RDMA_API_RSS4_IP_TCP": 3,
	}
)

func (x RdmaRss4) String() string {
	s, ok := RdmaRss4_name[uint32(x)]
	if ok {
		return s
	}
	return "RdmaRss4(" + strconv.Itoa(int(x)) + ")"
}

// RdmaRss6 defines enum 'rdma_rss6'.
type RdmaRss6 uint32

const (
	RDMA_API_RSS6_AUTO   RdmaRss6 = 0
	RDMA_API_RSS6_IP     RdmaRss6 = 1
	RDMA_API_RSS6_IP_UDP RdmaRss6 = 2
	RDMA_API_RSS6_IP_TCP RdmaRss6 = 3
)

var (
	RdmaRss6_name = map[uint32]string{
		0: "RDMA_API_RSS6_AUTO",
		1: "RDMA_API_RSS6_IP",
		2: "RDMA_API_RSS6_IP_UDP",
		3: "RDMA_API_RSS6_IP_TCP",
	}
	RdmaRss6_value = map[string]uint32{
		"RDMA_API_RSS6_AUTO":   0,
		"RDMA_API_RSS6_IP":     1,
		"RDMA_API_RSS6_IP_UDP": 2,
		"RDMA_API_RSS6_IP_TCP": 3,
	}
)

func (x RdmaRss6) String() string {
	s, ok := RdmaRss6_name[uint32(x)]
	if ok {
		return s
	}
	return "RdmaRss6(" + strconv.Itoa(int(x)) + ")"
}

// - client_index - opaque cookie to identify the sender
//   - host_if - Linux netdev interface name
//   - name - new rdma interface name
//   - rxq_num - number of receive queues (optional)
//   - rxq_size - receive queue size (optional)
//   - txq_size - transmit queue size (optional)
//   - mode - operation mode (optional)
//
// RdmaCreate defines message 'rdma_create'.
// Deprecated: 21.01
type RdmaCreate struct {
	HostIf  string   `binapi:"string[64],name=host_if" json:"host_if,omitempty"`
	Name    string   `binapi:"string[64],name=name" json:"name,omitempty"`
	RxqNum  uint16   `binapi:"u16,name=rxq_num,default=1" json:"rxq_num,omitempty"`
	RxqSize uint16   `binapi:"u16,name=rxq_size,default=1024" json:"rxq_size,omitempty"`
	TxqSize uint16   `binapi:"u16,name=txq_size,default=1024" json:"txq_size,omitempty"`
	Mode    RdmaMode `binapi:"rdma_mode,name=mode,default=0" json:"mode,omitempty"`
}

func (m *RdmaCreate) Reset()               { *m = RdmaCreate{} }
func (*RdmaCreate) GetMessageName() string { return "rdma_create" }
func (*RdmaCreate) GetCrcString() string   { return "076fe418" }
func (*RdmaCreate) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *RdmaCreate) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 64 // m.HostIf
	size += 64 // m.Name
	size += 2  // m.RxqNum
	size += 2  // m.RxqSize
	size += 2  // m.TxqSize
	size += 4  // m.Mode
	return size
}
func (m *RdmaCreate) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeString(m.HostIf, 64)
	buf.EncodeString(m.Name, 64)
	buf.EncodeUint16(m.RxqNum)
	buf.EncodeUint16(m.RxqSize)
	buf.EncodeUint16(m.TxqSize)
	buf.EncodeUint32(uint32(m.Mode))
	return buf.Bytes(), nil
}
func (m *RdmaCreate) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.HostIf = buf.DecodeString(64)
	m.Name = buf.DecodeString(64)
	m.RxqNum = buf.DecodeUint16()
	m.RxqSize = buf.DecodeUint16()
	m.TxqSize = buf.DecodeUint16()
	m.Mode = RdmaMode(buf.DecodeUint32())
	return nil
}

// - context - sender context, to match reply w/ request
//   - retval - return value for request
//   - sw_if_index - software index for the new rdma interface
//
// RdmaCreateReply defines message 'rdma_create_reply'.
// Deprecated: the message will be removed in the future versions
type RdmaCreateReply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *RdmaCreateReply) Reset()               { *m = RdmaCreateReply{} }
func (*RdmaCreateReply) GetMessageName() string { return "rdma_create_reply" }
func (*RdmaCreateReply) GetCrcString() string   { return "5383d31f" }
func (*RdmaCreateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *RdmaCreateReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *RdmaCreateReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *RdmaCreateReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// - client_index - opaque cookie to identify the sender
//   - host_if - Linux netdev interface name
//   - name - new rdma interface name
//   - rxq_num - number of receive queues (optional)
//   - rxq_size - receive queue size (optional)
//   - txq_size - transmit queue size (optional)
//   - mode - operation mode (optional)
//   - no_multi_seg (optional) - disable chained buffer RX
//   - max_pktlen (optional) - maximal RX packet size.
//
// RdmaCreateV2 defines message 'rdma_create_v2'.
// Deprecated: the message will be removed in the future versions
type RdmaCreateV2 struct {
	HostIf     string   `binapi:"string[64],name=host_if" json:"host_if,omitempty"`
	Name       string   `binapi:"string[64],name=name" json:"name,omitempty"`
	RxqNum     uint16   `binapi:"u16,name=rxq_num,default=1" json:"rxq_num,omitempty"`
	RxqSize    uint16   `binapi:"u16,name=rxq_size,default=1024" json:"rxq_size,omitempty"`
	TxqSize    uint16   `binapi:"u16,name=txq_size,default=1024" json:"txq_size,omitempty"`
	Mode       RdmaMode `binapi:"rdma_mode,name=mode,default=0" json:"mode,omitempty"`
	NoMultiSeg bool     `binapi:"bool,name=no_multi_seg,default=0" json:"no_multi_seg,omitempty"`
	MaxPktlen  uint16   `binapi:"u16,name=max_pktlen,default=0" json:"max_pktlen,omitempty"`
}

func (m *RdmaCreateV2) Reset()               { *m = RdmaCreateV2{} }
func (*RdmaCreateV2) GetMessageName() string { return "rdma_create_v2" }
func (*RdmaCreateV2) GetCrcString() string   { return "5826a4f3" }
func (*RdmaCreateV2) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *RdmaCreateV2) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 64 // m.HostIf
	size += 64 // m.Name
	size += 2  // m.RxqNum
	size += 2  // m.RxqSize
	size += 2  // m.TxqSize
	size += 4  // m.Mode
	size += 1  // m.NoMultiSeg
	size += 2  // m.MaxPktlen
	return size
}
func (m *RdmaCreateV2) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeString(m.HostIf, 64)
	buf.EncodeString(m.Name, 64)
	buf.EncodeUint16(m.RxqNum)
	buf.EncodeUint16(m.RxqSize)
	buf.EncodeUint16(m.TxqSize)
	buf.EncodeUint32(uint32(m.Mode))
	buf.EncodeBool(m.NoMultiSeg)
	buf.EncodeUint16(m.MaxPktlen)
	return buf.Bytes(), nil
}
func (m *RdmaCreateV2) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.HostIf = buf.DecodeString(64)
	m.Name = buf.DecodeString(64)
	m.RxqNum = buf.DecodeUint16()
	m.RxqSize = buf.DecodeUint16()
	m.TxqSize = buf.DecodeUint16()
	m.Mode = RdmaMode(buf.DecodeUint32())
	m.NoMultiSeg = buf.DecodeBool()
	m.MaxPktlen = buf.DecodeUint16()
	return nil
}

// - context - sender context, to match reply w/ request
//   - retval - return value for request
//   - sw_if_index - software index for the new rdma interface
//
// RdmaCreateV2Reply defines message 'rdma_create_v2_reply'.
// Deprecated: the message will be removed in the future versions
type RdmaCreateV2Reply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *RdmaCreateV2Reply) Reset()               { *m = RdmaCreateV2Reply{} }
func (*RdmaCreateV2Reply) GetMessageName() string { return "rdma_create_v2_reply" }
func (*RdmaCreateV2Reply) GetCrcString() string   { return "5383d31f" }
func (*RdmaCreateV2Reply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *RdmaCreateV2Reply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *RdmaCreateV2Reply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *RdmaCreateV2Reply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Same as v4, just not an autoendian (expect buggy handling of flag values).
//   - host_if - Linux netdev interface name
//   - name - new rdma interface name
//   - rxq_num - number of receive queues (optional)
//   - rxq_size - receive queue size (optional)
//   - txq_size - transmit queue size (optional)
//   - mode - operation mode (optional)
//   - no_multi_seg (optional) - disable chained buffer RX
//   - max_pktlen (optional) - maximal RX packet size.
//   - rss4 (optional) - IPv4 RSS
//   - rss6 (optional) - IPv6 RSS
//
// RdmaCreateV3 defines message 'rdma_create_v3'.
// Deprecated: the message will be removed in the future versions
type RdmaCreateV3 struct {
	HostIf     string   `binapi:"string[64],name=host_if" json:"host_if,omitempty"`
	Name       string   `binapi:"string[64],name=name" json:"name,omitempty"`
	RxqNum     uint16   `binapi:"u16,name=rxq_num,default=1" json:"rxq_num,omitempty"`
	RxqSize    uint16   `binapi:"u16,name=rxq_size,default=1024" json:"rxq_size,omitempty"`
	TxqSize    uint16   `binapi:"u16,name=txq_size,default=1024" json:"txq_size,omitempty"`
	Mode       RdmaMode `binapi:"rdma_mode,name=mode,default=0" json:"mode,omitempty"`
	NoMultiSeg bool     `binapi:"bool,name=no_multi_seg,default=0" json:"no_multi_seg,omitempty"`
	MaxPktlen  uint16   `binapi:"u16,name=max_pktlen,default=0" json:"max_pktlen,omitempty"`
	Rss4       RdmaRss4 `binapi:"rdma_rss4,name=rss4,default=0" json:"rss4,omitempty"`
	Rss6       RdmaRss6 `binapi:"rdma_rss6,name=rss6,default=0" json:"rss6,omitempty"`
}

func (m *RdmaCreateV3) Reset()               { *m = RdmaCreateV3{} }
func (*RdmaCreateV3) GetMessageName() string { return "rdma_create_v3" }
func (*RdmaCreateV3) GetCrcString() string   { return "c6287ea8" }
func (*RdmaCreateV3) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *RdmaCreateV3) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 64 // m.HostIf
	size += 64 // m.Name
	size += 2  // m.RxqNum
	size += 2  // m.RxqSize
	size += 2  // m.TxqSize
	size += 4  // m.Mode
	size += 1  // m.NoMultiSeg
	size += 2  // m.MaxPktlen
	size += 4  // m.Rss4
	size += 4  // m.Rss6
	return size
}
func (m *RdmaCreateV3) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeString(m.HostIf, 64)
	buf.EncodeString(m.Name, 64)
	buf.EncodeUint16(m.RxqNum)
	buf.EncodeUint16(m.RxqSize)
	buf.EncodeUint16(m.TxqSize)
	buf.EncodeUint32(uint32(m.Mode))
	buf.EncodeBool(m.NoMultiSeg)
	buf.EncodeUint16(m.MaxPktlen)
	buf.EncodeUint32(uint32(m.Rss4))
	buf.EncodeUint32(uint32(m.Rss6))
	return buf.Bytes(), nil
}
func (m *RdmaCreateV3) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.HostIf = buf.DecodeString(64)
	m.Name = buf.DecodeString(64)
	m.RxqNum = buf.DecodeUint16()
	m.RxqSize = buf.DecodeUint16()
	m.TxqSize = buf.DecodeUint16()
	m.Mode = RdmaMode(buf.DecodeUint32())
	m.NoMultiSeg = buf.DecodeBool()
	m.MaxPktlen = buf.DecodeUint16()
	m.Rss4 = RdmaRss4(buf.DecodeUint32())
	m.Rss6 = RdmaRss6(buf.DecodeUint32())
	return nil
}

// - client_index - opaque cookie to identify the sender
//   - sw_if_index - interface index
//
// RdmaCreateV3Reply defines message 'rdma_create_v3_reply'.
type RdmaCreateV3Reply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *RdmaCreateV3Reply) Reset()               { *m = RdmaCreateV3Reply{} }
func (*RdmaCreateV3Reply) GetMessageName() string { return "rdma_create_v3_reply" }
func (*RdmaCreateV3Reply) GetCrcString() string   { return "5383d31f" }
func (*RdmaCreateV3Reply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *RdmaCreateV3Reply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *RdmaCreateV3Reply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *RdmaCreateV3Reply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// - client_index - opaque cookie to identify the sender
//   - host_if - Linux netdev interface name
//   - name - new rdma interface name
//   - rxq_num - number of receive queues (optional)
//   - rxq_size - receive queue size (optional)
//   - txq_size - transmit queue size (optional)
//   - mode - operation mode (optional)
//   - no_multi_seg (optional) - disable chained buffer RX
//   - max_pktlen (optional) - maximal RX packet size.
//   - rss4 (optional) - IPv4 RSS
//   - rss6 (optional) - IPv6 RSS
//
// RdmaCreateV4 defines message 'rdma_create_v4'.
type RdmaCreateV4 struct {
	HostIf     string   `binapi:"string[64],name=host_if" json:"host_if,omitempty"`
	Name       string   `binapi:"string[64],name=name" json:"name,omitempty"`
	RxqNum     uint16   `binapi:"u16,name=rxq_num,default=1" json:"rxq_num,omitempty"`
	RxqSize    uint16   `binapi:"u16,name=rxq_size,default=1024" json:"rxq_size,omitempty"`
	TxqSize    uint16   `binapi:"u16,name=txq_size,default=1024" json:"txq_size,omitempty"`
	Mode       RdmaMode `binapi:"rdma_mode,name=mode,default=0" json:"mode,omitempty"`
	NoMultiSeg bool     `binapi:"bool,name=no_multi_seg,default=0" json:"no_multi_seg,omitempty"`
	MaxPktlen  uint16   `binapi:"u16,name=max_pktlen,default=0" json:"max_pktlen,omitempty"`
	Rss4       RdmaRss4 `binapi:"rdma_rss4,name=rss4,default=0" json:"rss4,omitempty"`
	Rss6       RdmaRss6 `binapi:"rdma_rss6,name=rss6,default=0" json:"rss6,omitempty"`
}

func (m *RdmaCreateV4) Reset()               { *m = RdmaCreateV4{} }
func (*RdmaCreateV4) GetMessageName() string { return "rdma_create_v4" }
func (*RdmaCreateV4) GetCrcString() string   { return "c6287ea8" }
func (*RdmaCreateV4) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *RdmaCreateV4) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 64 // m.HostIf
	size += 64 // m.Name
	size += 2  // m.RxqNum
	size += 2  // m.RxqSize
	size += 2  // m.TxqSize
	size += 4  // m.Mode
	size += 1  // m.NoMultiSeg
	size += 2  // m.MaxPktlen
	size += 4  // m.Rss4
	size += 4  // m.Rss6
	return size
}
func (m *RdmaCreateV4) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeString(m.HostIf, 64)
	buf.EncodeString(m.Name, 64)
	buf.EncodeUint16(m.RxqNum)
	buf.EncodeUint16(m.RxqSize)
	buf.EncodeUint16(m.TxqSize)
	buf.EncodeUint32(uint32(m.Mode))
	buf.EncodeBool(m.NoMultiSeg)
	buf.EncodeUint16(m.MaxPktlen)
	buf.EncodeUint32(uint32(m.Rss4))
	buf.EncodeUint32(uint32(m.Rss6))
	return buf.Bytes(), nil
}
func (m *RdmaCreateV4) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.HostIf = buf.DecodeString(64)
	m.Name = buf.DecodeString(64)
	m.RxqNum = buf.DecodeUint16()
	m.RxqSize = buf.DecodeUint16()
	m.TxqSize = buf.DecodeUint16()
	m.Mode = RdmaMode(buf.DecodeUint32())
	m.NoMultiSeg = buf.DecodeBool()
	m.MaxPktlen = buf.DecodeUint16()
	m.Rss4 = RdmaRss4(buf.DecodeUint32())
	m.Rss6 = RdmaRss6(buf.DecodeUint32())
	return nil
}

// - client_index - opaque cookie to identify the sender
//   - sw_if_index - interface index
//
// RdmaCreateV4Reply defines message 'rdma_create_v4_reply'.
type RdmaCreateV4Reply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *RdmaCreateV4Reply) Reset()               { *m = RdmaCreateV4Reply{} }
func (*RdmaCreateV4Reply) GetMessageName() string { return "rdma_create_v4_reply" }
func (*RdmaCreateV4Reply) GetCrcString() string   { return "5383d31f" }
func (*RdmaCreateV4Reply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *RdmaCreateV4Reply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *RdmaCreateV4Reply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *RdmaCreateV4Reply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// - client_index - opaque cookie to identify the sender
//   - sw_if_index - interface index
//
// RdmaDelete defines message 'rdma_delete'.
type RdmaDelete struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *RdmaDelete) Reset()               { *m = RdmaDelete{} }
func (*RdmaDelete) GetMessageName() string { return "rdma_delete" }
func (*RdmaDelete) GetCrcString() string   { return "f9e6675e" }
func (*RdmaDelete) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *RdmaDelete) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *RdmaDelete) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *RdmaDelete) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// RdmaDeleteReply defines message 'rdma_delete_reply'.
type RdmaDeleteReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *RdmaDeleteReply) Reset()               { *m = RdmaDeleteReply{} }
func (*RdmaDeleteReply) GetMessageName() string { return "rdma_delete_reply" }
func (*RdmaDeleteReply) GetCrcString() string   { return "e8d4e804" }
func (*RdmaDeleteReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *RdmaDeleteReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *RdmaDeleteReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *RdmaDeleteReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

func init() { file_rdma_binapi_init() }
func file_rdma_binapi_init() {
	api.RegisterMessage((*RdmaCreate)(nil), "rdma_create_076fe418")
	api.RegisterMessage((*RdmaCreateReply)(nil), "rdma_create_reply_5383d31f")
	api.RegisterMessage((*RdmaCreateV2)(nil), "rdma_create_v2_5826a4f3")
	api.RegisterMessage((*RdmaCreateV2Reply)(nil), "rdma_create_v2_reply_5383d31f")
	api.RegisterMessage((*RdmaCreateV3)(nil), "rdma_create_v3_c6287ea8")
	api.RegisterMessage((*RdmaCreateV3Reply)(nil), "rdma_create_v3_reply_5383d31f")
	api.RegisterMessage((*RdmaCreateV4)(nil), "rdma_create_v4_c6287ea8")
	api.RegisterMessage((*RdmaCreateV4Reply)(nil), "rdma_create_v4_reply_5383d31f")
	api.RegisterMessage((*RdmaDelete)(nil), "rdma_delete_f9e6675e")
	api.RegisterMessage((*RdmaDeleteReply)(nil), "rdma_delete_reply_e8d4e804")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*RdmaCreate)(nil),
		(*RdmaCreateReply)(nil),
		(*RdmaCreateV2)(nil),
		(*RdmaCreateV2Reply)(nil),
		(*RdmaCreateV3)(nil),
		(*RdmaCreateV3Reply)(nil),
		(*RdmaCreateV4)(nil),
		(*RdmaCreateV4Reply)(nil),
		(*RdmaDelete)(nil),
		(*RdmaDeleteReply)(nil),
	}
}
