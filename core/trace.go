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
	"sort"
	"sync"
	"time"

	"go.fd.io/govpp/api"
)

// default buffer size
const bufferSize = 100

// Trace is the default trace API implementation.
type Trace struct {
	*sync.RWMutex
	wg sync.WaitGroup

	records []*api.Record
	buffer  chan *api.Record
	index   int

	closeFunc func()
}

// NewTrace initializes the trace object, always bound to a GoVPP connection.
// The size limits the number of records stored.
// Initializing a new trace for the same connection replaces the old one and
// discards all values already collected.
func NewTrace(c *Connection, size int) (t *Trace) {
	t = &Trace{
		RWMutex: &sync.RWMutex{},
		records: make([]*api.Record, size),
		buffer:  make(chan *api.Record, bufferSize),
	}
	c.traceLock.Lock()
	c.trace = t
	c.traceLock.Unlock()
	t.closeFunc = func() {
		c.traceLock.Lock()
		c.trace = nil // no more records
		c.traceLock.Unlock()
		close(t.buffer)
	}
	go func() {
		for {
			record, ok := <-t.buffer
			if !ok {
				return
			}
			if t.index < len(t.records) {
				t.Lock()
				t.records[t.index] = record
				t.index++
				t.Unlock()
			}
			t.wg.Done()
		}
	}()
	return
}

func (t *Trace) GetRecords() (list []*api.Record) {
	// it is supposed to wait until all API messages sent to the
	// buffer are processed before returning the list
	t.wg.Wait()
	list = make([]*api.Record, t.index)
	t.RLock()
	copy(list, t.records[:t.index])
	t.RUnlock()
	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp.Before(list[j].Timestamp)
	})
	return list
}

func (t *Trace) GetRecordsForChannel(chId uint16) (list []*api.Record) {
	records := t.GetRecords()
	for _, record := range records {
		if record.ChannelID == chId {
			list = append(list, record)
		}
	}
	return list
}

func (t *Trace) Clear() {
	t.Lock()
	t.records = make([]*api.Record, len(t.records))
	t.index = 0
	t.Unlock()
}

func (t *Trace) Close() {
	t.closeFunc()
}

func (t *Trace) registerNew() (now time.Time, enabled bool) {
	if t != nil {
		t.wg.Add(1)
		enabled = true
	}
	return time.Now(), enabled
}

func (t *Trace) send(record *api.Record) {
	if t != nil {
		t.buffer <- record
	}
}
