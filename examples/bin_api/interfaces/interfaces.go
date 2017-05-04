// Package interfaces represents the VPP binary API of the 'interfaces' VPP module.
// DO NOT EDIT. Generated from 'bin_api/interface.api.json' on Thu, 04 May 2017 13:11:57 CEST.
package interfaces

import "gerrit.fd.io/r/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0x6857f668

// SwInterfaceSetFlags represents the VPP binary API message 'sw_interface_set_flags'.
// Generated from 'bin_api/interface.api.json', line 6:
//
//        ["sw_interface_set_flags",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "admin_up_down"],
//            ["u8", "link_up_down"],
//            ["u8", "deleted"],
//            {"crc" : "0xc230f9b1"}
//        ],
//
type SwInterfaceSetFlags struct {
	SwIfIndex   uint32
	AdminUpDown uint8
	LinkUpDown  uint8
	Deleted     uint8
}

func (*SwInterfaceSetFlags) GetMessageName() string {
	return "sw_interface_set_flags"
}
func (*SwInterfaceSetFlags) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceSetFlags) GetCrcString() string {
	return "c230f9b1"
}
func NewSwInterfaceSetFlags() api.Message {
	return &SwInterfaceSetFlags{}
}

// SwInterfaceSetFlagsReply represents the VPP binary API message 'sw_interface_set_flags_reply'.
// Generated from 'bin_api/interface.api.json', line 16:
//
//        ["sw_interface_set_flags_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xdfbf3afa"}
//        ],
//
type SwInterfaceSetFlagsReply struct {
	Retval int32
}

func (*SwInterfaceSetFlagsReply) GetMessageName() string {
	return "sw_interface_set_flags_reply"
}
func (*SwInterfaceSetFlagsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceSetFlagsReply) GetCrcString() string {
	return "dfbf3afa"
}
func NewSwInterfaceSetFlagsReply() api.Message {
	return &SwInterfaceSetFlagsReply{}
}

// SwInterfaceSetMtu represents the VPP binary API message 'sw_interface_set_mtu'.
// Generated from 'bin_api/interface.api.json', line 22:
//
//        ["sw_interface_set_mtu",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u16", "mtu"],
//            {"crc" : "0x535dab1d"}
//        ],
//
type SwInterfaceSetMtu struct {
	SwIfIndex uint32
	Mtu       uint16
}

func (*SwInterfaceSetMtu) GetMessageName() string {
	return "sw_interface_set_mtu"
}
func (*SwInterfaceSetMtu) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceSetMtu) GetCrcString() string {
	return "535dab1d"
}
func NewSwInterfaceSetMtu() api.Message {
	return &SwInterfaceSetMtu{}
}

// SwInterfaceSetMtuReply represents the VPP binary API message 'sw_interface_set_mtu_reply'.
// Generated from 'bin_api/interface.api.json', line 30:
//
//        ["sw_interface_set_mtu_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x0cc22552"}
//        ],
//
type SwInterfaceSetMtuReply struct {
	Retval int32
}

func (*SwInterfaceSetMtuReply) GetMessageName() string {
	return "sw_interface_set_mtu_reply"
}
func (*SwInterfaceSetMtuReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceSetMtuReply) GetCrcString() string {
	return "0cc22552"
}
func NewSwInterfaceSetMtuReply() api.Message {
	return &SwInterfaceSetMtuReply{}
}

// WantInterfaceEvents represents the VPP binary API message 'want_interface_events'.
// Generated from 'bin_api/interface.api.json', line 36:
//
//        ["want_interface_events",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "enable_disable"],
//            ["u32", "pid"],
//            {"crc" : "0xa0cbf57e"}
//        ],
//
type WantInterfaceEvents struct {
	EnableDisable uint32
	Pid           uint32
}

func (*WantInterfaceEvents) GetMessageName() string {
	return "want_interface_events"
}
func (*WantInterfaceEvents) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*WantInterfaceEvents) GetCrcString() string {
	return "a0cbf57e"
}
func NewWantInterfaceEvents() api.Message {
	return &WantInterfaceEvents{}
}

// WantInterfaceEventsReply represents the VPP binary API message 'want_interface_events_reply'.
// Generated from 'bin_api/interface.api.json', line 44:
//
//        ["want_interface_events_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x33788c73"}
//        ],
//
type WantInterfaceEventsReply struct {
	Retval int32
}

func (*WantInterfaceEventsReply) GetMessageName() string {
	return "want_interface_events_reply"
}
func (*WantInterfaceEventsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*WantInterfaceEventsReply) GetCrcString() string {
	return "33788c73"
}
func NewWantInterfaceEventsReply() api.Message {
	return &WantInterfaceEventsReply{}
}

// SwInterfaceDetails represents the VPP binary API message 'sw_interface_details'.
// Generated from 'bin_api/interface.api.json', line 50:
//
//        ["sw_interface_details",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u32", "sup_sw_if_index"],
//            ["u32", "l2_address_length"],
//            ["u8", "l2_address", 8],
//            ["u8", "interface_name", 64],
//            ["u8", "admin_up_down"],
//            ["u8", "link_up_down"],
//            ["u8", "link_duplex"],
//            ["u8", "link_speed"],
//            ["u16", "link_mtu"],
//            ["u32", "sub_id"],
//            ["u8", "sub_dot1ad"],
//            ["u8", "sub_dot1ah"],
//            ["u8", "sub_number_of_tags"],
//            ["u16", "sub_outer_vlan_id"],
//            ["u16", "sub_inner_vlan_id"],
//            ["u8", "sub_exact_match"],
//            ["u8", "sub_default"],
//            ["u8", "sub_outer_vlan_id_any"],
//            ["u8", "sub_inner_vlan_id_any"],
//            ["u32", "vtr_op"],
//            ["u32", "vtr_push_dot1q"],
//            ["u32", "vtr_tag1"],
//            ["u32", "vtr_tag2"],
//            ["u8", "tag", 64],
//            ["u16", "outer_tag"],
//            ["u8", "b_dmac", 6],
//            ["u8", "b_smac", 6],
//            ["u16", "b_vlanid"],
//            ["u32", "i_sid"],
//            {"crc" : "0xe2d855bb"}
//        ],
//
type SwInterfaceDetails struct {
	SwIfIndex         uint32
	SupSwIfIndex      uint32
	L2AddressLength   uint32
	L2Address         []byte `struc:"[8]byte"`
	InterfaceName     []byte `struc:"[64]byte"`
	AdminUpDown       uint8
	LinkUpDown        uint8
	LinkDuplex        uint8
	LinkSpeed         uint8
	LinkMtu           uint16
	SubID             uint32
	SubDot1ad         uint8
	SubDot1ah         uint8
	SubNumberOfTags   uint8
	SubOuterVlanID    uint16
	SubInnerVlanID    uint16
	SubExactMatch     uint8
	SubDefault        uint8
	SubOuterVlanIDAny uint8
	SubInnerVlanIDAny uint8
	VtrOp             uint32
	VtrPushDot1q      uint32
	VtrTag1           uint32
	VtrTag2           uint32
	Tag               []byte `struc:"[64]byte"`
	OuterTag          uint16
	BDmac             []byte `struc:"[6]byte"`
	BSmac             []byte `struc:"[6]byte"`
	BVlanid           uint16
	ISid              uint32
}

func (*SwInterfaceDetails) GetMessageName() string {
	return "sw_interface_details"
}
func (*SwInterfaceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceDetails) GetCrcString() string {
	return "e2d855bb"
}
func NewSwInterfaceDetails() api.Message {
	return &SwInterfaceDetails{}
}

// SwInterfaceDump represents the VPP binary API message 'sw_interface_dump'.
// Generated from 'bin_api/interface.api.json', line 85:
//
//        ["sw_interface_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "name_filter_valid"],
//            ["u8", "name_filter", 49],
//            {"crc" : "0x9a2f9d4d"}
//        ],
//
type SwInterfaceDump struct {
	NameFilterValid uint8
	NameFilter      []byte `struc:"[49]byte"`
}

func (*SwInterfaceDump) GetMessageName() string {
	return "sw_interface_dump"
}
func (*SwInterfaceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceDump) GetCrcString() string {
	return "9a2f9d4d"
}
func NewSwInterfaceDump() api.Message {
	return &SwInterfaceDump{}
}

// SwInterfaceAddDelAddress represents the VPP binary API message 'sw_interface_add_del_address'.
// Generated from 'bin_api/interface.api.json', line 93:
//
//        ["sw_interface_add_del_address",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_add"],
//            ["u8", "is_ipv6"],
//            ["u8", "del_all"],
//            ["u8", "address_length"],
//            ["u8", "address", 16],
//            {"crc" : "0x4e24d2df"}
//        ],
//
type SwInterfaceAddDelAddress struct {
	SwIfIndex     uint32
	IsAdd         uint8
	IsIpv6        uint8
	DelAll        uint8
	AddressLength uint8
	Address       []byte `struc:"[16]byte"`
}

func (*SwInterfaceAddDelAddress) GetMessageName() string {
	return "sw_interface_add_del_address"
}
func (*SwInterfaceAddDelAddress) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceAddDelAddress) GetCrcString() string {
	return "4e24d2df"
}
func NewSwInterfaceAddDelAddress() api.Message {
	return &SwInterfaceAddDelAddress{}
}

// SwInterfaceAddDelAddressReply represents the VPP binary API message 'sw_interface_add_del_address_reply'.
// Generated from 'bin_api/interface.api.json', line 105:
//
//        ["sw_interface_add_del_address_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xabe29452"}
//        ],
//
type SwInterfaceAddDelAddressReply struct {
	Retval int32
}

func (*SwInterfaceAddDelAddressReply) GetMessageName() string {
	return "sw_interface_add_del_address_reply"
}
func (*SwInterfaceAddDelAddressReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceAddDelAddressReply) GetCrcString() string {
	return "abe29452"
}
func NewSwInterfaceAddDelAddressReply() api.Message {
	return &SwInterfaceAddDelAddressReply{}
}

// SwInterfaceSetTable represents the VPP binary API message 'sw_interface_set_table'.
// Generated from 'bin_api/interface.api.json', line 111:
//
//        ["sw_interface_set_table",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_ipv6"],
//            ["u32", "vrf_id"],
//            {"crc" : "0xa94df510"}
//        ],
//
type SwInterfaceSetTable struct {
	SwIfIndex uint32
	IsIpv6    uint8
	VrfID     uint32
}

func (*SwInterfaceSetTable) GetMessageName() string {
	return "sw_interface_set_table"
}
func (*SwInterfaceSetTable) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceSetTable) GetCrcString() string {
	return "a94df510"
}
func NewSwInterfaceSetTable() api.Message {
	return &SwInterfaceSetTable{}
}

// SwInterfaceSetTableReply represents the VPP binary API message 'sw_interface_set_table_reply'.
// Generated from 'bin_api/interface.api.json', line 120:
//
//        ["sw_interface_set_table_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x99df273c"}
//        ],
//
type SwInterfaceSetTableReply struct {
	Retval int32
}

func (*SwInterfaceSetTableReply) GetMessageName() string {
	return "sw_interface_set_table_reply"
}
func (*SwInterfaceSetTableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceSetTableReply) GetCrcString() string {
	return "99df273c"
}
func NewSwInterfaceSetTableReply() api.Message {
	return &SwInterfaceSetTableReply{}
}

// SwInterfaceGetTable represents the VPP binary API message 'sw_interface_get_table'.
// Generated from 'bin_api/interface.api.json', line 126:
//
//        ["sw_interface_get_table",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "is_ipv6"],
//            {"crc" : "0xf5a1d557"}
//        ],
//
type SwInterfaceGetTable struct {
	SwIfIndex uint32
	IsIpv6    uint8
}

func (*SwInterfaceGetTable) GetMessageName() string {
	return "sw_interface_get_table"
}
func (*SwInterfaceGetTable) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceGetTable) GetCrcString() string {
	return "f5a1d557"
}
func NewSwInterfaceGetTable() api.Message {
	return &SwInterfaceGetTable{}
}

// SwInterfaceGetTableReply represents the VPP binary API message 'sw_interface_get_table_reply'.
// Generated from 'bin_api/interface.api.json', line 134:
//
//        ["sw_interface_get_table_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u32", "vrf_id"],
//            {"crc" : "0xab44111d"}
//        ],
//
type SwInterfaceGetTableReply struct {
	Retval int32
	VrfID  uint32
}

func (*SwInterfaceGetTableReply) GetMessageName() string {
	return "sw_interface_get_table_reply"
}
func (*SwInterfaceGetTableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceGetTableReply) GetCrcString() string {
	return "ab44111d"
}
func NewSwInterfaceGetTableReply() api.Message {
	return &SwInterfaceGetTableReply{}
}

// VnetInterfaceCounters represents the VPP binary API message 'vnet_interface_counters'.
// Generated from 'bin_api/interface.api.json', line 141:
//
//        ["vnet_interface_counters",
//            ["u16", "_vl_msg_id"],
//            ["u8", "vnet_counter_type"],
//            ["u8", "is_combined"],
//            ["u32", "first_sw_if_index"],
//            ["u32", "count"],
//            ["u8", "data", 0, "count"],
//            {"crc" : "0x312082b4"}
//        ],
//
type VnetInterfaceCounters struct {
	VnetCounterType uint8
	IsCombined      uint8
	FirstSwIfIndex  uint32
	Count           uint32 `struc:"sizeof=Data"`
	Data            []byte
}

func (*VnetInterfaceCounters) GetMessageName() string {
	return "vnet_interface_counters"
}
func (*VnetInterfaceCounters) GetMessageType() api.MessageType {
	return api.OtherMessage
}
func (*VnetInterfaceCounters) GetCrcString() string {
	return "312082b4"
}
func NewVnetInterfaceCounters() api.Message {
	return &VnetInterfaceCounters{}
}

// SwInterfaceSetUnnumbered represents the VPP binary API message 'sw_interface_set_unnumbered'.
// Generated from 'bin_api/interface.api.json', line 150:
//
//        ["sw_interface_set_unnumbered",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u32", "unnumbered_sw_if_index"],
//            ["u8", "is_add"],
//            {"crc" : "0xee0047b0"}
//        ],
//
type SwInterfaceSetUnnumbered struct {
	SwIfIndex           uint32
	UnnumberedSwIfIndex uint32
	IsAdd               uint8
}

func (*SwInterfaceSetUnnumbered) GetMessageName() string {
	return "sw_interface_set_unnumbered"
}
func (*SwInterfaceSetUnnumbered) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceSetUnnumbered) GetCrcString() string {
	return "ee0047b0"
}
func NewSwInterfaceSetUnnumbered() api.Message {
	return &SwInterfaceSetUnnumbered{}
}

// SwInterfaceSetUnnumberedReply represents the VPP binary API message 'sw_interface_set_unnumbered_reply'.
// Generated from 'bin_api/interface.api.json', line 159:
//
//        ["sw_interface_set_unnumbered_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x5b2275e1"}
//        ],
//
type SwInterfaceSetUnnumberedReply struct {
	Retval int32
}

func (*SwInterfaceSetUnnumberedReply) GetMessageName() string {
	return "sw_interface_set_unnumbered_reply"
}
func (*SwInterfaceSetUnnumberedReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceSetUnnumberedReply) GetCrcString() string {
	return "5b2275e1"
}
func NewSwInterfaceSetUnnumberedReply() api.Message {
	return &SwInterfaceSetUnnumberedReply{}
}

// SwInterfaceClearStats represents the VPP binary API message 'sw_interface_clear_stats'.
// Generated from 'bin_api/interface.api.json', line 165:
//
//        ["sw_interface_clear_stats",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            {"crc" : "0x9600fd50"}
//        ],
//
type SwInterfaceClearStats struct {
	SwIfIndex uint32
}

func (*SwInterfaceClearStats) GetMessageName() string {
	return "sw_interface_clear_stats"
}
func (*SwInterfaceClearStats) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceClearStats) GetCrcString() string {
	return "9600fd50"
}
func NewSwInterfaceClearStats() api.Message {
	return &SwInterfaceClearStats{}
}

// SwInterfaceClearStatsReply represents the VPP binary API message 'sw_interface_clear_stats_reply'.
// Generated from 'bin_api/interface.api.json', line 172:
//
//        ["sw_interface_clear_stats_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x21f50dd9"}
//        ],
//
type SwInterfaceClearStatsReply struct {
	Retval int32
}

func (*SwInterfaceClearStatsReply) GetMessageName() string {
	return "sw_interface_clear_stats_reply"
}
func (*SwInterfaceClearStatsReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceClearStatsReply) GetCrcString() string {
	return "21f50dd9"
}
func NewSwInterfaceClearStatsReply() api.Message {
	return &SwInterfaceClearStatsReply{}
}

// SwInterfaceTagAddDel represents the VPP binary API message 'sw_interface_tag_add_del'.
// Generated from 'bin_api/interface.api.json', line 178:
//
//        ["sw_interface_tag_add_del",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u8", "is_add"],
//            ["u32", "sw_if_index"],
//            ["u8", "tag", 64],
//            {"crc" : "0x50ae8d92"}
//        ],
//
type SwInterfaceTagAddDel struct {
	IsAdd     uint8
	SwIfIndex uint32
	Tag       []byte `struc:"[64]byte"`
}

func (*SwInterfaceTagAddDel) GetMessageName() string {
	return "sw_interface_tag_add_del"
}
func (*SwInterfaceTagAddDel) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*SwInterfaceTagAddDel) GetCrcString() string {
	return "50ae8d92"
}
func NewSwInterfaceTagAddDel() api.Message {
	return &SwInterfaceTagAddDel{}
}

// SwInterfaceTagAddDelReply represents the VPP binary API message 'sw_interface_tag_add_del_reply'.
// Generated from 'bin_api/interface.api.json', line 187:
//
//        ["sw_interface_tag_add_del_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x761cbcb0"}
//        ]
//
type SwInterfaceTagAddDelReply struct {
	Retval int32
}

func (*SwInterfaceTagAddDelReply) GetMessageName() string {
	return "sw_interface_tag_add_del_reply"
}
func (*SwInterfaceTagAddDelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*SwInterfaceTagAddDelReply) GetCrcString() string {
	return "761cbcb0"
}
func NewSwInterfaceTagAddDelReply() api.Message {
	return &SwInterfaceTagAddDelReply{}
}
