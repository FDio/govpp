// Package uflow represents the VPP binary API of the 'uflow' VPP module.
// DO NOT EDIT. Generated from 'bin_api/uflow.api.json' on Thu, 04 May 2017 13:11:57 CEST.
package uflow

import "gerrit.fd.io/r/govpp.git/api"

// VlApiVersion contains version of the API.
const VlAPIVersion = 0x85909300

// UflowIdx represents the VPP binary API data type 'uflow_idx'.
// Generated from 'bin_api/uflow.api.json', line 3:
//
//        ["uflow_idx",
//            ["u32", "vslot"],
//            ["u32", "md"],
//            ["u32", "sid"],
//            {"crc" : "0x3310d92c"}
//        ],
//
type UflowIdx struct {
	Vslot uint32
	Md    uint32
	Sid   uint32
}

func (*UflowIdx) GetTypeName() string {
	return "uflow_idx"
}
func (*UflowIdx) GetCrcString() string {
	return "3310d92c"
}

// UflowEnt represents the VPP binary API data type 'uflow_ent'.
// Generated from 'bin_api/uflow.api.json', line 9:
//
//        ["uflow_ent",
//            ["u32", "cm_dpidx"],
//            ["u32", "vbundle_dpidx"],
//            {"crc" : "0x50fa3f43"}
//        ],
//
type UflowEnt struct {
	CmDpidx      uint32
	VbundleDpidx uint32
}

func (*UflowEnt) GetTypeName() string {
	return "uflow_ent"
}
func (*UflowEnt) GetCrcString() string {
	return "50fa3f43"
}

// UflowRow represents the VPP binary API data type 'uflow_row'.
// Generated from 'bin_api/uflow.api.json', line 14:
//
//        ["uflow_row",
//            ["vl_api_uflow_idx_t", "idx"],
//            ["vl_api_uflow_ent_t", "ent"],
//            {"crc" : "0x3b73b975"}
//        ]
//
type UflowRow struct {
	Idx UflowIdx
	Ent UflowEnt
}

func (*UflowRow) GetTypeName() string {
	return "uflow_row"
}
func (*UflowRow) GetCrcString() string {
	return "3b73b975"
}

// UflowEnableDisable represents the VPP binary API message 'uflow_enable_disable'.
// Generated from 'bin_api/uflow.api.json', line 21:
//
//        ["uflow_enable_disable",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["u32", "sw_if_index"],
//            ["u8", "enable_disable"],
//            {"crc" : "0x4c7f1b8a"}
//        ],
//
type UflowEnableDisable struct {
	SwIfIndex     uint32
	EnableDisable uint8
}

func (*UflowEnableDisable) GetMessageName() string {
	return "uflow_enable_disable"
}
func (*UflowEnableDisable) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*UflowEnableDisable) GetCrcString() string {
	return "4c7f1b8a"
}
func NewUflowEnableDisable() api.Message {
	return &UflowEnableDisable{}
}

// UflowEnableDisableReply represents the VPP binary API message 'uflow_enable_disable_reply'.
// Generated from 'bin_api/uflow.api.json', line 29:
//
//        ["uflow_enable_disable_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xf47b6600"}
//        ],
//
type UflowEnableDisableReply struct {
	Retval int32
}

func (*UflowEnableDisableReply) GetMessageName() string {
	return "uflow_enable_disable_reply"
}
func (*UflowEnableDisableReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*UflowEnableDisableReply) GetCrcString() string {
	return "f47b6600"
}
func NewUflowEnableDisableReply() api.Message {
	return &UflowEnableDisableReply{}
}

// UflowSetEnt represents the VPP binary API message 'uflow_set_ent'.
// Generated from 'bin_api/uflow.api.json', line 35:
//
//        ["uflow_set_ent",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["vl_api_uflow_idx_t", "idx"],
//            ["vl_api_uflow_ent_t", "ent"],
//            {"crc" : "0x6bfeac11"}
//        ],
//
type UflowSetEnt struct {
	Idx UflowIdx
	Ent UflowEnt
}

func (*UflowSetEnt) GetMessageName() string {
	return "uflow_set_ent"
}
func (*UflowSetEnt) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*UflowSetEnt) GetCrcString() string {
	return "6bfeac11"
}
func NewUflowSetEnt() api.Message {
	return &UflowSetEnt{}
}

// UflowSetEntReply represents the VPP binary API message 'uflow_set_ent_reply'.
// Generated from 'bin_api/uflow.api.json', line 43:
//
//        ["uflow_set_ent_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0xc49943f5"}
//        ],
//
type UflowSetEntReply struct {
	Retval int32
}

func (*UflowSetEntReply) GetMessageName() string {
	return "uflow_set_ent_reply"
}
func (*UflowSetEntReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*UflowSetEntReply) GetCrcString() string {
	return "c49943f5"
}
func NewUflowSetEntReply() api.Message {
	return &UflowSetEntReply{}
}

// UflowClrEnt represents the VPP binary API message 'uflow_clr_ent'.
// Generated from 'bin_api/uflow.api.json', line 49:
//
//        ["uflow_clr_ent",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            ["vl_api_uflow_idx_t", "idx"],
//            {"crc" : "0x9c0b61a7"}
//        ],
//
type UflowClrEnt struct {
	Idx UflowIdx
}

func (*UflowClrEnt) GetMessageName() string {
	return "uflow_clr_ent"
}
func (*UflowClrEnt) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*UflowClrEnt) GetCrcString() string {
	return "9c0b61a7"
}
func NewUflowClrEnt() api.Message {
	return &UflowClrEnt{}
}

// UflowClrEntReply represents the VPP binary API message 'uflow_clr_ent_reply'.
// Generated from 'bin_api/uflow.api.json', line 56:
//
//        ["uflow_clr_ent_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            {"crc" : "0x6ca429f7"}
//        ],
//
type UflowClrEntReply struct {
	Retval int32
}

func (*UflowClrEntReply) GetMessageName() string {
	return "uflow_clr_ent_reply"
}
func (*UflowClrEntReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*UflowClrEntReply) GetCrcString() string {
	return "6ca429f7"
}
func NewUflowClrEntReply() api.Message {
	return &UflowClrEntReply{}
}

// UflowDump represents the VPP binary API message 'uflow_dump'.
// Generated from 'bin_api/uflow.api.json', line 62:
//
//        ["uflow_dump",
//            ["u16", "_vl_msg_id"],
//            ["u32", "client_index"],
//            ["u32", "context"],
//            {"crc" : "0xf0ac7601"}
//        ],
//
type UflowDump struct {
}

func (*UflowDump) GetMessageName() string {
	return "uflow_dump"
}
func (*UflowDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}
func (*UflowDump) GetCrcString() string {
	return "f0ac7601"
}
func NewUflowDump() api.Message {
	return &UflowDump{}
}

// UflowDumpReply represents the VPP binary API message 'uflow_dump_reply'.
// Generated from 'bin_api/uflow.api.json', line 68:
//
//        ["uflow_dump_reply",
//            ["u16", "_vl_msg_id"],
//            ["u32", "context"],
//            ["i32", "retval"],
//            ["u32", "num"],
//            ["vl_api_uflow_row_t", "row", 0, "num"],
//            {"crc" : "0x85b96451"}
//        ]
//
type UflowDumpReply struct {
	Retval int32
	Num    uint32 `struc:"sizeof=Row"`
	Row    []UflowRow
}

func (*UflowDumpReply) GetMessageName() string {
	return "uflow_dump_reply"
}
func (*UflowDumpReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}
func (*UflowDumpReply) GetCrcString() string {
	return "85b96451"
}
func NewUflowDumpReply() api.Message {
	return &UflowDumpReply{}
}
