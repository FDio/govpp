// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.4.0
//  VPP:              21.06-release
// source: /usr/share/vpp/api/core/sr_types.api.json

// Package sr_types contains generated bindings for API file sr_types.api.
//
// Contents:
//   3 enums
//
package sr_types

import (
	"strconv"

	api "git.fd.io/govpp.git/api"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

// SrBehavior defines enum 'sr_behavior'.
type SrBehavior uint8

const (
	SR_BEHAVIOR_API_END     SrBehavior = 1
	SR_BEHAVIOR_API_X       SrBehavior = 2
	SR_BEHAVIOR_API_T       SrBehavior = 3
	SR_BEHAVIOR_API_D_FIRST SrBehavior = 4
	SR_BEHAVIOR_API_DX2     SrBehavior = 5
	SR_BEHAVIOR_API_DX6     SrBehavior = 6
	SR_BEHAVIOR_API_DX4     SrBehavior = 7
	SR_BEHAVIOR_API_DT6     SrBehavior = 8
	SR_BEHAVIOR_API_DT4     SrBehavior = 9
	SR_BEHAVIOR_API_LAST    SrBehavior = 10
)

var (
	SrBehavior_name = map[uint8]string{
		1:  "SR_BEHAVIOR_API_END",
		2:  "SR_BEHAVIOR_API_X",
		3:  "SR_BEHAVIOR_API_T",
		4:  "SR_BEHAVIOR_API_D_FIRST",
		5:  "SR_BEHAVIOR_API_DX2",
		6:  "SR_BEHAVIOR_API_DX6",
		7:  "SR_BEHAVIOR_API_DX4",
		8:  "SR_BEHAVIOR_API_DT6",
		9:  "SR_BEHAVIOR_API_DT4",
		10: "SR_BEHAVIOR_API_LAST",
	}
	SrBehavior_value = map[string]uint8{
		"SR_BEHAVIOR_API_END":     1,
		"SR_BEHAVIOR_API_X":       2,
		"SR_BEHAVIOR_API_T":       3,
		"SR_BEHAVIOR_API_D_FIRST": 4,
		"SR_BEHAVIOR_API_DX2":     5,
		"SR_BEHAVIOR_API_DX6":     6,
		"SR_BEHAVIOR_API_DX4":     7,
		"SR_BEHAVIOR_API_DT6":     8,
		"SR_BEHAVIOR_API_DT4":     9,
		"SR_BEHAVIOR_API_LAST":    10,
	}
)

func (x SrBehavior) String() string {
	s, ok := SrBehavior_name[uint8(x)]
	if ok {
		return s
	}
	return "SrBehavior(" + strconv.Itoa(int(x)) + ")"
}

// SrPolicyOp defines enum 'sr_policy_op'.
type SrPolicyOp uint8

const (
	SR_POLICY_OP_API_NONE SrPolicyOp = 0
	SR_POLICY_OP_API_ADD  SrPolicyOp = 1
	SR_POLICY_OP_API_DEL  SrPolicyOp = 2
	SR_POLICY_OP_API_MOD  SrPolicyOp = 3
)

var (
	SrPolicyOp_name = map[uint8]string{
		0: "SR_POLICY_OP_API_NONE",
		1: "SR_POLICY_OP_API_ADD",
		2: "SR_POLICY_OP_API_DEL",
		3: "SR_POLICY_OP_API_MOD",
	}
	SrPolicyOp_value = map[string]uint8{
		"SR_POLICY_OP_API_NONE": 0,
		"SR_POLICY_OP_API_ADD":  1,
		"SR_POLICY_OP_API_DEL":  2,
		"SR_POLICY_OP_API_MOD":  3,
	}
)

func (x SrPolicyOp) String() string {
	s, ok := SrPolicyOp_name[uint8(x)]
	if ok {
		return s
	}
	return "SrPolicyOp(" + strconv.Itoa(int(x)) + ")"
}

// SrSteer defines enum 'sr_steer'.
type SrSteer uint8

const (
	SR_STEER_API_L2   SrSteer = 2
	SR_STEER_API_IPV4 SrSteer = 4
	SR_STEER_API_IPV6 SrSteer = 6
)

var (
	SrSteer_name = map[uint8]string{
		2: "SR_STEER_API_L2",
		4: "SR_STEER_API_IPV4",
		6: "SR_STEER_API_IPV6",
	}
	SrSteer_value = map[string]uint8{
		"SR_STEER_API_L2":   2,
		"SR_STEER_API_IPV4": 4,
		"SR_STEER_API_IPV6": 6,
	}
)

func (x SrSteer) String() string {
	s, ok := SrSteer_name[uint8(x)]
	if ok {
		return s
	}
	return "SrSteer(" + strconv.Itoa(int(x)) + ")"
}
