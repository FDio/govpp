// Copyright (c) 2018 Cisco and/or its affiliates.
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

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/bennyscetbun/jsongo"
)

// Package represents collection of objects parsed from VPP binary API JSON data
type Package struct {
	APIVersion string
	Enums      []Enum
	Unions     []Union
	Types      []Type
	Messages   []Message
	RefMap     map[string]string
}

// MessageType represents the type of a VPP message
type MessageType int

const (
	requestMessage MessageType = iota // VPP request message
	replyMessage                      // VPP reply message
	eventMessage                      // VPP event message
	otherMessage                      // other VPP message
)

// Message represents VPP binary API message
type Message struct {
	Name   string
	CRC    string
	Fields []Field
}

// Type represents VPP binary API type
type Type struct {
	Name   string
	CRC    string
	Fields []Field
}

// Union represents VPP binary API union
type Union struct {
	Name   string
	CRC    string
	Size   int
	Fields []Field
}

// Field represents VPP binary API object field
type Field struct {
	Name     string
	Type     string
	Length   int
	SizeFrom string
}

func (f *Field) IsArray() bool {
	return f.Length > 0 || f.SizeFrom != ""
}

// Enum represents VPP binary API enum
type Enum struct {
	Name    string
	Type    string
	Entries []EnumEntry
}

// EnumEntry represents VPP binary API enum entry
type EnumEntry struct {
	Name  string
	Value interface{}
}

func getSizeOfType(typ *Type) (size int) {
	for _, field := range typ.Fields {
		if n := getBinapiTypeSize(field.Type); n > 0 {
			if field.Length > 0 {
				size += n * field.Length
			} else {
				size += n
			}
		}
	}
	return size
}

func getTypeByRef(ctx *context, ref string) *Type {
	for _, typ := range ctx.packageData.Types {
		if ref == toApiType(typ.Name) {
			return &typ
		}
	}
	return nil
}

func getUnionSize(ctx *context, union *Union) (maxSize int) {
	for _, field := range union.Fields {
		if typ := getTypeByRef(ctx, field.Type); typ != nil {
			if size := getSizeOfType(typ); size > maxSize {
				maxSize = size
			}
		}
	}
	return
}

// toApiType returns name that is used as type reference in VPP binary API
func toApiType(name string) string {
	return fmt.Sprintf("vl_api_%s_t", name)
}

// parsePackage parses provided JSON data into objects prepared for code generation
func parsePackage(ctx *context, jsonRoot *jsongo.JSONNode) (*Package, error) {
	logf(" %s contains: %d services, %d messages, %d types, %d enums, %d unions (version: %s)",
		ctx.packageName,
		jsonRoot.Map("services").Len(),
		jsonRoot.Map("messages").Len(),
		jsonRoot.Map("types").Len(),
		jsonRoot.Map("enums").Len(),
		jsonRoot.Map("unions").Len(),
		jsonRoot.Map("vl_api_version").Get(),
	)

	pkg := Package{
		APIVersion: jsonRoot.Map("vl_api_version").Get().(string),
		RefMap:     make(map[string]string),
	}

	// load enums
	enums := jsonRoot.Map("enums")
	pkg.Enums = make([]Enum, enums.Len())
	for i := 0; i < enums.Len(); i++ {
		enumNode := enums.At(i)

		enum, err := parseEnum(ctx, enumNode)
		if err != nil {
			return nil, err
		}
		pkg.Enums[i] = *enum
		pkg.RefMap[toApiType(enum.Name)] = enum.Name
	}

	// load types
	types := jsonRoot.Map("types")
	pkg.Types = make([]Type, types.Len())
	for i := 0; i < types.Len(); i++ {
		typNode := types.At(i)

		typ, err := parseType(ctx, typNode)
		if err != nil {
			return nil, err
		}
		pkg.Types[i] = *typ
		pkg.RefMap[toApiType(typ.Name)] = typ.Name
	}

	// load unions
	unions := jsonRoot.Map("unions")
	pkg.Unions = make([]Union, unions.Len())
	for i := 0; i < unions.Len(); i++ {
		unionNode := unions.At(i)

		union, err := parseUnion(ctx, unionNode)
		if err != nil {
			return nil, err
		}
		pkg.Unions[i] = *union
		pkg.RefMap[toApiType(union.Name)] = union.Name
	}

	// load messages
	messages := jsonRoot.Map("messages")
	pkg.Messages = make([]Message, messages.Len())
	for i := 0; i < messages.Len(); i++ {
		msgNode := messages.At(i)

		msg, err := parseMessage(ctx, msgNode)
		if err != nil {
			return nil, err
		}
		pkg.Messages[i] = *msg
	}

	// TODO: load services

	printPackage(&pkg)

	return &pkg, nil
}

// printPackage prints all loaded objects for package
func printPackage(pkg *Package) {
	if len(pkg.Enums) > 0 {
		logf("loaded %d enums:", len(pkg.Enums))
		for k, enum := range pkg.Enums {
			logf(" - enum #%d\t%+v", k, enum)
		}
	}
	if len(pkg.Unions) > 0 {
		logf("loaded %d unions:", len(pkg.Unions))
		for k, union := range pkg.Unions {
			logf(" - union #%d\t%+v", k, union)
		}
	}
	if len(pkg.Types) > 0 {
		logf("loaded %d types:", len(pkg.Types))
		for _, typ := range pkg.Types {
			logf(" - type: %q (%d fields)", typ.Name, len(typ.Fields))
		}
	}
	if len(pkg.Messages) > 0 {
		logf("loaded %d messages:", len(pkg.Messages))
		for _, msg := range pkg.Messages {
			logf(" - message: %q (%d fields)", msg.Name, len(msg.Fields))
		}
	}
}

// parseEnum parses VPP binary API enum object from JSON node
func parseEnum(ctx *context, enumNode *jsongo.JSONNode) (*Enum, error) {
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

	enum := Enum{
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
			entryVal := entry.At(1).Get()

			enum.Entries = append(enum.Entries, EnumEntry{
				Name:  entryName,
				Value: entryVal,
			})
		}
	}

	return &enum, nil
}

// parseUnion parses VPP binary API union object from JSON node
func parseUnion(ctx *context, unionNode *jsongo.JSONNode) (*Union, error) {
	if unionNode.Len() == 0 || unionNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for union specified")
	}

	unionName, ok := unionNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("union name is %T, not a string", unionNode.At(0).Get())
	}
	unionCRC, ok := unionNode.At(unionNode.Len() - 1).At("crc").Get().(string)
	if !ok {
		return nil, fmt.Errorf("union crc invalid or missing")
	}

	union := Union{
		Name: unionName,
		CRC:  unionCRC,
	}

	// loop through union fields, skip first (name) and last (crc)
	for j := 1; j < unionNode.Len()-1; j++ {
		if unionNode.At(j).GetType() == jsongo.TypeArray {
			fieldNode := unionNode.At(j)

			field, err := parseField(ctx, fieldNode)
			if err != nil {
				return nil, err
			}

			union.Fields = append(union.Fields, *field)
		}
	}

	return &union, nil
}

// parseType parses VPP binary API type object from JSON node
func parseType(ctx *context, typeNode *jsongo.JSONNode) (*Type, error) {
	if typeNode.Len() == 0 || typeNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for type specified")
	}

	typeName, ok := typeNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("type name is %T, not a string", typeNode.At(0).Get())
	}
	typeCRC, ok := typeNode.At(typeNode.Len() - 1).At("crc").Get().(string)
	if !ok {
		return nil, fmt.Errorf("type crc invalid or missing")
	}

	typ := Type{
		Name: typeName,
		CRC:  typeCRC,
	}

	// loop through type fields, skip first (name) and last (crc)
	for j := 1; j < typeNode.Len()-1; j++ {
		if typeNode.At(j).GetType() == jsongo.TypeArray {
			fieldNode := typeNode.At(j)

			field, err := parseField(ctx, fieldNode)
			if err != nil {
				return nil, err
			}

			typ.Fields = append(typ.Fields, *field)
		}
	}

	return &typ, nil
}

// parseMessage parses VPP binary API message object from JSON node
func parseMessage(ctx *context, msgNode *jsongo.JSONNode) (*Message, error) {
	if msgNode.Len() == 0 || msgNode.At(0).GetType() != jsongo.TypeValue {
		return nil, errors.New("invalid JSON for message specified")
	}

	msgName, ok := msgNode.At(0).Get().(string)
	if !ok {
		return nil, fmt.Errorf("message name is %T, not a string", msgNode.At(0).Get())
	}
	msgCRC, ok := msgNode.At(msgNode.Len() - 1).At("crc").Get().(string)
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

			field, err := parseField(ctx, fieldNode)
			if err != nil {
				return nil, err
			}

			msg.Fields = append(msg.Fields, *field)
		}
	}

	return &msg, nil
}

// parseField parses VPP binary API object field from JSON node
func parseField(ctx *context, field *jsongo.JSONNode) (*Field, error) {
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
	var fieldLength float64
	if field.Len() >= 3 {
		fieldLength, ok = field.At(2).Get().(float64)
		if !ok {
			return nil, fmt.Errorf("field length is %T, not an int", field.At(2).Get())
		}
	}
	var fieldLengthFrom string
	if field.Len() >= 4 {
		fieldLengthFrom, ok = field.At(3).Get().(string)
		if !ok {
			return nil, fmt.Errorf("field length from is %T, not a string", field.At(3).Get())
		}
	}

	return &Field{
		Name:     fieldName,
		Type:     fieldType,
		Length:   int(fieldLength),
		SizeFrom: fieldLengthFrom,
	}, nil
}

// convertToGoType translates the VPP binary API type into Go type
func convertToGoType(ctx *context, binapiType string) (typ string) {
	if t, ok := binapiTypes[binapiType]; ok {
		// basic types
		typ = t
	} else if r, ok := ctx.packageData.RefMap[binapiType]; ok {
		// specific types (enums/types/unions)
		typ = camelCaseName(strings.Title(r))
	} else {
		// fallback type
		log.Printf("found unknown VPP binary API type %q, using byte", binapiType)
		typ = "byte"
	}

	return typ
}

// isNotSpace returns true if the rune is NOT a whitespace character.
func isNotSpace(r rune) bool {
	return !unicode.IsSpace(r)
}

// camelCaseName returns correct name identifier (camelCase).
func camelCaseName(name string) (should string) {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}
	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		return name
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word
		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// Keep consistent case, which is lowercase only at the start.
			if w == 0 && unicode.IsLower(runes[w]) {
				u = strings.ToLower(u)
			}
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))
		} else if w > 0 && strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}
