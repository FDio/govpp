//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package core

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/api"
)

type Stream struct {
	conn    *Connection
	ctx     context.Context
	channel *Channel
	// available options
	requestSize  int
	replySize    int
	replyTimeout time.Duration
	// per-request context
	pkgPath string
	sync.Mutex
}

func (c *Connection) NewStream(ctx context.Context, options ...api.StreamOption) (api.Stream, error) {
	if c == nil {
		return nil, errors.New("nil connection passed in")
	}
	s := &Stream{
		conn: c,
		ctx:  ctx,
		// default options
		requestSize:  RequestChanBufSize,
		replySize:    ReplyChanBufSize,
		replyTimeout: DefaultReplyTimeout,
	}

	// parse custom options
	for _, option := range options {
		option(s)
	}

	ch, err := c.newChannel(s.requestSize, s.replySize)
	if err != nil {
		return nil, err
	}
	s.channel = ch
	s.channel.SetReplyTimeout(s.replyTimeout)

	// Channel.watchRequests are not started here intentionally, because
	// requests are sent directly by SendMsg.

	return s, nil
}

func (c *Connection) Invoke(ctx context.Context, req api.Message, reply api.Message) error {
	stream, err := c.NewStream(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()
	if err := stream.SendMsg(req); err != nil {
		return err
	}
	s := stream.(*Stream)
	rep, err := s.recvReply()
	if err != nil {
		return err
	}
	if err := s.channel.msgCodec.DecodeMsg(rep.data, reply); err != nil {
		return err
	}
	return nil
}

type watcher struct {
	conn   *Connection
	ctx    context.Context
	sub    *subscriptionCtx
	events chan api.Message
	cancel context.CancelFunc
	quit   chan struct{}
}

func (w *watcher) Events() <-chan api.Message {
	return w.events
}

func (w *watcher) Close() {
	w.cancel()
}

func (w *watcher) watch() {
	log.WithField("event", w.sub.event.GetMessageName()).Debugf("starting event watcher")
	defer func() {
		if err := w.sub.Unsubscribe(); err != nil {
			log.Debugf("watcher unsubscribe error: %v", err)
		}
		close(w.events)
		log.WithField("event", w.sub.event.GetMessageName()).Debugf("event watcher done")
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		case e := <-w.sub.notifChan:
			// send to events
			select {
			case <-w.ctx.Done():
				return
			case w.events <- e:
				// received by events
			}
		}
	}
}

func (c *Connection) WatchEvent(ctx context.Context, event api.Message) (api.Watcher, error) {
	msgID, err := c.GetMessageID(event)
	if err != nil {
		log.WithFields(logrus.Fields{
			"msg_name": event.GetMessageName(),
			"msg_crc":  event.GetCrcString(),
		}).Debugf("unable to retrieve event message ID: %v", err)
		return nil, fmt.Errorf("unable to retrieve event message ID: %v", err)
	}

	cctx, cancel := context.WithCancel(ctx)

	w := &watcher{
		conn:   c,
		ctx:    cctx,
		cancel: cancel,
		events: make(chan api.Message),
		quit:   make(chan struct{}),
	}

	w.sub = &subscriptionCtx{
		conn:       c,
		notifChan:  make(chan api.Message, 10),
		msgID:      msgID,
		event:      event,
		msgFactory: getMsgFactory(event),
	}

	go w.watch()

	// add the subscription into map
	c.subscriptionsLock.Lock()
	defer c.subscriptionsLock.Unlock()

	c.subscriptions[msgID] = append(c.subscriptions[msgID], w.sub)

	return w, nil
}

func (s *Stream) Context() context.Context {
	return s.ctx
}

func (s *Stream) Close() error {
	if s.conn == nil {
		return errors.New("stream closed")
	}
	s.conn.releaseAPIChannel(s.channel)
	s.conn = nil
	return nil
}

func (s *Stream) SendMsg(msg api.Message) error {
	if s.conn == nil {
		return errors.New("stream closed")
	}
	req := s.channel.newRequest(msg, false)
	if err := s.conn.processRequest(s.channel, req); err != nil {
		return err
	}
	s.Lock()
	s.pkgPath = s.conn.GetMessagePath(msg)
	s.Unlock()
	return nil
}

func (s *Stream) RecvMsg() (api.Message, error) {
	reply, err := s.recvReply()
	if err != nil {
		return nil, err
	}
	// resolve message type
	s.Lock()
	path := s.pkgPath
	s.Unlock()
	msg, err := s.channel.msgIdentifier.LookupByID(path, reply.msgID)
	if err != nil {
		return nil, err
	}
	// allocate message instance
	msg = reflect.New(reflect.TypeOf(msg).Elem()).Interface().(api.Message)
	// decode message data
	if err := s.channel.msgCodec.DecodeMsg(reply.data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func WithRequestSize(size int) api.StreamOption {
	return func(stream api.Stream) {
		stream.(*Stream).requestSize = size
	}
}

func WithReplySize(size int) api.StreamOption {
	return func(stream api.Stream) {
		stream.(*Stream).replySize = size
	}
}

func WithReplyTimeout(timeout time.Duration) api.StreamOption {
	return func(stream api.Stream) {
		stream.(*Stream).replyTimeout = timeout
	}
}

func (s *Stream) recvReply() (*vppReply, error) {
	if s.conn == nil {
		return nil, errors.New("stream closed")
	}
	timeout := s.replyTimeout
	if timeout <= 0 {
		timeout = maxInt64
	}
	timeoutTimer := time.NewTimer(timeout)
	select {
	case reply, ok := <-s.channel.replyChan:
		if !ok {
			return nil, fmt.Errorf("reply channel closed")
		}
		if reply.err != nil {
			// this case should actually never happen for stream
			// since reply.err is only filled in watchRequests
			// and stream does not use it
			return nil, reply.err
		}
		return reply, nil
	case <-timeoutTimer.C:
		err := fmt.Errorf("no reply received within the timeout period %s", timeout)
		return nil, err
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	}
}
