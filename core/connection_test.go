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

/*
import (
	"testing"

	"git.fd.io/govpp.git/adapter/mock"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	"git.fd.io/govpp.git/examples/bin_api/interfaces"
	"git.fd.io/govpp.git/examples/bin_api/stats"

	"git.fd.io/govpp.git/codec"
	. "github.com/onsi/gomega"
)

type testCtx struct {
	mockVpp *mock.VppAdapter
	conn    *core.Connection
	ch      api.Channel
}

func setupTest(t *testing.T, bufferedChan bool) *testCtx {
	RegisterTestingT(t)

	ctx := &testCtx{}
	ctx.mockVpp = &mock.VppAdapter{}

	var err error
	ctx.conn, err = core.Connect(ctx.mockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	if bufferedChan {
		ctx.ch, err = ctx.conn.NewAPIChannelBuffered(100, 100)
	} else {
		ctx.ch, err = ctx.conn.NewAPIChannel()
	}
	Expect(err).ShouldNot(HaveOccurred())

	return ctx
}

func (ctx *testCtx) teardownTest() {
	ctx.ch.Close()
	ctx.conn.Disconnect()
}

func TestSimpleRequest(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	ctx.mockVpp.MockReply(&vpe.ControlPingReply{Retval: -5})

	req := &vpe.ControlPing{}
	reply := &vpe.ControlPingReply{}

	// send the request and receive a reply
	ctx.ch.GetRequestChannel() <- &api.VppRequest{Message: req}
	vppReply := <-ctx.ch.GetReplyChannel()

	Expect(vppReply).ShouldNot(BeNil())
	Expect(vppReply.Error).ShouldNot(HaveOccurred())

	// decode the message
	err := ctx.ch.GetMessageDecoder().DecodeMsg(vppReply.Data, reply)
	Expect(err).ShouldNot(HaveOccurred())

	Expect(reply.Retval).To(BeEquivalentTo(-5))
}

func TestMultiRequest(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	msgs := []api.Message{}
	for m := 0; m < 10; m++ {
		msgs = append(msgs, &interfaces.SwInterfaceDetails{})
	}
	ctx.mockVpp.MockReply(msgs...)
	ctx.mockVpp.MockReply(&vpe.ControlPingReply{})

	// send multipart request
	ctx.ch.GetRequestChannel() <- &api.VppRequest{Message: &interfaces.SwInterfaceDump{}, Multipart: true}

	cnt := 0
	for {
		// receive a reply
		vppReply := <-ctx.ch.GetReplyChannel()
		if vppReply.LastReplyReceived {
			break // break out of the loop
		}
		Expect(vppReply.Error).ShouldNot(HaveOccurred())

		// decode the message
		reply := &interfaces.SwInterfaceDetails{}
		err := ctx.ch.GetMessageDecoder().DecodeMsg(vppReply.Data, reply)
		Expect(err).ShouldNot(HaveOccurred())
		cnt++
	}

	Expect(cnt).To(BeEquivalentTo(10))
}

func TestNotifications(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	// subscribe for notification
	notifChan := make(chan api.Message, 1)
	subscription := &api.NotifSubscription{
		NotifChan:  notifChan,
		MsgFactory: interfaces.NewSwInterfaceSetFlags,
	}
	ctx.ch.GetNotificationChannel() <- &api.NotifSubscribeRequest{
		Subscription: subscription,
		Subscribe:    true,
	}
	err := <-ctx.ch.GetNotificationReplyChannel()
	Expect(err).ShouldNot(HaveOccurred())

	// mock the notification and force its delivery
	ctx.mockVpp.MockReply(&interfaces.SwInterfaceSetFlags{
		SwIfIndex:   3,
		AdminUpDown: 1,
	})
	ctx.mockVpp.SendMsg(0, []byte{0})

	// receive the notification
	notif := (<-notifChan).(*interfaces.SwInterfaceSetFlags)

	Expect(notif.SwIfIndex).To(BeEquivalentTo(3))

	// unsubscribe notification
	ctx.ch.GetNotificationChannel() <- &api.NotifSubscribeRequest{
		Subscription: subscription,
		Subscribe:    false,
	}
	err = <-ctx.ch.GetNotificationReplyChannel()
	Expect(err).ShouldNot(HaveOccurred())
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

func TestDoubleConnection(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	conn, err := core.Connect(ctx.mockVpp)
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("only one connection per process"))
	Expect(conn).Should(BeNil())
}

func TestAsyncConnection(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	ctx.conn.Disconnect()
	conn, ch, err := core.AsyncConnect(ctx.mockVpp)
	ctx.conn = conn

	Expect(err).ShouldNot(HaveOccurred())
	Expect(conn).ShouldNot(BeNil())

	ev := <-ch
	Expect(ev.State).Should(BeEquivalentTo(core.Connected))
}

func TestFullBuffer(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	// close the default API channel
	ctx.ch.Close()

	// create a new channel with limited buffer sizes
	var err error
	ctx.ch, err = ctx.conn.NewAPIChannelBuffered(10, 1)
	Expect(err).ShouldNot(HaveOccurred())

	// send multiple requests, only one reply should be read
	for i := 0; i < 20; i++ {
		ctx.mockVpp.MockReply(&vpe.ControlPingReply{})
		ctx.ch.GetRequestChannel() <- &api.VppRequest{Message: &vpe.ControlPing{}}
	}

	vppReply := <-ctx.ch.GetReplyChannel()
	Expect(vppReply).ShouldNot(BeNil())

	var received bool
	select {
	case <-ctx.ch.GetReplyChannel():
		received = true // this should not happen
	default:
		received = false // no reply to be received
	}
	Expect(received).Should(BeFalse(), "A reply has been recieved, should had been ignored.")
}

func TestCodec(t *testing.T) {
	RegisterTestingT(t)

	msgCodec := &codec.MsgCodec{}

	// request
	data, err := msgCodec.EncodeMsg(&interfaces.CreateLoopback{MacAddress: []byte{1, 2, 3, 4, 5, 6}}, 11)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(data).ShouldNot(BeEmpty())

	msg1 := &interfaces.CreateLoopback{}
	err = msgCodec.DecodeMsg(data, msg1)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(msg1.MacAddress).To(BeEquivalentTo([]byte{1, 2, 3, 4, 5, 6}))

	// reply
	data, err = msgCodec.EncodeMsg(&vpe.ControlPingReply{Retval: 55}, 22)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(data).ShouldNot(BeEmpty())

	msg2 := &vpe.ControlPingReply{}
	err = msgCodec.DecodeMsg(data, msg2)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(msg2.Retval).To(BeEquivalentTo(55))

	// other
	data, err = msgCodec.EncodeMsg(&stats.VnetIP4FibCounters{VrfID: 77}, 33)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(data).ShouldNot(BeEmpty())

	msg3 := &stats.VnetIP4FibCounters{}
	err = msgCodec.DecodeMsg(data, msg3)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(msg3.VrfID).To(BeEquivalentTo(77))
}

func TestCodecNegative(t *testing.T) {
	RegisterTestingT(t)

	msgCodec := &codec.MsgCodec{}

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
	err = msgCodec.DecodeMsg(nil, &vpe.ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("EOF"))
}

func TestSimpleRequestsWithSequenceNumbers(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	var reqCtx []api.RequestCtx
	for i := 0; i < 10; i++ {
		ctx.mockVpp.MockReply(&vpe.ControlPingReply{Retval: int32(i)})
		req := &vpe.ControlPing{}
		reqCtx = append(reqCtx, ctx.ch.SendRequest(req))
	}

	for i := 0; i < 10; i++ {
		reply := &vpe.ControlPingReply{}
		err := reqCtx[i].ReceiveReply(reply)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(reply.Retval).To(BeEquivalentTo(i))
	}
}

func TestMultiRequestsWithSequenceNumbers(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	msgs := []api.Message{}
	for i := 0; i < 10; i++ {
		msgs = append(msgs, &interfaces.SwInterfaceDetails{SwIfIndex: uint32(i)})
	}
	ctx.mockVpp.MockReply(msgs...)
	ctx.mockVpp.MockReply(&vpe.ControlPingReply{})

	// send multipart request
	reqCtx := ctx.ch.SendMultiRequest(&interfaces.SwInterfaceDump{})

	cnt := 0
	for {
		Expect(cnt < 11).To(BeTrue())

		// receive a reply
		reply := &interfaces.SwInterfaceDetails{}
		lastReplyReceived, err := reqCtx.ReceiveReply(reply)

		if lastReplyReceived {
			break // break out of the loop
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
		Msg:    &vpe.ControlPingReply{Retval: 1},
		SeqNum: 0,
	})

	// send reply later
	req1 := &vpe.ControlPing{}
	reqCtx1 := ctx.ch.SendRequest(req1)

	reply := &vpe.ControlPingReply{}
	err := reqCtx1.ReceiveReply(reply)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(HavePrefix("no reply received within the timeout period"))

	ctx.mockVpp.MockReplyWithContext(
		// reply for the previous request
		mock.MsgWithContext{
			Msg:    &vpe.ControlPingReply{Retval: 1},
			SeqNum: 1,
		},
		// reply for the next request
		mock.MsgWithContext{
			Msg:    &vpe.ControlPingReply{Retval: 2},
			SeqNum: 2,
		})

	// next request
	req2 := &vpe.ControlPing{}
	reqCtx2 := ctx.ch.SendRequest(req2)

	// second request should ignore the first reply and return the second one
	reply = &vpe.ControlPingReply{}
	err = reqCtx2.ReceiveReply(reply)
	Expect(err).To(BeNil())
	Expect(reply.Retval).To(BeEquivalentTo(2))
}

func TestSimpleRequestsWithMissingReply(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	// request without reply
	req1 := &vpe.ControlPing{}
	reqCtx1 := ctx.ch.SendRequest(req1)

	// another request without reply
	req2 := &vpe.ControlPing{}
	reqCtx2 := ctx.ch.SendRequest(req2)

	// third request with reply
	ctx.mockVpp.MockReplyWithContext(mock.MsgWithContext{
		Msg:    &vpe.ControlPingReply{Retval: 3},
		SeqNum: 3,
	})
	req3 := &vpe.ControlPing{}
	reqCtx3 := ctx.ch.SendRequest(req3)

	// the first two should fail, but not consume reply for the 3rd
	reply := &vpe.ControlPingReply{}
	err := reqCtx1.ReceiveReply(reply)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("missing binary API reply with sequence number: 1"))

	reply = &vpe.ControlPingReply{}
	err = reqCtx2.ReceiveReply(reply)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(Equal("missing binary API reply with sequence number: 2"))

	// the second request should succeed
	reply = &vpe.ControlPingReply{}
	err = reqCtx3.ReceiveReply(reply)
	Expect(err).To(BeNil())
	Expect(reply.Retval).To(BeEquivalentTo(3))
}

func TestMultiRequestsWithErrors(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	// replies for a previous timeouted requests to be ignored
	msgs := []mock.MsgWithContext{}
	msgs = append(msgs,
		mock.MsgWithContext{
			Msg:    &vpe.ControlPingReply{Retval: 1},
			SeqNum: 0xffff - 1,
		},
		mock.MsgWithContext{
			Msg:    &vpe.ControlPingReply{Retval: 1},
			SeqNum: 0xffff,
		},
		mock.MsgWithContext{
			Msg:    &vpe.ControlPingReply{Retval: 1},
			SeqNum: 0,
		})

	for i := 0; i < 10; i++ {
		msgs = append(msgs,
			mock.MsgWithContext{
				Msg:       &interfaces.SwInterfaceDetails{SwIfIndex: uint32(i)},
				SeqNum:    1,
				Multipart: true,
			})
	}
	// missing finalizing control ping

	// reply for a next request
	msgs = append(msgs,
		mock.MsgWithContext{
			Msg:       &vpe.ControlPingReply{Retval: 2},
			SeqNum:    2,
			Multipart: false,
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
	reqCtx2 := ctx.ch.SendRequest(&vpe.ControlPing{})
	reply2 := &vpe.ControlPingReply{}
	err = reqCtx2.ReceiveReply(reply2)
	Expect(err).To(BeNil())
	Expect(reply2.Retval).To(BeEquivalentTo(2))
}

func TestRequestsOrdering(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	// the orderings of SendRequest and ReceiveReply calls should match, otherwise
	// some replies will get thrown away

	// first request
	ctx.mockVpp.MockReply(&vpe.ControlPingReply{Retval: 1})
	req1 := &vpe.ControlPing{}
	reqCtx1 := ctx.ch.SendRequest(req1)

	// second request
	ctx.mockVpp.MockReply(&vpe.ControlPingReply{Retval: 2})
	req2 := &vpe.ControlPing{}
	reqCtx2 := ctx.ch.SendRequest(req2)

	// if reply for the second request is read first, the reply for the first
	// request gets thrown away.
	reply2 := &vpe.ControlPingReply{}
	err := reqCtx2.ReceiveReply(reply2)
	Expect(err).To(BeNil())
	Expect(reply2.Retval).To(BeEquivalentTo(2))

	// first request has already been considered closed
	reply1 := &vpe.ControlPingReply{}
	err = reqCtx1.ReceiveReply(reply1)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(HavePrefix("no reply received within the timeout period"))
}

func TestCycleOverSetOfSequenceNumbers(t *testing.T) {
	ctx := setupTest(t, true)
	defer ctx.teardownTest()

	numIters := 0xffff + 100
	reqCtx := make(map[int]api.RequestCtx)

	for i := 0; i < numIters+30; i++ {
		if i < numIters {
			ctx.mockVpp.MockReply(&vpe.ControlPingReply{Retval: int32(i)})
			req := &vpe.ControlPing{}
			reqCtx[i] = ctx.ch.SendRequest(req)
		}
		if i > 30 {
			reply := &vpe.ControlPingReply{}
			err := reqCtx[i-30].ReceiveReply(reply)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(reply.Retval).To(BeEquivalentTo(i - 30))
		}
	}
}
*/
