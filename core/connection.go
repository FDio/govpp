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
	"path"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/sirupsen/logrus"
	"go.fd.io/govpp/core/genericpool"

	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/codec"
)

const (
	DefaultReconnectInterval    = time.Second / 2 // default interval between reconnect attempts
	DefaultMaxReconnectAttempts = 3               // default maximum number of reconnect attempts
)

var (
	RequestChanBufSize      = 100 // default size of the request channel buffer
	ReplyChanBufSize        = 100 // default size of the reply channel buffer
	NotificationChanBufSize = 100 // default size of the notification channel buffer
)

var (
	HealthCheckProbeInterval = time.Second            // default health check probe interval
	HealthCheckReplyTimeout  = time.Millisecond * 250 // timeout for reply to a health check probe
	HealthCheckThreshold     = 2                      // number of failed health checks until the error is reported
	DefaultReplyTimeout      = time.Second            // default timeout for replies from VPP
)

// ConnectionState represents the current state of the connection to VPP.
type ConnectionState int

const (
	// Connected represents state in which the connection has been successfully established.
	Connected ConnectionState = iota

	// NotResponding represents a state where the VPP socket accepts messages but replies are received with delay,
	// or not at all. GoVPP treats this state internally the same as disconnected.
	NotResponding

	// Disconnected represents state in which the VPP socket is closed and the connection is considered dropped.
	Disconnected

	// Failed represents state in which the reconnecting failed after exceeding maximum number of attempts.
	Failed
)

func (s ConnectionState) String() string {
	switch s {
	case Connected:
		return "Connected"
	case NotResponding:
		return "NotResponding"
	case Disconnected:
		return "Disconnected"
	case Failed:
		return "Failed"
	default:
		return fmt.Sprintf("UnknownState(%d)", s)
	}
}

// ConnectionEvent is a notification about change in the VPP connection state.
type ConnectionEvent struct {
	// Timestamp holds the time when the event has been created.
	Timestamp time.Time

	// State holds the new state of the connection at the time when the event has been created.
	State ConnectionState

	// Error holds error if any encountered.
	Error error
}

// Connection represents a shared memory connection to VPP via vppAdapter.
type Connection struct {
	vppClient adapter.VppAPI // VPP binary API client

	maxAttempts int           // interval for reconnect attempts
	recInterval time.Duration // maximum number of reconnect attempts

	vppConnected uint32 // non-zero if the adapter is connected to VPP

	connChan        chan ConnectionEvent // connection status events are sent to this channel
	healthCheckDone chan struct{}        // used to terminate health check loop

	codec        MessageCodec                      // message codec
	msgIDs       map[string]uint16                 // map of message IDs indexed by message name + CRC
	msgMapByPath map[string]map[uint16]api.Message // map of messages indexed by message ID which are indexed by path

	channelsLock sync.RWMutex        // lock for the channels map and the channel ID
	channels     map[uint16]*Channel // map of all API channels indexed by the channel ID
	channelPool  *genericpool.Pool[*Channel]

	subscriptionsLock sync.RWMutex                  // lock for the subscriptions map
	subscriptions     map[uint16][]*subscriptionCtx // map od all notification subscriptions indexed by message ID

	pingReqID   uint16 // ID if the ControlPing message
	pingReplyID uint16 // ID of the ControlPingReply message

	lastReplyLock sync.Mutex // lock for the last reply
	lastReply     time.Time  // time of the last received reply from VPP

	msgControlPing      api.Message
	msgControlPingReply api.Message

	apiTrace *trace // API tracer (disabled by default)
}

func newConnection(binapi adapter.VppAPI, attempts int, interval time.Duration) *Connection {
	if attempts == 0 {
		attempts = DefaultMaxReconnectAttempts
	}
	if interval == 0 {
		interval = DefaultReconnectInterval
	}

	c := &Connection{
		vppClient:           binapi,
		maxAttempts:         attempts,
		recInterval:         interval,
		connChan:            make(chan ConnectionEvent, NotificationChanBufSize),
		healthCheckDone:     make(chan struct{}),
		codec:               codec.DefaultCodec,
		msgIDs:              make(map[string]uint16),
		msgMapByPath:        make(map[string]map[uint16]api.Message),
		channels:            make(map[uint16]*Channel),
		subscriptions:       make(map[uint16][]*subscriptionCtx),
		msgControlPing:      msgControlPing,
		msgControlPingReply: msgControlPingReply,
		apiTrace: &trace{
			list: make([]*api.Record, 0),
			mux:  &sync.Mutex{},
		},
	}

	var nextChannelID uint32
	c.channelPool = genericpool.New[*Channel](func() *Channel {
		chID := atomic.AddUint32(&nextChannelID, 1)
		if chID > 0x7fff {
			return nil
		}
		// create new channel
		return &Channel{
			id:                  uint16(chID),
			conn:                c,
			msgCodec:            c.codec,
			msgIdentifier:       c,
			reqChan:             make(chan *vppRequest, RequestChanBufSize),
			replyChan:           make(chan *vppReply, ReplyChanBufSize),
			replyTimeout:        DefaultReplyTimeout,
			receiveReplyTimeout: ReplyChannelTimeout,
		}
	})

	binapi.SetMsgCallback(c.msgCallback)
	return c
}

// Connect connects to VPP API using specified adapter and returns a connection handle.
// This call blocks until it is either connected, or an error occurs.
// Only one connection attempt will be performed.
func Connect(binapi adapter.VppAPI) (*Connection, error) {
	// create new connection handle
	c := newConnection(binapi, DefaultMaxReconnectAttempts, DefaultReconnectInterval)

	// blocking attempt to connect to VPP
	if err := c.connectVPP(); err != nil {
		return nil, err
	}

	return c, nil
}

// AsyncConnect asynchronously connects to VPP using specified VPP adapter and returns the connection handle
// and ConnectionState channel. This call does not block until connection is established, it
// returns immediately. The caller is supposed to watch the returned ConnectionState channel for
// Connected/Disconnected events. In case of disconnect, the library will asynchronously try to reconnect.
func AsyncConnect(binapi adapter.VppAPI, attempts int, interval time.Duration) (*Connection, chan ConnectionEvent, error) {
	// create new connection handle
	c := newConnection(binapi, attempts, interval)

	// asynchronously attempt to connect to VPP
	go c.connectLoop()

	return c, c.connChan, nil
}

// connectVPP performs blocking attempt to connect to VPP.
func (c *Connection) connectVPP() error {
	log.Debug("Connecting to VPP..")

	// blocking connect
	if err := c.vppClient.Connect(); err != nil {
		return err
	}
	log.Debugf("Connected to VPP")

	if err := c.retrieveMessageIDs(); err != nil {
		if err := c.vppClient.Disconnect(); err != nil {
			log.Debugf("disconnecting vpp client failed: %v", err)
		}
		return fmt.Errorf("VPP is incompatible: %v", err)
	}

	// store connected state
	atomic.StoreUint32(&c.vppConnected, 1)

	return nil
}

// Disconnect disconnects from VPP API and releases all connection-related resources.
func (c *Connection) Disconnect() {
	if c == nil {
		return
	}
	if c.vppClient != nil {
		c.disconnectVPP(true)
	}
}

// disconnectVPP disconnects from VPP in case it is connected. terminate tells
// that disconnectVPP() was called from Close(), so healthCheckLoop() can be
// terminated.
func (c *Connection) disconnectVPP(terminate bool) {
	if atomic.CompareAndSwapUint32(&c.vppConnected, 1, 0) {
		if terminate {
			close(c.healthCheckDone)
		}
		log.Debug("Disconnecting from VPP..")

		if err := c.vppClient.Disconnect(); err != nil {
			log.Debugf("Disconnect from VPP failed: %v", err)
		}
		log.Debug("Disconnected from VPP")
	}
}

func (c *Connection) NewAPIChannel() (api.Channel, error) {
	return c.newAPIChannel(RequestChanBufSize, ReplyChanBufSize)
}

func (c *Connection) NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (api.Channel, error) {
	return c.newAPIChannel(reqChanBufSize, replyChanBufSize)
}

// NewAPIChannelBuffered returns a new API channel for communication with VPP via govpp core.
// It allows to specify custom buffer sizes for the request and reply Go channels.
func (c *Connection) newAPIChannel(reqChanBufSize, replyChanBufSize int) (*Channel, error) {
	if c == nil {
		return nil, errors.New("nil connection passed in")
	}

	channel, err := c.newChannel(reqChanBufSize, replyChanBufSize)
	if err != nil {
		return nil, err
	}

	// start watching on the request channel
	go c.watchRequests(channel)

	return channel, nil
}

// releaseAPIChannel releases API channel that needs to be closed.
func (c *Connection) releaseAPIChannel(ch *Channel) {
	log.WithFields(logger.Fields{
		"channel": ch.id,
	}).Debug("API channel released")

	c.channelPool.Put(ch)

	// delete the channel from channels map
	c.channelsLock.Lock()
	delete(c.channels, ch.id)
	c.channelsLock.Unlock()
}

// connectLoop attempts to connect to VPP until it succeeds.
// Then it continues with healthCheckLoop.
func (c *Connection) connectLoop() {
	var reconnectAttempts int

	// loop until connected
	for {
		if err := c.vppClient.WaitReady(); err != nil {
			log.Debugf("wait ready failed: %v", err)
		}
		if err := c.connectVPP(); err == nil {
			// signal connected event
			c.sendConnEvent(ConnectionEvent{Timestamp: time.Now(), State: Connected})
			break
		} else if reconnectAttempts < c.maxAttempts {
			reconnectAttempts++
			log.Warnf("connecting failed (attempt %d/%d): %v", reconnectAttempts, c.maxAttempts, err)
			time.Sleep(c.recInterval)
		} else {
			c.sendConnEvent(ConnectionEvent{Timestamp: time.Now(), State: Failed, Error: err})
			return
		}
	}

	// we are now connected, continue with health check loop
	c.healthCheckLoop()
}

// healthCheckLoop checks whether connection to VPP is alive. In case of disconnect,
// it continues with connectLoop and tries to reconnect.
func (c *Connection) healthCheckLoop() {
	// create a separate API channel for health check probes
	ch, err := c.newAPIChannel(1, 1)
	if err != nil {
		log.Error("Failed to create health check API channel, health check will be disabled:", err)
		return
	}
	defer ch.Close()

	var (
		sinceLastReply time.Duration
		failedChecks   int
	)

	// send health check probes until an error or timeout occurs
	probeInterval := time.NewTicker(HealthCheckProbeInterval)
	defer probeInterval.Stop()

HealthCheck:
	for {
		select {
		case <-c.healthCheckDone:
			// Terminate the health check loop on connection disconnect
			log.Debug("Disconnected on request, exiting health check loop.")
			return
		case <-probeInterval.C:
			// try draining probe replies from previous request before sending next one
			select {
			case <-ch.replyChan:
				log.Debug("drained old probe reply from reply channel")
			default:
			}

			// send the control ping request
			ch.reqChan <- &vppRequest{msg: c.msgControlPing}

			for {
				// expect response within timeout period
				select {
				case vppReply := <-ch.replyChan:
					err = vppReply.err

				case <-time.After(HealthCheckReplyTimeout):
					err = ErrProbeTimeout

					// check if time since last reply from any other
					// channel is less than health check reply timeout
					c.lastReplyLock.Lock()
					sinceLastReply = time.Since(c.lastReply)
					c.lastReplyLock.Unlock()

					if sinceLastReply < HealthCheckReplyTimeout {
						log.Warnf("VPP health check probe timing out, but some request on other channel was received %v ago, continue waiting!", sinceLastReply)
						continue
					}
				}
				break
			}

			if err == ErrProbeTimeout {
				failedChecks++
				log.Warnf("VPP health check probe timed out after %v (%d. timeout)", HealthCheckReplyTimeout, failedChecks)
				if failedChecks > HealthCheckThreshold {
					// in case of exceeded failed check threshold, assume VPP unresponsive
					log.Errorf("VPP does not responding, the health check exceeded threshold for timeouts (>%d)", HealthCheckThreshold)
					c.sendConnEvent(ConnectionEvent{Timestamp: time.Now(), State: NotResponding})
					break HealthCheck
				}
			} else if err != nil {
				// in case of error, assume VPP disconnected
				log.Errorf("VPP health check probe failed: %v", err)
				c.sendConnEvent(ConnectionEvent{Timestamp: time.Now(), State: Disconnected, Error: err})
				break HealthCheck
			} else if failedChecks > 0 {
				// in case of success after failed checks, clear failed check counter
				failedChecks = 0
				log.Infof("VPP health check probe OK")
			}
		}
	}

	// cleanup
	c.disconnectVPP(false)

	// we are now disconnected, start connect loop
	c.connectLoop()
}

func getMsgNameWithCrc(x api.Message) string {
	return getMsgID(x.GetMessageName(), x.GetCrcString())
}

func getMsgID(name, crc string) string {
	return name + "_" + crc
}

func getMsgFactory(msg api.Message) func() api.Message {
	return func() api.Message {
		return reflect.New(reflect.TypeOf(msg).Elem()).Interface().(api.Message)
	}
}

// GetMessageID returns message identifier of given API message.
func (c *Connection) GetMessageID(msg api.Message) (uint16, error) {
	if c == nil {
		return 0, errors.New("nil connection passed in")
	}
	pkgPath := c.GetMessagePath(msg)
	msgID, err := c.vppClient.GetMsgID(msg.GetMessageName(), msg.GetCrcString())
	if err != nil {
		return 0, err
	}
	if pathMsgs, pathOk := c.msgMapByPath[pkgPath]; !pathOk {
		c.msgMapByPath[pkgPath] = make(map[uint16]api.Message)
		c.msgMapByPath[pkgPath][msgID] = msg
	} else if _, msgOk := pathMsgs[msgID]; !msgOk {
		c.msgMapByPath[pkgPath][msgID] = msg
	}
	if _, ok := c.msgIDs[getMsgNameWithCrc(msg)]; ok {
		return msgID, nil
	}
	c.msgIDs[getMsgNameWithCrc(msg)] = msgID
	return msgID, nil
}

// LookupByID looks up message name and crc by ID.
func (c *Connection) LookupByID(path string, msgID uint16) (api.Message, error) {
	if c == nil {
		return nil, errors.New("nil connection passed in")
	}
	if msg, ok := c.msgMapByPath[path][msgID]; ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unknown message ID %d for path '%s'", msgID, path)
}

// GetMessagePath returns path for the given message
func (c *Connection) GetMessagePath(msg api.Message) string {
	return path.Dir(reflect.TypeOf(msg).Elem().PkgPath())
}

// retrieveMessageIDs retrieves IDs for all registered messages and stores them in map
func (c *Connection) retrieveMessageIDs() (err error) {
	t := time.Now()

	msgsByPath := api.GetRegisteredMessages()

	var n int
	for pkgPath, msgs := range msgsByPath {
		for _, msg := range msgs {
			msgID, err := c.GetMessageID(msg)
			if err != nil {
				if debugMsgIDs {
					log.Debugf("retrieving message ID for %s.%s failed: %v",
						pkgPath, msg.GetMessageName(), err)
				}
				continue
			}
			n++

			if c.pingReqID == 0 && msg.GetMessageName() == c.msgControlPing.GetMessageName() {
				c.pingReqID = msgID
				c.msgControlPing = reflect.New(reflect.TypeOf(msg).Elem()).Interface().(api.Message)
			} else if c.pingReplyID == 0 && msg.GetMessageName() == c.msgControlPingReply.GetMessageName() {
				c.pingReplyID = msgID
				c.msgControlPingReply = reflect.New(reflect.TypeOf(msg).Elem()).Interface().(api.Message)
			}

			if debugMsgIDs {
				log.Debugf("message %q (%s) has ID: %d", msg.GetMessageName(), getMsgNameWithCrc(msg), msgID)
			}
		}
		log.WithField("took", time.Since(t)).
			Debugf("retrieved IDs for %d messages (registered %d) from path %s", n, len(msgs), pkgPath)
	}

	return nil
}

func (c *Connection) sendConnEvent(event ConnectionEvent) {
	select {
	case c.connChan <- event:
	default:
		log.Warn("Connection state channel is full, discarding value.")
	}
}

// Trace gives access to the API trace interface
func (c *Connection) Trace() api.Trace {
	return c.apiTrace
}

// trace records api message
func (c *Connection) trace(msg api.Message, chId uint16, t time.Time, isReceived bool) {
	if atomic.LoadInt32(&c.apiTrace.isEnabled) == 0 {
		return
	}
	entry := &api.Record{
		Message:    msg,
		Timestamp:  t,
		IsReceived: isReceived,
		ChannelID:  chId,
	}
	c.apiTrace.mux.Lock()
	c.apiTrace.list = append(c.apiTrace.list, entry)
	c.apiTrace.mux.Unlock()
}
