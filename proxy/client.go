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

package proxy

import (
	"context"
	"fmt"
	"net/rpc"
	"reflect"
	"time"

	"go.fd.io/govpp/api"
	"go.fd.io/govpp/core"
)

type Client struct {
	serverAddr string
	rpc        *rpc.Client
}

// Connect dials remote proxy server on given address and
// returns new client if successful.
func Connect(addr string) (*Client, error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("connection error:%v", err)
	}
	c := &Client{
		serverAddr: addr,
		rpc:        client,
	}
	return c, nil
}

// NewStatsClient returns new StatsClient which implements api.StatsProvider.
func (c *Client) NewStatsClient() (*StatsClient, error) {
	stats := &StatsClient{
		rpc: c.rpc,
	}
	return stats, nil
}

// NewBinapiClient returns new BinapiClient which implements api.Channel.
func (c *Client) NewBinapiClient() (*BinapiClient, error) {
	binapi := &BinapiClient{
		rpc:     c.rpc,
		timeout: core.DefaultReplyTimeout,
	}
	return binapi, nil
}

type StatsClient struct {
	rpc *rpc.Client
}

func (s *StatsClient) GetSystemStats(sysStats *api.SystemStats) error {
	// we need to start with a clean, zeroed item before decoding
	// 'cause if the new values are 'zero' for the type, they will be ignored
	// by the decoder. (i.e the old values will be left unchanged).
	req := StatsRequest{StatsType: "system"}
	resp := StatsResponse{SysStats: new(api.SystemStats)}
	if err := s.rpc.Call("StatsRPC.GetStats", req, &resp); err != nil {
		return err
	}
	*sysStats = *resp.SysStats
	return nil
}

func (s *StatsClient) GetNodeStats(nodeStats *api.NodeStats) error {
	req := StatsRequest{StatsType: "node"}
	resp := StatsResponse{NodeStats: new(api.NodeStats)}
	if err := s.rpc.Call("StatsRPC.GetStats", req, &resp); err != nil {
		return err
	}
	*nodeStats = *resp.NodeStats
	return nil
}

func (s *StatsClient) GetInterfaceStats(ifaceStats *api.InterfaceStats) error {
	req := StatsRequest{StatsType: "interface"}
	resp := StatsResponse{IfaceStats: new(api.InterfaceStats)}
	if err := s.rpc.Call("StatsRPC.GetStats", req, &resp); err != nil {
		return err
	}
	*ifaceStats = *resp.IfaceStats
	return nil
}

func (s *StatsClient) GetErrorStats(errStats *api.ErrorStats) error {
	req := StatsRequest{StatsType: "error"}
	resp := StatsResponse{ErrStats: new(api.ErrorStats)}
	if err := s.rpc.Call("StatsRPC.GetStats", req, &resp); err != nil {
		return err
	}
	*errStats = *resp.ErrStats
	return nil
}

func (s *StatsClient) GetBufferStats(bufStats *api.BufferStats) error {
	req := StatsRequest{StatsType: "buffer"}
	resp := StatsResponse{BufStats: new(api.BufferStats)}
	if err := s.rpc.Call("StatsRPC.GetStats", req, &resp); err != nil {
		return err
	}
	*bufStats = *resp.BufStats
	return nil
}

func (s *StatsClient) GetMemoryStats(memStats *api.MemoryStats) error {
	req := StatsRequest{StatsType: "memory"}
	resp := StatsResponse{MemStats: new(api.MemoryStats)}
	if err := s.rpc.Call("StatsRPC.GetStats", req, &resp); err != nil {
		return err
	}
	*memStats = *resp.MemStats
	return nil
}

type BinapiClient struct {
	rpc     *rpc.Client
	timeout time.Duration
}

// RPCStream is a stream for forwarding requests to BinapiRPC's stream.
type RPCStream struct {
	ctx context.Context
	rpc *rpc.Client
	id  uint32
}

func (s *RPCStream) Context() context.Context {
	return s.ctx
}

func (s *RPCStream) SendMsg(msg api.Message) error {
	req := RPCStreamReqResp{
		ID:  s.id,
		Msg: msg,
	}
	resp := RPCStreamReqResp{}
	err := s.rpc.Call("BinapiRPC.SendMessage", req, &resp)
	if err != nil {
		return fmt.Errorf("RPC SendMessage call failed: %v", err)
	}
	return nil
}

func (s *RPCStream) RecvMsg() (api.Message, error) {
	req := RPCStreamReqResp{
		ID: s.id,
	}
	resp := RPCStreamReqResp{}
	err := s.rpc.Call("BinapiRPC.ReceiveMessage", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("RPC ReceiveMessage call failed: %v", err)
	}
	return resp.Msg, nil
}

func (s *RPCStream) Close() error {
	req := RPCStreamReqResp{
		ID: s.id,
	}
	resp := RPCStreamReqResp{}
	err := s.rpc.Call("BinapiRPC.CloseStream", req, &resp)
	if err != nil {
		return fmt.Errorf("RPC CloseStream call failed: %v", err)
	}
	return nil
}

func (b *BinapiClient) NewStream(_ context.Context, _ ...api.StreamOption) (api.Stream, error) {
	stream := &RPCStream{
		rpc: b.rpc,
	}
	req := RPCStreamReqResp{}
	resp := RPCStreamReqResp{}
	err := stream.rpc.Call("BinapiRPC.NewAPIStream", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("RPC NewAPIStream call failed: %v", err)
	}
	stream.id = resp.ID
	return stream, err
}

func (b *BinapiClient) Invoke(_ context.Context, request api.Message, reply api.Message) error {
	return invokeInternal(b.rpc, request, reply, b.timeout)
}

func (b *BinapiClient) SendRequest(msg api.Message) api.RequestCtx {
	req := &requestCtx{
		rpc:     b.rpc,
		timeout: b.timeout,
		req:     msg,
	}
	log.Debugf("SendRequest: %T %+v", msg, msg)
	return req
}

type requestCtx struct {
	rpc     *rpc.Client
	req     api.Message
	timeout time.Duration
}

func (r *requestCtx) ReceiveReply(msg api.Message) error {
	return invokeInternal(r.rpc, r.req, msg, r.timeout)
}

func invokeInternal(rpc *rpc.Client, msgIn, msgOut api.Message, timeout time.Duration) error {
	req := BinapiRequest{
		Msg:      msgIn,
		ReplyMsg: msgOut,
		Timeout:  timeout,
	}
	resp := BinapiResponse{}

	err := rpc.Call("BinapiRPC.Invoke", req, &resp)
	if err != nil {
		return fmt.Errorf("RPC call failed: %v", err)
	}

	// we set the value of msg to the value from response
	reflect.ValueOf(msgOut).Elem().Set(reflect.ValueOf(resp.Msg).Elem())

	return nil
}

func (b *BinapiClient) SendMultiRequest(msg api.Message) api.MultiRequestCtx {
	req := &multiRequestCtx{
		rpc:     b.rpc,
		timeout: b.timeout,
		req:     msg,
	}
	log.Debugf("SendMultiRequest: %T %+v", msg, msg)
	return req
}

type multiRequestCtx struct {
	rpc     *rpc.Client
	req     api.Message
	timeout time.Duration

	index   int
	replies []api.Message
}

func (r *multiRequestCtx) ReceiveReply(msg api.Message) (stop bool, err error) {
	// we call Invoke only on first ReceiveReply
	if r.index == 0 {
		req := BinapiRequest{
			Msg:      r.req,
			ReplyMsg: msg,
			IsMulti:  true,
			Timeout:  r.timeout,
		}
		resp := BinapiResponse{}

		err := r.rpc.Call("BinapiRPC.Invoke", req, &resp)
		if err != nil {
			return false, fmt.Errorf("RPC call failed: %v", err)
		}

		r.replies = resp.Msgs
	}

	if r.index >= len(r.replies) {
		return true, nil
	}

	// we set the value of msg to the value from response
	reflect.ValueOf(msg).Elem().Set(reflect.ValueOf(r.replies[r.index]).Elem())
	r.index++

	return false, nil
}

func (b *BinapiClient) SubscribeNotification(notifChan chan api.Message, event api.Message) (api.SubscriptionCtx, error) {
	panic("implement me")
}

func (b *BinapiClient) SetReplyTimeout(timeout time.Duration) {
	b.timeout = timeout
}

func (b *BinapiClient) CheckCompatiblity(msgs ...api.Message) error {
	msgNamesCrscs := make([]string, 0, len(msgs))

	for _, msg := range msgs {
		msgNamesCrscs = append(msgNamesCrscs, msg.GetMessageName()+"_"+msg.GetCrcString())
	}

	req := BinapiCompatibilityRequest{MsgNameCrcs: msgNamesCrscs}
	resp := BinapiCompatibilityResponse{}

	if err := b.rpc.Call("BinapiRPC.Compatibility", req, &resp); err != nil {
		return err
	}

	return nil
}

func (b *BinapiClient) Close() {
	b.rpc.Close()
}
