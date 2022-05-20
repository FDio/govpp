package core

import (
	"context"
	"testing"
	"time"

	"git.fd.io/govpp.git/adapter/mock"
	. "github.com/onsi/gomega"
)

type streamCtx struct {
	mockVpp *mock.VppAdapter
	conn    *Connection
	stream  *Stream
}

func setupStreamTest(t *testing.T) *streamCtx {
	RegisterTestingT(t)

	ctx := &streamCtx{
		mockVpp: mock.NewVppAdapter(),
	}

	var err error
	ctx.conn, err = Connect(ctx.mockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	stream, err := ctx.conn.NewStream(context.TODO())
	Expect(err).ShouldNot(HaveOccurred())

	ctx.stream = stream.(*Stream)
	return ctx
}

func (ctx *streamCtx) teardownTest() {
	err := ctx.stream.Close()
	Expect(err).ShouldNot(HaveOccurred())
	ctx.conn.Disconnect()
}

func TestStreamSetReplyTimeout(t *testing.T) {
	ctx := setupStreamTest(t)
	defer ctx.teardownTest()

	ctx.stream.replyTimeout = time.Millisecond

	// mock reply
	ctx.mockVpp.MockReply(&ControlPingReply{})

	// first one request should work
	err := ctx.stream.SendMsg(&ControlPing{})
	Expect(err).ShouldNot(HaveOccurred())
	_, err = ctx.stream.RecvMsg()
	Expect(err).ShouldNot(HaveOccurred())

	// no other reply ready - expect timeout
	err = ctx.stream.SendMsg(&ControlPing{})
	Expect(err).ShouldNot(HaveOccurred())
	_, err = ctx.stream.RecvMsg()
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("timeout"))
}

func TestStreamReceiveReplyAfterTimeout(t *testing.T) {
	ctx := setupStreamTest(t)
	defer ctx.teardownTest()

	ctx.stream.replyTimeout = time.Millisecond

	// mock reply
	ctx.mockVpp.MockReply(&ControlPingReply{})

	// first one request should work
	err := ctx.stream.SendMsg(&ControlPing{})
	Expect(err).ShouldNot(HaveOccurred())
	_, err = ctx.stream.RecvMsg()
	Expect(err).ShouldNot(HaveOccurred())

	// no other reply ready - expect timeout
	err = ctx.stream.SendMsg(&ControlPing{})
	Expect(err).ShouldNot(HaveOccurred())
	_, err = ctx.stream.RecvMsg()
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("timeout"))

	// mock late reply
	ctx.mockVpp.MockReply(&ControlPingReply{})

	// should succeed
	err = ctx.stream.SendMsg(&ControlPing{})
	Expect(err).ShouldNot(HaveOccurred())
	_, err = ctx.stream.RecvMsg()
	Expect(err).ShouldNot(HaveOccurred())
}
