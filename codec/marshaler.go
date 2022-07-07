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
	"reflect"

	"github.com/lunixbochs/struc"

	"go.fd.io/govpp/api"
)

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

type Wrapper struct {
	api.Message
}

func (w Wrapper) Size() int {
	if size, err := struc.Sizeof(w.Message); err != nil {
		return 0
	} else {
		return size
	}
}

func (w Wrapper) Marshal(b []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	if reflect.TypeOf(w.Message).Elem().NumField() > 0 {
		if err := struc.Pack(buf, w.Message); err != nil {
			return nil, err
		}
	}
	if b != nil {
		copy(b, buf.Bytes())
	}
	return buf.Bytes(), nil
}

func (w Wrapper) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	if err := struc.Unpack(buf, w.Message); err != nil {
		return err
	}
	return nil
}
