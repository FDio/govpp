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

package vppapi

type File struct {
	Name string
	Path string

	CRC     string
	Options map[string]string `json:",omitempty"`

	Imports []string `json:",omitempty"`

	AliasTypes  []AliasType  `json:",omitempty"`
	EnumTypes   []EnumType   `json:",omitempty"`
	StructTypes []StructType `json:",omitempty"`
	UnionTypes  []UnionType  `json:",omitempty"`
	Messages    []Message    `json:",omitempty"`
	Service     *Service     `json:",omitempty"`
}

func (x File) Version() string {
	if x.Options != nil {
		return x.Options[fileOptionVersion]
	}
	return ""
}

type AliasType struct {
	Name   string
	Type   string
	Length int `json:",omitempty"`
}

type EnumType struct {
	Name    string
	Type    string
	Entries []EnumEntry
}

type EnumEntry struct {
	Name  string
	Value uint32
}

type StructType struct {
	Name   string
	Fields []Field
}

type UnionType struct {
	Name   string
	Fields []Field
}

type Message struct {
	Name   string
	Fields []Field
	CRC    string
}

type Field struct {
	Name     string
	Type     string
	Length   int                    `json:",omitempty"`
	Array    bool                   `json:",omitempty"`
	SizeFrom string                 `json:",omitempty"`
	Meta     map[string]interface{} `json:",omitempty"`
}

type Service struct {
	RPCs []RPC `json:",omitempty"`
}

type RPC struct {
	Name       string
	RequestMsg string
	ReplyMsg   string
	Stream     bool     `json:",omitempty"`
	StreamMsg  string   `json:",omitempty"`
	Events     []string `json:",omitempty"`
}
