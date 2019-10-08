package proxy

import (
	"fmt"
	"log"
	"reflect"

	"git.fd.io/govpp.git/api"
)

type StatsRequest struct {
	StatsType string
}

type StatsResponse struct {
	SysStats   *api.SystemStats
	NodeStats  *api.NodeStats
	IfaceStats *api.InterfaceStats
	ErrStats   *api.ErrorStats
	BufStats   *api.BufferStats
}

// StatsRPC is a RPC server for proxying client request to api.StatsProvider.
type StatsRPC struct {
	stats api.StatsProvider
}

// NewStatsRPC returns new StatsRPC to be used as RPC server
// proxying request to given api.StatsProvider.
func NewStatsRPC(stats api.StatsProvider) *StatsRPC {
	return &StatsRPC{stats: stats}
}

func (s *StatsRPC) GetStats(req StatsRequest, resp *StatsResponse) error {
	log.Printf("StatsRPC.GetStats - REQ: %+v", req)

	switch req.StatsType {
	case "system":
		resp.SysStats = new(api.SystemStats)
		return s.stats.GetSystemStats(resp.SysStats)
	case "node":
		resp.NodeStats = new(api.NodeStats)
		return s.stats.GetNodeStats(resp.NodeStats)
	case "interface":
		resp.IfaceStats = new(api.InterfaceStats)
		return s.stats.GetInterfaceStats(resp.IfaceStats)
	case "error":
		resp.ErrStats = new(api.ErrorStats)
		return s.stats.GetErrorStats(resp.ErrStats)
	case "buffer":
		resp.BufStats = new(api.BufferStats)
		return s.stats.GetBufferStats(resp.BufStats)
	default:
		return fmt.Errorf("unknown stats type: %s", req.StatsType)
	}
}

type BinapiRequest struct {
	Msg      api.Message
	IsMulti  bool
	ReplyMsg api.Message
}

type BinapiResponse struct {
	Msg  api.Message
	Msgs []api.Message
}

// BinapiRPC is a RPC server for proxying client request to api.Channel.
type BinapiRPC struct {
	binapi api.Channel
}

// NewBinapiRPC returns new BinapiRPC to be used as RPC server
// proxying request to given api.Channel.
func NewBinapiRPC(binapi api.Channel) *BinapiRPC {
	return &BinapiRPC{binapi: binapi}
}

func (s *BinapiRPC) Invoke(req BinapiRequest, resp *BinapiResponse) error {
	log.Printf("BinapiRPC.Invoke - REQ: %#v", req)

	if req.IsMulti {
		multi := s.binapi.SendMultiRequest(req.Msg)
		for {
			// create new message in response of type ReplyMsg
			msg := reflect.New(reflect.TypeOf(req.ReplyMsg).Elem()).Interface().(api.Message)

			stop, err := multi.ReceiveReply(msg)
			if err != nil {
				return err
			} else if stop {
				break
			}

			resp.Msgs = append(resp.Msgs, msg)
		}
	} else {
		// create new message in response of type ReplyMsg
		resp.Msg = reflect.New(reflect.TypeOf(req.ReplyMsg).Elem()).Interface().(api.Message)

		err := s.binapi.SendRequest(req.Msg).ReceiveReply(resp.Msg)
		if err != nil {
			return err
		}
	}

	return nil
}
