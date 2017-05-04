// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ifcounters

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/lunixbochs/struc"
)

// VnetInterfaceCounters is the input data type defined in the 'interface.api', with binary encoded Data field,
// that can be decoded into the InterfaceCounter or CombinedInterfaceCounter struct using this package.
type VnetInterfaceCounters struct {
	VnetCounterType uint8
	IsCombined      uint8
	FirstSwIfIndex  uint32
	Count           uint32 `struc:"sizeof=Data"`
	Data            []byte
}

// CounterType is the basic counter type - contains only packet statistics.
type CounterType int

// constants as defined in the vnet_interface_counter_type_t enum in 'vnet/interface.h'
const (
	Drop    CounterType = 0
	Punt                = 1
	IPv4                = 2
	IPv6                = 3
	RxNoBuf             = 4
	RxMiss              = 5
	RxError             = 6
	TxError             = 7
	MPLS                = 8
)

// CombinedCounterType is the extended counter type - contains both packet and byte statistics.
type CombinedCounterType int

// constants as defined in the vnet_interface_counter_type_t enum in 'vnet/interface.h'
const (
	Rx CombinedCounterType = 0
	Tx                     = 1
)

// InterfaceCounter contains basic counter data (contains only packet statistics).
type InterfaceCounter struct {
	Type      CounterType
	SwIfIndex uint32
	Packets   uint64
}

// CombinedInterfaceCounter contains extended counter data (contains both packet and byte statistics).
type CombinedInterfaceCounter struct {
	Type      CombinedCounterType
	SwIfIndex uint32
	Packets   uint64
	Bytes     uint64
}

type counterData struct {
	Packets uint64
}

type counter struct {
	Count uint32 `struc:"sizeof=Data"`
	Data  []counterData
}

type combinedCounterData struct {
	Packets uint64
	Bytes   uint64
}

type combinedCounter struct {
	Count uint32 `struc:"sizeof=Data"`
	Data  []combinedCounterData
}

// DecodeCounters decodes VnetInterfaceCounters struct content into the slice of InterfaceCounter structs.
func DecodeCounters(vnetCounters VnetInterfaceCounters) ([]InterfaceCounter, error) {
	if vnetCounters.IsCombined == 1 {
		return nil, errors.New("invalid argument - combined counter passed in")
	}

	// decode into internal struct
	var c counter
	buf := bytes.NewReader(vnetCounters.Data)
	err := struc.Unpack(buf, &c)
	if err != nil {
		return nil, fmt.Errorf("unable to decode counter data: %v", err)
	}

	// prepare the slice
	res := make([]InterfaceCounter, c.Count)

	// fill in the slice
	for i := uint32(0); i < c.Count; i++ {
		res[i].Type = CounterType(vnetCounters.VnetCounterType)
		res[i].SwIfIndex = vnetCounters.FirstSwIfIndex + i
		res[i].Packets = c.Data[i].Packets
	}

	return res, nil
}

// DecodeCombinedCounters decodes VnetInterfaceCounters struct content into the slice of CombinedInterfaceCounter structs.
func DecodeCombinedCounters(vnetCounters VnetInterfaceCounters) ([]CombinedInterfaceCounter, error) {
	if vnetCounters.IsCombined != 1 {
		return nil, errors.New("invalid argument - simple counter passed in")
	}

	// decode into internal struct
	var c combinedCounter
	buf := bytes.NewReader(vnetCounters.Data)
	err := struc.Unpack(buf, &c)
	if err != nil {
		return nil, fmt.Errorf("unable to decode counter data: %v", err)
	}

	// prepare the slice
	res := make([]CombinedInterfaceCounter, c.Count)

	// fill in the slice
	for i := uint32(0); i < c.Count; i++ {
		res[i].Type = CombinedCounterType(vnetCounters.VnetCounterType)
		res[i].SwIfIndex = vnetCounters.FirstSwIfIndex + i
		res[i].Packets = c.Data[i].Packets
		res[i].Bytes = c.Data[i].Bytes
	}

	return res, nil
}
