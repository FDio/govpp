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
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/core"
)

const (
	binapiErrorMsg = `
------------------------------------------------------------
 received binapi request while VPP connection is down!
  - is VPP running ?
  - have you called Connect on the binapi RPC ?
------------------------------------------------------------
`
	statsErrorMsg = `
------------------------------------------------------------
 received stats request while stats connection is down!
  - is VPP running ?
  - is the correct socket name configured ?
  - have you called Connect on the stats RPC ?
------------------------------------------------------------
`
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
	MemStats   *api.MemoryStats
}

// StatsRPC is a RPC server for proxying client request to api.StatsProvider.
type StatsRPC struct {
	statsConn *core.StatsConnection
	stats     adapter.StatsAPI

	done chan struct{}
	// non-zero if the RPC service is available
	available uint32
	// non-zero if connected to stats file.
	isConnected uint32
	// synchronizes access to statsConn.
	mu sync.Mutex
}

// NewStatsRPC returns new StatsRPC to be used as RPC server
// proxying request to given api.StatsProvider.
func NewStatsRPC(stats adapter.StatsAPI) (*StatsRPC, error) {
	rpc := new(StatsRPC)
	if err := rpc.connect(stats); err != nil {
		return nil, err
	}
	return rpc, nil
}

func (s *StatsRPC) watchConnection() {
	heartbeatTicker := time.NewTicker(10 * time.Second).C
	atomic.StoreUint32(&s.available, 1)
	log.Debugln("enabling statsRPC service")

	count := 0
	prev := new(api.SystemStats)

	s.mu.Lock()
	if err := s.statsConn.GetSystemStats(prev); err != nil {
		atomic.StoreUint32(&s.available, 0)
		log.Warnf("disabling statsRPC service, reason: %v", err)
	}
	s.mu.Unlock()

	for {
		select {
		case <-heartbeatTicker:
			// If disconnect was called exit.
			if atomic.LoadUint32(&s.isConnected) == 0 {
				atomic.StoreUint32(&s.available, 0)
				return
			}

			curr := new(api.SystemStats)

			s.mu.Lock()
			if err := s.statsConn.GetSystemStats(curr); err != nil {
				atomic.StoreUint32(&s.available, 0)
				log.Warnf("disabling statsRPC service, reason: %v", err)
			}
			s.mu.Unlock()

			if curr.Heartbeat <= prev.Heartbeat {
				count++
				// vpp might have crashed/reset... try reconnecting
				if count == 5 {
					count = 0
					atomic.StoreUint32(&s.available, 0)
					log.Warnln("disabling statsRPC service, reason: vpp might have crashed/reset...")
					s.statsConn.Disconnect()
					for {
						var err error
						s.statsConn, err = core.ConnectStats(s.stats)
						if err == nil {
							atomic.StoreUint32(&s.available, 1)
							log.Debugln("enabling statsRPC service")
							break
						}
						time.Sleep(5 * time.Second)
					}
				}
			} else {
				count = 0
			}

			prev = curr
		case <-s.done:
			return
		}
	}
}

func (s *StatsRPC) connect(stats adapter.StatsAPI) error {
	if atomic.LoadUint32(&s.isConnected) == 1 {
		return errors.New("connection already exists")
	}
	s.stats = stats
	var err error
	s.statsConn, err = core.ConnectStats(s.stats)
	if err != nil {
		return err
	}
	s.done = make(chan struct{})
	atomic.StoreUint32(&s.isConnected, 1)

	go s.watchConnection()
	return nil
}

func (s *StatsRPC) disconnect() {
	if atomic.LoadUint32(&s.isConnected) == 1 {
		atomic.StoreUint32(&s.isConnected, 0)
		close(s.done)
		s.statsConn.Disconnect()
		s.statsConn = nil
	}
}

func (s *StatsRPC) serviceAvailable() bool {
	return atomic.LoadUint32(&s.available) == 1
}

func (s *StatsRPC) GetStats(req StatsRequest, resp *StatsResponse) error {
	if !s.serviceAvailable() {
		log.Print(statsErrorMsg)
		return errors.New("server does not support 'get stats' at this time, try again later")
	}
	log.Debugf("StatsRPC.GetStats - REQ: %+v", req)

	s.mu.Lock()
	defer s.mu.Unlock()

	switch req.StatsType {
	case "system":
		resp.SysStats = new(api.SystemStats)
		return s.statsConn.GetSystemStats(resp.SysStats)
	case "node":
		resp.NodeStats = new(api.NodeStats)
		return s.statsConn.GetNodeStats(resp.NodeStats)
	case "interface":
		resp.IfaceStats = new(api.InterfaceStats)
		return s.statsConn.GetInterfaceStats(resp.IfaceStats)
	case "error":
		resp.ErrStats = new(api.ErrorStats)
		return s.statsConn.GetErrorStats(resp.ErrStats)
	case "buffer":
		resp.BufStats = new(api.BufferStats)
		return s.statsConn.GetBufferStats(resp.BufStats)
	case "memory":
		resp.MemStats = new(api.MemoryStats)
		return s.statsConn.GetMemoryStats(resp.MemStats)
	default:
		return fmt.Errorf("unknown stats type: %s", req.StatsType)
	}
}

type BinapiRequest struct {
	Msg      api.Message
	IsMulti  bool
	ReplyMsg api.Message
	Timeout  time.Duration
}

type BinapiResponse struct {
	Msg  api.Message
	Msgs []api.Message
}

type BinapiCompatibilityRequest struct {
	MsgNameCrcs []string
}

type BinapiCompatibilityResponse struct {
	CompatibleMsgs   map[string][]string
	IncompatibleMsgs map[string][]string
}

// BinapiRPC is a RPC server for proxying client request to api.Channel
// or api.Stream.
type BinapiRPC struct {
	binapiConn *core.Connection
	binapi     adapter.VppAPI

	streamsLock sync.Mutex
	// local ID, different from api.Stream ID
	maxStreamID uint32
	streams     map[uint32]api.Stream

	events chan core.ConnectionEvent
	done   chan struct{}
	// non-zero if the RPC service is available
	available uint32
	// non-zero if connected to vpp.
	isConnected uint32
}

// NewBinapiRPC returns new BinapiRPC to be used as RPC server
// proxying request to given api.Channel.
func NewBinapiRPC(binapi adapter.VppAPI) (*BinapiRPC, error) {
	rpc := new(BinapiRPC)
	if err := rpc.connect(binapi); err != nil {
		return nil, err
	}
	return rpc, nil
}

func (s *BinapiRPC) watchConnection() {
	for {
		select {
		case e := <-s.events:
			// If disconnect was called exit.
			if atomic.LoadUint32(&s.isConnected) == 0 {
				atomic.StoreUint32(&s.available, 0)
				return
			}

			switch e.State {
			case core.Connected:
				if !s.serviceAvailable() {
					atomic.StoreUint32(&s.available, 1)
					log.Debugln("enabling binapiRPC service")
				}
			case core.Disconnected:
				if s.serviceAvailable() {
					atomic.StoreUint32(&s.available, 0)
					log.Warnf("disabling binapiRPC, reason: %v\n", e.Error)
				}
			case core.Failed:
				if s.serviceAvailable() {
					atomic.StoreUint32(&s.available, 0)
					log.Warnf("disabling binapiRPC, reason: %v\n", e.Error)
				}
				// vpp might have crashed/reset... reconnect
				s.binapiConn.Disconnect()

				var err error
				s.binapiConn, s.events, err = core.AsyncConnect(s.binapi, 3, 5*time.Second)
				if err != nil {
					log.Println(err)
				}
			}
		case <-s.done:
			return
		}
	}
}

func (s *BinapiRPC) connect(binapi adapter.VppAPI) error {
	if atomic.LoadUint32(&s.isConnected) == 1 {
		return errors.New("connection already exists")
	}
	s.binapi = binapi
	var err error
	s.binapiConn, s.events, err = core.AsyncConnect(binapi, 3, time.Second)
	if err != nil {
		return err
	}
	s.done = make(chan struct{})
	atomic.StoreUint32(&s.isConnected, 1)

	go s.watchConnection()
	return nil
}

func (s *BinapiRPC) disconnect() {
	if atomic.LoadUint32(&s.isConnected) == 1 {
		atomic.StoreUint32(&s.isConnected, 0)
		close(s.done)
		s.binapiConn.Disconnect()
		s.binapiConn = nil
	}
}

func (s *BinapiRPC) serviceAvailable() bool {
	return atomic.LoadUint32(&s.available) == 1
}

type RPCStreamReqResp struct {
	ID  uint32
	Msg api.Message
}

func (s *BinapiRPC) NewAPIStream(req RPCStreamReqResp, resp *RPCStreamReqResp) error {
	if !s.serviceAvailable() {
		log.Print(binapiErrorMsg)
		return errors.New("server does not support RPC calls at this time, try again later")
	}
	log.Debugf("BinapiRPC.NewAPIStream - REQ: %#v", req)

	stream, err := s.binapiConn.NewStream(context.Background())
	if err != nil {
		return err
	}

	if s.streams == nil {
		s.streams = make(map[uint32]api.Stream)
	}

	s.streamsLock.Lock()
	s.maxStreamID++
	s.streams[s.maxStreamID] = stream
	resp.ID = s.maxStreamID
	s.streamsLock.Unlock()

	return nil
}

func (s *BinapiRPC) SendMessage(req RPCStreamReqResp, resp *RPCStreamReqResp) error {
	if !s.serviceAvailable() {
		log.Print(binapiErrorMsg)
		return errors.New("server does not support RPC calls at this time, try again later")
	}
	log.Debugf("BinapiRPC.SendMessage - REQ: %#v", req)

	stream, err := s.getStream(req.ID)
	if err != nil {
		return err
	}

	return stream.SendMsg(req.Msg)
}

func (s *BinapiRPC) ReceiveMessage(req RPCStreamReqResp, resp *RPCStreamReqResp) error {
	if !s.serviceAvailable() {
		log.Print(binapiErrorMsg)
		return errors.New("server does not support RPC calls at this time, try again later")
	}
	log.Debugf("BinapiRPC.ReceiveMessage - REQ: %#v", req)

	stream, err := s.getStream(req.ID)
	if err != nil {
		return err
	}

	resp.Msg, err = stream.RecvMsg()
	return err
}

func (s *BinapiRPC) CloseStream(req RPCStreamReqResp, resp *RPCStreamReqResp) error {
	if !s.serviceAvailable() {
		log.Print(binapiErrorMsg)
		return errors.New("server does not support RPC calls at this time, try again later")
	}
	log.Debugf("BinapiRPC.CloseStream - REQ: %#v", req)

	stream, err := s.getStream(req.ID)
	if err != nil {
		return err
	}

	s.streamsLock.Lock()
	delete(s.streams, req.ID)
	s.streamsLock.Unlock()

	return stream.Close()
}

func (s *BinapiRPC) getStream(id uint32) (api.Stream, error) {
	s.streamsLock.Lock()
	stream := s.streams[id]
	s.streamsLock.Unlock()

	if stream == nil || reflect.ValueOf(stream).IsNil() {
		s.streamsLock.Lock()
		// delete the stream in case it is still in the map
		delete(s.streams, id)
		s.streamsLock.Unlock()
		return nil, errors.New("BinapiRPC stream closed")
	}
	return stream, nil
}

func (s *BinapiRPC) Invoke(req BinapiRequest, resp *BinapiResponse) error {
	if !s.serviceAvailable() {
		log.Print(binapiErrorMsg)
		return errors.New("server does not support 'invoke' at this time, try again later")
	}
	log.Debugf("BinapiRPC.Invoke - REQ: %#v", req)

	ch, err := s.binapiConn.NewAPIChannel()
	if err != nil {
		return err
	}
	defer ch.Close()
	ch.SetReplyTimeout(req.Timeout)

	if req.IsMulti {
		multi := ch.SendMultiRequest(req.Msg)
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

		err := ch.SendRequest(req.Msg).ReceiveReply(resp.Msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *BinapiRPC) Compatibility(req BinapiCompatibilityRequest, resp *BinapiCompatibilityResponse) error {
	if !s.serviceAvailable() {
		log.Print(binapiErrorMsg)
		return errors.New("server does not support 'compatibility check' at this time, try again later")
	}
	log.Debugf("BinapiRPC.Compatiblity - REQ: %#v", req)

	ch, err := s.binapiConn.NewAPIChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	resp.CompatibleMsgs = make(map[string][]string)
	resp.IncompatibleMsgs = make(map[string][]string)

	for path, messages := range api.GetRegisteredMessages() {
		resp.IncompatibleMsgs[path] = make([]string, 0, len(req.MsgNameCrcs))
		resp.CompatibleMsgs[path] = make([]string, 0, len(req.MsgNameCrcs))

		for _, msg := range req.MsgNameCrcs {
			val, ok := messages[msg]
			if !ok {
				resp.IncompatibleMsgs[path] = append(resp.IncompatibleMsgs[path], msg)
				continue
			}
			if err = ch.CheckCompatiblity(val); err != nil {
				resp.IncompatibleMsgs[path] = append(resp.IncompatibleMsgs[path], msg)
			} else {
				resp.CompatibleMsgs[path] = append(resp.CompatibleMsgs[path], msg)
			}
		}
	}

	compatible := false
	for path, incompatibleMsgs := range resp.IncompatibleMsgs {
		if len(incompatibleMsgs) == 0 {
			compatible = true
		} else {
			log.Debugf("messages are incompatible for path %s", path)
		}
	}
	if !compatible {
		return errors.New("compatibility check failed")
	}

	return nil
}
