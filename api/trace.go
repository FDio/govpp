//  Copyright (c) 2021 Cisco and/or its affiliates.
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

package api

import (
	"time"
)

// Trace is the GoVPP utility tool capturing processed API messages. The trace is not operational
// by default.
// Enable trace for a given connection by calling `NewTrace(connection, size)`
type Trace interface {
	// GetRecords returns all API messages from all channels captured since the trace
	// was initialized or cleared up to the point of the call.
	GetRecords() []*Record

	// GetRecordsForChannel returns all API messages recorded by the given channel.
	GetRecordsForChannel(chId uint16) []*Record

	// Clear erases messages captured so far.
	Clear()

	// Close the tracer and release associated resources
	Close()
}

// Record contains essential information about the traced message, its timestamp and whether
// the message was received or sent
type Record struct {
	Message    Message
	Timestamp  time.Time
	IsReceived bool
	ChannelID  uint16
	Succeeded  bool
}
