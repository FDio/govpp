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

/*
import (
	"testing"

	"go.fd.io/govpp/codec"
)

var Data []byte

func BenchmarkEncodeNew(b *testing.B) {
	m := NewTestAllMsg()
	c := codec.DefaultCodec

	var err error
	var data []byte

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err = c.EncodeMsg(m, 100)
		if err != nil {
			b.Fatalf("expected nil error, got: %v", err)
		}
	}
	Data = data
}

func BenchmarkEncodeWrapper(b *testing.B) {
	m := codec.Wrapper{NewTestAllMsg()}
	c := codec.DefaultCodec

	var err error
	var data []byte

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err = c.EncodeMsg(m, 100)
		if err != nil {
			b.Fatalf("expected nil error, got: %v", err)
		}
	}
	Data = data
}

func BenchmarkEncodeHard(b *testing.B) {
	m := NewTestAllMsg()

	var err error
	var data []byte

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err = m.Marshal(nil)
		if err != nil {
			b.Fatalf("expected nil error, got: %v", err)
		}
	}
	Data = data
}
*/
