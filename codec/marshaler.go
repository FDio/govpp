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

package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"unsafe"

	"github.com/lunixbochs/struc"

	"git.fd.io/govpp.git/api"
)

var DefaultCodec = &NewCodec{} // &MsgCodec{}

// Marshaler is the interface implemented by the binary API messages that can
// marshal itself into binary form for the wire.
type Marshaler interface {
	Size() int
	Marshal([]byte) ([]byte, error)
}

// Unmarshaler is the interface implemented by the binary API messages that can
// unmarshal a binary representation of itself from the wire.
type Unmarshaler interface {
	Unmarshal([]byte) error
}

type NewCodec struct{}

func (*NewCodec) EncodeMsg(msg api.Message, msgID uint16) (data []byte, err error) {
	if msg == nil {
		return nil, errors.New("nil message passed in")
	}
	marshaller, ok := msg.(Marshaler)
	if !ok {
		return nil, fmt.Errorf("message %s does not implement marshaller", msg.GetMessageName())
	}

	size := marshaller.Size()
	offset := getOffset(msg)
	//fmt.Printf("size=%d offset=%d\n", size, offset)

	b := make([]byte, size+offset)
	b[0] = byte(msgID >> 8)
	b[1] = byte(msgID)

	//fmt.Printf("len(b)=%d cap(b)=%d\n", len(b), cap(b))
	//b = append(b, byte(msgID>>8), byte(msgID))

	//buf := new(bytes.Buffer)
	//buf.Grow(size)

	// encode msg ID
	//buf.WriteByte(byte(msgID >> 8))
	//buf.WriteByte(byte(msgID))

	data, err = marshaller.Marshal(b[offset:])
	if err != nil {
		return nil, err
	}
	//buf.Write(b)

	return b[0:len(b):len(b)], nil
}

func getOffset(msg api.Message) (offset int) {
	switch msg.GetMessageType() {
	case api.RequestMessage:
		return 10
	case api.ReplyMessage:
		return 6
	case api.EventMessage:
		return 6
	}
	return 2
}

func (*NewCodec) DecodeMsg(data []byte, msg api.Message) (err error) {
	if msg == nil {
		return errors.New("nil message passed in")
	}
	marshaller, ok := msg.(Unmarshaler)
	if !ok {
		return fmt.Errorf("message %s does not implement marshaller", msg.GetMessageName())
	}

	offset := getOffset(msg)

	err = marshaller.Unmarshal(data[offset:len(data)])
	if err != nil {
		return err
	}

	return nil
}

func (*NewCodec) DecodeMsgContext(data []byte, msg api.Message) (context uint32, err error) {
	if msg == nil {
		return 0, errors.New("nil message passed in")
	}

	switch msg.GetMessageType() {
	case api.RequestMessage:
		return order.Uint32(data[6:10]), nil
	case api.ReplyMessage:
		return order.Uint32(data[2:6]), nil
	}

	return 0, nil
}

type OldCodec struct{}

func (c *OldCodec) Marshal(v interface{}) (b []byte, err error) {
	buf := new(bytes.Buffer)
	if err := struc.Pack(buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *OldCodec) Unmarshal(data []byte, v interface{}) error {
	buf := bytes.NewReader(data)
	if err := struc.Unpack(buf, v); err != nil {
		return err
	}
	return nil
}

/*type CodecNew struct{}

func (c *CodecNew) Marshal(v interface{}) (b []byte, err error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}*/

func EncodeBool(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func MarshalValue(value interface{}) []byte {
	switch value.(type) {
	case int8:
		return []byte{byte(value.(int8))}
	case uint8:
		return []byte{byte(value.(uint8))}
	}
	return nil
}

var order = binary.BigEndian

type Buffer struct {
	pos int
	buf []byte
}

func (b *Buffer) Bytes() []byte {
	return b.buf[:b.pos]
}

func (b *Buffer) EncodeUint8(v uint8) {
	b.buf[b.pos] = v
	b.pos += 1
}

func (b *Buffer) EncodeUint16(v uint16) {
	order.PutUint16(b.buf[b.pos:b.pos+2], v)
	b.pos += 2
}

func (b *Buffer) EncodeUint32(v uint32) {
	order.PutUint32(b.buf[b.pos:b.pos+4], v)
	b.pos += 4
}

func (b *Buffer) EncodeUint64(v uint64) {
	order.PutUint64(b.buf[b.pos:b.pos+8], v)
	b.pos += 8
}

func (b *Buffer) EncodeFloat64(v float64) {
	order.PutUint64(b.buf[b.pos:b.pos+8], math.Float64bits(v))
	b.pos += 8
}

func (b *Buffer) EncodeBool(v bool) {
	if v {
		b.buf[b.pos] = 1
	} else {
		b.buf[b.pos] = 0
	}
	b.pos += 1
}

func (b *Buffer) EncodeString(v string) {

	b.pos += 1
}

func DecodeString(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	stringHeader := reflect.StringHeader{Data: sliceHeader.Data, Len: sliceHeader.Len}
	return *(*string)(unsafe.Pointer(&stringHeader))
}
