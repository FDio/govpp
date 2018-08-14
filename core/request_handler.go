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

package core

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	logger "github.com/sirupsen/logrus"
)

var (
	ErrNotConnected = errors.New("not connected to VPP, ignoring the request")
	ErrProbeTimeout = errors.New("probe reply not received within timeout period")
)

// watchRequests watches for requests on the request API channel and forwards them as messages to VPP.
func (c *Connection) watchRequests(ch *channel) {
	for {
		select {
		case req, ok := <-ch.reqChan:
			// new request on the request channel
			if !ok {
				// after closing the request channel, release API channel and return
				c.releaseAPIChannel(ch)
				return
			}
			c.processRequest(ch, req)

		case req := <-ch.notifSubsChan:
			// new request on the notification subscribe channel
			c.processSubscriptionRequest(ch, req)
		}
	}
}

// processRequest processes a single request received on the request channel.
func (c *Connection) processRequest(ch *channel, req *vppRequest) error {
	// check whether we are connected to VPP
	if atomic.LoadUint32(&c.connected) == 0 {
		err := ErrNotConnected
		log.Errorf("processing request failed: %v", err)
		sendReplyError(ch, req, err)
		return err
	}

	// retrieve message ID
	msgID, err := c.GetMessageID(req.msg)
	if err != nil {
		err = fmt.Errorf("unable to retrieve message ID: %v", err)
		log.WithFields(logger.Fields{
			"msg_name": req.msg.GetMessageName(),
			"msg_crc":  req.msg.GetCrcString(),
			"seq_num":  req.seqNum,
		}).Error(err)
		sendReplyError(ch, req, err)
		return err
	}

	// encode the message into binary
	data, err := c.codec.EncodeMsg(req.msg, msgID)
	if err != nil {
		err = fmt.Errorf("unable to encode the messge: %v", err)
		log.WithFields(logger.Fields{
			"channel":  ch.id,
			"msg_id":   msgID,
			"msg_name": req.msg.GetMessageName(),
			"seq_num":  req.seqNum,
		}).Error(err)
		sendReplyError(ch, req, err)
		return err
	}

	// get context
	context := packRequestContext(ch.id, req.multi, req.seqNum)
	if log.Level == logger.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		log.WithFields(logger.Fields{
			"channel":  ch.id,
			"context":  context,
			"is_multi": req.multi,
			"msg_id":   msgID,
			"msg_name": req.msg.GetMessageName(),
			"msg_size": len(data),
			"seq_num":  req.seqNum,
		}).Debug(" -> Sending a message to VPP.")
	}

	// send the request to VPP
	err = c.vpp.SendMsg(context, data)
	if err != nil {
		err = fmt.Errorf("unable to send the message: %v", err)
		log.WithFields(logger.Fields{
			"context": context,
			"msg_id":  msgID,
			"seq_num": req.seqNum,
		}).Error(err)
		sendReplyError(ch, req, err)
		return err
	}

	if req.multi {
		// send a control ping to determine end of the multipart response
		pingData, _ := c.codec.EncodeMsg(msgControlPing, c.pingReqID)

		log.WithFields(logger.Fields{
			"channel":  ch.id,
			"context":  context,
			"msg_id":   c.pingReqID,
			"msg_size": len(pingData),
			"seq_num":  req.seqNum,
		}).Debug(" -> Sending a control ping to VPP.")

		if err := c.vpp.SendMsg(context, pingData); err != nil {
			log.WithFields(logger.Fields{
				"context": context,
				"msg_id":  msgID,
				"seq_num": req.seqNum,
			}).Warnf("unable to send control ping: %v", err)
		}
	}

	return nil
}

// msgCallback is called whenever any binary API message comes from VPP.
func (c *Connection) msgCallback(msgID uint16, context uint32, data []byte) {
	connLock.RLock()
	defer connLock.RUnlock()

	if c == nil {
		log.Warn("Already disconnected, ignoring the message.")
		return
	}

	msg, ok := c.msgMap[msgID]
	if !ok {
		log.Warnf("Unknown message received, ID: %d", msgID)
		return
	}

	// decode message context to fix for special cases of messages,
	// for example:
	// - replies that don't have context as first field (comes as zero)
	// - events that don't have context at all (comes as non zero)
	//
	msgContext, err := c.codec.DecodeMsgContext(data, msg)
	if err == nil {
		if context != msgContext {
			log.Warnf("different context was decoded from message (%d -> %d)", context, msgContext)
			context = msgContext
		}
	} else {
		log.Errorf("decoding context failed: %v", err)
	}

	chanID, isMulti, seqNum := unpackRequestContext(context)
	if log.Level == logger.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		log.WithFields(logger.Fields{
			"context":  context,
			"msg_id":   msgID,
			"msg_name": msg.GetMessageName(),
			"msg_size": len(data),
			"channel":  chanID,
			"is_multi": isMulti,
			"seq_num":  seqNum,
		}).Debug(" <- Received a message from VPP.")
	}

	if context == 0 || c.isNotificationMessage(msgID) {
		// process the message as a notification
		c.sendNotifications(msgID, data)
		return
	}

	// match ch according to the context
	c.channelsLock.RLock()
	ch, ok := c.channels[chanID]
	c.channelsLock.RUnlock()
	if !ok {
		log.WithFields(logger.Fields{
			"channel": chanID,
			"msg_id":  msgID,
		}).Error("Channel ID not known, ignoring the message.")
		return
	}

	// if this is a control ping reply to a multipart request,
	// treat this as a last part of the reply
	lastReplyReceived := isMulti && msgID == c.pingReplyID

	// send the data to the channel
	sendReply(ch, &vppReply{
		msgID:        msgID,
		seqNum:       seqNum,
		data:         data,
		lastReceived: lastReplyReceived,
	})

	// store actual time of this reply
	c.lastReplyLock.Lock()
	c.lastReply = time.Now()
	c.lastReplyLock.Unlock()
}

// sendReply sends the reply into the go channel, if it cannot be completed without blocking, otherwise
// it logs the error and do not send the message.
func sendReply(ch *channel, reply *vppReply) {
	select {
	case ch.replyChan <- reply:
		// reply sent successfully
	case <-time.After(time.Millisecond * 100):
		// receiver still not ready
		log.WithFields(logger.Fields{
			"channel": ch,
			"msg_id":  reply.msgID,
			"seq_num": reply.seqNum,
		}).Warn("Unable to send the reply, reciever end not ready.")
	}
}

func sendReplyError(ch *channel, req *vppRequest, err error) {
	sendReply(ch, &vppReply{seqNum: req.seqNum, err: err})
}

// +------------------+-------------------+-----------------------+
// | 15b = channel ID | 1b = is multipart | 16b = sequence number |
// +------------------+-------------------+-----------------------+
func packRequestContext(chanID uint16, isMultipart bool, seqNum uint16) uint32 {
	context := uint32(chanID) << 17
	if isMultipart {
		context |= 1 << 16
	}
	context |= uint32(seqNum)
	return context
}

func unpackRequestContext(context uint32) (chanID uint16, isMulipart bool, seqNum uint16) {
	chanID = uint16(context >> 17)
	if ((context >> 16) & 0x1) != 0 {
		isMulipart = true
	}
	seqNum = uint16(context & 0xffff)
	return
}

// compareSeqNumbers returns -1, 0, 1 if sequence number <seqNum1> precedes, equals to,
// or succeeds seq. number <seqNum2>.
// Since sequence numbers cycle in the finite set of size 2^16, the function
// must assume that the distance between compared sequence numbers is less than
// (2^16)/2 to determine the order.
func compareSeqNumbers(seqNum1, seqNum2 uint16) int {
	// calculate distance from seqNum1 to seqNum2
	var dist uint16
	if seqNum1 <= seqNum2 {
		dist = seqNum2 - seqNum1
	} else {
		dist = 0xffff - (seqNum1 - seqNum2 - 1)
	}
	if dist == 0 {
		return 0
	} else if dist <= 0x8000 {
		return -1
	}
	return 1
}
