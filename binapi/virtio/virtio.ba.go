// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.11.0
//  VPP:              25.02-release
// source: core/virtio.api.json

// Package virtio contains generated bindings for API file virtio.api.
//
// Contents:
// -  1 enum
// -  8 messages
package virtio

import (
	"strconv"

	api "go.fd.io/govpp/api"
	ethernet_types "go.fd.io/govpp/binapi/ethernet_types"
	interface_types "go.fd.io/govpp/binapi/interface_types"
	pci_types "go.fd.io/govpp/binapi/pci_types"
	codec "go.fd.io/govpp/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "virtio"
	APIVersion = "3.0.0"
	VersionCrc = 0xa507d784
)

// VirtioFlags defines enum 'virtio_flags'.
type VirtioFlags uint32

const (
	VIRTIO_API_FLAG_GSO          VirtioFlags = 1
	VIRTIO_API_FLAG_CSUM_OFFLOAD VirtioFlags = 2
	VIRTIO_API_FLAG_GRO_COALESCE VirtioFlags = 4
	VIRTIO_API_FLAG_PACKED       VirtioFlags = 8
	VIRTIO_API_FLAG_IN_ORDER     VirtioFlags = 16
	VIRTIO_API_FLAG_BUFFERING    VirtioFlags = 32
	VIRTIO_API_FLAG_RSS          VirtioFlags = 64
)

var (
	VirtioFlags_name = map[uint32]string{
		1:  "VIRTIO_API_FLAG_GSO",
		2:  "VIRTIO_API_FLAG_CSUM_OFFLOAD",
		4:  "VIRTIO_API_FLAG_GRO_COALESCE",
		8:  "VIRTIO_API_FLAG_PACKED",
		16: "VIRTIO_API_FLAG_IN_ORDER",
		32: "VIRTIO_API_FLAG_BUFFERING",
		64: "VIRTIO_API_FLAG_RSS",
	}
	VirtioFlags_value = map[string]uint32{
		"VIRTIO_API_FLAG_GSO":          1,
		"VIRTIO_API_FLAG_CSUM_OFFLOAD": 2,
		"VIRTIO_API_FLAG_GRO_COALESCE": 4,
		"VIRTIO_API_FLAG_PACKED":       8,
		"VIRTIO_API_FLAG_IN_ORDER":     16,
		"VIRTIO_API_FLAG_BUFFERING":    32,
		"VIRTIO_API_FLAG_RSS":          64,
	}
)

func (x VirtioFlags) String() string {
	s, ok := VirtioFlags_name[uint32(x)]
	if ok {
		return s
	}
	str := func(n uint32) string {
		s, ok := VirtioFlags_name[uint32(n)]
		if ok {
			return s
		}
		return "VirtioFlags(" + strconv.Itoa(int(n)) + ")"
	}
	for i := uint32(0); i <= 32; i++ {
		val := uint32(x)
		if val&(1<<i) != 0 {
			if s != "" {
				s += "|"
			}
			s += str(1 << i)
		}
	}
	if s == "" {
		return str(uint32(x))
	}
	return s
}

// Reply for virtio pci interface dump request
//   - sw_if_index - software index of virtio pci interface
//   - pci_addr - pci address
//   - mac_addr - native virtio device mac address
//   - tx_ring_sz - the number of entries of TX ring
//   - rx_ring_sz - the number of entries of RX ring
//   - features - the virtio features which driver have negotiated with device
//
// SwInterfaceVirtioPciDetails defines message 'sw_interface_virtio_pci_details'.
type SwInterfaceVirtioPciDetails struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	PciAddr   pci_types.PciAddress           `binapi:"pci_address,name=pci_addr" json:"pci_addr,omitempty"`
	MacAddr   ethernet_types.MacAddress      `binapi:"mac_address,name=mac_addr" json:"mac_addr,omitempty"`
	TxRingSz  uint16                         `binapi:"u16,name=tx_ring_sz" json:"tx_ring_sz,omitempty"`
	RxRingSz  uint16                         `binapi:"u16,name=rx_ring_sz" json:"rx_ring_sz,omitempty"`
	Features  uint64                         `binapi:"u64,name=features" json:"features,omitempty"`
}

func (m *SwInterfaceVirtioPciDetails) Reset()               { *m = SwInterfaceVirtioPciDetails{} }
func (*SwInterfaceVirtioPciDetails) GetMessageName() string { return "sw_interface_virtio_pci_details" }
func (*SwInterfaceVirtioPciDetails) GetCrcString() string   { return "6ca9c167" }
func (*SwInterfaceVirtioPciDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwInterfaceVirtioPciDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4     // m.SwIfIndex
	size += 2     // m.PciAddr.Domain
	size += 1     // m.PciAddr.Bus
	size += 1     // m.PciAddr.Slot
	size += 1     // m.PciAddr.Function
	size += 1 * 6 // m.MacAddr
	size += 2     // m.TxRingSz
	size += 2     // m.RxRingSz
	size += 8     // m.Features
	return size
}
func (m *SwInterfaceVirtioPciDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint16(m.PciAddr.Domain)
	buf.EncodeUint8(m.PciAddr.Bus)
	buf.EncodeUint8(m.PciAddr.Slot)
	buf.EncodeUint8(m.PciAddr.Function)
	buf.EncodeBytes(m.MacAddr[:], 6)
	buf.EncodeUint16(m.TxRingSz)
	buf.EncodeUint16(m.RxRingSz)
	buf.EncodeUint64(m.Features)
	return buf.Bytes(), nil
}
func (m *SwInterfaceVirtioPciDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.PciAddr.Domain = buf.DecodeUint16()
	m.PciAddr.Bus = buf.DecodeUint8()
	m.PciAddr.Slot = buf.DecodeUint8()
	m.PciAddr.Function = buf.DecodeUint8()
	copy(m.MacAddr[:], buf.DecodeBytes(6))
	m.TxRingSz = buf.DecodeUint16()
	m.RxRingSz = buf.DecodeUint16()
	m.Features = buf.DecodeUint64()
	return nil
}

// Dump virtio pci interfaces request
// SwInterfaceVirtioPciDump defines message 'sw_interface_virtio_pci_dump'.
type SwInterfaceVirtioPciDump struct{}

func (m *SwInterfaceVirtioPciDump) Reset()               { *m = SwInterfaceVirtioPciDump{} }
func (*SwInterfaceVirtioPciDump) GetMessageName() string { return "sw_interface_virtio_pci_dump" }
func (*SwInterfaceVirtioPciDump) GetCrcString() string   { return "51077d14" }
func (*SwInterfaceVirtioPciDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwInterfaceVirtioPciDump) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *SwInterfaceVirtioPciDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *SwInterfaceVirtioPciDump) Unmarshal(b []byte) error {
	return nil
}

// Initialize a new virtio pci interface with the given parameters
//   - pci_addr - pci address
//   - use_random_mac - let the system generate a unique mac address
//   - mac_address - mac addr to assign to the interface if use_random not set
//   - gso_enabled - enable gso feature if available, 1 to enable
//   - checksum_offload_enabled - enable checksum feature if available, 1 to enable
//   - features - the virtio features which driver should negotiate with device
//
// VirtioPciCreate defines message 'virtio_pci_create'.
// Deprecated: the message will be removed in the future versions
type VirtioPciCreate struct {
	PciAddr                pci_types.PciAddress      `binapi:"pci_address,name=pci_addr" json:"pci_addr,omitempty"`
	UseRandomMac           bool                      `binapi:"bool,name=use_random_mac" json:"use_random_mac,omitempty"`
	MacAddress             ethernet_types.MacAddress `binapi:"mac_address,name=mac_address" json:"mac_address,omitempty"`
	GsoEnabled             bool                      `binapi:"bool,name=gso_enabled" json:"gso_enabled,omitempty"`
	ChecksumOffloadEnabled bool                      `binapi:"bool,name=checksum_offload_enabled" json:"checksum_offload_enabled,omitempty"`
	Features               uint64                    `binapi:"u64,name=features" json:"features,omitempty"`
}

func (m *VirtioPciCreate) Reset()               { *m = VirtioPciCreate{} }
func (*VirtioPciCreate) GetMessageName() string { return "virtio_pci_create" }
func (*VirtioPciCreate) GetCrcString() string   { return "1944f8db" }
func (*VirtioPciCreate) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *VirtioPciCreate) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 2     // m.PciAddr.Domain
	size += 1     // m.PciAddr.Bus
	size += 1     // m.PciAddr.Slot
	size += 1     // m.PciAddr.Function
	size += 1     // m.UseRandomMac
	size += 1 * 6 // m.MacAddress
	size += 1     // m.GsoEnabled
	size += 1     // m.ChecksumOffloadEnabled
	size += 8     // m.Features
	return size
}
func (m *VirtioPciCreate) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint16(m.PciAddr.Domain)
	buf.EncodeUint8(m.PciAddr.Bus)
	buf.EncodeUint8(m.PciAddr.Slot)
	buf.EncodeUint8(m.PciAddr.Function)
	buf.EncodeBool(m.UseRandomMac)
	buf.EncodeBytes(m.MacAddress[:], 6)
	buf.EncodeBool(m.GsoEnabled)
	buf.EncodeBool(m.ChecksumOffloadEnabled)
	buf.EncodeUint64(m.Features)
	return buf.Bytes(), nil
}
func (m *VirtioPciCreate) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.PciAddr.Domain = buf.DecodeUint16()
	m.PciAddr.Bus = buf.DecodeUint8()
	m.PciAddr.Slot = buf.DecodeUint8()
	m.PciAddr.Function = buf.DecodeUint8()
	m.UseRandomMac = buf.DecodeBool()
	copy(m.MacAddress[:], buf.DecodeBytes(6))
	m.GsoEnabled = buf.DecodeBool()
	m.ChecksumOffloadEnabled = buf.DecodeBool()
	m.Features = buf.DecodeUint64()
	return nil
}

// Reply for virtio pci create reply
//   - retval - return code
//   - sw_if_index - software index allocated for the new virtio pci interface
//
// VirtioPciCreateReply defines message 'virtio_pci_create_reply'.
// Deprecated: the message will be removed in the future versions
type VirtioPciCreateReply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *VirtioPciCreateReply) Reset()               { *m = VirtioPciCreateReply{} }
func (*VirtioPciCreateReply) GetMessageName() string { return "virtio_pci_create_reply" }
func (*VirtioPciCreateReply) GetCrcString() string   { return "5383d31f" }
func (*VirtioPciCreateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *VirtioPciCreateReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *VirtioPciCreateReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *VirtioPciCreateReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Initialize a new virtio pci interface with the given parameters
//   - pci_addr - pci address
//   - use_random_mac - let the system generate a unique mac address
//   - mac_address - mac addr to assign to the interface if use_random not set
//   - virtio_flags - feature flags to enable
//   - features - the virtio features which driver should negotiate with device
//
// VirtioPciCreateV2 defines message 'virtio_pci_create_v2'.
type VirtioPciCreateV2 struct {
	PciAddr      pci_types.PciAddress      `binapi:"pci_address,name=pci_addr" json:"pci_addr,omitempty"`
	UseRandomMac bool                      `binapi:"bool,name=use_random_mac" json:"use_random_mac,omitempty"`
	MacAddress   ethernet_types.MacAddress `binapi:"mac_address,name=mac_address" json:"mac_address,omitempty"`
	VirtioFlags  VirtioFlags               `binapi:"virtio_flags,name=virtio_flags" json:"virtio_flags,omitempty"`
	Features     uint64                    `binapi:"u64,name=features" json:"features,omitempty"`
}

func (m *VirtioPciCreateV2) Reset()               { *m = VirtioPciCreateV2{} }
func (*VirtioPciCreateV2) GetMessageName() string { return "virtio_pci_create_v2" }
func (*VirtioPciCreateV2) GetCrcString() string   { return "5d096e1a" }
func (*VirtioPciCreateV2) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *VirtioPciCreateV2) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 2     // m.PciAddr.Domain
	size += 1     // m.PciAddr.Bus
	size += 1     // m.PciAddr.Slot
	size += 1     // m.PciAddr.Function
	size += 1     // m.UseRandomMac
	size += 1 * 6 // m.MacAddress
	size += 4     // m.VirtioFlags
	size += 8     // m.Features
	return size
}
func (m *VirtioPciCreateV2) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint16(m.PciAddr.Domain)
	buf.EncodeUint8(m.PciAddr.Bus)
	buf.EncodeUint8(m.PciAddr.Slot)
	buf.EncodeUint8(m.PciAddr.Function)
	buf.EncodeBool(m.UseRandomMac)
	buf.EncodeBytes(m.MacAddress[:], 6)
	buf.EncodeUint32(uint32(m.VirtioFlags))
	buf.EncodeUint64(m.Features)
	return buf.Bytes(), nil
}
func (m *VirtioPciCreateV2) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.PciAddr.Domain = buf.DecodeUint16()
	m.PciAddr.Bus = buf.DecodeUint8()
	m.PciAddr.Slot = buf.DecodeUint8()
	m.PciAddr.Function = buf.DecodeUint8()
	m.UseRandomMac = buf.DecodeBool()
	copy(m.MacAddress[:], buf.DecodeBytes(6))
	m.VirtioFlags = VirtioFlags(buf.DecodeUint32())
	m.Features = buf.DecodeUint64()
	return nil
}

// Reply for virtio pci create reply
//   - retval - return code
//   - sw_if_index - software index allocated for the new virtio pci interface
//
// VirtioPciCreateV2Reply defines message 'virtio_pci_create_v2_reply'.
type VirtioPciCreateV2Reply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *VirtioPciCreateV2Reply) Reset()               { *m = VirtioPciCreateV2Reply{} }
func (*VirtioPciCreateV2Reply) GetMessageName() string { return "virtio_pci_create_v2_reply" }
func (*VirtioPciCreateV2Reply) GetCrcString() string   { return "5383d31f" }
func (*VirtioPciCreateV2Reply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *VirtioPciCreateV2Reply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *VirtioPciCreateV2Reply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *VirtioPciCreateV2Reply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Delete virtio pci interface
//   - sw_if_index - interface index of existing virtio pci interface
//
// VirtioPciDelete defines message 'virtio_pci_delete'.
type VirtioPciDelete struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *VirtioPciDelete) Reset()               { *m = VirtioPciDelete{} }
func (*VirtioPciDelete) GetMessageName() string { return "virtio_pci_delete" }
func (*VirtioPciDelete) GetCrcString() string   { return "f9e6675e" }
func (*VirtioPciDelete) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *VirtioPciDelete) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *VirtioPciDelete) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *VirtioPciDelete) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// VirtioPciDeleteReply defines message 'virtio_pci_delete_reply'.
type VirtioPciDeleteReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *VirtioPciDeleteReply) Reset()               { *m = VirtioPciDeleteReply{} }
func (*VirtioPciDeleteReply) GetMessageName() string { return "virtio_pci_delete_reply" }
func (*VirtioPciDeleteReply) GetCrcString() string   { return "e8d4e804" }
func (*VirtioPciDeleteReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *VirtioPciDeleteReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *VirtioPciDeleteReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *VirtioPciDeleteReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

func init() { file_virtio_binapi_init() }
func file_virtio_binapi_init() {
	api.RegisterMessage((*SwInterfaceVirtioPciDetails)(nil), "sw_interface_virtio_pci_details_6ca9c167")
	api.RegisterMessage((*SwInterfaceVirtioPciDump)(nil), "sw_interface_virtio_pci_dump_51077d14")
	api.RegisterMessage((*VirtioPciCreate)(nil), "virtio_pci_create_1944f8db")
	api.RegisterMessage((*VirtioPciCreateReply)(nil), "virtio_pci_create_reply_5383d31f")
	api.RegisterMessage((*VirtioPciCreateV2)(nil), "virtio_pci_create_v2_5d096e1a")
	api.RegisterMessage((*VirtioPciCreateV2Reply)(nil), "virtio_pci_create_v2_reply_5383d31f")
	api.RegisterMessage((*VirtioPciDelete)(nil), "virtio_pci_delete_f9e6675e")
	api.RegisterMessage((*VirtioPciDeleteReply)(nil), "virtio_pci_delete_reply_e8d4e804")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*SwInterfaceVirtioPciDetails)(nil),
		(*SwInterfaceVirtioPciDump)(nil),
		(*VirtioPciCreate)(nil),
		(*VirtioPciCreateReply)(nil),
		(*VirtioPciCreateV2)(nil),
		(*VirtioPciCreateV2Reply)(nil),
		(*VirtioPciDelete)(nil),
		(*VirtioPciDeleteReply)(nil),
	}
}
