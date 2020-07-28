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
	"reflect"
	"sync/atomic"
	"time"

	logger "github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/api"
)

var ReplyChannelTimeout = time.Millisecond * 100

var (
	ErrNotConnected = errors.New("not connected to VPP, ignoring the request")
	ErrProbeTimeout = errors.New("probe reply not received within timeout period")
)

// watchRequests watches for requests on the request API channel and forwards them as messages to VPP.
func (c *Connection) watchRequests(ch *Channel) {
	for {
		select {
		case req, ok := <-ch.reqChan:
			// new request on the request channel
			if !ok {
				// after closing the request channel, release API channel and return
				c.releaseAPIChannel(ch)
				return
			}
			if err := c.processRequest(ch, req); err != nil {
				sendReply(ch, &vppReply{
					seqNum: req.seqNum,
					err:    fmt.Errorf("unable to process request: %w", err),
				})
			}
		}
	}
}

// processRequest processes a single request received on the request channel.
func (c *Connection) sendMessage(context uint32, msg api.Message) error {
	// check whether we are connected to VPP
	if atomic.LoadUint32(&c.vppConnected) == 0 {
		return ErrNotConnected
	}

	/*log := log.WithFields(logger.Fields{
		"context":  context,
		"msg_name": msg.GetMessageName(),
		"msg_crc":  msg.GetCrcString(),
	})*/

	// retrieve message ID
	msgID, err := c.GetMessageID(msg)
	if err != nil {
		//log.WithError(err).Debugf("unable to retrieve message ID: %#v", msg)
		return err
	}

	//log = log.WithField("msg_id", msgID)

	// encode the message
	data, err := c.codec.EncodeMsg(msg, msgID)
	if err != nil {
		log.WithError(err).Debugf("unable to encode message: %#v", msg)
		return err
	}

	//log = log.WithField("msg_length", len(data))

	if log.Level >= logger.DebugLevel {
		log.Debugf("--> SEND: MSG %T %+v", msg, msg)
	}

	// send message to VPP
	err = c.vppClient.SendMsg(context, data)
	if err != nil {
		log.WithError(err).Debugf("unable to send message: %#v", msg)
		return err
	}

	return nil
}

// processRequest processes a single request received on the request channel.
func (c *Connection) processRequest(ch *Channel, req *vppRequest) error {
	// check whether we are connected to VPP
	if atomic.LoadUint32(&c.vppConnected) == 0 {
		err := ErrNotConnected
		log.WithFields(logger.Fields{
			"channel":  ch.id,
			"seq_num":  req.seqNum,
			"msg_name": req.msg.GetMessageName(),
			"msg_crc":  req.msg.GetCrcString(),
			"error":    err,
		}).Warnf("Unable to process request")
		return err
	}

	// retrieve message ID
	msgID, err := c.GetMessageID(req.msg)
	if err != nil {
		log.WithFields(logger.Fields{
			"channel":  ch.id,
			"msg_name": req.msg.GetMessageName(),
			"msg_crc":  req.msg.GetCrcString(),
			"seq_num":  req.seqNum,
			"error":    err,
		}).Warnf("Unable to retrieve message ID")
		return err
	}

	// encode the message into binary
	data, err := c.codec.EncodeMsg(req.msg, msgID)
	if err != nil {
		log.WithFields(logger.Fields{
			"channel":  ch.id,
			"msg_id":   msgID,
			"msg_name": req.msg.GetMessageName(),
			"msg_crc":  req.msg.GetCrcString(),
			"seq_num":  req.seqNum,
			"error":    err,
		}).Warnf("Unable to encode message: %T %+v", req.msg, req.msg)
		return err
	}

	context := packRequestContext(ch.id, req.multi, req.seqNum)

	if log.Level >= logger.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		log.WithFields(logger.Fields{
			"channel":  ch.id,
			"msg_id":   msgID,
			"msg_name": req.msg.GetMessageName(),
			"msg_crc":  req.msg.GetCrcString(),
			"seq_num":  req.seqNum,
			"is_multi": req.multi,
			"context":  context,
			"data_len": len(data),
		}).Debugf("--> SEND MSG: %T %+v", req.msg, req.msg)
	}

	// send the request to VPP
	err = c.vppClient.SendMsg(context, data)
	if err != nil {
		log.WithFields(logger.Fields{
			"channel":  ch.id,
			"msg_id":   msgID,
			"msg_name": req.msg.GetMessageName(),
			"msg_crc":  req.msg.GetCrcString(),
			"seq_num":  req.seqNum,
			"is_multi": req.multi,
			"context":  context,
			"data_len": len(data),
			"error":    err,
		}).Warnf("Unable to send message")
		return err
	}

	if req.multi {
		// send a control ping to determine end of the multipart response
		pingData, _ := c.codec.EncodeMsg(c.msgControlPing, c.pingReqID)

		if log.Level >= logger.DebugLevel {
			log.WithFields(logger.Fields{
				"channel":  ch.id,
				"msg_id":   c.pingReqID,
				"msg_name": c.msgControlPing.GetMessageName(),
				"msg_crc":  c.msgControlPing.GetCrcString(),
				"seq_num":  req.seqNum,
				"context":  context,
				"data_len": len(pingData),
			}).Debugf(" -> SEND MSG: %T", c.msgControlPing)
		}

		if err := c.vppClient.SendMsg(context, pingData); err != nil {
			log.WithFields(logger.Fields{
				"context": context,
				"seq_num": req.seqNum,
				"error":   err,
			}).Warnf("unable to send control ping")
		}
	}

	return nil
}

// msgCallback is called whenever any binary API message comes from VPP.
func (c *Connection) msgCallback(msgID uint16, data []byte) {
	if c == nil {
		log.WithField(
			"msg_id", msgID,
		).Warn("Connection already disconnected, ignoring the message.")
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
	context, err := c.codec.DecodeMsgContext(data, msg)
	if err != nil {
		log.WithField("msg_id", msgID).Warnf("Unable to decode message context: %v", err)
		return
	}

	chanID, isMulti, seqNum := unpackRequestContext(context)

	if log.Level == logger.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		msg = reflect.New(reflect.TypeOf(msg).Elem()).Interface().(api.Message)

		// decode the message
		if err = c.codec.DecodeMsg(data, msg); err != nil {
			err = fmt.Errorf("decoding message failed: %w", err)
			return
		}

		log.WithFields(logger.Fields{
			"context":  context,
			"msg_id":   msgID,
			"msg_size": len(data),
			"channel":  chanID,
			"is_multi": isMulti,
			"seq_num":  seqNum,
			"msg_crc":  msg.GetCrcString(),
		}).Debugf("<-- govpp RECEIVE: %s %+v", msg.GetMessageName(), msg)
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

	// send the data to the channel, it needs to be copied,
	// because it will be freed after this function returns
	sendReply(ch, &vppReply{
		msgID:        msgID,
		seqNum:       seqNum,
		data:         append([]byte(nil), data...),
		lastReceived: lastReplyReceived,
	})

	// store actual time of this reply
	c.lastReplyLock.Lock()
	c.lastReply = time.Now()
	c.lastReplyLock.Unlock()
}

// sendReply sends the reply into the go channel, if it cannot be completed without blocking, otherwise
// it logs the error and do not send the message.
func sendReply(ch *Channel, reply *vppReply) {
	// first try to avoid creating timer
	select {
	case ch.replyChan <- reply:
		return // reply sent ok
	default:
		// reply channel full
	}
	if ch.receiveReplyTimeout == 0 {
		log.WithFields(logger.Fields{
			"channel": ch.id,
			"msg_id":  reply.msgID,
			"seq_num": reply.seqNum,
			"err":     reply.err,
		}).Warn("Reply channel full, dropping reply.")
		return
	}
	select {
	case ch.replyChan <- reply:
		return // reply sent ok
	case <-time.After(ch.receiveReplyTimeout):
		// receiver still not ready
		log.WithFields(logger.Fields{
			"channel": ch.id,
			"msg_id":  reply.msgID,
			"seq_num": reply.seqNum,
			"err":     reply.err,
		}).Warnf("Unable to send reply (reciever end not ready in %v).", ch.receiveReplyTimeout)
	}
}

// isNotificationMessage returns true if someone has subscribed to provided message ID.
func (c *Connection) isNotificationMessage(msgID uint16) bool {
	c.subscriptionsLock.RLock()
	defer c.subscriptionsLock.RUnlock()

	_, exists := c.subscriptions[msgID]
	return exists
}

// sendNotifications send a notification message to all subscribers subscribed for that message.
func (c *Connection) sendNotifications(msgID uint16, data []byte) {
	c.subscriptionsLock.RLock()

	matched := false

	// send to notification to each subscriber
	for _, sub := range c.subscriptions[msgID] {
		log.WithFields(logger.Fields{
			"msg_name": sub.event.GetMessageName(),
			"msg_id":   msgID,
			"msg_size": len(data),
		}).Debug("Sending a notification to the subscription channel.")

		event := sub.msgFactory()
		if err := c.codec.DecodeMsg(data, event); err != nil {
			log.WithFields(logger.Fields{
				"msg_name": sub.event.GetMessageName(),
				"msg_id":   msgID,
				"msg_size": len(data),
				"error":    err,
			}).Warnf("Unable to decode the notification message")
			continue
		}

		matched = true

		if sub.notifFn != nil {
			defer sub.notifFn(event) // defer until the lock is released
			continue
		}

		// send the message into the go channel of the subscription
		select {
		case sub.notifChan <- event:
			// message sent successfully
		default:
			// unable to write into the channel without blocking
			log.WithFields(logger.Fields{
				"msg_name": sub.event.GetMessageName(),
				"msg_id":   msgID,
				"msg_size": len(data),
			}).Warn("Unable to deliver the notification, reciever end not ready.")
		}
	}

	if !matched {
		log.WithFields(logger.Fields{
			"msg_id":   msgID,
			"msg_size": len(data),
		}).Info("No subscription found for the notification message.")
	}

	c.subscriptionsLock.RUnlock()
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
