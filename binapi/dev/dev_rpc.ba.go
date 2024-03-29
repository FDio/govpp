// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package dev

import (
	"context"

	api "go.fd.io/govpp/api"
)

// RPCService defines RPC service dev.
type RPCService interface {
	DevAttach(ctx context.Context, in *DevAttach) (*DevAttachReply, error)
	DevCreatePortIf(ctx context.Context, in *DevCreatePortIf) (*DevCreatePortIfReply, error)
	DevDetach(ctx context.Context, in *DevDetach) (*DevDetachReply, error)
	DevRemovePortIf(ctx context.Context, in *DevRemovePortIf) (*DevRemovePortIfReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) DevAttach(ctx context.Context, in *DevAttach) (*DevAttachReply, error) {
	out := new(DevAttachReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) DevCreatePortIf(ctx context.Context, in *DevCreatePortIf) (*DevCreatePortIfReply, error) {
	out := new(DevCreatePortIfReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) DevDetach(ctx context.Context, in *DevDetach) (*DevDetachReply, error) {
	out := new(DevDetachReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) DevRemovePortIf(ctx context.Context, in *DevRemovePortIf) (*DevRemovePortIfReply, error) {
	out := new(DevRemovePortIfReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}
