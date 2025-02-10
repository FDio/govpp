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

	"github.com/bennyscetbun/jsongo"
)

const (
	// root keys
	fileVersionCrc = "vl_api_version"
	fileOptions    = "options"
	fileTypes      = "types"
	fileMessages   = "messages"
	fileUnions     = "unions"
	fileEnums      = "enums"
	fileEnumflags  = "enumflags"
	fileAliases    = "aliases"
	fileServices   = "services"
	fileImports    = "imports"
	fileCounters   = "counters"
	filePaths      = "paths"

	// message keys
	messageCrc     = "crc"
	messageOptions = "options"
	messageComment = "comment"

	// type keys
	enumType    = "enumtype"
	aliasLength = "length"
	aliasType   = "type"

	// counters
	counter         = "counter"
	counterName     = "name"
	counterPath     = "path"
	elements        = "elements"
	elementName     = "name"
	elementSeverity = "severity"
	elementType     = "type"
	elementUnits    = "units"
	elementDesc     = "description"

	// service
	serviceReply     = "reply"
	serviceStream    = "stream"
	serviceStreamMsg = "stream_msg"
	serviceEvents    = "events"
)

const (
	OptFileVersion = "version"
)

func parseApiJsonFile(data []byte) (file *File, err error) {
	// parse root json node
	rootNode := new(jsongo.Node)
	if err := json.Unmarshal(data, rootNode); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON failed: %w", err)
	}

	// print contents
	logf("file contents:")
	for _, rootKey := range rootNode.GetKeys() {
		keyNode := rootNode.At(rootKey)
		length := keyNode.Len()
		logf(" - %-15s %3d\t(type: %v)", rootKey, length, keyNode.GetType())
	}

	file = new(File)

	logf("parsing file CRC")

	// parse file CRC
	crcNode := rootNode.At(fileVersionCrc)
	if crcNode.GetType() == jsongo.TypeValue {
		file.CRC = crcNode.MustGetString()
	} else {
		logf("key %q expected to be TypeValue, but is %v", fileVersionCrc, crcNode.GetType())
	}

	logf("parsing file options")

	// parse file options
	optNode := rootNode.Map(fileOptions)
	if optNode.GetType() == jsongo.TypeMap {
		file.Options = make(map[string]string)
		for _, key := range optNode.GetKeys() {
			optionKey := key.(string)
			optionVal := optNode.At(key).MustGetString()
			file.Options[optionKey] = optionVal
		}
	} else {
		logf("key %q expected to be TypeMap, but is %v", fileOptions, optNode.GetType())
	}

	logf("parsing file imports")

	// parse imports
	importsNode := rootNode.Map(fileImports)
	if importsNode.GetType() == jsongo.TypeArray {
		file.Imports = make([]string, 0, importsNode.Len())
		uniq := make(map[string]struct{})
		for i := 0; i < importsNode.Len(); i++ {
			importNode := importsNode.At(i)
			imp := importNode.MustGetString()
			if _, ok := uniq[imp]; ok {
				logf("duplicate import found: %v", imp)
				continue
			}
			uniq[imp] = struct{}{}
			file.Imports = append(file.Imports, imp)
		}
	} else {
		logf("key %q expected to be TypeArray, but is %v", fileImports, importsNode.GetType())
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

	logf("parsing enums")

	// parse enum types
	enumsNode := rootNode.Map(fileEnums)
	if enumsNode.GetType() == jsongo.TypeArray {
		file.EnumTypes = make([]EnumType, 0)
		for i := 0; i < enumsNode.Len(); i++ {
			enum, err := parseEnum(enumsNode.At(i))
			if err != nil {
				return nil, err
			}
			if exists(enum.Name) {
				continue
			}
			file.EnumTypes = append(file.EnumTypes, *enum)
		}
	} else {
		logf("key %q expected to be TypeArray, but is %v", fileEnums, enumsNode.GetType())
	}

	logf("parsing enum flags")

	// parse enumflags types
	enumflagsNode := rootNode.Map(fileEnumflags)
	if enumflagsNode.GetType() == jsongo.TypeArray {
		file.EnumflagTypes = make([]EnumType, 0)
		for i := 0; i < enumflagsNode.Len(); i++ {
			enumflag, err := parseEnum(enumflagsNode.At(i))
			if err != nil {
				return nil, err
			}
			if exists(enumflag.Name) {
				continue
			}
			file.EnumflagTypes = append(file.EnumflagTypes, *enumflag)
		}
	} else {
		logf("key %q expected to be TypeArray, but is %v", fileEnumflags, enumflagsNode.GetType())
	}

	logf("parsing aliases")

	// parse alias types
	aliasesNode := rootNode.Map(fileAliases)
	if aliasesNode.GetType() == jsongo.TypeMap {
		file.AliasTypes = make([]AliasType, 0)
		for _, key := range aliasesNode.GetKeys() {
			aliasName := key.(string)
			alias, err := parseAlias(aliasName, aliasesNode.At(key))
			if err != nil {
				return nil, err
			}
			if exists(alias.Name) {
				continue
			}
			file.AliasTypes = append(file.AliasTypes, *alias)
		}
	} else {
		logf("key %q expected to be TypeMap, but is %v", fileAliases, aliasesNode.GetType())
	}

	logf("parsing types")

	// parse struct types
	typesNode := rootNode.Map(fileTypes)
	if typesNode.GetType() == jsongo.TypeArray {
		file.StructTypes = make([]StructType, 0)
		for i := 0; i < typesNode.Len(); i++ {
			structyp, err := parseStruct(typesNode.At(i))
			if err != nil {
				return nil, err
			}
			if exists(structyp.Name) {
				continue
			}
			file.StructTypes = append(file.StructTypes, *structyp)
		}
	} else {
		logf("key %q expected to be TypeArray, but is %v", fileTypes, typesNode.GetType())
	}

	logf("parsing unions")

	// parse union types
	unionsNode := rootNode.Map(fileUnions)
	if unionsNode.GetType() == jsongo.TypeArray {
		file.UnionTypes = make([]UnionType, 0)
		for i := 0; i < unionsNode.Len(); i++ {
			union, err := parseUnion(unionsNode.At(i))
			if err != nil {
				return nil, err
			}
			if exists(union.Name) {
				continue
			}
			file.UnionTypes = append(file.UnionTypes, *union)
		}
	} else {
		logf("key %q expected to be TypeArray, but is %v", fileUnions, unionsNode.GetType())
	}

	logf("parsing messages")

	// parse messages
	messagesNode := rootNode.Map(fileMessages)
	if messagesNode.GetType() == jsongo.TypeArray {
		file.Messages = make([]Message, messagesNode.Len())
		for i := 0; i < messagesNode.Len(); i++ {
			msg, err := parseMessage(messagesNode.At(i))
			if err != nil {
				return nil, err
			}
			file.Messages[i] = *msg
		}
	} else {
		logf("key %q expected to be TypeArray, but is %v", fileMessages, messagesNode.GetType())
	}

	logf("parsing services")

	// parse services
	servicesNode := rootNode.Map(fileServices)
	if servicesNode.GetType() == jsongo.TypeMap {
		file.Service = &Service{
			RPCs: make([]RPC, servicesNode.Len()),
		}
		for i, key := range servicesNode.GetKeys() {
			rpcName := key.(string)
			svc, err := parseServiceRPC(rpcName, servicesNode.At(key))
			if err != nil {
				return nil, err
			}
			file.Service.RPCs[i] = *svc
		}
	} else {
		logf("key %q expected to be TypeMap, but is %v", fileServices, servicesNode.GetType())
	}

	logf("parsing counters")

	// parse counters
	countersNode := rootNode.Map(fileCounters)
	if countersNode.GetType() == jsongo.TypeArray {
		file.Counters = make([]Counter, countersNode.Len())
		for i := 0; i < countersNode.Len(); i++ {
			c, err := parseCounter(countersNode.At(i))
			if err != nil {
				return nil, err
			}
			file.Counters[i] = *c
		}
	} else {
		logf("key %q expected to be TypeArray, but is %v", fileCounters, countersNode.GetType())
	}

	logf("parsing paths")

	// parse paths
	pathsNode := rootNode.Map(filePaths)
	if pathsNode.GetType() == jsongo.TypeArray {
		file.Paths = make([]CounterPaths, 0)
		for i := 0; i < pathsNode.Len(); i++ {
			counterPaths, err := parsePath(pathsNode.At(i))
			if err != nil {
				return nil, err
			} else if counterPaths == nil {
				continue
			}
			file.Paths = append(file.Paths, *counterPaths...)
		}
	} else {
		logf("key %q expected to be TypeArray, but is %v", filePaths, pathsNode.GetType())
	}

	return file, nil
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

	msgMeta := msgNode.At(msgNode.Len() - 1)

	// parse message CRC
	msgCRC, ok := msgMeta.At(messageCrc).Get().(string)
	if !ok {
		return nil, fmt.Errorf("message crc invalid or missing")
	}

	// parse message options
	var msgOpts map[string]string
	msgOptsNode := msgMeta.Map(messageOptions)
	if msgOptsNode.GetType() == jsongo.TypeMap {
		msgOpts = make(map[string]string)
		for _, opt := range msgOptsNode.GetKeys() {
			if _, ok := opt.(string); !ok {
				logf("invalid message option key, expected string")
				continue
			}
			msgOpts[opt.(string)] = ""
			if msgOptsNode.At(opt).Get() != nil {
				if optMsgStr, ok := msgOptsNode.At(opt).Get().(string); ok {
					msgOpts[opt.(string)] = optMsgStr
				} else {
					logf("invalid message option value, expected string")
				}
			}
		}
	}

	// parse message comment
	var msgComment string
	msgCommentNode := msgMeta.At(messageComment)
	if msgCommentNode.GetType() == jsongo.TypeValue {
		comment, ok := msgCommentNode.Get().(string)
		if !ok {
			logf("message comment value invalid, expected string")
		} else {
			msgComment = comment
		}
	}

	msg := Message{
		Name:    msgName,
		CRC:     msgCRC,
		Options: msgOpts,
		Comment: msgComment,
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

	// field length (array) or field meta
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
	// field size from
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

// parseCounter parses VPP binary API service object from JSON node
func parseCounter(counterNode *jsongo.Node) (*Counter, error) {
	// counter name
	c := &Counter{}
	if counterNameNode := counterNode.At(counterName); counterNameNode.GetType() == jsongo.TypeValue {
		var ok bool
		if c.Name, ok = counterNameNode.Get().(string); !ok {
			return nil, fmt.Errorf("counter name is %T, not a string", counterNameNode.Get())
		}
	}

	// counter elements
	if elementsNode := counterNode.At(elements); elementsNode.GetType() == jsongo.TypeArray {
		c.Elements = make([]Element, elementsNode.Len())
		for i := 0; i < elementsNode.Len(); i++ {
			if elementNode := elementsNode.At(i); elementNode.GetType() == jsongo.TypeMap {
				c.Elements[i] = Element{
					Name:        elementNode.At(elementName).Get().(string),
					Severity:    elementNode.At(elementSeverity).Get().(string),
					Type:        elementNode.At(elementType).Get().(string),
					Units:       elementNode.At(elementUnits).Get().(string),
					Description: elementNode.At(elementDesc).Get().(string),
				}
			}
		}
	}

	return c, nil
}

// parseCounter parses VPP binary API service object from JSON node
func parsePath(pathsNode *jsongo.Node) (*[]CounterPaths, error) {
	paths := make([]CounterPaths, 0)
	if pathsNode.GetType() == jsongo.TypeMap {
		// TODO: fix parsing for this case
		logf("PATHS NODE IS MAP")
		return nil, nil
	} /*else if pathsNode.GetType() == jsongo.TypeArray {
		// expected
	}*/
NodeLoop:
	for i := 0; i < pathsNode.Len(); i++ {
		pathNode := pathsNode.At(i)
		if pathValues := pathNode; pathValues.GetType() == jsongo.TypeMap {
			for j, path := range paths {
				if path.Name == pathNode.At(counter).Get().(string) {
					paths[j].Paths = append(path.Paths, pathNode.At(counterPath).Get().(string))
					continue NodeLoop
				}
			}
			path := CounterPaths{
				Name:  pathNode.At(counter).Get().(string),
				Paths: []string{pathNode.At(counterPath).Get().(string)},
			}
			paths = append(paths, path)
		}
	}
	return &paths, nil
}
