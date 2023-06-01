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

type (
	// Schema represents a collection of API files for a specific VPP version.
	Schema struct {
		// Files is a list of File objects that are part of this scheme.
		Files []File
		// Version is a VPP version of this schema.
		Version string
	}

	// File is a single API file and its contents.
	File struct {
		Name string // Name is the name of this API file (without any extension).
		Path string // Path is the location of thi API file relative to API directory.
		CRC  string // CRC is a checksum for this API file.

		// Options is a map of string key-value pairs that provides additional options for this file.
		Options map[string]string `json:",omitempty"`
		// Imports is a list of strings representing the names of API files that are imported by this file.
		Imports []string `json:",omitempty"`

		AliasTypes    []AliasType  `json:",omitempty"`
		EnumTypes     []EnumType   `json:",omitempty"`
		EnumflagTypes []EnumType   `json:",omitempty"`
		StructTypes   []StructType `json:",omitempty"`
		UnionTypes    []UnionType  `json:",omitempty"`

		// Messages is a list of Message objects representing messages used in the API.
		Messages []Message `json:",omitempty"`
		// Service is an object representing the RPC services used in the API.
		// In case there is not any services defined for this File, the Service is nil.
		Service *Service `json:",omitempty"`

		Counters []Counter      `json:",omitempty"`
		Paths    []CounterPaths `json:",omitempty"`
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
		Name    string
		Fields  []Field
		CRC     string
		Options map[string]string `json:",omitempty"`
		Comment string            `json:",omitempty"`
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

	Counter struct {
		Name     string
		Elements []Element `json:",omitempty"`
	}

	Element struct {
		Name        string
		Severity    string
		Type        string
		Units       string
		Description string
	}

	CounterPaths struct {
		Name  string
		Paths []string `json:",omitempty"`
	}
)
