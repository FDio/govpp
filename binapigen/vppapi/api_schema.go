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

// Package vppapi parses VPP API files without any additional processing.
package vppapi

type (
	File struct {
		Name string
		Path string
		CRC  string

		Options map[string]string `json:",omitempty"`
		Imports []string          `json:",omitempty"`

		AliasTypes  []AliasType  `json:",omitempty"`
		EnumTypes   []EnumType   `json:",omitempty"`
		StructTypes []StructType `json:",omitempty"`
		UnionTypes  []UnionType  `json:",omitempty"`

		Messages []Message `json:",omitempty"`
		Service  *Service  `json:",omitempty"`
	}

	AliasType struct {
		Name   string
		Type   string
		Length int `json:",omitempty"`
	}

	EnumType struct {
		Name    string
		Type    string
		Entries []EnumEntry
	}

	EnumEntry struct {
		Name  string
		Value uint32
	}

	StructType struct {
		Name   string
		Fields []Field
	}

	UnionType struct {
		Name   string
		Fields []Field
	}

	Message struct {
		Name   string
		Fields []Field
		CRC    string
	}

	Field struct {
		Name     string
		Type     string
		Length   int                    `json:",omitempty"`
		Array    bool                   `json:",omitempty"`
		SizeFrom string                 `json:",omitempty"`
		Meta     map[string]interface{} `json:",omitempty"`
	}

	Service struct {
		RPCs []RPC `json:",omitempty"`
	}

	RPC struct {
		Request   string
		Reply     string
		Stream    bool     `json:",omitempty"`
		StreamMsg string   `json:",omitempty"`
		Events    []string `json:",omitempty"`
	}
)
