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

const traceSize = 10

func TestTraceEnabled(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	trace := core.NewTrace(ctx.conn, traceSize)
	Expect(trace).ToNot(BeNil())
	defer trace.Close()

	request := []api.Message{
		&interfaces.CreateLoopback{},
		&memif.MemifCreate{},
		&l2.BridgeDomainAddDelV2{},
		&ip.IPTableAddDel{},
	}
	reply := []api.Message{
		&interfaces.CreateLoopbackReply{},
		&memif.MemifCreateReply{},
		&l2.BridgeDomainAddDelV2Reply{},
		&ip.IPTableAddDelReply{},
	}

	for i := 0; i < len(request); i++ {
		ctx.mockVpp.MockReply(reply[i])
		err := ctx.ch.SendRequest(request[i]).ReceiveReply(reply[i])
		Expect(err).To(BeNil())
	}
	records := trace.GetRecords()
	Expect(records).ToNot(BeNil())
	Expect(records).To(HaveLen(len(request) + len(reply)))
	for i, entry := range records {
		Expect(entry.Timestamp).ToNot(BeNil())
		Expect(entry.Message.GetMessageName()).ToNot(Equal(""))
		Expect(entry.Succeeded).To(BeTrue())
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
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	trace := core.NewTrace(ctx.conn, traceSize)
	Expect(trace).ToNot(BeNil())
	defer trace.Close()

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

	records := trace.GetRecords()
	Expect(records).ToNot(BeNil())
	Expect(records).To(HaveLen(6))
	for eIdx, entry := range records {
		Expect(entry.Timestamp).ToNot(BeNil())
		Expect(entry.Message.GetMessageName()).ToNot(Equal(""))
		Expect(entry.Succeeded).To(BeTrue())
		if strings.HasSuffix(entry.Message.GetMessageName(), "_reply") ||
			strings.HasSuffix(entry.Message.GetMessageName(), "_details") {
			Expect(entry.IsReceived).To(BeTrue())
		} else {
			Expect(entry.IsReceived).To(BeFalse())
		}
		if eIdx == 0 {
			Expect(request[0].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		} else if eIdx == len(records)-2 {
			msg := memclnt.ControlPing{}
			Expect(msg.GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		} else if eIdx == len(records)-1 {
			msg := memclnt.ControlPingReply{}
			Expect(msg.GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		} else {
			Expect(reply[i-1].GetMessageName()).To(Equal(entry.Message.GetMessageName()))
		}
	}
}

func TestTraceDisabled(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	// do not enable trace

	request := []api.Message{
		&interfaces.CreateLoopback{},
		&memif.MemifCreate{},
		&l2.BridgeDomainAddDelV2{},
		&ip.IPTableAddDel{},
	}
	reply := []api.Message{
		&interfaces.CreateLoopbackReply{},
		&memif.MemifCreateReply{},
		&l2.BridgeDomainAddDelV2Reply{},
		&ip.IPTableAddDelReply{},
	}

	for i := 0; i < len(request); i++ {
		ctx.mockVpp.MockReply(reply[i])
		err := ctx.ch.SendRequest(request[i]).ReceiveReply(reply[i])
		Expect(err).To(BeNil())
	}

	trace := core.NewTrace(ctx.conn, traceSize)
	Expect(trace).ToNot(BeNil())
	defer trace.Close()

	records := trace.GetRecords()
	Expect(records).ToNot(BeNil())
	Expect(records).To(BeEmpty())
}

func TestTracePerChannel(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	trace := core.NewTrace(ctx.conn, traceSize)
	Expect(trace).ToNot(BeNil())
	defer trace.Close()

	ch1 := ctx.ch
	ch2, err := ctx.conn.NewAPIChannel()
	Expect(err).ToNot(HaveOccurred())

	requestCh1 := []api.Message{
		&interfaces.CreateLoopback{},
		&memif.MemifCreate{},
		&l2.BridgeDomainAddDelV2{},
	}
	replyCh1 := []api.Message{
		&interfaces.CreateLoopbackReply{},
		&memif.MemifCreateReply{},
		&l2.BridgeDomainAddDelV2Reply{},
	}
	requestCh2 := []api.Message{
		&ip.IPTableAddDel{},
	}
	replyCh2 := []api.Message{
		&ip.IPTableAddDelReply{},
	}

	for i := 0; i < len(requestCh1); i++ {
		ctx.mockVpp.MockReply(replyCh1[i])
		err = ch1.SendRequest(requestCh1[i]).ReceiveReply(replyCh1[i])
		Expect(err).To(BeNil())
	}
	for i := 0; i < len(requestCh2); i++ {
		ctx.mockVpp.MockReply(replyCh2[i])
		err = ch2.SendRequest(requestCh2[i]).ReceiveReply(replyCh2[i])
		Expect(err).To(BeNil())
	}

	records := trace.GetRecords()
	Expect(records).ToNot(BeNil())
	Expect(records).To(HaveLen(8))

	// per channel
	channel1, ok := ch1.(*core.Channel)
	Expect(ok).To(BeTrue())
	channel2, ok := ch2.(*core.Channel)
	Expect(ok).To(BeTrue())

	recordsCh1 := trace.GetRecordsForChannel(channel1.GetID())
	Expect(recordsCh1).ToNot(BeNil())
	Expect(recordsCh1).To(HaveLen(6))
	for i, entry := range recordsCh1 {
		Expect(entry.Timestamp).ToNot(BeNil())
		Expect(entry.Message.GetMessageName()).ToNot(Equal(""))
		Expect(entry.Succeeded).To(BeTrue())
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

	recordsCh2 := trace.GetRecordsForChannel(channel2.GetID())
	Expect(recordsCh2).ToNot(BeNil())
	Expect(recordsCh2).To(HaveLen(2))
	for i, entry := range recordsCh2 {
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
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	trace := core.NewTrace(ctx.conn, traceSize)
	Expect(trace).ToNot(BeNil())
	defer trace.Close()

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

	records := trace.GetRecords()
	Expect(records).ToNot(BeNil())
	Expect(records).To(HaveLen(4))

	trace.Clear()

	for i := 0; i < len(request); i++ {
		ctx.mockVpp.MockReply(reply[i])
		err := ctx.ch.SendRequest(request[i]).ReceiveReply(reply[i])
		Expect(err).To(BeNil())
	}
	records = trace.GetRecords()
	Expect(records).ToNot(BeNil())
	Expect(records).To(HaveLen(4))

	trace.Clear()

	records = trace.GetRecords()
	Expect(records).To(BeEmpty())
}

func TestTraceUseIfClosed(t *testing.T) {
	ctx := setupTest(t, false)
	defer ctx.teardownTest()

	trace := core.NewTrace(ctx.conn, traceSize)
	Expect(trace).ToNot(BeNil())
	trace.Close()

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

	records := trace.GetRecords()
	Expect(records).ToNot(BeNil())
	Expect(records).To(BeEmpty())
}
