// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package pppoe

import (
	"context"
	"fmt"
	"io"

	api "git.fd.io/govpp.git/api"
	vpe "git.fd.io/govpp.git/binapi/vpe"
)

// RPCService defines RPC service pppoe.
type RPCService interface {
	PppoeAddDelCp(ctx context.Context, in *PppoeAddDelCp) (*PppoeAddDelCpReply, error)
	PppoeAddDelSession(ctx context.Context, in *PppoeAddDelSession) (*PppoeAddDelSessionReply, error)
	PppoeSessionDump(ctx context.Context, in *PppoeSessionDump) (RPCService_PppoeSessionDumpClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) PppoeAddDelCp(ctx context.Context, in *PppoeAddDelCp) (*PppoeAddDelCpReply, error) {
	out := new(PppoeAddDelCpReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) PppoeAddDelSession(ctx context.Context, in *PppoeAddDelSession) (*PppoeAddDelSessionReply, error) {
	out := new(PppoeAddDelSessionReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) PppoeSessionDump(ctx context.Context, in *PppoeSessionDump) (RPCService_PppoeSessionDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_PppoeSessionDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_PppoeSessionDumpClient interface {
	Recv() (*PppoeSessionDetails, error)
	api.Stream
}

type serviceClient_PppoeSessionDumpClient struct {
	api.Stream
}

func (c *serviceClient_PppoeSessionDumpClient) Recv() (*PppoeSessionDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *PppoeSessionDetails:
		return m, nil
	case *vpe.ControlPingReply:
		err = c.Stream.Close()
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}
