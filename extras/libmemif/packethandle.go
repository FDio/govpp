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

package libmemif

import (
	"io"
	"sync"
	"time"

	"github.com/google/gopacket"
)

type memoizedPacket struct {
	data RawPacketData
	ci   gopacket.CaptureInfo
}

type MemifPacketHandle struct {
	memif   *Memif
	queueId uint8
	rxCount uint16

	// Used for caching packets when larger rxburst is called
	packetQueue []*memoizedPacket

	// Used for synchronization of read/write calls
	readMu  sync.Mutex
	writeMu sync.Mutex
	closeMu sync.Mutex
	stop    bool
}

// Create new GoPacket packet handle from libmemif queue. rxCount determines how many packets will be read
// at once, minimum value is 1
func (memif *Memif) NewPacketHandle(queueId uint8, rxCount uint16) *MemifPacketHandle {
	if rxCount == 0 {
		rxCount = 1
	}

	return &MemifPacketHandle{
		memif:   memif,
		queueId: queueId,
		rxCount: rxCount,
	}
}

// Reads packet data from memif in bursts, based on previously configured rxCount parameter. Then caches the
// resulting packets and returns them 1 by 1 from this method until queue is empty then tries to call new rx burst
// to read more data. If no data is returned, io.EOF error is thrown and caller should stop reading.
func (handle *MemifPacketHandle) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	handle.readMu.Lock()
	defer handle.readMu.Unlock()

	if handle.stop {
		err = io.EOF
		return
	}

	queueLen := len(handle.packetQueue)

	if queueLen == 0 {
		packets, burstErr := handle.memif.RxBurst(handle.queueId, handle.rxCount)
		packetsLen := len(packets)

		if burstErr != nil {
			err = burstErr
			return
		}

		if packetsLen == 0 {
			err = io.EOF
			return
		}

		handle.packetQueue = make([]*memoizedPacket, packetsLen)

		for i, packet := range packets {
			packetLen := len(packet)

			handle.packetQueue[i] = &memoizedPacket{
				data: []byte(packet),
				ci: gopacket.CaptureInfo{
					Timestamp:     time.Now(),
					CaptureLength: packetLen,
					Length:        packetLen,
				},
			}
		}
	}

	packet := handle.packetQueue[0]
	handle.packetQueue = handle.packetQueue[1:]
	data = packet.data
	ci = packet.ci

	return
}

// Writes packet data to memif in burst of 1 packet. In case no packet is sent, this method throws io.EOF error and
// called should stop trying to write packets.
func (handle *MemifPacketHandle) WritePacketData(data []byte) (err error) {
	handle.writeMu.Lock()
	defer handle.writeMu.Unlock()

	if handle.stop {
		err = io.EOF
		return
	}

	count, err := handle.memif.TxBurst(handle.queueId, []RawPacketData{data})

	if err != nil {
		return
	}

	if count == 0 {
		err = io.EOF
	}

	return
}

// Waits for all read and write operations to finish and then prevents more from occurring. Handle can be closed only
// once and then can never be opened again.
func (handle *MemifPacketHandle) Close() {
	handle.closeMu.Lock()
	defer handle.closeMu.Unlock()

	// wait for packet reader to stop
	handle.readMu.Lock()
	defer handle.readMu.Unlock()

	// wait for packet writer to stop
	handle.writeMu.Lock()
	defer handle.writeMu.Unlock()

	// stop reading and writing
	handle.stop = true
}
