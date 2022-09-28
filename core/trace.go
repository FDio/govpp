// Copyright (c) 2021 Cisco and/or its affiliates.
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

package core

import (
	"go.fd.io/govpp/api"
	"sort"
	"sync"
	"sync/atomic"
)

// trace is the API tracer object synchronizing and keeping recoded messages.
type trace struct {
	list []*api.Record
	mux  *sync.Mutex

	isEnabled int32
}

func (c *trace) Enable(enable bool) {
	if enable && atomic.CompareAndSwapInt32(&c.isEnabled, 0, 1) {
		log.Debugf("API trace enabled")
	} else if atomic.CompareAndSwapInt32(&c.isEnabled, 1, 0) {
		log.Debugf("API trace disabled")
	}
}

func (c *trace) GetRecords() (list []*api.Record) {
	c.mux.Lock()
	list = append(list, c.list...)
	c.mux.Unlock()
	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp.Before(list[j].Timestamp)
	})
	return list
}

func (c *trace) GetRecordsForChannel(chId uint16) (list []*api.Record) {
	c.mux.Lock()
	for _, entry := range c.list {
		if entry.ChannelID == chId {
			list = append(list, entry)
		}
	}
	c.mux.Unlock()
	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp.Before(list[j].Timestamp)
	})
	return list
}

func (c *trace) Clear() {
	c.mux.Lock()
	c.list = make([]*api.Record, 0)
	c.mux.Unlock()
}
