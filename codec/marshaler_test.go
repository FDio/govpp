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
	"reflect"
	"testing"

	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/binapi/ip_types"
	"git.fd.io/govpp.git/binapi/sr"
	"git.fd.io/govpp.git/codec"
	"git.fd.io/govpp.git/internal/testbinapi/binapi2001/interfaces"
	"git.fd.io/govpp.git/internal/testbinapi/binapi2001/ip"
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

func TestNewCodecEncodeDecode2(t *testing.T) {
	m := &sr.SrPoliciesDetails{
		Bsid:        ip_types.IP6Address{00, 11, 22, 33, 44, 55, 66, 77, 88, 99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		IsSpray:     true,
		IsEncap:     false,
		FibTable:    33,
		NumSidLists: 1,
		SidLists: []sr.Srv6SidList{
			{
				Weight:  555,
				NumSids: 2,
				Sids: [16]ip_types.IP6Address{
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
}

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
