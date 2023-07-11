// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.8.0
//  VPP:              23.06-release
// source: core/pci_types.api.json

// Package pci_types contains generated bindings for API file pci_types.api.
//
// Contents:
// -  1 struct
package pci_types

import (
	api "go.fd.io/govpp/api"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "pci_types"
	APIVersion = "1.0.0"
	VersionCrc = 0x5d418665
)

// PciAddress defines type 'pci_address'.
type PciAddress struct {
	Domain   uint16 `binapi:"u16,name=domain" json:"domain,omitempty"`
	Bus      uint8  `binapi:"u8,name=bus" json:"bus,omitempty"`
	Slot     uint8  `binapi:"u8,name=slot" json:"slot,omitempty"`
	Function uint8  `binapi:"u8,name=function" json:"function,omitempty"`
}
