package core

import (
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/codec"
)

var (
	msgControlPing      api.Message = new(ControlPing)
	msgControlPingReply api.Message = new(ControlPingReply)
)

// SetControlPing sets the control ping message used by core.
func SetControlPing(m api.Message) {
	msgControlPing = m
}

// SetControlPingReply sets the control ping reply message used by core.
func SetControlPingReply(m api.Message) {
	msgControlPingReply = m
}

// Control ping from client to api server request
// ControlPing defines message 'control_ping'.
type ControlPing struct{}

func (m *ControlPing) Reset()               { *m = ControlPing{} }
func (*ControlPing) GetMessageName() string { return "control_ping" }
func (*ControlPing) GetCrcString() string   { return "51077d14" }
func (*ControlPing) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *ControlPing) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *ControlPing) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *ControlPing) Unmarshal(b []byte) error {
	return nil
}

// Control ping from the client to the server response
//   - retval - return code for the request
//   - vpe_pid - the pid of the vpe, returned by the server
//
// ControlPingReply defines message 'control_ping_reply'.
type ControlPingReply struct {
	Retval      int32  `binapi:"i32,name=retval" json:"retval,omitempty"`
	ClientIndex uint32 `binapi:"u32,name=client_index" json:"client_index,omitempty"`
	VpePID      uint32 `binapi:"u32,name=vpe_pid" json:"vpe_pid,omitempty"`
}

func (m *ControlPingReply) Reset()               { *m = ControlPingReply{} }
func (*ControlPingReply) GetMessageName() string { return "control_ping_reply" }
func (*ControlPingReply) GetCrcString() string   { return "f6b0b8ca" }
func (*ControlPingReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *ControlPingReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.ClientIndex
	size += 4 // m.VpePID
	return size
}
func (m *ControlPingReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(m.ClientIndex)
	buf.EncodeUint32(m.VpePID)
	return buf.Bytes(), nil
}
func (m *ControlPingReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.ClientIndex = buf.DecodeUint32()
	m.VpePID = buf.DecodeUint32()
	return nil
}
