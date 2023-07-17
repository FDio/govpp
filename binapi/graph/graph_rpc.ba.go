// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package graph

import (
	"context"
	"fmt"
	"io"

	api "go.fd.io/govpp/api"
)

// RPCService defines RPC service graph.
type RPCService interface {
	GraphNodeGet(ctx context.Context, in *GraphNodeGet) (RPCService_GraphNodeGetClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) GraphNodeGet(ctx context.Context, in *GraphNodeGet) (RPCService_GraphNodeGetClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_GraphNodeGetClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_GraphNodeGetClient interface {
	Recv() (*GraphNodeDetails, *GraphNodeGetReply, error)
	api.Stream
}

type serviceClient_GraphNodeGetClient struct {
	api.Stream
}

func (c *serviceClient_GraphNodeGetClient) Recv() (*GraphNodeDetails, *GraphNodeGetReply, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, nil, err
	}
	switch m := msg.(type) {
	case *GraphNodeDetails:
		return m, nil, nil
	case *GraphNodeGetReply:
		if err := api.RetvalToVPPApiError(m.Retval); err != nil {
			return nil, m, err
		}
		err = c.Stream.Close()
		if err != nil {
			return nil, m, err
		}
		return nil, m, io.EOF
	default:
		return nil, nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}
