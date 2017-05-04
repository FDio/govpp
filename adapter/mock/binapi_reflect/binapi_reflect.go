// Copyright (c) 2017 Cisco and/or its affiliates.
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

// Package binapi_reflect is a helper package for generic handling of VPP binary API messages
// in the mock adapter and integration tests.
package binapi_reflect

import (
	"reflect"
)

const SwIfIndex = "SwIfIndex"
const Retval = "Retval"
const Reply = "_reply"

// TODO comment
func FindFieldOfType(reply reflect.Type, fieldName string) (reflect.StructField, bool) {
	if reply.Kind() == reflect.Struct {
		field, found := reply.FieldByName(fieldName)
		return field, found
	} else if reply.Kind() == reflect.Ptr && reply.Elem().Kind() == reflect.Struct {
		field, found := reply.Elem().FieldByName(fieldName)
		return field, found
	}
	return reflect.StructField{}, false
}

// TODO comment
func FindFieldOfValue(reply reflect.Value, fieldName string) (reflect.Value, bool) {
	if reply.Kind() == reflect.Struct {
		field := reply.FieldByName(fieldName)
		return field, field.IsValid()
	} else if reply.Kind() == reflect.Ptr && reply.Elem().Kind() == reflect.Struct {
		field := reply.Elem().FieldByName(fieldName)
		return field, field.IsValid()
	}
	return reflect.Value{}, false
}

// TODO comment
func IsReplySwIfIdx(reply reflect.Type) bool {
	_, found := FindFieldOfType(reply, SwIfIndex)
	return found
}

// TODO comment
func SetSwIfIdx(reply reflect.Value, swIfIndex uint32) {
	if field, found := FindFieldOfValue(reply, SwIfIndex); found {
		field.Set(reflect.ValueOf(swIfIndex))
	}
}

// TODO comment
func SetRetVal(reply reflect.Value, retVal int32) {
	if field, found := FindFieldOfValue(reply, Retval); found {
		field.Set(reflect.ValueOf(retVal))
	}
}

// TODO comment
func ReplyNameFor(request string) (string, bool) {
	return request + Reply, true
}
