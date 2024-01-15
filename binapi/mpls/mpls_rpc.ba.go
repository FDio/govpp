// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package mpls

import (
	"context"
	"fmt"
	"io"

	api "go.fd.io/govpp/api"
	memclnt "go.fd.io/govpp/binapi/memclnt"
)

// RPCService defines RPC service mpls.
type RPCService interface {
	MplsInterfaceDump(ctx context.Context, in *MplsInterfaceDump) (RPCService_MplsInterfaceDumpClient, error)
	MplsIPBindUnbind(ctx context.Context, in *MplsIPBindUnbind) (*MplsIPBindUnbindReply, error)
	MplsRouteAddDel(ctx context.Context, in *MplsRouteAddDel) (*MplsRouteAddDelReply, error)
	MplsRouteDump(ctx context.Context, in *MplsRouteDump) (RPCService_MplsRouteDumpClient, error)
	MplsTableAddDel(ctx context.Context, in *MplsTableAddDel) (*MplsTableAddDelReply, error)
	MplsTableDump(ctx context.Context, in *MplsTableDump) (RPCService_MplsTableDumpClient, error)
	MplsTunnelAddDel(ctx context.Context, in *MplsTunnelAddDel) (*MplsTunnelAddDelReply, error)
	MplsTunnelDump(ctx context.Context, in *MplsTunnelDump) (RPCService_MplsTunnelDumpClient, error)
	SwInterfaceSetMplsEnable(ctx context.Context, in *SwInterfaceSetMplsEnable) (*SwInterfaceSetMplsEnableReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) MplsInterfaceDump(ctx context.Context, in *MplsInterfaceDump) (RPCService_MplsInterfaceDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_MplsInterfaceDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_MplsInterfaceDumpClient interface {
	Recv() (*MplsInterfaceDetails, error)
	api.Stream
}

type serviceClient_MplsInterfaceDumpClient struct {
	api.Stream
}

func (c *serviceClient_MplsInterfaceDumpClient) Recv() (*MplsInterfaceDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *MplsInterfaceDetails:
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

func (c *serviceClient) MplsIPBindUnbind(ctx context.Context, in *MplsIPBindUnbind) (*MplsIPBindUnbindReply, error) {
	out := new(MplsIPBindUnbindReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) MplsRouteAddDel(ctx context.Context, in *MplsRouteAddDel) (*MplsRouteAddDelReply, error) {
	out := new(MplsRouteAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) MplsRouteDump(ctx context.Context, in *MplsRouteDump) (RPCService_MplsRouteDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_MplsRouteDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_MplsRouteDumpClient interface {
	Recv() (*MplsRouteDetails, error)
	api.Stream
}

type serviceClient_MplsRouteDumpClient struct {
	api.Stream
}

func (c *serviceClient_MplsRouteDumpClient) Recv() (*MplsRouteDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *MplsRouteDetails:
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

func (c *serviceClient) MplsTableAddDel(ctx context.Context, in *MplsTableAddDel) (*MplsTableAddDelReply, error) {
	out := new(MplsTableAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) MplsTableDump(ctx context.Context, in *MplsTableDump) (RPCService_MplsTableDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_MplsTableDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_MplsTableDumpClient interface {
	Recv() (*MplsTableDetails, error)
	api.Stream
}

type serviceClient_MplsTableDumpClient struct {
	api.Stream
}

func (c *serviceClient_MplsTableDumpClient) Recv() (*MplsTableDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *MplsTableDetails:
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

func (c *serviceClient) MplsTunnelAddDel(ctx context.Context, in *MplsTunnelAddDel) (*MplsTunnelAddDelReply, error) {
	out := new(MplsTunnelAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) MplsTunnelDump(ctx context.Context, in *MplsTunnelDump) (RPCService_MplsTunnelDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_MplsTunnelDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_MplsTunnelDumpClient interface {
	Recv() (*MplsTunnelDetails, error)
	api.Stream
}

type serviceClient_MplsTunnelDumpClient struct {
	api.Stream
}

func (c *serviceClient_MplsTunnelDumpClient) Recv() (*MplsTunnelDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *MplsTunnelDetails:
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

func (c *serviceClient) SwInterfaceSetMplsEnable(ctx context.Context, in *SwInterfaceSetMplsEnable) (*SwInterfaceSetMplsEnableReply, error) {
	out := new(SwInterfaceSetMplsEnableReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}
