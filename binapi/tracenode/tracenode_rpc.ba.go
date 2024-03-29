// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package tracenode

import (
	"context"

	api "go.fd.io/govpp/api"
)

// RPCService defines RPC service tracenode.
type RPCService interface {
	TracenodeEnableDisable(ctx context.Context, in *TracenodeEnableDisable) (*TracenodeEnableDisableReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) TracenodeEnableDisable(ctx context.Context, in *TracenodeEnableDisable) (*TracenodeEnableDisableReply, error) {
	out := new(TracenodeEnableDisableReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}
