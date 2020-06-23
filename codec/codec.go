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
	"encoding/binary"
	"math"
	"reflect"
	"unsafe"
)

var order = binary.BigEndian

type Buffer struct {
	pos int
	buf []byte
}

func (b *Buffer) Bytes() []byte {
	return b.buf[:b.pos]
}

func (b *Buffer) EncodeBool(v bool) {
	if v {
		b.buf[b.pos] = 1
	} else {
		b.buf[b.pos] = 0
	}
	b.pos += 1
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

func (b *Buffer) EncodeString(v string, length int) {
	copy(b.buf[b.pos:b.pos+length], v)
	b.pos += length
}

func DecodeString(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	stringHeader := reflect.StringHeader{Data: sliceHeader.Data, Len: sliceHeader.Len}
	return *(*string)(unsafe.Pointer(&stringHeader))
}
