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

package core

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"go.fd.io/govpp/adapter/mock"
	"go.fd.io/govpp/api"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/interface_types"
	"go.fd.io/govpp/binapi/memif"
	"go.fd.io/govpp/binapi/vpe"
)

type testCtx struct {
	mockVpp *mock.VppAdapter
	conn    *Connection
	ch      api.Channel
}

func setupTest(t *testing.T) *testCtx {
	RegisterTestingT(t)

	ctx := &testCtx{
		mockVpp: mock.NewVppAdapter(),
	}

	var err error
	ctx.conn, err = Connect(ctx.mockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	ctx.ch, err = ctx.conn.NewAPIChannel()
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

func TestChannelReset(t *testing.T) {
	RegisterTestingT(t)

	mockVpp := mock.NewVppAdapter()

	conn, err := Connect(mockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	ch, err := conn.NewAPIChannel()
	Expect(err).ShouldNot(HaveOccurred())

	Ch := ch.(*Channel)
	Ch.replyChan <- &vppReply{seqNum: 1}

	id := Ch.GetID()
	Expect(id).To(BeNumerically(">", 0))

	active := func() bool {
		conn.channelsLock.RLock()
		_, ok := conn.channels[id]
		conn.channelsLock.RUnlock()
		return ok
	}
	Expect(active()).To(BeTrue())

	Expect(Ch.replyChan).To(HaveLen(1))

	ch.Close()

	Eventually(active).Should(BeFalse())
	Eventually(func() int {
		return len(Ch.replyChan)
	}).Should(BeZero())
}

func TestRequestReplyMemifCreate(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	ctx.mockVpp.MockReply(&memif.MemifCreateReply{
		SwIfIndex: 4,
	})

	request := &memif.MemifCreate{
		Role:       10,
		ID:         12,
		RingSize:   8000,
		BufferSize: 50,
	}
	reply := &memif.MemifCreateReply{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reply.Retval).To(BeEquivalentTo(0),
		"Incorrect Retval value for MemifCreate")
	Expect(reply.SwIfIndex).To(BeEquivalentTo(4),
		"Incorrect SwIfIndex value for MemifCreate")
}

func TestRequestReplyMemifDelete(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	ctx.mockVpp.MockReply(&memif.MemifDeleteReply{})

	request := &memif.MemifDelete{
		SwIfIndex: 15,
	}
	reply := &memif.MemifDeleteReply{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
}

func TestRequestReplyMemifDetails(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	ctx.mockVpp.MockReply(&memif.MemifDetails{
		SwIfIndex: 25,
		IfName:    "memif-name",
		Role:      0,
	})

	request := &memif.MemifDump{}
	reply := &memif.MemifDetails{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reply.SwIfIndex).To(BeEquivalentTo(25),
		"Incorrect SwIfIndex value for MemifDetails")
	Expect(reply.IfName).ToNot(BeEmpty(),
		"MemifDetails IfName is empty byte array")
	Expect(reply.Role).To(BeEquivalentTo(0),
		"Incorrect Role value for MemifDetails")
}

func TestMultiRequestReplySwInterfaceMemifDump(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	var msgs []api.Message
	for i := 1; i <= 10; i++ {
		msgs = append(msgs, &memif.MemifDetails{
			SwIfIndex: interface_types.InterfaceIndex(i),
		})
	}
	ctx.mockVpp.MockReply(msgs...)
	ctx.mockVpp.MockReply(&ControlPingReply{})

	reqCtx := ctx.ch.SendMultiRequest(&memif.MemifDump{})
	cnt := 0
	for {
		msg := &memif.MemifDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		Expect(err).ShouldNot(HaveOccurred())
		cnt++
	}
	Expect(cnt).To(BeEquivalentTo(10))
}

func TestNotificationEvent(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// subscribe for notification
	notifChan := make(chan api.Message, 1)
	sub, err := ctx.ch.SubscribeNotification(notifChan, &interfaces.SwInterfaceEvent{})
	Expect(err).ShouldNot(HaveOccurred())

	// mock event and force its delivery
	ctx.mockVpp.MockReply(&interfaces.SwInterfaceEvent{
		SwIfIndex: 2,
		Flags:     interface_types.IF_STATUS_API_FLAG_LINK_UP,
	})
	err = ctx.mockVpp.SendMsg(0, []byte(""))
	Expect(err).ShouldNot(HaveOccurred())

	// receive the notification
	var notif *interfaces.SwInterfaceEvent
	Eventually(func() *interfaces.SwInterfaceEvent {
		select {
		case n := <-notifChan:
			notif = n.(*interfaces.SwInterfaceEvent)
			return notif
		default:
			return nil
		}
	}).ShouldNot(BeNil())

	// verify the received notifications
	Expect(notif.SwIfIndex).To(BeEquivalentTo(2), "Incorrect SwIfIndex value for SwInterfaceSetFlags")
	Expect(notif.Flags).To(BeEquivalentTo(interface_types.IF_STATUS_API_FLAG_LINK_UP), "Incorrect LinkUpDown value for SwInterfaceSetFlags")

	err = sub.Unsubscribe()
	Expect(err).ShouldNot(HaveOccurred())
}

func TestSetReplyTimeout(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	ctx.mockVpp.MockReply(&ControlPingReply{})

	// first one request should work
	err := ctx.ch.SendRequest(&ControlPing{}).ReceiveReply(&ControlPingReply{})
	Expect(err).ShouldNot(HaveOccurred())

	ctx.ch.SetReplyTimeout(time.Millisecond * 1)

	// no other reply ready - expect timeout
	err = ctx.ch.SendRequest(&ControlPing{}).ReceiveReply(&ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("timeout"))
}

func TestSetReplyTimeoutMultiRequest(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	ctx.mockVpp.MockReply(
		&interfaces.SwInterfaceDetails{
			SwIfIndex:     1,
			InterfaceName: "if-name-test",
		},
		&interfaces.SwInterfaceDetails{
			SwIfIndex:     2,
			InterfaceName: "if-name-test",
		},
		&interfaces.SwInterfaceDetails{
			SwIfIndex:     3,
			InterfaceName: "if-name-test",
		},
	)
	ctx.mockVpp.MockReply(&ControlPingReply{})

	cnt := 0
	sendMultiRequest := func() error {
		reqCtx := ctx.ch.SendMultiRequest(&interfaces.SwInterfaceDump{})
		for {
			msg := &interfaces.SwInterfaceDetails{}
			stop, err := reqCtx.ReceiveReply(msg)
			if err != nil {
				return err
			}
			if stop {
				break
			}
			cnt++
		}
		return nil
	}

	// first one request should work
	err := sendMultiRequest()
	Expect(err).ShouldNot(HaveOccurred())

	ctx.ch.SetReplyTimeout(time.Millisecond * 1)

	// no other reply ready - expect timeout
	err = sendMultiRequest()
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("timeout"))

	Expect(cnt).To(BeEquivalentTo(3))
}

func TestReceiveReplyNegative(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// invalid context 1
	reqCtx1 := &requestCtx{}
	err := reqCtx1.ReceiveReply(&ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid request context"))

	// invalid context 2
	reqCtx2 := &multiRequestCtx{}
	_, err = reqCtx2.ReceiveReply(&ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid request context"))

	// NU
	reqCtx3 := &requestCtx{}
	err = reqCtx3.ReceiveReply(nil)
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid request context"))
}

func TestMultiRequestDouble(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	var msgs []mock.MsgWithContext
	for i := 1; i <= 3; i++ {
		msgs = append(msgs, mock.MsgWithContext{
			Msg: &interfaces.SwInterfaceDetails{
				SwIfIndex:     interface_types.InterfaceIndex(i),
				InterfaceName: "if-name-test",
			},
			Multipart: true,
			SeqNum:    1,
		})
	}
	msgs = append(msgs, mock.MsgWithContext{Msg: &ControlPingReply{}, Multipart: true, SeqNum: 1})

	for i := 1; i <= 3; i++ {
		msgs = append(msgs,
			mock.MsgWithContext{
				Msg: &interfaces.SwInterfaceDetails{
					SwIfIndex:     interface_types.InterfaceIndex(i),
					InterfaceName: "if-name-test",
				},
				Multipart: true,
				SeqNum:    2,
			})
	}
	msgs = append(msgs, mock.MsgWithContext{Msg: &ControlPingReply{}, Multipart: true, SeqNum: 2})

	ctx.mockVpp.MockReplyWithContext(msgs...)

	cnt := 0
	var sendMultiRequest = func() error {
		reqCtx := ctx.ch.SendMultiRequest(&interfaces.SwInterfaceDump{})
		for {
			msg := &interfaces.SwInterfaceDetails{}
			stop, err := reqCtx.ReceiveReply(msg)
			if stop {
				break
			}
			if err != nil {
				return err
			}
			cnt++
		}
		return nil
	}

	err := sendMultiRequest()
	Expect(err).ShouldNot(HaveOccurred())

	err = sendMultiRequest()
	Expect(err).ShouldNot(HaveOccurred())

	Expect(cnt).To(BeEquivalentTo(6))
}

func TestReceiveReplyAfterTimeout(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	ctx.mockVpp.MockReplyWithContext(mock.MsgWithContext{Msg: &ControlPingReply{}, SeqNum: 1})

	// first request should succeed
	err := ctx.ch.SendRequest(&ControlPing{}).ReceiveReply(&ControlPingReply{})
	Expect(err).ShouldNot(HaveOccurred())

	// second request should fail with timeout
	ctx.ch.SetReplyTimeout(time.Millisecond * 1)
	req := ctx.ch.SendRequest(&ControlPing{})
	time.Sleep(time.Millisecond * 2)
	err = req.ReceiveReply(&ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("timeout"))

	ctx.mockVpp.MockReplyWithContext(
		// late reply from previous request
		mock.MsgWithContext{
			Msg:    &ControlPingReply{},
			SeqNum: 2,
		},
		// normal reply for next request
		mock.MsgWithContext{
			Msg:    &interfaces.SwInterfaceSetFlagsReply{},
			SeqNum: 3,
		},
	)

	reply := &interfaces.SwInterfaceSetFlagsReply{}

	ctx.resetReplyTimeout()

	// third request should succeed
	err = ctx.ch.SendRequest(&interfaces.SwInterfaceSetFlags{
		SwIfIndex: 1,
		Flags:     interface_types.IF_STATUS_API_FLAG_ADMIN_UP,
	}).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
}

func TestReceiveReplyAfterTimeoutMultiRequest(t *testing.T) {
	/*
		TODO: fix mock adapter
		This test will fail because mock adapter will stop sending replies
		when it encounters control_ping_reply from multi request,
		thus never sending reply for next request
	*/
	t.Skip()

	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	ctx.mockVpp.MockReplyWithContext(mock.MsgWithContext{Msg: &ControlPingReply{}, SeqNum: 1})

	// first one request should work
	err := ctx.ch.SendRequest(&ControlPing{}).ReceiveReply(&ControlPingReply{})
	Expect(err).ShouldNot(HaveOccurred())

	ctx.ch.SetReplyTimeout(time.Millisecond * 1)

	cnt := 0
	var sendMultiRequest = func() error {
		reqCtx := ctx.ch.SendMultiRequest(&interfaces.SwInterfaceDump{})
		for {
			msg := &interfaces.SwInterfaceDetails{}
			stop, err := reqCtx.ReceiveReply(msg)
			if stop {
				break
			}
			if err != nil {
				return err
			}
			cnt++
		}
		return nil
	}
	err = sendMultiRequest()
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("timeout"))
	Expect(cnt).To(BeEquivalentTo(0))

	ctx.resetReplyTimeout()

	// simulating late replies
	var msgs []mock.MsgWithContext
	for i := 1; i <= 3; i++ {
		msgs = append(msgs, mock.MsgWithContext{
			Msg: &interfaces.SwInterfaceDetails{
				SwIfIndex:     interface_types.InterfaceIndex(i),
				InterfaceName: "if-name-test",
			},
			Multipart: true,
			SeqNum:    2,
		})
	}
	msgs = append(msgs, mock.MsgWithContext{Msg: &ControlPingReply{}, Multipart: true, SeqNum: 2})
	ctx.mockVpp.MockReplyWithContext(msgs...)

	// normal reply for next request
	ctx.mockVpp.MockReplyWithContext(mock.MsgWithContext{Msg: &interfaces.SwInterfaceSetFlagsReply{}, SeqNum: 3})

	req := &interfaces.SwInterfaceSetFlags{
		SwIfIndex: 1,
		Flags:     interface_types.IF_STATUS_API_FLAG_ADMIN_UP,
	}
	reply := &interfaces.SwInterfaceSetFlagsReply{}

	// should succeed
	err = ctx.ch.SendRequest(req).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
}

func TestInvalidMessageID(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	ctx.mockVpp.MockReply(&vpe.ShowVersionReply{})
	ctx.mockVpp.MockReply(&vpe.ShowVersionReply{})

	// first one request should work
	err := ctx.ch.SendRequest(&vpe.ShowVersion{}).ReceiveReply(&vpe.ShowVersionReply{})
	Expect(err).ShouldNot(HaveOccurred())

	// second should fail with error invalid message ID
	err = ctx.ch.SendRequest(&ControlPing{}).ReceiveReply(&ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("unexpected message"))
}
