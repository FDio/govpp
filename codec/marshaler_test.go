//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package codec_test

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/lunixbochs/struc"

	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/codec"
	"git.fd.io/govpp.git/codec/testdata/binapi2001/interfaces"
	"git.fd.io/govpp.git/codec/testdata/binapi2001/ip"
)

// CliInband represents VPP binary API message 'cli_inband'.
type CliInband struct {
	XXX_CmdLen uint32 `struc:"sizeof=Cmd"`
	Cmd        string
}

func (m *CliInband) Reset()                        { *m = CliInband{} }
func (*CliInband) GetMessageName() string          { return "cli_inband" }
func (*CliInband) GetCrcString() string            { return "f8377302" }
func (*CliInband) GetMessageType() api.MessageType { return api.RequestMessage }

// CliInbandReply represents VPP binary API message 'cli_inband_reply'.
type CliInbandReply struct {
	Retval       int32
	XXX_ReplyLen uint32 `struc:"sizeof=Reply"`
	Reply        string
}

func (m *CliInbandReply) Reset()                        { *m = CliInbandReply{} }
func (*CliInbandReply) GetMessageName() string          { return "cli_inband_reply" }
func (*CliInbandReply) GetCrcString() string            { return "05879051" }
func (*CliInbandReply) GetMessageType() api.MessageType { return api.ReplyMessage }

func TestWrapperEncode(t *testing.T) {
	msg := &CliInband{
		XXX_CmdLen: 5,
		Cmd:        "abcde",
	}
	expectedData := []byte{
		0x00, 0x64,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x05,
		0x61, 0x62, 0x63, 0x64, 0x65,
	}

	c := codec.DefaultCodec

	data, err := c.EncodeMsg(msg, 100)
	if err != nil {
		t.Fatalf("EncodeMsg failed: %v", err)
	}
	if !bytes.Equal(data, expectedData) {
		t.Fatalf("unexpected encoded data,\nexpected: % 02x\n     got: % 02x\n", expectedData, data)
	}
}

func TestWrapperDecode(t *testing.T) {
	data := []byte{
		0x00, 0x64,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x05,
		0x61, 0x62, 0x63, 0x64, 0x65,
	}
	expectedMsg := &CliInbandReply{
		Retval:       0,
		XXX_ReplyLen: 5,
		Reply:        "abcde",
	}

	c := codec.DefaultCodec

	msg := new(CliInbandReply)
	err := c.DecodeMsg(data, msg)
	if err != nil {
		t.Fatalf("DecodeMsg failed: %v", err)
	}
	if !reflect.DeepEqual(msg, expectedMsg) {
		t.Fatalf("unexpected decoded msg,\nexpected: %+v\n     got: %+v\n", expectedMsg, msg)
	}
}

/*func TestNewCodecEncodeDecode(t *testing.T) {
	tests := []struct {
		name string
		msg  Codec
	}{
		{
			"", &TestAllMsg{
				Bool:        true,
				AliasUint32: 5,
				AliasArray:  MacAddress{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
				BaseArray:   []uint32{0x00, 0x00, 0x00, 0x00},
				Enum:        IF_STATUS_API_FLAG_LINK_UP,
				Uint8:       8,
				Uint16:      16,
				Uint32:      32,
				Int8:        88,
				Int16:       1616,
				Int32:       3232,
				Slice:       []byte{10, 20, 30, 40, 0, 0, 0},
				String:      "abcdefghikl",
				SizeOf:      2,
				VariableSlice: []SliceType{
					{Proto: IP_API_PROTO_AH},
					{Proto: IP_API_PROTO_ESP},
				},
				TypeUnion: Address{
					Af: ADDRESS_IP4,
					Un: AddressUnionIP4(IP4Address{1, 2, 3, 4}),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := test.msg.Marshal()
			if err != nil {
				t.Fatalf("expected nil error, got: %v", err)
			}

			var m2 TestAllMsg
			if err := m2.Unmarshal(data); err != nil {
				t.Fatalf("expected nil error, got: %v", err)
			}

			t.Logf("Data:\nOLD: %+v\nNEW: %+v", m, &m2)

			if !reflect.DeepEqual(m, &m2) {
				t.Fatalf("newData differs from oldData")
			}
		})
	}
}*/

func NewTestAllMsg() *TestAllMsg {
	return &TestAllMsg{
		Bool:        true,
		AliasUint32: 5,
		AliasArray:  MacAddress{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
		BaseArray:   []uint32{0x00, 0x00, 0x00, 0x00},
		Enum:        IF_STATUS_API_FLAG_LINK_UP,
		Uint8:       8,
		Uint16:      16,
		Uint32:      32,
		Int8:        88,
		Int16:       1616,
		Int32:       3232,
		Slice:       []byte{10, 20, 30, 40, 0, 0, 0},
		String:      "abcdefghikl",
		SizeOf:      2,
		VariableSlice: []SliceType{
			{Proto: IP_API_PROTO_AH},
			{Proto: IP_API_PROTO_ESP},
		},
		TypeUnion: Address{
			Af: ADDRESS_IP4,
			Un: AddressUnionIP4(IP4Address{1, 2, 3, 4}),
		},
	}
}

func TestNewCodecEncodeDecode_(t *testing.T) {
	m := NewTestAllMsg()

	data, err := m.Marshal(nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	var m2 TestAllMsg
	if err := m2.Unmarshal(data); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	t.Logf("Data:\nOLD: %+v\nNEW: %+v", m, &m2)

	if !reflect.DeepEqual(m, &m2) {
		t.Fatalf("newData differs from oldData")
	}
}

// -------------

func TestNewCodecEncodeDecode3(t *testing.T) {
	m := NewTestAllMsg()

	data, err := m.Marshal(nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	var m2 TestAllMsg
	if err := m2.Unmarshal(data); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	t.Logf("Data:\nOLD: %+v\nNEW: %+v", m, &m2)

	if !reflect.DeepEqual(m, &m2) {
		t.Fatalf("newData differs from oldData")
	}
}
func TestNewCodecEncodeDecode4(t *testing.T) {
	m := &interfaces.SwInterfaceSetRxMode{
		Mode:         interfaces.RX_MODE_API_POLLING,
		QueueID:      70000,
		QueueIDValid: true,
		SwIfIndex:    300,
	}

	b := make([]byte, 2+m.Size())

	data, err := m.Marshal(b[2:])
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	t.Logf("ENCODED DATA(%d): % 03x", len(data), data)

	var m2 interfaces.SwInterfaceSetRxMode
	if err := m2.Unmarshal(b[2:]); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	t.Logf("Data:\nOLD: %+v\nNEW: %+v", m, &m2)

	if !reflect.DeepEqual(m, &m2) {
		t.Fatalf("newData differs from oldData")
	}
}

/*func TestNewCodecEncodeDecode2(t *testing.T) {
	m := &sr.SrPoliciesDetails{
		Bsid:        sr.IP6Address{00, 11, 22, 33, 44, 55, 66, 77, 88, 99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		IsSpray:     true,
		IsEncap:     false,
		FibTable:    33,
		NumSidLists: 1,
		SidLists: []sr.Srv6SidList{
			{
				Weight:  555,
				NumSids: 2,
				Sids: [16]sr.IP6Address{
					{99},
					{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				},
			},
		},
	}

	b := make([]byte, m.Size())
	data, err := m.Marshal(b)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	t.Logf("ENCODED DATA(%d): % 03x", len(data), data)

	var m2 sr.SrPoliciesDetails
	if err := m2.Unmarshal(data); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	t.Logf("Data:\nOLD: %+v\nNEW: %+v", m, &m2)

	if !reflect.DeepEqual(m, &m2) {
		t.Fatalf("newData differs from oldData")
	}
}*/

func TestNewCodecEncode(t *testing.T) {
	m := NewIPRouteLookupReply()
	/*m := &sr.SrPoliciesDetails{
		Bsid:        sr.IP6Address{00, 11, 22, 33, 44, 55, 66, 77, 88, 99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		IsSpray:     true,
		IsEncap:     false,
		FibTable:    33,
		NumSidLists: 1,
		SidLists: []sr.Srv6SidList{
			{
				Weight:  555,
				NumSids: 2,
				Sids: [16]sr.IP6Address{
					{99},
					{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				},
			},
		},
	}*/

	var err error
	var oldData, newData []byte
	{
		w := codec.Wrapper{m}
		size := m.Size()
		t.Logf("wrapper size: %d", size)
		oldData, err = w.Marshal(nil)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	}
	{
		size := m.Size()
		t.Logf("size: %d", size)
		newData, err = m.Marshal(nil)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	}
	t.Logf("Data:\nOLD[%d]: % 03x\nNEW[%d]: % 03x", len(oldData), oldData, len(newData), newData)

	if !bytes.Equal(oldData, newData) {
		t.Fatalf("newData differs from oldData")
	}
}

func TestNewCodecDecode(t *testing.T) {
	data := []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x03,
		0x00, 0x00, 0x00, 0x05, 0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x01, 0x00,
		0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x06, 0x00,
		0x00, 0x00, 0x08, 0x09, 0x0a, 0x00, 0x00, 0x00,
		0x0b, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00, 0x02, 0x01, 0x09, 0x00,
		0x00, 0x00, 0x08, 0x07, 0x06, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	var err error
	var oldData, newData ip.IPRouteAddDel
	{
		w := codec.Wrapper{&oldData}
		err = w.Unmarshal(data)
		if err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}
	}
	{
		err = newData.Unmarshal(data)
		if err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}
	}
	t.Logf("Data:\nOLD: %+v\nNEW: %+v", oldData, newData)

	if !reflect.DeepEqual(oldData, newData) {
		t.Fatalf("newData differs from oldData")
	}
}

func NewIPRouteLookupReply() *ip.IPRouteAddDel {
	return &ip.IPRouteAddDel{
		Route: ip.IPRoute{
			TableID:    3,
			StatsIndex: 5,
			Prefix: ip.Prefix{
				Address: ip.Address{
					Af: ip.ADDRESS_IP6,
					Un: ip.AddressUnion{},
				},
				Len: 24,
			},
			NPaths: 2,
			Paths: []ip.FibPath{
				{
					SwIfIndex:  5,
					TableID:    6,
					RpfID:      8,
					Weight:     9,
					Preference: 10,
					Type:       11,
					Flags:      1,
					Proto:      2,
					Nh: ip.FibPathNh{
						Address:            ip.AddressUnion{},
						ViaLabel:           3,
						ObjID:              1,
						ClassifyTableIndex: 2,
					},
					NLabels: 1,
					LabelStack: [16]ip.FibMplsLabel{
						{
							IsUniform: 9,
							Label:     8,
							TTL:       7,
							Exp:       6,
						},
					},
				},
				{
					SwIfIndex:  7,
					TableID:    6,
					RpfID:      8,
					Weight:     9,
					Preference: 10,
					Type:       11,
					Flags:      1,
					Proto:      1,
					Nh: ip.FibPathNh{
						Address:            ip.AddressUnion{},
						ViaLabel:           3,
						ObjID:              1,
						ClassifyTableIndex: 2,
					},
					NLabels: 2,
					LabelStack: [16]ip.FibMplsLabel{
						{
							IsUniform: 9,
							Label:     8,
							TTL:       7,
							Exp:       6,
						},
						{
							IsUniform: 10,
							Label:     8,
							TTL:       7,
							Exp:       6,
						},
					},
				},
			},
		},
	}
}

func TestSize(t *testing.T) {
	m := NewTestAllMsg()
	size := binary.Size(*m)
	t.Logf("size: %v", size)
}

func (m *TestAllMsg) Marshal(b []byte) ([]byte, error) {
	order := binary.BigEndian
	tmp := make([]byte, 143)
	pos := 0

	tmp[pos] = boolToUint(m.Bool)
	pos += 1

	tmp[pos] = m.Uint8
	pos += 1

	order.PutUint16(tmp[pos:pos+2], m.Uint16)
	pos += 2

	order.PutUint32(tmp[pos:pos+4], m.Uint32)
	pos += 4

	tmp[pos] = byte(m.Int8)
	pos += 1

	order.PutUint16(tmp[pos:pos+2], uint16(m.Int16))
	pos += 2

	order.PutUint32(tmp[pos:pos+4], uint32(m.Int32))
	pos += 4

	order.PutUint32(tmp[pos:pos+4], uint32(m.AliasUint32))
	pos += 4

	copy(tmp[pos:pos+6], m.AliasArray[:])
	pos += 6

	order.PutUint32(tmp[pos:pos+4], uint32(m.Enum))
	pos += 4

	for i := 0; i < 4; i++ {
		var x uint32
		if i < len(m.BaseArray) {
			x = m.BaseArray[i]
		}
		order.PutUint32(tmp[pos:pos+4], uint32(x))
		pos += 4
	}

	copy(tmp[pos:pos+7], m.Slice)
	pos += 7

	copy(tmp[pos:pos+64], m.String)
	pos += 64

	order.PutUint32(tmp[pos:pos+4], uint32(len(m.VlaStr)) /*m.SizeOf*/)
	pos += 4

	copy(tmp[pos:pos+len(m.VlaStr)], m.VlaStr[:])
	pos += len(m.VlaStr)

	order.PutUint32(tmp[pos:pos+4], uint32(len(m.VariableSlice)) /*m.SizeOf*/)
	pos += 4

	for i := range m.VariableSlice {
		tmp[pos+i*1] = uint8(m.VariableSlice[i].Proto)
		//copy(tmp[102+i:103+i], []byte{byte(m.VariableSlice[i].Proto)})
	}
	pos += len(m.VariableSlice) * 1

	tmp[pos] = uint8(m.TypeUnion.Af)
	pos += 1

	copy(tmp[pos:pos+16], m.TypeUnion.Un.XXX_UnionData[:])
	pos += 16

	return tmp, nil
}

func (m *TestAllMsg) Unmarshal(tmp []byte) error {
	order := binary.BigEndian

	//tmp := make([]byte, 143)
	pos := 0

	m.Bool = tmp[pos] != 0
	pos += 1

	//tmp[pos] = m.Uint8
	m.Uint8 = tmp[pos]
	pos += 1

	//order.PutUint16(tmp[pos:pos+2], m.Uint16)
	m.Uint16 = order.Uint16(tmp[pos : pos+2])
	pos += 2

	//order.PutUint32(tmp[pos:pos+4], m.Uint32)
	m.Uint32 = order.Uint32(tmp[pos : pos+4])
	pos += 4

	//tmp[pos] = byte(m.Int8)
	m.Int8 = int8(tmp[pos])
	pos += 1

	//order.PutUint16(tmp[pos:pos+2], uint16(m.Int16))
	m.Int16 = int16(order.Uint16(tmp[pos : pos+2]))
	pos += 2

	//order.PutUint32(tmp[pos:pos+4], uint32(m.Int32))
	m.Int32 = int32(order.Uint32(tmp[pos : pos+4]))
	pos += 4

	//order.PutUint32(tmp[pos:pos+4], uint32(m.AliasUint32))
	m.AliasUint32 = InterfaceIndex(order.Uint32(tmp[pos : pos+4]))
	pos += 4

	//copy(tmp[pos:pos+6], m.AliasArray[:])
	copy(m.AliasArray[:], tmp[pos:pos+6])
	pos += 6

	//order.PutUint32(tmp[pos:pos+4], uint32(m.Enum))
	m.Enum = IfStatusFlags(order.Uint32(tmp[pos : pos+4]))
	pos += 4

	m.BaseArray = make([]uint32, 4)
	for i := 0; i < 4; i++ {
		/*var x uint32
		if i < len(m.BaseArray) {
			x = m.BaseArray[i]
		}
		order.PutUint32(tmp[pos:pos+4], uint32(x))*/
		m.BaseArray[i] = order.Uint32(tmp[pos : pos+4])
		pos += 4
	}

	m.Slice = make([]byte, 7)
	copy(m.Slice[:7], tmp[pos:pos+7])
	//copy(tmp[pos:pos+7], m.Slice)
	pos += 7

	i := bytes.Index(tmp[pos:pos+64], []byte{0x00})
	m.String = string(tmp[pos : pos+i])
	//copy(tmp[pos:pos+64], m.String)
	pos += 64

	//order.PutUint32(tmp[pos:pos+4], uint32(len(m.VlaStr)) /*m.SizeOf*/)
	VlaStrLen := int(order.Uint32(tmp[pos : pos+4]))
	pos += 4

	m.VlaStr = string(tmp[pos : pos+VlaStrLen])
	//copy(m.VlaStr[pos:pos+VlaStrLen], tmp[pos:pos+64])
	pos += len(m.VlaStr)

	m.SizeOf = uint32(order.Uint32(tmp[pos : pos+4]))
	pos += 4

	/*order.PutUint32(tmp[pos:pos+4], uint32(len(m.VariableSlice)))
	m.VariableSlice = IfStatusFlags(order.Uint32(tmp[pos : pos+4]))
	pos += 4*/

	m.VariableSlice = make([]SliceType, m.SizeOf)
	for i := range m.VariableSlice {
		//tmp[pos+i*1] = uint8(m.VariableSlice[i].Proto)
		m.VariableSlice[i].Proto = IPProto(tmp[pos+i*1])
		//copy(tmp[102+i:103+i], []byte{byte(m.VariableSlice[i].Proto)})
	}
	pos += len(m.VariableSlice) * 1

	//tmp[pos] = uint8(m.TypeUnion.Af)
	m.TypeUnion.Af = AddressFamily(tmp[pos])
	pos += 1

	//copy(tmp[pos:pos+16], m.TypeUnion.Un.XXX_UnionData[:])
	copy(m.TypeUnion.Un.XXX_UnionData[:], tmp[pos:pos+16])
	pos += 16

	return nil
}

func boolToUint(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// SwInterfaceDetails represents VPP binary API message 'sw_interface_details'.
type TestAllMsg struct {
	Bool          bool
	Uint8         uint8
	Uint16        uint16
	Uint32        uint32
	Int8          int8
	Int16         int16
	Int32         int32
	AliasUint32   InterfaceIndex
	AliasArray    MacAddress
	Enum          IfStatusFlags
	BaseArray     []uint32 `struc:"[4]uint32"`
	Slice         []byte   `struc:"[7]byte"`
	String        string   `struc:"[64]byte"`
	XXX_VlaStrLen uint32   `struc:"sizeof=VlaStr"`
	VlaStr        string
	SizeOf        uint32 `struc:"sizeof=VariableSlice"`
	VariableSlice []SliceType
	TypeUnion     Address
}

type InterfaceIndex uint32
type MacAddress [6]uint8
type IfStatusFlags uint32

const (
	IF_STATUS_API_FLAG_ADMIN_UP IfStatusFlags = 1
	IF_STATUS_API_FLAG_LINK_UP  IfStatusFlags = 2
)

// Address represents VPP binary API type 'address'.
type Address struct {
	Af AddressFamily
	Un AddressUnion
}

// AddressFamily represents VPP binary API enum 'address_family'.
type AddressFamily uint8

const (
	ADDRESS_IP4 AddressFamily = 0
	ADDRESS_IP6 AddressFamily = 1
)

// AddressUnion represents VPP binary API union 'address_union'.
type AddressUnion struct {
	XXX_UnionData [16]byte
}

func (*AddressUnion) GetTypeName() string { return "address_union" }

func AddressUnionIP4(a IP4Address) (u AddressUnion) {
	u.SetIP4(a)
	return
}
func (u *AddressUnion) SetIP4(a IP4Address) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.XXX_UnionData[:], b.Bytes())
}
func (u *AddressUnion) GetIP4() (a IP4Address) {
	var b = bytes.NewReader(u.XXX_UnionData[:])
	struc.Unpack(b, &a)
	return
}

func AddressUnionIP6(a IP6Address) (u AddressUnion) {
	u.SetIP6(a)
	return
}
func (u *AddressUnion) SetIP6(a IP6Address) {
	var b = new(bytes.Buffer)
	if err := struc.Pack(b, &a); err != nil {
		return
	}
	copy(u.XXX_UnionData[:], b.Bytes())
}
func (u *AddressUnion) GetIP6() (a IP6Address) {
	var b = bytes.NewReader(u.XXX_UnionData[:])
	struc.Unpack(b, &a)
	return
}

// IP4Address represents VPP binary API alias 'ip4_address'.
type IP4Address [4]uint8

// IP6Address represents VPP binary API alias 'ip6_address'.
type IP6Address [16]uint8

type SliceType struct {
	Proto IPProto
}

type IPProto uint8

const (
	IP_API_PROTO_HOPOPT   IPProto = 0
	IP_API_PROTO_ICMP     IPProto = 1
	IP_API_PROTO_IGMP     IPProto = 2
	IP_API_PROTO_TCP      IPProto = 6
	IP_API_PROTO_UDP      IPProto = 17
	IP_API_PROTO_GRE      IPProto = 47
	IP_API_PROTO_ESP      IPProto = 50
	IP_API_PROTO_AH       IPProto = 51
	IP_API_PROTO_ICMP6    IPProto = 58
	IP_API_PROTO_EIGRP    IPProto = 88
	IP_API_PROTO_OSPF     IPProto = 89
	IP_API_PROTO_SCTP     IPProto = 132
	IP_API_PROTO_RESERVED IPProto = 255
)

func (m *TestAllMsg) Reset()                        { *m = TestAllMsg{} }
func (*TestAllMsg) GetMessageName() string          { return "sw_interface_details" }
func (*TestAllMsg) GetCrcString() string            { return "17b69fa2" }
func (*TestAllMsg) GetMessageType() api.MessageType { return api.ReplyMessage }
