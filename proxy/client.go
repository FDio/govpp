package proxy

import (
	"fmt"
	"log"
	"net/rpc"
	"reflect"
	"time"

	"git.fd.io/govpp.git/api"
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
		log.Fatal("Connection error: ", err)
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
		rpc: c.rpc,
	}
	return binapi, nil
}

type StatsClient struct {
	rpc *rpc.Client
}

func (s *StatsClient) GetSystemStats(sysStats *api.SystemStats) error {
	req := StatsRequest{StatsType: "system"}
	resp := StatsResponse{SysStats: sysStats}
	return s.rpc.Call("StatsRPC.GetStats", req, &resp)
}

func (s *StatsClient) GetNodeStats(nodeStats *api.NodeStats) error {
	req := StatsRequest{StatsType: "node"}
	resp := StatsResponse{NodeStats: nodeStats}
	return s.rpc.Call("StatsRPC.GetStats", req, &resp)
}

func (s *StatsClient) GetInterfaceStats(ifaceStats *api.InterfaceStats) error {
	req := StatsRequest{StatsType: "interface"}
	resp := StatsResponse{IfaceStats: ifaceStats}
	return s.rpc.Call("StatsRPC.GetStats", req, &resp)
}

func (s *StatsClient) GetErrorStats(errStats *api.ErrorStats) error {
	req := StatsRequest{StatsType: "error"}
	resp := StatsResponse{ErrStats: errStats}
	return s.rpc.Call("StatsRPC.GetStats", req, &resp)
}

func (s *StatsClient) GetBufferStats(bufStats *api.BufferStats) error {
	req := StatsRequest{StatsType: "buffer"}
	resp := StatsResponse{BufStats: bufStats}
	return s.rpc.Call("StatsRPC.GetStats", req, &resp)
}

type BinapiClient struct {
	rpc *rpc.Client
}

func (b *BinapiClient) SendRequest(msg api.Message) api.RequestCtx {
	req := &requestCtx{
		rpc: b.rpc,
		req: msg,
	}
	log.Printf("SendRequest: %T %+v", msg, msg)
	return req
}

type requestCtx struct {
	rpc *rpc.Client
	req api.Message
}

func (r *requestCtx) ReceiveReply(msg api.Message) error {
	req := BinapiRequest{
		Msg:      r.req,
		ReplyMsg: msg,
	}
	resp := BinapiResponse{}

	err := r.rpc.Call("BinapiRPC.Invoke", req, &resp)
	if err != nil {
		return fmt.Errorf("RPC call failed: %v", err)
	}

	// we set the value of msg to the value from response
	reflect.ValueOf(msg).Elem().Set(reflect.ValueOf(resp.Msg).Elem())

	return nil
}

func (b *BinapiClient) SendMultiRequest(msg api.Message) api.MultiRequestCtx {
	req := &multiRequestCtx{
		rpc: b.rpc,
		req: msg,
	}
	log.Printf("SendMultiRequest: %T %+v", msg, msg)
	return req
}

type multiRequestCtx struct {
	rpc *rpc.Client
	req api.Message

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
	panic("implement me")
}

func (b *BinapiClient) CheckCompatiblity(msgs ...api.Message) error {
	return nil // TODO: proxy this
}

func (b *BinapiClient) Close() {
	b.rpc.Close()
}
