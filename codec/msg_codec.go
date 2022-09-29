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

package codec

import (
	"encoding/binary"
	"errors"
	"fmt"

	"go.fd.io/govpp/api"
)

var DefaultCodec = new(MsgCodec)

func EncodeMsg(msg api.Message, msgID uint16) (data []byte, err error) {
	return DefaultCodec.EncodeMsg(msg, msgID)
}
func DecodeMsg(data []byte, msg api.Message) (err error) {
	return DefaultCodec.DecodeMsg(data, msg)
}
func DecodeMsgContext(data []byte, msgType api.MessageType) (context uint32, err error) {
	return DefaultCodec.DecodeMsgContext(data, msgType)
}

// MsgCodec provides encoding and decoding functionality of `api.Message` structs into/from
// binary format as accepted by VPP.
type MsgCodec struct{}

func (*MsgCodec) EncodeMsg(msg api.Message, msgID uint16) (data []byte, err error) {
	if msg == nil {
		return nil, errors.New("nil message passed in")
	}

	// try to recover panic which might possibly occur
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); !ok {
				err = fmt.Errorf("%v", r)
			}
			err = fmt.Errorf("panic occurred during encoding message %s: %v", msg.GetMessageName(), err)
		}
	}()

	marshaller, ok := msg.(Marshaler)
	if !ok {
		marshaller = Wrapper{msg}
	}

	size := marshaller.Size()
	offset := getOffset(msg)

	// encode msg ID
	b := make([]byte, size+offset)
	b[0] = byte(msgID >> 8)
	b[1] = byte(msgID)

	data, err = marshaller.Marshal(b[offset:])
	if err != nil {
		return nil, err
	}

	return b[0:len(b):len(b)], nil
}

func (*MsgCodec) DecodeMsg(data []byte, msg api.Message) (err error) {
	if msg == nil {
		return errors.New("nil message passed in")
	}

	// try to recover panic which might possibly occur
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); !ok {
				err = fmt.Errorf("%v", r)
			}
			err = fmt.Errorf("panic occurred during decoding message %s: %v", msg.GetMessageName(), err)
		}
	}()

	marshaller, ok := msg.(Unmarshaler)
	if !ok {
		marshaller = Wrapper{msg}
	}

	offset := getOffset(msg)

	err = marshaller.Unmarshal(data[offset:])
	if err != nil {
		return err
	}

	return nil
}

func (*MsgCodec) DecodeMsgContext(data []byte, msgType api.MessageType) (context uint32, err error) {
	switch msgType {
	case api.RequestMessage:
		return binary.BigEndian.Uint32(data[6:10]), nil
	case api.ReplyMessage:
		return binary.BigEndian.Uint32(data[2:6]), nil
	}

	return 0, nil
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
