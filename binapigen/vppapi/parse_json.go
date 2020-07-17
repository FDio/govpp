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

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bennyscetbun/jsongo"
)

var debug = strings.Contains(os.Getenv("DEBUG_GOVPP"), "parser")

func logf(f string, v ...interface{}) {
	if debug {
		log.Printf(f, v...)
	}
}

const (
	// root keys
	fileAPIVersion = "vl_api_version"
	fileOptions    = "options"
	fileTypes      = "types"
	fileMessages   = "messages"
	fileUnions     = "unions"
	fileEnums      = "enums"
	fileAliases    = "aliases"
	fileServices   = "services"
	fileImports    = "imports"
	// type keys
	messageCrc  = "crc"
	enumType    = "enumtype"
	aliasLength = "length"
	aliasType   = "type"
	// service
	serviceReply     = "reply"
	serviceStream    = "stream"
	serviceStreamMsg = "stream_msg"
	serviceEvents    = "events"
)

func parseJSON(data []byte) (module *File, err error) {
	// parse root
	jsonRoot := new(jsongo.Node)
	if err := json.Unmarshal(data, jsonRoot); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON failed: %v", err)
	}

	logf("file contains:")
	for _, key := range jsonRoot.GetKeys() {
		if jsonRoot.At(key).Len() > 0 {
			logf("  - %2d %s", jsonRoot.At(key).Len(), key)
		}
	}

	module = new(File)

	// parse CRC
	crc := jsonRoot.At(fileAPIVersion)
	if crc.GetType() == jsongo.TypeValue {
		module.CRC = crc.MustGetString()
	}

	// parse options
	opt := jsonRoot.Map(fileOptions)
	if opt.GetType() == jsongo.TypeMap {
		module.Options = make(map[string]string)
		for _, key := range opt.GetKeys() {
			optionKey := key.(string)
			optionVal := opt.At(key).MustGetString()
			module.Options[optionKey] = optionVal
		}
	}

	// parse imports
	importsNode := jsonRoot.Map(fileImports)
	module.Imports = make([]string, 0, importsNode.Len())
	uniq := make(map[string]struct{})
	for i := 0; i < importsNode.Len(); i++ {
		importNode := importsNode.At(i)
		imp := importNode.MustGetString()
		if _, ok := uniq[imp]; ok {
			logf("duplicate import found: %v", imp)
			continue
		}
		uniq[imp] = struct{}{}
		module.Imports = append(module.Imports, imp)
	}

	// avoid duplicate objects
	known := make(map[string]struct{})
	exists := func(name string) bool {
		if _, ok := known[name]; ok {
			logf("duplicate object found: %v", name)
			return true
		}
		known[name] = struct{}{}
		return false
	}

	// parse enum types
	enumsNode := jsonRoot.Map(fileEnums)
	module.EnumTypes = make([]EnumType, 0)
	for i := 0; i < enumsNode.Len(); i++ {
		enum, err := parseEnum(enumsNode.At(i))
		if err != nil {
			return nil, err
		}
		if exists(enum.Name) {
			continue
		}
		module.EnumTypes = append(module.EnumTypes, *enum)
	}

	// parse alias types
	aliasesNode := jsonRoot.Map(fileAliases)
	if aliasesNode.GetType() == jsongo.TypeMap {
		module.AliasTypes = make([]AliasType, 0)
		for _, key := range aliasesNode.GetKeys() {
			aliasName := key.(string)
			alias, err := parseAlias(aliasName, aliasesNode.At(key))
			if err != nil {
				return nil, err
			}
			if exists(alias.Name) {
				continue
			}
			module.AliasTypes = append(module.AliasTypes, *alias)
		}
	}

	// parse struct types
	typesNode := jsonRoot.Map(fileTypes)
	module.StructTypes = make([]StructType, 0)
	for i := 0; i < typesNode.Len(); i++ {
		structyp, err := parseStruct(typesNode.At(i))
		if err != nil {
			return nil, err
		}
		if exists(structyp.Name) {
			continue
		}
		module.StructTypes = append(module.StructTypes, *structyp)
	}

	// parse union types
	unionsNode := jsonRoot.Map(fileUnions)
	module.UnionTypes = make([]UnionType, 0)
	for i := 0; i < unionsNode.Len(); i++ {
		union, err := parseUnion(unionsNode.At(i))
		if err != nil {
			return nil, err
		}
		if exists(union.Name) {
			continue
		}
		module.UnionTypes = append(module.UnionTypes, *union)
	}

	// parse messages
	messagesNode := jsonRoot.Map(fileMessages)
	if messagesNode.GetType() == jsongo.TypeArray {
		module.Messages = make([]Message, messagesNode.Len())
		for i := 0; i < messagesNode.Len(); i++ {
			msg, err := parseMessage(messagesNode.At(i))
			if err != nil {
				return nil, err
			}
			module.Messages[i] = *msg
		}
	}

	// parse services
	servicesNode := jsonRoot.Map(fileServices)
	if servicesNode.GetType() == jsongo.TypeMap {
		module.Service = &Service{
			RPCs: make([]RPC, servicesNode.Len()),
		}
		for i, key := range servicesNode.GetKeys() {
			rpcName := key.(string)
			svc, err := parseServiceRPC(rpcName, servicesNode.At(key))
			if err != nil {
				return nil, err
			}
			module.Service.RPCs[i] = *svc
		}
	}

	return module, nil
}

// parseEnum parses VPP binary API enum object from JSON node
func parseEnum(enumNode *jsongo.Node) (*EnumType, error) {
	if enumNode.Len() == 0 || enumNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for enum specified")
	}

	enumName, ok := enumNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("enum name is %T, not a string", enumNode.At(0).Get())
	}
	enumType, ok := enumNode.At(enumNode.Len() - 1).At(enumType).Get().(string)
	if !ok {
		return nil, fmt.Errorf("enum type invalid or missing")
	}

	enum := EnumType{
		Name: enumName,
		Type: enumType,
	}

	// loop through enum entries, skip first (name) and last (enumtype)
	for j := 1; j < enumNode.Len()-1; j++ {
		if enumNode.At(j).GetType() == jsongo.TypeArray {
			entry := enumNode.At(j)

			if entry.Len() < 2 || entry.At(0).GetType() != jsongo.TypeValue || entry.At(1).GetType() != jsongo.TypeValue {
				return nil, errors.New("invalid JSON for enum entry specified")
			}

			entryName, ok := entry.At(0).Get().(string)
			if !ok {
				return nil, fmt.Errorf("enum entry name is %T, not a string", entry.At(0).Get())
			}
			entryVal := entry.At(1).Get().(float64)

			enum.Entries = append(enum.Entries, EnumEntry{
				Name:  entryName,
				Value: uint32(entryVal),
			})
		}
	}

	return &enum, nil
}

// parseUnion parses VPP binary API union object from JSON node
func parseUnion(unionNode *jsongo.Node) (*UnionType, error) {
	if unionNode.Len() == 0 || unionNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for union specified")
	}

	unionName, ok := unionNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("union name is %T, not a string", unionNode.At(0).Get())
	}

	union := UnionType{
		Name: unionName,
	}

	// loop through union fields, skip first (name)
	for j := 1; j < unionNode.Len(); j++ {
		if unionNode.At(j).GetType() == jsongo.TypeArray {
			fieldNode := unionNode.At(j)

			field, err := parseField(fieldNode)
			if err != nil {
				return nil, err
			}

			union.Fields = append(union.Fields, *field)
		}
	}

	return &union, nil
}

// parseStruct parses VPP binary API type object from JSON node
func parseStruct(typeNode *jsongo.Node) (*StructType, error) {
	if typeNode.Len() == 0 || typeNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for type specified")
	}

	typeName, ok := typeNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("type name is %T, not a string", typeNode.At(0).Get())
	}

	typ := StructType{
		Name: typeName,
	}

	// loop through type fields, skip first (name)
	for j := 1; j < typeNode.Len(); j++ {
		if typeNode.At(j).GetType() == jsongo.TypeArray {
			fieldNode := typeNode.At(j)

			field, err := parseField(fieldNode)
			if err != nil {
				return nil, err
			}

			typ.Fields = append(typ.Fields, *field)
		}
	}

	return &typ, nil
}

// parseAlias parses VPP binary API alias object from JSON node
func parseAlias(aliasName string, aliasNode *jsongo.Node) (*AliasType, error) {
	if aliasNode.Len() == 0 || aliasNode.At(aliasType).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for alias specified")
	}

	alias := AliasType{
		Name: aliasName,
	}

	if typeNode := aliasNode.At(aliasType); typeNode.GetType() == jsongo.TypeValue {
		typ, ok := typeNode.Get().(string)
		if !ok {
			return nil, fmt.Errorf("alias type is %T, not a string", typeNode.Get())
		}
		if typ != "null" {
			alias.Type = typ
		}
	}

	if lengthNode := aliasNode.At(aliasLength); lengthNode.GetType() == jsongo.TypeValue {
		length, ok := lengthNode.Get().(float64)
		if !ok {
			return nil, fmt.Errorf("alias length is %T, not a float64", lengthNode.Get())
		}
		alias.Length = int(length)
	}

	return &alias, nil
}

// parseMessage parses VPP binary API message object from JSON node
func parseMessage(msgNode *jsongo.Node) (*Message, error) {
	if msgNode.Len() < 2 || msgNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for message specified")
	}

	msgName, ok := msgNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("message name is %T, not a string", msgNode.At(0).Get())
	}
	msgCRC, ok := msgNode.At(msgNode.Len() - 1).At(messageCrc).Get().(string)
	if !ok {
		return nil, fmt.Errorf("message crc invalid or missing")
	}

	msg := Message{
		Name: msgName,
		CRC:  msgCRC,
	}

	// loop through message fields, skip first (name) and last (crc)
	for j := 1; j < msgNode.Len()-1; j++ {
		if msgNode.At(j).GetType() == jsongo.TypeArray {
			fieldNode := msgNode.At(j)

			field, err := parseField(fieldNode)
			if err != nil {
				return nil, err
			}

			msg.Fields = append(msg.Fields, *field)
		}
	}

	return &msg, nil
}

// parseField parses VPP binary API object field from JSON node
func parseField(field *jsongo.Node) (*Field, error) {
	if field.Len() < 2 || field.At(0).GetType() != jsongo.TypeValue || field.At(1).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for field specified")
	}

	fieldType, ok := field.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("field type is %T, not a string", field.At(0).Get())
	}
	fieldName, ok := field.At(1).Get().(string)
	if !ok {
		return nil, fmt.Errorf("field name is %T, not a string", field.At(1).Get())
	}

	f := &Field{
		Name: fieldName,
		Type: fieldType,
	}

	if field.Len() >= 3 {
		switch field.At(2).GetType() {
		case jsongo.TypeValue:
			fieldLength, ok := field.At(2).Get().(float64)
			if !ok {
				return nil, fmt.Errorf("field length is %T, not float64", field.At(2).Get())
			}
			f.Length = int(fieldLength)
			f.Array = true

		case jsongo.TypeMap:
			fieldMeta := field.At(2)
			if fieldMeta.Len() == 0 {
				break
			}
			f.Meta = map[string]interface{}{}
			for _, key := range fieldMeta.GetKeys() {
				metaName := key.(string)
				metaValue := fieldMeta.At(key).Get()
				f.Meta[metaName] = metaValue
			}

		default:
			return nil, errors.New("invalid JSON for field specified")
		}
	}
	if field.Len() >= 4 {
		fieldLengthFrom, ok := field.At(3).Get().(string)
		if !ok {
			return nil, fmt.Errorf("field length from is %T, not a string", field.At(3).Get())
		}
		f.SizeFrom = fieldLengthFrom
	}

	return f, nil
}

// parseServiceRPC parses VPP binary API service object from JSON node
func parseServiceRPC(rpcName string, rpcNode *jsongo.Node) (*RPC, error) {
	if rpcNode.Len() == 0 || rpcNode.At(serviceReply).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for service RPC specified")
	}

	rpc := RPC{
		Request: rpcName,
	}

	if replyNode := rpcNode.At(serviceReply); replyNode.GetType() == jsongo.TypeValue {
		reply, ok := replyNode.Get().(string)
		if !ok {
			return nil, fmt.Errorf("service RPC reply is %T, not a string", replyNode.Get())
		}
		rpc.Reply = reply
	}

	// is stream (dump)
	if streamNode := rpcNode.At(serviceStream); streamNode.GetType() == jsongo.TypeValue {
		var ok bool
		rpc.Stream, ok = streamNode.Get().(bool)
		if !ok {
			return nil, fmt.Errorf("service RPC stream is %T, not a boolean", streamNode.Get())
		}
	}

	// stream message
	if streamMsgNode := rpcNode.At(serviceStreamMsg); streamMsgNode.GetType() == jsongo.TypeValue {
		var ok bool
		rpc.StreamMsg, ok = streamMsgNode.Get().(string)
		if !ok {
			return nil, fmt.Errorf("service RPC stream msg is %T, not a string", streamMsgNode.Get())
		}
	}

	// events service (event subscription)
	if eventsNode := rpcNode.At(serviceEvents); eventsNode.GetType() == jsongo.TypeArray {
		for j := 0; j < eventsNode.Len(); j++ {
			event := eventsNode.At(j).Get().(string)
			rpc.Events = append(rpc.Events, event)
		}
	}

	return &rpc, nil
}
