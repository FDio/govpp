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

	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/api"
)

var ReplyChannelTimeout = time.Millisecond * 100

var (
	ErrNotConnected = errors.New("not connected to VPP, ignoring the request")
	ErrProbeTimeout = errors.New("probe reply not received within timeout period")
	ErrReplyTimeout = errors.New("no reply received within the timeout period")
)

// watchRequests watches for requests on the request API channel and forwards them as messages to VPP.
func (c *Connection) watchRequests(ch *Channel) {
	for {
		req, ok := <-ch.reqChan
		// new request on the request channel
		if !ok {
			// after closing the request channel, release API channel and return
			c.releaseAPIChannel(ch)
			return
		}
		if err := c.processRequest(ch, req); err != nil {
			sendReply(ch.logger, ch, &vppReply{
				seqNum: req.seqNum,
				err:    fmt.Errorf("unable to process request: %w", err),
			})
		}
	}
}

// processRequest processes a single request received on the request channel.
func (c *Connection) processRequest(ch *Channel, req *vppRequest) error {
	l := ch.logger.WithFields(logrus.Fields{
		"chanId":  ch.id,
		"seqNum":  req.seqNum,
		"msgName": req.msg.GetMessageName(),
		"msgCrc":  req.msg.GetCrcString(),
		"isMulti": req.multi,
	})

	// check whether we are connected to VPP
	if atomic.LoadUint32(&c.vppConnected) == 0 {
		err := ErrNotConnected
		l.WithField("error", err).Warnf("Unable to process request")
		return err
	}

	// retrieve message ID
	msgID, err := c.GetMessageID(req.msg)
	if err != nil {
		l.WithField("error", err).Warnf("Unable to retrieve message ID")
		return err
	}

	l = l.WithFields(logrus.Fields{
		"msgId": msgID,
	})

	// encode the message into binary
	data, err := c.codec.EncodeMsg(req.msg, msgID)
	if err != nil {
		l.WithField("error", err).Warnf("Unable to encode message: %T %+v", req.msg, req.msg)
		return err
	}

	context := packRequestContext(ch.id, req.multi, req.seqNum)

	l = l.WithFields(logrus.Fields{
		"context": context,
		"msgLen":  len(data),
	})

	if log.Level >= logrus.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		l.Debugf("-->govpp SEND: %T %+v", req.msg, req.msg)
	}

	var timestamp time.Time
	{
		c.traceLock.Lock()
		if c.trace != nil {
			timestamp, _ = c.trace.registerNew()
		}
		c.traceLock.Unlock()
		// send the request to VPP
		if err := c.vppClient.SendMsg(context, data); err != nil {
			c.traceLock.Lock()
			if c.trace != nil {
				c.trace.send(&api.Record{
					Message:   req.msg,
					Timestamp: timestamp,
					ChannelID: ch.id,
					Succeeded: false,
				})
			}
			c.traceLock.Unlock()
			l.WithField("error", err).Warnf("Unable to send message")
			return err
		}
	}

	c.traceLock.Lock()
	if c.trace != nil {
		c.trace.send(&api.Record{
			Message:   req.msg,
			Timestamp: timestamp,
			ChannelID: ch.id,
			Succeeded: true,
		})
	}
	c.traceLock.Unlock()

	if req.multi {
		// send a control ping to determine end of the multipart response
		pingData, _ := c.codec.EncodeMsg(c.msgControlPing, c.pingReqID)

		if log.Level >= logrus.DebugLevel {
			l.WithField("error", err).Debugf("-->govpp SEND PING: %T", c.msgControlPing)
		}
		c.traceLock.Lock()
		if c.trace != nil {
			timestamp, _ = c.trace.registerNew()
		}
		c.traceLock.Unlock()
		// send the control ping request to VPP
		if err := c.vppClient.SendMsg(context, pingData); err != nil {
			if c.trace != nil {
				c.traceLock.Lock()
				c.trace.send(&api.Record{
					Message:   c.msgControlPing,
					Timestamp: timestamp,
					ChannelID: ch.id,
					Succeeded: false,
				})
				c.traceLock.Unlock()
			}
			l.WithField("error", err).Warnf("unable to send control ping")
		} else {
			if c.trace != nil {
				c.traceLock.Lock()
				c.trace.send(&api.Record{
					Message:   c.msgControlPing,
					Timestamp: timestamp,
					ChannelID: ch.id,
					Succeeded: true,
				})
				c.traceLock.Unlock()
			}
		}
	}

	return nil
}

// msgCallback is called whenever any binary API message comes from VPP.
func (c *Connection) msgCallback(msgID uint16, data []byte) {
	var l logrus.Ext1FieldLogger
	if c.logger == nil {
		l = logrus.StandardLogger()
	} else {
		l = c.logger.WithFields(logrus.Fields{
			"msgId":  msgID,
			"msgLen": len(data),
		})
	}

	if c == nil {
		l.Warn("Connection already disconnected, ignoring the message.")
		return
	}

	msg, err := c.getMessageByID(msgID)
	if err != nil {
		c.logger.Warnln("Unable to get message by ID", err)
		return
	}

	l = l.WithFields(logrus.Fields{
		"msgName": msg.GetMessageName(),
		"msgCrc":  msg.GetCrcString(),
	})

	// decode message context to fix for special cases of messages,
	// for example:
	// - replies that don't have context as first field (comes as zero)
	// - events that don't have context at all (comes as non zero)
	//
	context, err := c.codec.DecodeMsgContext(data, msg.GetMessageType())
	if err != nil {
		l.Warnf("Unable to decode message context: %v", err)
		return
	}

	l = l.WithField("context", context)

	chanID, isMulti, seqNum := unpackRequestContext(context)

	l = l.WithFields(logrus.Fields{
		"chanId":  chanID,
		"isMulti": isMulti,
		"seqNum":  seqNum,
	})

	var decoded bool

	// decode and trace the message
	c.traceLock.Lock()
	if c.trace != nil {
		var timestamp time.Time
		timestamp, _ = c.trace.registerNew()
		decoded = true
		msg = reflect.New(reflect.TypeOf(msg).Elem()).Interface().(api.Message)
		if err := c.codec.DecodeMsg(data, msg); err != nil {
			l.Debugf("Unable to decode message: %v", err)
		} else {
			c.trace.send(&api.Record{
				Message:    msg,
				Timestamp:  timestamp,
				IsReceived: true,
				ChannelID:  chanID,
				Succeeded:  err == nil,
			})
		}
	}
	c.traceLock.Unlock()

	if log.Level >= logrus.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		if !decoded {
			decoded = true
			msg = reflect.New(reflect.TypeOf(msg).Elem()).Interface().(api.Message)
			if err := c.codec.DecodeMsg(data, msg); err != nil {
				l.Debugf("Unable to decode message: %v", err)
			}
		}
		l.Debugf("<--govpp RECV: %T %+v", msg, msg)
	}

	if context == 0 || c.isNotificationMessage(msgID) {
		// process the message as a notification
		c.sendNotifications(l, msgID, data)
		return
	}

	// match ch according to the context
	c.channelsLock.RLock()
	ch, ok := c.channels[chanID]
	c.channelsLock.RUnlock()
	if !ok {
		if !decoded {
			msg = reflect.New(reflect.TypeOf(msg).Elem()).Interface().(api.Message)
			if err := c.codec.DecodeMsg(data, msg); err != nil {
				l.Debugf("Unable to decode message: %v", err)
			}
		}
		l.Errorf("Channel ID not known, ignoring the message: %T %+v", msg, msg)
		return
	}

	// if this is a control ping reply to a multipart request,
	// treat this as a last part of the reply
	lastReplyReceived := isMulti && msgID == c.pingReplyID

	// send the data to the channel, it needs to be copied,
	// because it will be freed after this function returns
	sendReply(l, ch, &vppReply{
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
func sendReply(l logrus.Ext1FieldLogger, ch *Channel, reply *vppReply) {
	// first try to avoid creating timer
	select {
	case ch.replyChan <- reply:
		return // reply sent ok
	default:
		// reply channel full
	}
	if ch.receiveReplyTimeout == 0 {
		l.WithField("error", reply.err).Warn("Reply channel full, dropping reply.")
		return
	}
	replyTimeoutTimer := time.NewTimer(ch.receiveReplyTimeout)
	defer replyTimeoutTimer.Stop()
	select {
	case ch.replyChan <- reply:
		return // reply sent ok
	case <-replyTimeoutTimer.C:
		// receiver still isn't ready
		l.WithField("error", reply.err).Warnf("Unable to send reply (reciever end not ready in %v).", ch.receiveReplyTimeout)
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
func (c *Connection) sendNotifications(l logrus.Ext1FieldLogger, msgID uint16, data []byte) {
	c.subscriptionsLock.RLock()
	defer c.subscriptionsLock.RUnlock()

	matched := false

	// send to notification to each subscriber
	for _, sub := range c.subscriptions[msgID] {
		l.Debug("Sending a notification to the subscription channel.")

		event := sub.msgFactory()
		if err := c.codec.DecodeMsg(data, event); err != nil {
			l.WithField("error", err).Warnf("Unable to decode the notification message")
			continue
		}

		// send the message into the go channel of the subscription
		select {
		case sub.notifChan <- event:
			// message sent successfully
		default:
			// unable to write into the channel without blocking
			l.Warn("Unable to deliver the notification, reciever end not ready.")
		}

		matched = true
	}

	if !matched {
		l.Info("No subscription found for the notification message.")
	}
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

// Returns message based on the message ID not depending on message path.
func (c *Connection) getMessageByID(msgID uint16) (msg api.Message, err error) {
	c.msgMapByPathLock.RLock()
	defer c.msgMapByPathLock.RUnlock()

	var ok bool
	for _, messages := range c.msgMapByPath {
		if msg, ok = messages[msgID]; ok {
			return msg, nil
		}
	}
	return nil, fmt.Errorf("unknown message received, ID: %d", msgID)
}
