//  Copyright (c) 2021 Cisco and/or its affiliates.
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

package api

import (
	"context"
	"time"
)

// Connection represents the client connection to VPP API.
type Connection interface {
	// NewStream creates a new stream for sending and receiving messages.
	// Context can be used to close the stream using cancel or timeout.
	NewStream(ctx context.Context, options ...StreamOption) (Stream, error)

	// Invoke can be used for a simple request-reply RPC.
	// It creates stream and calls SendMsg with req and RecvMsg which returns
	// reply.
	Invoke(ctx context.Context, req Message, reply Message) error

	// WatchEvent creates a new watcher for watching events of type specified by
	// event parameter. Context can be used to close the watcher.
	WatchEvent(ctx context.Context, event Message) (Watcher, error)
}

// Watcher provides access to watched event messages. It can be created by calling Connection.WatchEvent.
type Watcher interface {
	// Events returns a channel where events are sent. The channel is closed when
	// watcher context is canceled or when Close is called.
	Events() <-chan Message

	// Close closes the watcher along with the events channel.
	Close()
}

// Stream provides low-level access for sending and receiving messages.
// Users should handle correct type and ordering of messages.
//
// It is not safe to call these methods on the same stream in different goroutines.
type Stream interface {
	// Context returns the context for this stream.
	Context() context.Context

	// SendMsg sends a message to the client.
	// It blocks until message is sent to the transport.
	SendMsg(Message) error

	// RecvMsg blocks until a message is received or error occurs.
	RecvMsg() (Message, error)

	// Close closes the stream. Calling SendMsg and RecvMsg will return error
	// after closing stream.
	Close() error
}

// StreamOption allows customizing a Stream.
type StreamOption func(Stream)

// ChannelProvider provides the communication channel with govpp core.
//
// DEPRECATED: Use Connection instead.
type ChannelProvider interface {
	// NewAPIChannel returns a new channel for communication with VPP via govpp core.
	// It uses default buffer sizes for the request and reply Go channels.
	NewAPIChannel() (Channel, error)

	// NewAPIChannelBuffered returns a new channel for communication with VPP via govpp core.
	// It allows to specify custom buffer sizes for the request and reply Go channels.
	NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (Channel, error)
}

// Channel provides methods for direct communication with VPP channel.
//
// DEPRECATED: Use Stream instead.
type Channel interface {
	// SendRequest asynchronously sends a request to VPP. Returns a request context, that can be used to call ReceiveReply.
	// In case of any errors by sending, the error will be delivered to ReplyChan (and returned by ReceiveReply).
	SendRequest(msg Message) RequestCtx

	// SendMultiRequest asynchronously sends a multipart request (request to which multiple responses are expected) to VPP.
	// Returns a multipart request context, that can be used to call ReceiveReply.
	// In case of any errors by sending, the error will be delivered to ReplyChan (and returned by ReceiveReply).
	SendMultiRequest(msg Message) MultiRequestCtx

	// SubscribeNotification subscribes for receiving of the specified notification messages via provided Go channel.
	// Note that the caller is responsible for creating the Go channel with preferred buffer size. If the channel's
	// buffer is full, the notifications will not be delivered into it.
	SubscribeNotification(notifChan chan Message, event Message) (SubscriptionCtx, error)

	// SetReplyTimeout sets the timeout for replies from VPP. It represents the maximum time the client waits for a reply
	// from VPP before returning a timeout error. Setting the reply timeout to 0 disables it. The initial reply timeout is
	//set to the value of core.DefaultReplyTimeout.
	SetReplyTimeout(timeout time.Duration)

	// CheckCompatibility checks the compatiblity for the given messages.
	// It will return an error if any of the given messages are not compatible.
	CheckCompatiblity(msgs ...Message) error

	// Close closes the API channel and releases all API channel-related resources
	// in the ChannelProvider.
	Close()
}

// RequestCtx is helper interface which allows to receive reply on request.
//
// DEPRECATED: Use Stream instead.
type RequestCtx interface {
	// ReceiveReply receives a reply from VPP (blocks until a reply is delivered
	// from VPP, or until an error occurs). The reply will be decoded into the msg
	// argument. Error will be returned if the response cannot be received or decoded.
	ReceiveReply(msg Message) error
}

// MultiRequestCtx is helper interface which allows to receive reply on multi-request.
//
// DEPRECATED: Use Stream instead.
type MultiRequestCtx interface {
	// ReceiveReply receives a reply from VPP (blocks until a reply is delivered
	// from VPP, or until an error occurs).The reply will be decoded into the msg
	// argument. If the last reply has been already consumed, lastReplyReceived is
	// set to true. Do not use the message itself if lastReplyReceived is
	// true - it won't be filled with actual data.Error will be returned if the
	// response cannot be received or decoded.
	ReceiveReply(msg Message) (lastReplyReceived bool, err error)
}

// SubscriptionCtx is helper interface which allows to control subscription for
// notification events.
//
// DEPRECATED: Use Connection instead.
type SubscriptionCtx interface {
	// Unsubscribe unsubscribes from receiving the notifications tied to the
	// subscription context.
	Unsubscribe() error
}
