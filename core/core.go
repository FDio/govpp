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

//go:generate binapi-generator --input-file=/usr/share/vpp/api/vpe.api.json --output-dir=./bin_api

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	logger "github.com/Sirupsen/logrus"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core/bin_api/vpe"
)

const (
	requestChannelBufSize = 100 // default size of the request channel buffers
	replyChannelBufSize   = 100 // default size of the reply channel buffers
)

// Connection represents a shared memory connection to VPP via vppAdapter.
type Connection struct {
	vpp   adapter.VppAdapter // VPP adapter
	codec *MsgCodec          // message codec

	msgIDs     map[string]uint16 // map of message IDs indexed by message name + CRC
	msgIDsLock sync.RWMutex      // lock for the message IDs map

	channels     map[uint32]*api.Channel // map of all API channels indexed by the channel ID
	channelsLock sync.RWMutex            // lock for the channels map

	notifSubscriptions     map[uint16][]*api.NotifSubscription // map od all notification subscriptions indexed by message ID
	notifSubscriptionsLock sync.RWMutex                        // lock for the subscriptions map

	maxChannelID uint32 // maximum used client ID
	pingReqID    uint16 // ID if the ControlPing message
	pingReplyID  uint16 // ID of the ControlPingReply message
}

// channelMetadata contains core-local metadata of an API channel.
type channelMetadata struct {
	id        uint32 // channel ID
	multipart uint32 // 1 if multipart request is being processed, 0 otherwise
}

var (
	log      *logger.Logger // global logger
	conn     *Connection    // global handle to the Connection (used in the message receive callback)
	connLock sync.RWMutex   // lock for the global connection
)

// init initializes global logger, which logs debug level messages to stdout.
func init() {
	log = logger.New()
	log.Out = os.Stdout
	log.Level = logger.DebugLevel
}

// SetLogger sets global logger to provided one.
func SetLogger(l *logger.Logger) {
	log = l
}

// Connect connects to VPP using specified VPP adapter and returns the connection handle.
func Connect(vppAdapter adapter.VppAdapter) (*Connection, error) {
	connLock.Lock()
	defer connLock.Unlock()

	if conn != nil {
		return nil, errors.New("only one connection per process is supported")
	}

	conn = &Connection{vpp: vppAdapter, codec: &MsgCodec{}}
	conn.channels = make(map[uint32]*api.Channel)
	conn.msgIDs = make(map[string]uint16)
	conn.notifSubscriptions = make(map[uint16][]*api.NotifSubscription)

	conn.vpp.SetMsgCallback(msgCallback)

	logger.Debug("Connecting to VPP...")

	err := conn.vpp.Connect()
	if err != nil {
		return nil, err
	}

	// store control ping IDs
	conn.pingReqID, _ = conn.GetMessageID(&vpe.ControlPing{})
	conn.pingReplyID, _ = conn.GetMessageID(&vpe.ControlPingReply{})

	logger.Debug("VPP connected.")

	return conn, nil
}

// Disconnect disconnects from VPP.
func (c *Connection) Disconnect() {
	if c == nil {
		return
	}
	connLock.Lock()
	defer connLock.Unlock()

	if c != nil && c.vpp != nil {
		c.vpp.Disconnect()
	}
	conn = nil
}

// NewAPIChannel returns a new API channel for communication with VPP via govpp core.
// It uses default buffer sizes for the request and reply Go channels.
func (c *Connection) NewAPIChannel() (*api.Channel, error) {
	if c == nil {
		return nil, errors.New("nil connection passed in")
	}
	return c.NewAPIChannelBuffered(requestChannelBufSize, replyChannelBufSize)
}

// NewAPIChannelBuffered returns a new API channel for communication with VPP via govpp core.
// It allows to specify custom buffer sizes for the request and reply Go channels.
func (c *Connection) NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (*api.Channel, error) {
	if c == nil {
		return nil, errors.New("nil connection passed in")
	}
	chID := atomic.AddUint32(&c.maxChannelID, 1)
	chMeta := &channelMetadata{id: chID}

	ch := api.NewChannelInternal(chMeta)
	ch.MsgDecoder = c.codec
	ch.MsgIdentifier = c

	// create the communication channels
	ch.ReqChan = make(chan *api.VppRequest, reqChanBufSize)
	ch.ReplyChan = make(chan *api.VppReply, replyChanBufSize)
	ch.NotifSubsChan = make(chan *api.NotifSubscribeRequest, reqChanBufSize)
	ch.NotifSubsReplyChan = make(chan error, replyChanBufSize)

	// store API channel within the client
	c.channelsLock.Lock()
	c.channels[chID] = ch
	c.channelsLock.Unlock()

	// start watching on the request channel
	go c.watchRequests(ch, chMeta)

	return ch, nil
}

// watchRequests watches for requests on the request API channel and forwards them as messages to VPP.
func (c *Connection) watchRequests(ch *api.Channel, chMeta *channelMetadata) {
	for {
		select {
		case req, ok := <-ch.ReqChan:
			// new request on the request channel
			if !ok {
				// after closing the request channel, release API channel and return
				c.releaseAPIChannel(ch, chMeta)
				return
			}
			c.processRequest(ch, chMeta, req)

		case req := <-ch.NotifSubsChan:
			// new request on the notification subscribe channel
			c.processNotifSubscribeRequest(ch, req)
		}
	}
}

// processRequest processes a single request received on the request channel.
func (c *Connection) processRequest(ch *api.Channel, chMeta *channelMetadata, req *api.VppRequest) error {
	// retrieve message ID
	msgID, err := c.GetMessageID(req.Message)
	if err != nil {
		error := fmt.Errorf("unable to retrieve message ID: %v", err)
		log.WithFields(logger.Fields{
			"msg_name": req.Message.GetMessageName(),
			"msg_crc":  req.Message.GetCrcString(),
		}).Errorf("unable to retrieve message ID: %v", err)
		sendReply(ch, &api.VppReply{Error: error})
		return error
	}

	// encode the message into binary
	data, err := c.codec.EncodeMsg(req.Message, msgID)
	if err != nil {
		error := fmt.Errorf("unable to encode the messge: %v", err)
		log.WithFields(logger.Fields{
			"context": chMeta.id,
			"msg_id":  msgID,
		}).Errorf("%v", error)
		sendReply(ch, &api.VppReply{Error: error})
		return error
	}

	// send the message
	log.WithFields(logger.Fields{
		"context":  chMeta.id,
		"msg_id":   msgID,
		"msg_size": len(data),
	}).Debug("Sending a message to VPP.")

	if req.Multipart {
		// expect multipart response
		atomic.StoreUint32(&chMeta.multipart, 1)
	}

	// send the request to VPP
	c.vpp.SendMsg(chMeta.id, data)

	if req.Multipart {
		// send a control ping to determine end of the multipart response
		ping := &vpe.ControlPing{}
		pingData, _ := c.codec.EncodeMsg(ping, c.pingReqID)

		log.WithFields(logger.Fields{
			"context":  chMeta.id,
			"msg_id":   c.pingReqID,
			"msg_size": len(pingData),
		}).Debug("Sending a control ping to VPP.")

		c.vpp.SendMsg(chMeta.id, pingData)
	}

	return nil
}

// releaseAPIChannel releases API channel that needs to be closed.
func (c *Connection) releaseAPIChannel(ch *api.Channel, chMeta *channelMetadata) {
	log.WithFields(logger.Fields{
		"context": chMeta.id,
	}).Debug("API channel closed.")

	// delete the channel from channels map
	c.channelsLock.Lock()
	delete(c.channels, chMeta.id)
	c.channelsLock.Unlock()
}

// msgCallback is called whenever any binary API message comes from VPP.
func msgCallback(context uint32, msgID uint16, data []byte) {
	connLock.RLock()
	defer connLock.RUnlock()

	if conn == nil {
		log.Warn("Already disconnected, ignoring the message.")
		return
	}

	log.WithFields(logger.Fields{
		"context":  context,
		"msg_id":   msgID,
		"msg_size": len(data),
	}).Debug("Received a message from VPP.")

	if context == 0 || conn.isNotificationMessage(msgID) {
		// process the message as a notification
		conn.sendNotifications(msgID, data)
		return
	}

	// match ch according to the context
	conn.channelsLock.RLock()
	ch, ok := conn.channels[context]
	conn.channelsLock.RUnlock()

	if !ok {
		log.WithFields(logger.Fields{
			"context": context,
			"msg_id":  msgID,
		}).Error("Context ID not known, ignoring the message.")
		return
	}

	chMeta := ch.Metadata().(*channelMetadata)
	lastReplyReceived := false
	// if this is a control ping reply and multipart request is being processed, treat this as a last part of the reply
	if msgID == conn.pingReplyID && atomic.CompareAndSwapUint32(&chMeta.multipart, 1, 0) {
		lastReplyReceived = true
	}

	// send the data to the channel
	sendReply(ch, &api.VppReply{
		MessageID:         msgID,
		Data:              data,
		LastReplyReceived: lastReplyReceived,
	})
}

// sendReply sends the reply into the go channel, if it cannot be completed without blocking, otherwise
// it logs the error and do not send the message.
func sendReply(ch *api.Channel, reply *api.VppReply) {
	select {
	case ch.ReplyChan <- reply:
		// reply sent successfully
	default:
		// unable to write into the channel without blocking
		log.WithFields(logger.Fields{
			"channel": ch,
			"msg_id":  reply.MessageID,
		}).Warn("Unable to send the reply, reciever end not ready.")
	}
}

// GetMessageID returns message identifier of given API message.
func (c *Connection) GetMessageID(msg api.Message) (uint16, error) {
	if c == nil {
		return 0, errors.New("nil connection passed in")
	}
	return c.messageNameToID(msg.GetMessageName(), msg.GetCrcString())
}

// messageNameToID returns message ID of a message identified by its name and CRC.
func (c *Connection) messageNameToID(msgName string, msgCrc string) (uint16, error) {
	// try to get the ID from the map
	c.msgIDsLock.RLock()
	id, ok := c.msgIDs[msgName+msgCrc]
	c.msgIDsLock.RUnlock()
	if ok {
		return id, nil
	}

	// get the ID using VPP API
	id, err := c.vpp.GetMsgID(msgName, msgCrc)
	if err != nil {
		error := fmt.Errorf("unable to retrieve message ID: %v", err)
		log.WithFields(logger.Fields{
			"msg_name": msgName,
			"msg_crc":  msgCrc,
		}).Errorf("unable to retrieve message ID: %v", err)
		return id, error
	}

	c.msgIDsLock.Lock()
	c.msgIDs[msgName+msgCrc] = id
	c.msgIDsLock.Unlock()

	return id, nil
}
