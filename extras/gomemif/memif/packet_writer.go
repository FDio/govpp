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

// WritePacket writes one packet to the shared memory and
// returns the number of bytes written
func (q *Queue) WritePacket(pkt []byte) int {
	var mask uint32 = uint32(q.ring.size - 1)
	var slot int
	var nFree uint16
	var packetBufferSize uint32 = q.i.run.PacketBufferSize

	if q.i.args.IsMaster {
		slot = q.readTail()
		nFree = uint16(q.readHead() - slot)
	} else {
		slot = q.readHead()
		nFree = uint16(q.ring.size - slot + q.readTail())
	}

	if nFree == 0 {
		q.interrupt()
		return 0
	}

	// copy descriptor from shm
	desc := newDescBuf()
	q.getDescBuf(slot&int(mask), desc)
	// reset flags
	desc.setFlags(0)
	// reset length
	if q.i.args.IsMaster {
		packetBufferSize = desc.getLength()
	}
	desc.setLength(0)
	offset := desc.getOffset()

	// write packet into memif buffer
	n := copy(q.i.regions[desc.getRegion()].data[offset:offset+packetBufferSize], pkt[:])
	desc.setLength(uint32(n))
	for n < len(pkt) {
		nFree--
		if nFree == 0 {
			q.interrupt()
			return 0
		}
		desc.setFlags(descFlagNext)
		q.putDescBuf(slot&int(mask), desc)
		slot++

		// copy descriptor from shm
		q.getDescBuf(slot&int(mask), desc)
		// reset flags
		desc.setFlags(0)
		// reset length
		if q.i.args.IsMaster {
			packetBufferSize = desc.getLength()
		}
		desc.setLength(0)
		offset := desc.getOffset()

		tmp := copy(q.i.regions[desc.getRegion()].data[offset:offset+packetBufferSize], pkt[:])
		desc.setLength(uint32(tmp))
		n += tmp
	}

	// copy descriptor to shm
	q.putDescBuf(slot&int(mask), desc)
	slot++

	if q.i.args.IsMaster {
		q.writeTail(slot)
	} else {
		q.writeHead(slot)
	}

	q.interrupt()

	return n
}

func (q *Queue) Tx_burst(pkt []MemifPacketBuffer) int {
	var mask uint32 = uint32(q.ring.size - 1)
	var slot int
	var nFree uint16
	var packetBufferSize uint32 = q.i.run.PacketBufferSize

	if q.i.args.IsMaster {
		slot = q.readTail()
		nFree = uint16(q.readHead() - slot)
	} else {
		slot = q.readHead()
		nFree = uint16(q.ring.size - slot + q.readTail())
	}

	if nFree == 0 {
		q.interrupt()
		return 0
	}

	var n int
	for i := 0; i < len(pkt); i++ {
		// copy descriptor from shm
		desc := newDescBuf()
		q.getDescBuf(slot&int(mask), desc)
		// reset flags
		desc.setFlags(0)
		// reset length
		if q.i.args.IsMaster {
			packetBufferSize = desc.getLength()
		}
		desc.setLength(0)
		offset := desc.getOffset()
		// write packet into memif buffer
		n = copy(q.i.regions[desc.getRegion()].data[offset:offset+packetBufferSize], pkt[i].Buf[:])
		desc.setLength(uint32(n))
		for n < len(pkt[i].Buf) {
			nFree--
			if nFree == 0 {
				q.interrupt()
				return 0
			}
			desc.setFlags(descFlagNext)
			q.putDescBuf(slot&int(mask), desc)
			slot++

			// copy descriptor from shm
			q.getDescBuf(slot&int(mask), desc)
			// reset flags
			desc.setFlags(0)
			// reset length
			if q.i.args.IsMaster {
				packetBufferSize = desc.getLength()
			}
			desc.setLength(0)
			offset := desc.getOffset()

			tmp := copy(q.i.regions[desc.getRegion()].data[offset:offset+packetBufferSize], pkt[i].Buf[:])
			desc.setLength(uint32(tmp))
			n += tmp
		}

		// copy descriptor to shm
		q.putDescBuf(slot&int(mask), desc)
		slot++
	}
	if q.i.args.IsMaster {
		q.writeTail(slot)
	} else {
		q.writeHead(slot)
	}

	q.interrupt()

	return n
}
