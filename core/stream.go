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
	"sync/atomic"
	"time"

	"git.fd.io/govpp.git/api"
)

type Stream struct {
	id      uint32
	conn    *Connection
	ctx     context.Context
	channel *Channel
	// available options
	requestSize  int
	replySize    int
	replyTimeout time.Duration
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
	// create and store a new channel
	s.id = atomic.AddUint32(&c.maxChannelID, 1) & 0x7fff
	s.channel = newChannel(uint16(s.id), c, c.codec, c, s.requestSize, s.replySize)
	s.channel.SetReplyTimeout(s.replyTimeout)

	// store API channel within the client
	c.channelsLock.Lock()
	c.channels[uint16(s.id)] = s.channel
	c.channelsLock.Unlock()

	// Channel.watchRequests are not started here intentionally, because
	// requests are sent directly by SendMsg.

	return s, nil
}

func (c *Connection) Invoke(ctx context.Context, req api.Message, reply api.Message) error {
	stream, err := c.NewStream(ctx)
	if err != nil {
		return err
	}
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
	return nil
}

func (s *Stream) RecvMsg() (api.Message, error) {
	reply, err := s.recvReply()
	if err != nil {
		return nil, err
	}
	// resolve message type
	msg, err := s.channel.msgIdentifier.LookupByID(reply.msgID)
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

	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	}
}
