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
	"os"
	"strings"

	"github.com/bennyscetbun/jsongo"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	if strings.Contains(os.Getenv("DEBUG_GOVPP"), "parser") {
		Logger = logrus.StandardLogger()
	}
}

func logf(f string, v ...interface{}) {
	if Logger != nil {
		Logger.Printf(f, v...)
	}
}

const (
	// file
	objAPIVersion = "vl_api_version"
	objTypes      = "types"
	objMessages   = "messages"
	objUnions     = "unions"
	objEnums      = "enums"
	objServices   = "services"
	objAliases    = "aliases"
	objOptions    = "options"
	objImports    = "imports"

	// message
	messageFieldCrc = "crc"

	// alias
	aliasFieldLength = "length"
	aliasFieldType   = "type"

	// service
	serviceFieldReply     = "reply"
	serviceFieldStream    = "stream"
	serviceFieldStreamMsg = "stream_msg"
	serviceFieldEvents    = "events"
)

const (
	// file
	fileOptionVersion = "version"

	// field
	fieldOptionLimit   = "limit"
	fieldOptionDefault = "default"

	// service
	serviceReplyNull = "null"
)

func parseJSON(data []byte) (module *File, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("recovered panic: %v", e)
		}
	}()

	// parse JSON data into objects
	jsonRoot := new(jsongo.Node)
	if err := json.Unmarshal(data, jsonRoot); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON failed: %v", err)
	}

	logf("parsed module contains:")
	for _, key := range jsonRoot.GetKeys() {
		logf("  - %2d %s", jsonRoot.At(key).Len(), key)
	}

	module = new(File)

	// parse CRC
	if crc := jsonRoot.At(objAPIVersion); crc.GetType() == jsongo.TypeValue {
		module.CRC = crc.Get().(string)
	}

	// parse options
	opt := jsonRoot.Map(objOptions)
	if opt.GetType() == jsongo.TypeMap {
		module.Options = make(map[string]string, 0)
		for _, key := range opt.GetKeys() {
			optionsNode := opt.At(key)
			optionKey := key.(string)
			optionValue := optionsNode.Get().(string)
			module.Options[optionKey] = optionValue
		}
	}

	// parse imports
	imports := jsonRoot.Map(objImports)
	module.Imports = make([]string, 0)
	imported := make(map[string]struct{})
	for i := 0; i < imports.Len(); i++ {
		importNode := imports.At(i)
		imp, err := parseImport(importNode)
		if err != nil {
			return nil, err
		}
		if _, ok := imported[*imp]; ok {
			logf("duplicate import found: %v", *imp)
			continue
		}
		imported[*imp] = struct{}{}
		module.Imports = append(module.Imports, *imp)
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
	enumsNode := jsonRoot.Map(objEnums)
	module.EnumTypes = make([]EnumType, 0)
	for i := 0; i < enumsNode.Len(); i++ {
		enumNode := enumsNode.At(i)
		enum, err := parseEnum(enumNode)
		if err != nil {
			return nil, err
		}
		if exists(enum.Name) {
			continue
		}
		module.EnumTypes = append(module.EnumTypes, *enum)
	}

	// parse alias types
	aliasesNode := jsonRoot.Map(objAliases)
	if aliasesNode.GetType() == jsongo.TypeMap {
		module.AliasTypes = make([]AliasType, 0)
		for _, key := range aliasesNode.GetKeys() {
			aliasNode := aliasesNode.At(key)
			aliasName := key.(string)
			alias, err := parseAlias(aliasName, aliasNode)
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
	typesNode := jsonRoot.Map(objTypes)
	module.StructTypes = make([]StructType, 0)
	for i := 0; i < typesNode.Len(); i++ {
		typNode := typesNode.At(i)
		structyp, err := parseStruct(typNode)
		if err != nil {
			return nil, err
		}
		if exists(structyp.Name) {
			continue
		}
		module.StructTypes = append(module.StructTypes, *structyp)
	}

	// parse union types
	unionsNode := jsonRoot.Map(objUnions)
	module.UnionTypes = make([]UnionType, 0)
	for i := 0; i < unionsNode.Len(); i++ {
		unionNode := unionsNode.At(i)
		union, err := parseUnion(unionNode)
		if err != nil {
			return nil, err
		}
		if exists(union.Name) {
			continue
		}
		module.UnionTypes = append(module.UnionTypes, *union)
	}

	// parse messages
	messagesNode := jsonRoot.Map(objMessages)
	if messagesNode.GetType() == jsongo.TypeArray {
		module.Messages = make([]Message, messagesNode.Len())
		for i := 0; i < messagesNode.Len(); i++ {
			msgNode := messagesNode.At(i)
			msg, err := parseMessage(msgNode)
			if err != nil {
				return nil, err
			}
			module.Messages[i] = *msg
		}
	}

	// parse services
	servicesNode := jsonRoot.Map(objServices)
	if servicesNode.GetType() == jsongo.TypeMap {
		module.Service = &Service{
			RPCs: make([]RPC, servicesNode.Len()),
		}
		for i, key := range servicesNode.GetKeys() {
			rpcNode := servicesNode.At(key)
			rpcName := key.(string)
			svc, err := parseServiceRPC(rpcName, rpcNode)
			if err != nil {
				return nil, err
			}
			module.Service.RPCs[i] = *svc
		}
	}

	return module, nil
}

// parseImport parses VPP binary API import from JSON node
func parseImport(importNode *jsongo.Node) (*string, error) {
	if importNode.GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for import specified")
	}

	importName, ok := importNode.Get().(string)
	if !ok {
		return nil, fmt.Errorf("import name is %T, not a string", importNode.Get())
	}

	return &importName, nil
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
	enumType, ok := enumNode.At(enumNode.Len() - 1).At("enumtype").Get().(string)
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
	if aliasNode.Len() == 0 || aliasNode.At(aliasFieldType).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for alias specified")
	}

	alias := AliasType{
		Name: aliasName,
	}

	if typeNode := aliasNode.At(aliasFieldType); typeNode.GetType() == jsongo.TypeValue {
		typ, ok := typeNode.Get().(string)
		if !ok {
			return nil, fmt.Errorf("alias type is %T, not a string", typeNode.Get())
		}
		if typ != "null" {
			alias.Type = typ
		}
	}

	if lengthNode := aliasNode.At(aliasFieldLength); lengthNode.GetType() == jsongo.TypeValue {
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
	if msgNode.Len() == 0 || msgNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for message specified")
	}

	msgName, ok := msgNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("message name is %T, not a string", msgNode.At(0).Get())
	}
	msgCRC, ok := msgNode.At(msgNode.Len() - 1).At(messageFieldCrc).Get().(string)
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

			for _, key := range fieldMeta.GetKeys() {
				metaNode := fieldMeta.At(key)
				metaName := key.(string)
				metaValue := metaNode.Get()

				switch metaName {
				case fieldOptionLimit:
					metaValue = int(metaNode.Get().(float64))
				case fieldOptionDefault:
					metaValue = metaNode.Get()
				default:
					logrus.Warnf("unknown meta info (%s=%v) for field (%s)", metaName, metaValue, fieldName)
				}

				if f.Meta == nil {
					f.Meta = map[string]interface{}{}
				}
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
	if rpcNode.Len() == 0 || rpcNode.At(serviceFieldReply).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for service RPC specified")
	}

	rpc := RPC{
		Name:       rpcName,
		RequestMsg: rpcName,
	}

	if replyNode := rpcNode.At(serviceFieldReply); replyNode.GetType() == jsongo.TypeValue {
		reply, ok := replyNode.Get().(string)
		if !ok {
			return nil, fmt.Errorf("service RPC reply is %T, not a string", replyNode.Get())
		}
		if reply != serviceReplyNull {
			rpc.ReplyMsg = reply
		}
	}

	// is stream (dump)
	if streamNode := rpcNode.At(serviceFieldStream); streamNode.GetType() == jsongo.TypeValue {
		var ok bool
		rpc.Stream, ok = streamNode.Get().(bool)
		if !ok {
			return nil, fmt.Errorf("service RPC stream is %T, not a boolean", streamNode.Get())
		}
	}

	// stream message
	if streamMsgNode := rpcNode.At(serviceFieldStreamMsg); streamMsgNode.GetType() == jsongo.TypeValue {
		var ok bool
		rpc.StreamMsg, ok = streamMsgNode.Get().(string)
		if !ok {
			return nil, fmt.Errorf("service RPC stream msg is %T, not a string", streamMsgNode.Get())
		}
	}

	// events service (event subscription)
	if eventsNode := rpcNode.At(serviceFieldEvents); eventsNode.GetType() == jsongo.TypeArray {
		for j := 0; j < eventsNode.Len(); j++ {
			event := eventsNode.At(j).Get().(string)
			rpc.Events = append(rpc.Events, event)
		}
	}

	return &rpc, nil
}
