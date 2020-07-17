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

	"git.fd.io/govpp.git/api"
)

type Stream struct {
	id      uint32
	conn    *Connection
	ctx     context.Context
	channel *Channel
}

func (c *Connection) NewStream(ctx context.Context) (api.Stream, error) {
	if c == nil {
		return nil, errors.New("nil connection passed in")
	}
	// TODO: add stream options as variadic parameters for customizing:
	// - request/reply channel size
	// - reply timeout
	// - retries
	// - ???

	// create new channel
	chID := uint16(atomic.AddUint32(&c.maxChannelID, 1) & 0x7fff)
	channel := newChannel(chID, c, c.codec, c, 10, 10)

	// store API channel within the client
	c.channelsLock.Lock()
	c.channels[chID] = channel
	c.channelsLock.Unlock()

	// Channel.watchRequests are not started here intentionally, because
	// requests are sent directly by SendMsg.

	return &Stream{
		id:      uint32(chID),
		conn:    c,
		ctx:     ctx,
		channel: channel,
	}, nil
}

func (c *Connection) Invoke(ctx context.Context, req api.Message, reply api.Message) error {
	stream, err := c.NewStream(ctx)
	if err != nil {
		return err
	}
	if err := stream.SendMsg(req); err != nil {
		return err
	}
	msg, err := stream.RecvMsg()
	if err != nil {
		return err
	}
	if msg.GetMessageName() != reply.GetMessageName() ||
		msg.GetCrcString() != reply.GetCrcString() {
		return fmt.Errorf("unexpected reply: %T %+v", msg, msg)
	}
	reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(msg).Elem())
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

	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	}
}
