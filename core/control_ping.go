package core

import (
	"go.fd.io/govpp/api"
	"go.fd.io/govpp/binapi/memclnt"
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

type (
	ControlPing      = memclnt.ControlPing
	ControlPingReply = memclnt.ControlPingReply
)
