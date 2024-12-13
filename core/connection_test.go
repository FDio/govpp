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

package core_test

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"go.fd.io/govpp/adapter/mock"
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/binapi/ethernet_types"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/interface_types"
	"go.fd.io/govpp/binapi/memclnt"
	"go.fd.io/govpp/codec"
	"go.fd.io/govpp/core"
)

type testCtx struct {
	mockVpp *mock.VppAdapter
	conn    *core.Connection
	ch      api.Channel
}

func setupTest(t *testing.T, bufferedChan bool) *testCtx {
	RegisterTestingT(t)

	ctx := &testCtx{
		mockVpp: mock.NewVppAdapter(),
	}

	var err error
	ctx.conn, err = core.Connect(ctx.mockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	if bufferedChan {
		ctx.ch, err = ctx.conn.NewAPIChannelBuffered(100, 100)
	} else {
		ctx.ch, err = ctx.conn.NewAPIChannel()
	}
	Expect(err).ShouldNot(HaveOccurred())

	ctx.resetReplyTimeout()

	return ctx
}

func (ctx *testCtx) resetReplyTimeout() {
	// setting reply timeout to non-zero value to fail fast on potential deadlocks
	ctx.ch.SetReplyTimeout(time.Second * 5)
}

func (ctx *testCtx) teardownTest() {
	ctx.ch.Close()
	ctx.conn.Disconnect()
}

func TestNilConnection(t *testing.T) {
	RegisterTestingT(t)
	var conn *core.Connection

	ch, err := conn.NewAPIChannel()
	Expect(ch).Should(BeNil())
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("nil"))

	ch, err = conn.NewAPIChannelBuffered(1, 1)
	Expect(ch).Should(BeNil())
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("nil"))
}

func TestAsyncConnection(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	var (
		connectCalled    int
		disconnectCalled int
	)

	ctx.conn.Disconnect()

	ctx.mockVpp.SetConnectCallback(func() { connectCalled++ })
	ctx.mockVpp.SetDisconnectCallback(func() { disconnectCalled++ })
	conn, statusChan, err := core.AsyncConnect(ctx.mockVpp, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	ctx.conn = conn

	Expect(err).ShouldNot(HaveOccurred())
	Expect(conn).ShouldNot(BeNil())

	ev := <-statusChan
	Expect(ev.State).Should(BeEquivalentTo(core.Connected))

	conn.Disconnect()
	Expect(connectCalled).Should(BeEquivalentTo(1))
	Expect(disconnectCalled).Should(BeEquivalentTo(1))
}

func TestAsyncConnectionProcessesVppTimeout(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	var (
		connectCalled    int
		disconnectCalled int
	)

	ctx.conn.Disconnect()

	ctx.mockVpp.SetConnectCallback(func() {
		if connectCalled == 0 {
			ctx.mockVpp.MockConnectError(fmt.Errorf("no VPP present"))
		} else {
			ctx.mockVpp.MockConnectError(nil)
		}
		connectCalled++
	})
	ctx.mockVpp.SetDisconnectCallback(func() { disconnectCalled++ })
	conn, statusChan, err := core.AsyncConnect(ctx.mockVpp, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	ctx.conn = conn

	Expect(err).ShouldNot(HaveOccurred())
	Expect(conn).ShouldNot(BeNil())

	ev := <-statusChan
	Expect(ev.State).Should(BeEquivalentTo(core.Connected))

	conn.Disconnect()
	Expect(connectCalled).Should(BeEquivalentTo(2))
	Expect(disconnectCalled).Should(BeEquivalentTo(1))
}

func TestAsyncConnectionEarlyDisconnect(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	var (
		connectCalled    int
		disconnectCalled int
	)

	timeout := 100 * time.Millisecond

	ctx.conn.Disconnect()

	ctx.mockVpp.MockConnectError(fmt.Errorf("no VPP present"))
	ctx.mockVpp.SetConnectCallback(func() { connectCalled++ })
	ctx.mockVpp.SetDisconnectCallback(func() { disconnectCalled++ })

	conn, statusChan, err := core.AsyncConnect(ctx.mockVpp, math.MaxInt, timeout)
	ctx.conn = conn

	Expect(err).ShouldNot(HaveOccurred())
	Expect(conn).ShouldNot(BeNil())

	timer := time.NewTimer(8 * timeout)
	time.Sleep(3 * timeout)
	conn.Disconnect()

	// Check if disconnect disconnects
	ctx.mockVpp.MockConnectError(nil)
	time.Sleep(2 * timeout)

	var connected bool
	select {
	case <-statusChan:
		connected = true
	case <-timer.C:
		break
	}

	Expect(connected).Should(BeFalse())
	Expect(connectCalled).Should(BeNumerically(">", 1))
	Expect(disconnectCalled).Should(BeEquivalentTo(0))
}

func TestAsyncConnectionDoubleDisconnect(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	var (
		connectCalled    int
		disconnectCalled int
	)

	ctx.conn.Disconnect()

	ctx.mockVpp.SetConnectCallback(func() { connectCalled++ })
	ctx.mockVpp.SetDisconnectCallback(func() { disconnectCalled++ })
	conn, statusChan, err := core.AsyncConnect(ctx.mockVpp, core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
	ctx.conn = conn

	Expect(err).ShouldNot(HaveOccurred())
	Expect(conn).ShouldNot(BeNil())

	ev := <-statusChan
	Expect(ev.State).Should(BeEquivalentTo(core.Connected))

	conn.Disconnect()
	conn.Disconnect()
	Expect(connectCalled).Should(BeEquivalentTo(1))
	Expect(disconnectCalled).Should(BeEquivalentTo(1))
}

func TestCodec(t *testing.T) {
	RegisterTestingT(t)

	var msgCodec = codec.DefaultCodec

	// request
	data, err := msgCodec.EncodeMsg(&interfaces.CreateLoopback{MacAddress: ethernet_types.MacAddress{1, 2, 3, 4, 5, 6}}, 11)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(data).ShouldNot(BeEmpty())

	msg1 := &interfaces.CreateLoopback{}
	err = msgCodec.DecodeMsg(data, msg1)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(msg1.MacAddress).To(BeEquivalentTo(ethernet_types.MacAddress{1, 2, 3, 4, 5, 6}))

	// reply
	data, err = msgCodec.EncodeMsg(&memclnt.ControlPingReply{Retval: 55}, 22)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(data).ShouldNot(BeEmpty())

	msg2 := &memclnt.ControlPingReply{}
	err = msgCodec.DecodeMsg(data, msg2)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(msg2.Retval).To(BeEquivalentTo(55))
}

func TestCodecNegative(t *testing.T) {
	RegisterTestingT(t)

	var msgCodec = codec.DefaultCodec

	// nil message for encoding
	data, err := msgCodec.EncodeMsg(nil, 15)
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("nil message"))
	Expect(data).Should(BeNil())

	// nil message for decoding
	err = msgCodec.DecodeMsg(data, nil)
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("nil message"))

	// nil data for decoding
	err = msgCodec.DecodeMsg(nil, &memclnt.ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("panic"))
}

func TestSimpleRequestsWithSequenceNumbers(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	var reqCtx []api.RequestCtx
	for i := 0; i < 10; i++ {
		ctx.mockVpp.MockReply(&memclnt.ControlPingReply{})
		req := &memclnt.ControlPing{}
		reqCtx = append(reqCtx, ctx.ch.SendRequest(req))
	}

	for i := 0; i < 10; i++ {
		reply := &memclnt.ControlPingReply{}
		err := reqCtx[i].ReceiveReply(reply)
		Expect(err).ShouldNot(HaveOccurred())
	}
}

func TestMultiRequestsWithSequenceNumbers(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	var msgs []api.Message
	for i := 0; i < 10; i++ {
		msgs = append(msgs, &interfaces.SwInterfaceDetails{SwIfIndex: interface_types.InterfaceIndex(i)})
	}
	ctx.mockVpp.MockReply(msgs...)
	ctx.mockVpp.MockReply(&memclnt.ControlPingReply{})

	// send multipart request
	reqCtx := ctx.ch.SendMultiRequest(&interfaces.SwInterfaceDump{})

	cnt := 0
	for {
		Expect(cnt < 11).To(BeTrue())

		// receive a reply
		reply := &interfaces.SwInterfaceDetails{}
		lastReplyReceived, err := reqCtx.ReceiveReply(reply)

		if lastReplyReceived {
			break
		}

		Expect(err).ShouldNot(HaveOccurred())
		Expect(reply.SwIfIndex).To(BeEquivalentTo(cnt))

		cnt++
	}

	Expect(cnt).To(BeEquivalentTo(10))
}

func TestSimpleRequestWithTimeout(t *testing.T) {
	ctx := setupTest(t, true)
	defer ctx.teardownTest()

	// reply for a previous timeouted requests to be ignored
	ctx.mockVpp.MockReplyWithContext(mock.MsgWithContext{
		Msg:    &memclnt.ControlPingReply{},
		SeqNum: 0,
	})

	// send reply later
	req1 := &memclnt.ControlPing{}
	reqCtx1 := ctx.ch.SendRequest(req1)

	ctx.ch.SetReplyTimeout(time.Millisecond)
	time.Sleep(time.Millisecond)

	reply := &memclnt.ControlPingReply{}
	err := reqCtx1.ReceiveReply(reply)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(HavePrefix("no reply received within the timeout period"))
	Expect(errors.Is(err, core.ErrReplyTimeout)).To(Equal(true))

	ctx.mockVpp.MockReplyWithContext(
		// reply for the previous request
		mock.MsgWithContext{
			Msg:    &memclnt.ControlPingReply{},
			SeqNum: 1,
		},
		// reply for the next request
		mock.MsgWithContext{
			Msg:    &memclnt.ControlPingReply{},
			SeqNum: 2,
		})

	// next request
	req2 := &memclnt.ControlPing{}
	reqCtx2 := ctx.ch.SendRequest(req2)

	ctx.resetReplyTimeout()

	// second request should ignore the first reply and return the second one
	reply = &memclnt.ControlPingReply{}
	err = reqCtx2.ReceiveReply(reply)
	Expect(err).To(BeNil())
}

func TestSimpleRequestsWithMissingReply(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	// request without reply
	req1 := &memclnt.ControlPing{}
	reqCtx1 := ctx.ch.SendRequest(req1)

	// another request without reply
	req2 := &memclnt.ControlPing{}
	reqCtx2 := ctx.ch.SendRequest(req2)

	// third request with reply
	ctx.mockVpp.MockReplyWithContext(mock.MsgWithContext{
		Msg:    &memclnt.ControlPingReply{},
		SeqNum: 3,
	})
	req3 := &memclnt.ControlPing{}
	reqCtx3 := ctx.ch.SendRequest(req3)

	// the first two should fail, but not consume reply for the 3rd
	reply := &memclnt.ControlPingReply{}
	err := reqCtx1.ReceiveReply(reply)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("missing binary API reply with sequence number: 1"))

	reply = &memclnt.ControlPingReply{}
	err = reqCtx2.ReceiveReply(reply)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("missing binary API reply with sequence number: 2"))

	// the second request should succeed
	reply = &memclnt.ControlPingReply{}
	err = reqCtx3.ReceiveReply(reply)
	Expect(err).To(BeNil())
}

func TestMultiRequestsWithErrors(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	// replies for a previous timeouted requests to be ignored
	msgs := []mock.MsgWithContext{
		{Msg: &memclnt.ControlPingReply{}, SeqNum: 0xffff - 1},
		{Msg: &memclnt.ControlPingReply{}, SeqNum: 0xffff},
		{Msg: &memclnt.ControlPingReply{}, SeqNum: 0},
	}
	for i := 0; i < 10; i++ {
		msgs = append(msgs, mock.MsgWithContext{
			Msg:       &interfaces.SwInterfaceDetails{SwIfIndex: interface_types.InterfaceIndex(i)},
			SeqNum:    1,
			Multipart: true,
		})
	}
	// missing finalizing control ping

	// reply for a next request
	msgs = append(msgs, mock.MsgWithContext{
		Msg:    &memclnt.ControlPingReply{},
		SeqNum: 2,
	})

	// queue replies
	ctx.mockVpp.MockReplyWithContext(msgs...)

	// send multipart request
	reqCtx := ctx.ch.SendMultiRequest(&interfaces.SwInterfaceDump{})

	for i := 0; i < 10; i++ {
		// receive multi-part replies
		reply := &interfaces.SwInterfaceDetails{}
		lastReplyReceived, err := reqCtx.ReceiveReply(reply)

		Expect(lastReplyReceived).To(BeFalse())
		Expect(err).ShouldNot(HaveOccurred())
		Expect(reply.SwIfIndex).To(BeEquivalentTo(i))
	}

	// missing closing control ping
	reply := &interfaces.SwInterfaceDetails{}
	_, err := reqCtx.ReceiveReply(reply)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("missing binary API reply with sequence number: 1"))

	// try again - still fails and nothing consumed
	_, err = reqCtx.ReceiveReply(reply)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("missing binary API reply with sequence number: 1"))

	// reply for the second request has not been consumed
	reqCtx2 := ctx.ch.SendRequest(&memclnt.ControlPing{})
	reply2 := &memclnt.ControlPingReply{}
	err = reqCtx2.ReceiveReply(reply2)
	Expect(err).To(BeNil())
}

func TestRequestsOrdering(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	// the orderings of SendRequest and ReceiveReply calls should match, otherwise
	// some replies will get thrown away

	// first request
	ctx.mockVpp.MockReply(&memclnt.ControlPingReply{})
	req1 := &memclnt.ControlPing{}
	reqCtx1 := ctx.ch.SendRequest(req1)

	// second request
	ctx.mockVpp.MockReply(&memclnt.ControlPingReply{})
	req2 := &memclnt.ControlPing{}
	reqCtx2 := ctx.ch.SendRequest(req2)

	// if reply for the second request is read first, the reply for the first
	// request gets thrown away.
	reply2 := &memclnt.ControlPingReply{}
	err := reqCtx2.ReceiveReply(reply2)
	Expect(err).To(BeNil())

	ctx.ch.SetReplyTimeout(time.Millisecond)

	// first request has already been considered closed
	reply1 := &memclnt.ControlPingReply{}
	err = reqCtx1.ReceiveReply(reply1)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(HavePrefix("no reply received within the timeout period"))
	Expect(errors.Is(err, core.ErrReplyTimeout)).To(Equal(true))
}

func TestCycleOverSetOfSequenceNumbers(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	numIters := 0xffff + 100
	reqCtx := make(map[int]api.RequestCtx)

	for i := 0; i < numIters+30; i++ {
		if i < numIters {
			ctx.mockVpp.MockReply(&memclnt.ControlPingReply{})
			req := &memclnt.ControlPing{}
			reqCtx[i] = ctx.ch.SendRequest(req)
		}
		if i > 30 {
			reply := &memclnt.ControlPingReply{}
			err := reqCtx[i-30].ReceiveReply(reply)
			Expect(err).ShouldNot(HaveOccurred())
		}
	}
}
