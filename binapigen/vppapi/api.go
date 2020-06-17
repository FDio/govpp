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

// Schema represents a collection of VPP binary API modules for single VPP revision.
/*type Schema struct {
	Version string // VPP version

	Core    map[string]File
	Plugins map[string]File
}*/

type File struct {
	Name string
	Path string

	CRC string

	Options map[string]string `json:",omitempty"`
	Imports []string          `json:",omitempty"`

	AliasTypes  []AliasType  `json:",omitempty"`
	EnumTypes   []EnumType   `json:",omitempty"`
	StructTypes []StructType `json:",omitempty"`
	UnionTypes  []UnionType  `json:",omitempty"`

	Messages []Message `json:",omitempty"`
	Service  *Service  `json:",omitempty"`
}

func (x File) Version() string {
	if x.Options != nil {
		return x.Options[fileOptionVersion]
	}
	return ""
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

type AliasType struct {
	Name   string
	Type   string
	Length int `json:",omitempty"`
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

/*func (x File) String() string {
	return fmt.Sprintf("File: %s %s (%v)", x.Name, x.Version(), x.CRC)
}

func (x Service) String() string {
	return fmt.Sprintf("Service: %v", x.RPCs)
}

func (x RPC) String() string {
	var buf strings.Builder
	buf.WriteString("RPC: ")
	buf.WriteString(fmt.Sprintf("%s ", x.RequestMsg))
	if x.Stream {
		if x.StreamMsg != "" {
			buf.WriteString(fmt.Sprintf("stream (%s)", x.StreamMsg))
			buf.WriteString(fmt.Sprintf("reply (%s)", x.ReplyMsg))
		} else {
			buf.WriteString(fmt.Sprintf("stream (%s)", x.ReplyMsg))
		}
	} else {
		buf.WriteString(fmt.Sprintf("reply (%s)", x.ReplyMsg))
	}
	if len(x.Events) > 0 {
		buf.WriteString(fmt.Sprintf("events %s", x.Events))
	}
	return buf.String()

}

func (x EnumType) String() string {
	return fmt.Sprintf("EnumType: %s/%s (%d entries) %v", x.Name, x.Type, len(x.Entries), x.Entries)
}

func (x EnumEntry) String() string {
	return fmt.Sprintf("%s=%v", x.Name, x.Value)
}

func (x AliasType) String() string {
	if x.Length > 0 {
		return fmt.Sprintf("AliasType: %s/%s[%d]", x.Name, x.Type, x.Length)
	}
	return fmt.Sprintf("AliasType: %s/%s", x.Name, x.Type)
}

func (x Type) String() string {
	return fmt.Sprintf("Type: %s (%d fields) %v", x.Name, len(x.Fields), x.Fields)
}

func (x UnionType) String() string {
	return fmt.Sprintf("UnionType: %s (%d fields) %v", x.Name, len(x.Fields), x.Fields)
}

func (x Message) String() string {
	return fmt.Sprintf("Message: %s (%d fields) %v", x.Name, len(x.Fields), x.Fields)
}

func (x Field) String() string {
	if x.Array {
		length := ""
		if x.Length > 0 {
			length = fmt.Sprint(x.Length)
		}
		return fmt.Sprintf("%s/%v[%v]", x.Name, x.Type, length)
	}
	return fmt.Sprintf("%s/%v", x.Name, x.Type)
}*/
