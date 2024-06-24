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
	"math"
)

// Buffer provides buffer for encoding and decoding data on wire.
type Buffer struct {
	buf []byte
	pos int
}

// NewBuffer creates new buffer using b as data.
func NewBuffer(b []byte) *Buffer {
	return &Buffer{
		buf: b,
	}
}

// Bytes returns buffer data up to current position.
func (b *Buffer) Bytes() []byte {
	return b.buf[:b.pos]
}

func (b *Buffer) EncodeBytes(v []byte, length int) {
	if length == 0 {
		length = len(v)
	}
	copy(b.buf[b.pos:b.pos+length], v)
	b.pos += length
}

func (b *Buffer) DecodeBytes(length int) []byte {
	v := b.buf[b.pos : b.pos+length]
	b.pos += length
	return v
}

func (b *Buffer) EncodeBool(v bool) {
	if v {
		b.buf[b.pos] = 1
	} else {
		b.buf[b.pos] = 0
	}
	b.pos += 1
}

func (b *Buffer) DecodeBool() bool {
	v := b.buf[b.pos] != 0
	b.pos += 1
	return v
}

func (b *Buffer) EncodeUint8(v uint8) {
	b.buf[b.pos] = byte(v)
	b.pos += 1
}

func (b *Buffer) DecodeUint8() uint8 {
	v := uint8(b.buf[b.pos])
	b.pos += 1
	return v
}

func (b *Buffer) EncodeUint16(v uint16) {
	binary.BigEndian.PutUint16(b.buf[b.pos:b.pos+2], v)
	b.pos += 2
}

func (b *Buffer) DecodeUint16() uint16 {
	v := binary.BigEndian.Uint16(b.buf[b.pos : b.pos+2])
	b.pos += 2
	return v
}

func (b *Buffer) EncodeUint32(v uint32) {
	binary.BigEndian.PutUint32(b.buf[b.pos:b.pos+4], v)
	b.pos += 4
}

func (b *Buffer) DecodeUint32() uint32 {
	v := binary.BigEndian.Uint32(b.buf[b.pos : b.pos+4])
	b.pos += 4
	return v
}

func (b *Buffer) EncodeUint64(v uint64) {
	binary.BigEndian.PutUint64(b.buf[b.pos:b.pos+8], v)
	b.pos += 8
}

func (b *Buffer) DecodeUint64() uint64 {
	v := binary.BigEndian.Uint64(b.buf[b.pos : b.pos+8])
	b.pos += 8
	return v
}

func (b *Buffer) EncodeInt8(v int8) {
	b.buf[b.pos] = byte(v)
	b.pos += 1
}

func (b *Buffer) DecodeInt8() int8 {
	v := int8(b.buf[b.pos])
	b.pos += 1
	return v
}

func (b *Buffer) EncodeInt16(v int16) {
	binary.BigEndian.PutUint16(b.buf[b.pos:b.pos+2], uint16(v))
	b.pos += 2
}

func (b *Buffer) DecodeInt16() int16 {
	v := int16(binary.BigEndian.Uint16(b.buf[b.pos : b.pos+2]))
	b.pos += 2
	return v
}

func (b *Buffer) EncodeInt32(v int32) {
	binary.BigEndian.PutUint32(b.buf[b.pos:b.pos+4], uint32(v))
	b.pos += 4
}

func (b *Buffer) DecodeInt32() int32 {
	v := int32(binary.BigEndian.Uint32(b.buf[b.pos : b.pos+4]))
	b.pos += 4
	return v
}

func (b *Buffer) EncodeInt64(v int64) {
	binary.BigEndian.PutUint64(b.buf[b.pos:b.pos+8], uint64(v))
	b.pos += 8
}

func (b *Buffer) DecodeInt64() int64 {
	v := int64(binary.BigEndian.Uint64(b.buf[b.pos : b.pos+8]))
	b.pos += 8
	return v
}

func (b *Buffer) EncodeFloat64(v float64) {
	binary.LittleEndian.PutUint64(b.buf[b.pos:b.pos+8], math.Float64bits(v))
	b.pos += 8
}

func (b *Buffer) DecodeFloat64() float64 {
	v := math.Float64frombits(binary.LittleEndian.Uint64(b.buf[b.pos : b.pos+8]))
	b.pos += 8
	return v
}

func (b *Buffer) EncodeString(v string, length int) {
	if length == 0 {
		b.EncodeUint32(uint32(len(v)))
		length = len(v)
	}
	copy(b.buf[b.pos:b.pos+length], v)
	b.pos += length
}

func (b *Buffer) DecodeString(length int) string {
	var v []byte
	if length == 0 {
		length = int(b.DecodeUint32())
		v = b.buf[b.pos : b.pos+length]
	} else {
		v = b.buf[b.pos : b.pos+length]
		if nul := bytes.Index(v, []byte{0x00}); nul >= 0 {
			v = v[:nul]
		}
	}
	b.pos += length
	return string(v)
}
