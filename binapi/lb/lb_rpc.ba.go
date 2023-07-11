// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package lb

import (
	"context"
	"fmt"
	"io"

	api "go.fd.io/govpp/api"
	memclnt "go.fd.io/govpp/binapi/memclnt"
)

// RPCService defines RPC service lb.
type RPCService interface {
	LbAddDelAs(ctx context.Context, in *LbAddDelAs) (*LbAddDelAsReply, error)
	LbAddDelIntfNat4(ctx context.Context, in *LbAddDelIntfNat4) (*LbAddDelIntfNat4Reply, error)
	LbAddDelIntfNat6(ctx context.Context, in *LbAddDelIntfNat6) (*LbAddDelIntfNat6Reply, error)
	LbAddDelVip(ctx context.Context, in *LbAddDelVip) (*LbAddDelVipReply, error)
	LbAddDelVipV2(ctx context.Context, in *LbAddDelVipV2) (*LbAddDelVipV2Reply, error)
	LbAsDump(ctx context.Context, in *LbAsDump) (RPCService_LbAsDumpClient, error)
	LbConf(ctx context.Context, in *LbConf) (*LbConfReply, error)
	LbFlushVip(ctx context.Context, in *LbFlushVip) (*LbFlushVipReply, error)
	LbVipDump(ctx context.Context, in *LbVipDump) (RPCService_LbVipDumpClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) LbAddDelAs(ctx context.Context, in *LbAddDelAs) (*LbAddDelAsReply, error) {
	out := new(LbAddDelAsReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) LbAddDelIntfNat4(ctx context.Context, in *LbAddDelIntfNat4) (*LbAddDelIntfNat4Reply, error) {
	out := new(LbAddDelIntfNat4Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) LbAddDelIntfNat6(ctx context.Context, in *LbAddDelIntfNat6) (*LbAddDelIntfNat6Reply, error) {
	out := new(LbAddDelIntfNat6Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) LbAddDelVip(ctx context.Context, in *LbAddDelVip) (*LbAddDelVipReply, error) {
	out := new(LbAddDelVipReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) LbAddDelVipV2(ctx context.Context, in *LbAddDelVipV2) (*LbAddDelVipV2Reply, error) {
	out := new(LbAddDelVipV2Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) LbAsDump(ctx context.Context, in *LbAsDump) (RPCService_LbAsDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_LbAsDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_LbAsDumpClient interface {
	Recv() (*LbAsDetails, error)
	api.Stream
}

type serviceClient_LbAsDumpClient struct {
	api.Stream
}

func (c *serviceClient_LbAsDumpClient) Recv() (*LbAsDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *LbAsDetails:
		return m, nil
	case *memclnt.ControlPingReply:
		err = c.Stream.Close()
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}

func (c *serviceClient) LbConf(ctx context.Context, in *LbConf) (*LbConfReply, error) {
	out := new(LbConfReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) LbFlushVip(ctx context.Context, in *LbFlushVip) (*LbFlushVipReply, error) {
	out := new(LbFlushVipReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) LbVipDump(ctx context.Context, in *LbVipDump) (RPCService_LbVipDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_LbVipDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_LbVipDumpClient interface {
	Recv() (*LbVipDetails, error)
	api.Stream
}

type serviceClient_LbVipDumpClient struct {
	api.Stream
}

func (c *serviceClient_LbVipDumpClient) Recv() (*LbVipDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *LbVipDetails:
		return m, nil
	case *memclnt.ControlPingReply:
		err = c.Stream.Close()
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}
