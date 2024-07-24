/*
 *------------------------------------------------------------------
 * Copyright (c) 2020 Cisco and/or its affiliates.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *------------------------------------------------------------------
 */

package memif

import (
	"fmt"
	"syscall"
)

type MemifPacketBuffer struct {
	Buf    []byte
	Buflen int
}

// ReadPacket reads one packet form the shared memory and
// returns the number of bytes read
func (q *Queue) ReadPacket(pkt []byte) (uint32, error) {
	var mask uint32 = uint32(q.ring.size - 1)
	var slot int
	var lastSlot int
	var length uint32
	var offset uint32
	var pktOffset uint32 = 0
	var nSlots uint16
	var desc descBuf = newDescBuf()

	if q.i.args.IsMaster {
		slot = int(q.lastHead)
		lastSlot = q.readHead()
	} else {
		slot = int(q.lastTail)
		lastSlot = q.readTail()
	}

	nSlots = uint16(lastSlot - slot)
	if nSlots == 0 {
		goto refill
	}

	// copy descriptor from shm
	q.getDescBuf(slot&int(mask), desc)
	length = desc.getLength()
	offset = desc.getOffset()

	copy(pkt[:], q.i.regions[desc.getRegion()].data[offset:offset+length])
	pktOffset += length

	slot++
	nSlots--

	for (desc.getFlags() & descFlagNext) == descFlagNext {
		if nSlots == 0 {
			return 0, fmt.Errorf("Incomplete chained buffer, may suggest peer error.")
		}

		q.getDescBuf(slot&int(mask), desc)
		length = desc.getLength()
		offset = desc.getOffset()

		copy(pkt[pktOffset:], q.i.regions[desc.getRegion()].data[offset:offset+length])
		pktOffset += length

		slot++
		nSlots--
	}

refill:
	if q.i.args.IsMaster {
		q.lastHead = uint16(slot)
		q.writeTail(slot)
	} else {
		q.lastTail = uint16(slot)

		head := q.readHead()

		for nSlots := uint16(q.ring.size - head + int(q.lastTail)); nSlots > 0; nSlots-- {
			q.setDescLength(head&int(mask), int(q.i.run.PacketBufferSize))
			head++
		}
		q.writeHead(head)
	}

	return pktOffset, nil
}

// ReadPacket reads one packet form the shared memory and
// returns the number of packets
func (q *Queue) Rx_burst(pkt []MemifPacketBuffer) (uint16, error) {
	var mask uint32 = uint32(q.ring.size - 1)
	var slot int
	var lastSlot int
	var length uint32
	var offset uint32
	var nSlots uint16
	var readQueueInterrupt bool = true
	var desc descBuf = newDescBuf()

	if q.i.args.IsMaster {
		slot = int(q.lastHead)
		lastSlot = q.readHead()
	} else {
		slot = int(q.lastTail)
		lastSlot = q.readTail()
	}

	nSlots = uint16(lastSlot - slot)
	if nSlots == 0 {
		b := make([]byte, 8)
		syscall.Read(int(q.interruptFd), b)
		return 0, nil
	}

	if nSlots > uint16(len(pkt)) {
		nSlots = uint16(len(pkt))
		readQueueInterrupt = false
	}

	rx := 0
	for nSlots > 0 {
		// copy descriptor from shm
		q.getDescBuf(slot&int(mask), desc)
		length = desc.getLength()
		offset = desc.getOffset()
		copy(pkt[rx].Buf[:], q.i.regions[desc.getRegion()].data[offset:offset+length])
		pkt[rx].Buflen = int(length)
		rx++
		nSlots--
		slot++
	}

	if q.i.args.IsMaster {
		q.lastHead += uint16(rx)
	} else {
		q.lastTail += uint16(rx)

	}

	if readQueueInterrupt {
		b := make([]byte, 8)
		syscall.Read(int(q.interruptFd), b)
	}

	return uint16(rx), nil
}

func (q *Queue) Refill(count int) {
	var mask int = q.ring.size - 1

	counter := 0
	if q.i.args.IsMaster {
		if q.readTail()+count <= int(q.lastHead) {
			q.writeTail(q.readTail() + count)
		} else {
			q.writeTail(int(q.lastHead))
		}
	}

	head := q.readHead()
	slot := head
	ns := (1 << q.ring.log2Size) - head + int(q.lastTail)

	if count >= ns {
		count = ns
	}

	for counter < count {
		slot++
		counter++
	}
	for nSlots := uint16(q.ring.size - head + int(q.lastTail)); nSlots > 0; nSlots-- {
		q.setDescLength(head&mask, int(q.i.run.PacketBufferSize))
		head++
	}

	if !q.i.args.IsMaster {
		q.writeHead(head) //slot

	}
}
