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

// Trace gives access to the API trace tool, capturing outcoming and incoming messages
// to and from GoVPP.
type Trace interface {
	// Enable allows to enable or disable API trace for a connection.
	Enable(enable bool)

	// GetRecords retrieves all messages collected (from all channels if they are used)
	// since the point the trace was enabled or cleared.
	GetRecords() []*Record

	// GetRecordsForChannel retrieves messages collected by the given channel since
	// the point the trace was enabled or cleared.
	GetRecordsForChannel(chId uint16) []*Record

	// Clear erases messages captured so far.
	Clear()
}

// Record contains essential information about traced message, its timestamp and whether
// the message was received or sent
type Record struct {
	Message    Message
	Timestamp  time.Time
	IsReceived bool
	ChannelID  uint16
}
