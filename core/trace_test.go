package core_test

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"go.fd.io/govpp/api"
	interfaces "go.fd.io/govpp/binapi/interface"
	"go.fd.io/govpp/binapi/ip"
	"go.fd.io/govpp/binapi/l2"
	"go.fd.io/govpp/binapi/memclnt"
	"go.fd.io/govpp/binapi/memif"
	"go.fd.io/govpp/core"
)

func TestTraceEnabled(t *testing.T) {
	t.Skipf("these randomly fail, see integration tests")

	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	Expect(ctx.conn.Trace()).ToNot(BeNil())
	ctx.conn.Trace().Enable(true)

	request := []api.Message{
		&interfaces.CreateLoopback{},
		&memif.MemifCreate{},
		&l2.BridgeDomainAddDel{},
		&ip.IPTableAddDel{},
	}
	reply := []api.Message{
		&interfaces.CreateLoopbackReply{},
		&memif.MemifCreateReply{},
		&l2.BridgeDomainAddDelReply{},
		&ip.IPTableAddDelReply{},
	}

	for i := 0; i < len(request); i++ {
		ctx.mockVpp.MockReply(reply[i])
		err := ctx.ch.SendRequest(request[i]).ReceiveReply(reply[i])
		Expect(err).To(BeNil())
	}

	traced := ctx.conn.Trace().GetRecords()
	Expect(traced).ToNot(BeNil())
	Expect(traced).To(HaveLen(8))
	for i, entry := range traced {
		Expect(entry.Timestamp).ToNot(BeNil())
		Expect(entry.Message.GetMessageName()).ToNot(Equal(""))
		if strings.HasSuffix(entry.Message.GetMessageName(), "_reply") ||
			strings.HasSuffix(entry.Message.GetMessageName(), "_details") {
			Expect(entry.IsReceived).To(BeTrue())
		} else {
			Expect(entry.IsReceived).To(BeFalse())
		}
		if i%2 == 0 {
			Expect(request[i/2].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		} else {
			Expect(reply[i/2].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		}
	}
}

func TestMultiRequestTraceEnabled(t *testing.T) {
	t.Skipf("these randomly fail, see integration tests")

	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	ctx.conn.Trace().Enable(true)

	request := []api.Message{
		&interfaces.SwInterfaceDump{},
	}
	reply := []api.Message{
		&interfaces.SwInterfaceDetails{
			SwIfIndex: 1,
		},
		&interfaces.SwInterfaceDetails{
			SwIfIndex: 2,
		},
		&interfaces.SwInterfaceDetails{
			SwIfIndex: 3,
		},
		&memclnt.ControlPingReply{},
	}

	ctx.mockVpp.MockReply(reply[0 : len(reply)-1]...)
	ctx.mockVpp.MockReply(reply[len(reply)-1])
	multiCtx := ctx.ch.SendMultiRequest(request[0])

	i := 0
	for {
		last, err := multiCtx.ReceiveReply(reply[i])
		Expect(err).ToNot(HaveOccurred())
		if last {
			break
		}
		i++
	}

	traced := ctx.conn.Trace().GetRecords()
	Expect(traced).ToNot(BeNil())
	Expect(traced).To(HaveLen(6))
	for _, entry := range traced {
		Expect(entry.Timestamp).ToNot(BeNil())
		Expect(entry.Message.GetMessageName()).ToNot(Equal(""))
		if strings.HasSuffix(entry.Message.GetMessageName(), "_reply") ||
			strings.HasSuffix(entry.Message.GetMessageName(), "_details") {
			Expect(entry.IsReceived).To(BeTrue())
		} else {
			Expect(entry.IsReceived).To(BeFalse())
		}
		// FIXME: the way mock adapter works now prevents having the exact same order for each execution
		/*if i == 0 {
		  	Expect(request[0].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		  } else if i == len(traced)-1 {
		  	msg := memclnt.ControlPing{}
		  	Expect(msg.GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		  } else {
		  	Expect(reply[i-1].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		  }*/
	}
}

func TestTraceDisabled(t *testing.T) {
	t.Skipf("these randomly fail, see integration tests")

	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	ctx.conn.Trace().Enable(false)

	request := []api.Message{
		&interfaces.CreateLoopback{},
		&memif.MemifCreate{},
		&l2.BridgeDomainAddDel{},
		&ip.IPTableAddDel{},
	}
	reply := []api.Message{
		&interfaces.CreateLoopbackReply{},
		&memif.MemifCreateReply{},
		&l2.BridgeDomainAddDelReply{},
		&ip.IPTableAddDelReply{},
	}

	for i := 0; i < len(request); i++ {
		ctx.mockVpp.MockReply(reply[i])
		err := ctx.ch.SendRequest(request[i]).ReceiveReply(reply[i])
		Expect(err).To(BeNil())
	}

	traced := ctx.conn.Trace().GetRecords()
	Expect(traced).To(BeNil())
}

func TestTracePerChannel(t *testing.T) {
	t.Skipf("these randomly fail, see integration tests")

	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	ctx.conn.Trace().Enable(true)

	ch1 := ctx.ch
	ch2, err := ctx.conn.NewAPIChannel()
	Expect(err).ToNot(HaveOccurred())

	requestCh1 := []api.Message{
		&interfaces.CreateLoopback{},
		&memif.MemifCreate{},
		&l2.BridgeDomainAddDel{},
	}
	replyCh1 := []api.Message{
		&interfaces.CreateLoopbackReply{},
		&memif.MemifCreateReply{},
		&l2.BridgeDomainAddDelReply{},
	}
	requestCh2 := []api.Message{
		&ip.IPTableAddDel{},
	}
	replyCh2 := []api.Message{
		&ip.IPTableAddDelReply{},
	}

	for i := 0; i < len(requestCh1); i++ {
		ctx.mockVpp.MockReply(replyCh1[i])
		err := ch1.SendRequest(requestCh1[i]).ReceiveReply(replyCh1[i])
		Expect(err).To(BeNil())
	}
	for i := 0; i < len(requestCh2); i++ {
		ctx.mockVpp.MockReply(replyCh2[i])
		err := ch2.SendRequest(requestCh2[i]).ReceiveReply(replyCh2[i])
		Expect(err).To(BeNil())
	}

	trace := ctx.conn.Trace().GetRecords()
	Expect(trace).ToNot(BeNil())
	Expect(trace).To(HaveLen(8))

	// per channel
	channel1, ok := ch1.(*core.Channel)
	Expect(ok).To(BeTrue())
	channel2, ok := ch2.(*core.Channel)
	Expect(ok).To(BeTrue())

	tracedCh1 := ctx.conn.Trace().GetRecordsForChannel(channel1.GetID())
	Expect(tracedCh1).ToNot(BeNil())
	Expect(tracedCh1).To(HaveLen(6))
	for i, entry := range tracedCh1 {
		Expect(entry.Timestamp).ToNot(BeNil())
		Expect(entry.Message.GetMessageName()).ToNot(Equal(""))
		if strings.HasSuffix(entry.Message.GetMessageName(), "_reply") ||
			strings.HasSuffix(entry.Message.GetMessageName(), "_details") {
			Expect(entry.IsReceived).To(BeTrue())
		} else {
			Expect(entry.IsReceived).To(BeFalse())
		}
		if i%2 == 0 {
			Expect(requestCh1[i/2].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		} else {
			Expect(replyCh1[i/2].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		}
	}

	tracedCh2 := ctx.conn.Trace().GetRecordsForChannel(channel2.GetID())
	Expect(tracedCh2).ToNot(BeNil())
	Expect(tracedCh2).To(HaveLen(2))
	for i, entry := range tracedCh2 {
		Expect(entry.Timestamp).ToNot(BeNil())
		Expect(entry.Message.GetMessageName()).ToNot(Equal(""))
		if strings.HasSuffix(entry.Message.GetMessageName(), "_reply") ||
			strings.HasSuffix(entry.Message.GetMessageName(), "_details") {
			Expect(entry.IsReceived).To(BeTrue())
		} else {
			Expect(entry.IsReceived).To(BeFalse())
		}
		if i%2 == 0 {
			Expect(requestCh2[i/2].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		} else {
			Expect(replyCh2[i/2].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		}
	}
}

func TestTraceClear(t *testing.T) {
	t.Skipf("these randomly fail, see integration tests")

	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	ctx.conn.Trace().Enable(true)

	request := []api.Message{
		&interfaces.CreateLoopback{},
		&memif.MemifCreate{},
	}
	reply := []api.Message{
		&interfaces.CreateLoopbackReply{},
		&memif.MemifCreateReply{},
	}

	for i := 0; i < len(request); i++ {
		ctx.mockVpp.MockReply(reply[i])
		err := ctx.ch.SendRequest(request[i]).ReceiveReply(reply[i])
		Expect(err).To(BeNil())
	}

	traced := ctx.conn.Trace().GetRecords()
	Expect(traced).ToNot(BeNil())
	Expect(traced).To(HaveLen(4))

	ctx.conn.Trace().Clear()
	traced = ctx.conn.Trace().GetRecords()
	Expect(traced).To(BeNil())
	Expect(traced).To(BeEmpty())
}
