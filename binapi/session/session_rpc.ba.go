// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package session

import (
	"context"
	"fmt"
	"io"

	api "go.fd.io/govpp/api"
	memclnt "go.fd.io/govpp/binapi/memclnt"
)

// RPCService defines RPC service session.
type RPCService interface {
	AppAddCertKeyPair(ctx context.Context, in *AppAddCertKeyPair) (*AppAddCertKeyPairReply, error)
	AppAttach(ctx context.Context, in *AppAttach) (*AppAttachReply, error)
	AppDelCertKeyPair(ctx context.Context, in *AppDelCertKeyPair) (*AppDelCertKeyPairReply, error)
	AppNamespaceAddDel(ctx context.Context, in *AppNamespaceAddDel) (*AppNamespaceAddDelReply, error)
	AppNamespaceAddDelV2(ctx context.Context, in *AppNamespaceAddDelV2) (*AppNamespaceAddDelV2Reply, error)
	AppNamespaceAddDelV3(ctx context.Context, in *AppNamespaceAddDelV3) (*AppNamespaceAddDelV3Reply, error)
	AppNamespaceAddDelV4(ctx context.Context, in *AppNamespaceAddDelV4) (*AppNamespaceAddDelV4Reply, error)
	AppWorkerAddDel(ctx context.Context, in *AppWorkerAddDel) (*AppWorkerAddDelReply, error)
	ApplicationDetach(ctx context.Context, in *ApplicationDetach) (*ApplicationDetachReply, error)
	SessionEnableDisable(ctx context.Context, in *SessionEnableDisable) (*SessionEnableDisableReply, error)
	SessionEnableDisableV2(ctx context.Context, in *SessionEnableDisableV2) (*SessionEnableDisableV2Reply, error)
	SessionRuleAddDel(ctx context.Context, in *SessionRuleAddDel) (*SessionRuleAddDelReply, error)
	SessionRulesDump(ctx context.Context, in *SessionRulesDump) (RPCService_SessionRulesDumpClient, error)
	SessionRulesV2Dump(ctx context.Context, in *SessionRulesV2Dump) (RPCService_SessionRulesV2DumpClient, error)
	SessionSapiEnableDisable(ctx context.Context, in *SessionSapiEnableDisable) (*SessionSapiEnableDisableReply, error)
	SessionSdlAddDel(ctx context.Context, in *SessionSdlAddDel) (*SessionSdlAddDelReply, error)
	SessionSdlAddDelV2(ctx context.Context, in *SessionSdlAddDelV2) (*SessionSdlAddDelV2Reply, error)
	SessionSdlDump(ctx context.Context, in *SessionSdlDump) (RPCService_SessionSdlDumpClient, error)
	SessionSdlV2Dump(ctx context.Context, in *SessionSdlV2Dump) (RPCService_SessionSdlV2DumpClient, error)
	SessionSdlV3Dump(ctx context.Context, in *SessionSdlV3Dump) (RPCService_SessionSdlV3DumpClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) AppAddCertKeyPair(ctx context.Context, in *AppAddCertKeyPair) (*AppAddCertKeyPairReply, error) {
	out := new(AppAddCertKeyPairReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) AppAttach(ctx context.Context, in *AppAttach) (*AppAttachReply, error) {
	out := new(AppAttachReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) AppDelCertKeyPair(ctx context.Context, in *AppDelCertKeyPair) (*AppDelCertKeyPairReply, error) {
	out := new(AppDelCertKeyPairReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) AppNamespaceAddDel(ctx context.Context, in *AppNamespaceAddDel) (*AppNamespaceAddDelReply, error) {
	out := new(AppNamespaceAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) AppNamespaceAddDelV2(ctx context.Context, in *AppNamespaceAddDelV2) (*AppNamespaceAddDelV2Reply, error) {
	out := new(AppNamespaceAddDelV2Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) AppNamespaceAddDelV3(ctx context.Context, in *AppNamespaceAddDelV3) (*AppNamespaceAddDelV3Reply, error) {
	out := new(AppNamespaceAddDelV3Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) AppNamespaceAddDelV4(ctx context.Context, in *AppNamespaceAddDelV4) (*AppNamespaceAddDelV4Reply, error) {
	out := new(AppNamespaceAddDelV4Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) AppWorkerAddDel(ctx context.Context, in *AppWorkerAddDel) (*AppWorkerAddDelReply, error) {
	out := new(AppWorkerAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) ApplicationDetach(ctx context.Context, in *ApplicationDetach) (*ApplicationDetachReply, error) {
	out := new(ApplicationDetachReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) SessionEnableDisable(ctx context.Context, in *SessionEnableDisable) (*SessionEnableDisableReply, error) {
	out := new(SessionEnableDisableReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) SessionEnableDisableV2(ctx context.Context, in *SessionEnableDisableV2) (*SessionEnableDisableV2Reply, error) {
	out := new(SessionEnableDisableV2Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) SessionRuleAddDel(ctx context.Context, in *SessionRuleAddDel) (*SessionRuleAddDelReply, error) {
	out := new(SessionRuleAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) SessionRulesDump(ctx context.Context, in *SessionRulesDump) (RPCService_SessionRulesDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_SessionRulesDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_SessionRulesDumpClient interface {
	Recv() (*SessionRulesDetails, error)
	api.Stream
}

type serviceClient_SessionRulesDumpClient struct {
	api.Stream
}

func (c *serviceClient_SessionRulesDumpClient) Recv() (*SessionRulesDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *SessionRulesDetails:
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

func (c *serviceClient) SessionRulesV2Dump(ctx context.Context, in *SessionRulesV2Dump) (RPCService_SessionRulesV2DumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_SessionRulesV2DumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_SessionRulesV2DumpClient interface {
	Recv() (*SessionRulesV2Details, error)
	api.Stream
}

type serviceClient_SessionRulesV2DumpClient struct {
	api.Stream
}

func (c *serviceClient_SessionRulesV2DumpClient) Recv() (*SessionRulesV2Details, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *SessionRulesV2Details:
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

func (c *serviceClient) SessionSapiEnableDisable(ctx context.Context, in *SessionSapiEnableDisable) (*SessionSapiEnableDisableReply, error) {
	out := new(SessionSapiEnableDisableReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) SessionSdlAddDel(ctx context.Context, in *SessionSdlAddDel) (*SessionSdlAddDelReply, error) {
	out := new(SessionSdlAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) SessionSdlAddDelV2(ctx context.Context, in *SessionSdlAddDelV2) (*SessionSdlAddDelV2Reply, error) {
	out := new(SessionSdlAddDelV2Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) SessionSdlDump(ctx context.Context, in *SessionSdlDump) (RPCService_SessionSdlDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_SessionSdlDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_SessionSdlDumpClient interface {
	Recv() (*SessionSdlDetails, error)
	api.Stream
}

type serviceClient_SessionSdlDumpClient struct {
	api.Stream
}

func (c *serviceClient_SessionSdlDumpClient) Recv() (*SessionSdlDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *SessionSdlDetails:
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

func (c *serviceClient) SessionSdlV2Dump(ctx context.Context, in *SessionSdlV2Dump) (RPCService_SessionSdlV2DumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_SessionSdlV2DumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_SessionSdlV2DumpClient interface {
	Recv() (*SessionSdlV2Details, error)
	api.Stream
}

type serviceClient_SessionSdlV2DumpClient struct {
	api.Stream
}

func (c *serviceClient_SessionSdlV2DumpClient) Recv() (*SessionSdlV2Details, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *SessionSdlV2Details:
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

func (c *serviceClient) SessionSdlV3Dump(ctx context.Context, in *SessionSdlV3Dump) (RPCService_SessionSdlV3DumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_SessionSdlV3DumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_SessionSdlV3DumpClient interface {
	Recv() (*SessionSdlV3Details, error)
	api.Stream
}

type serviceClient_SessionSdlV3DumpClient struct {
	api.Stream
}

func (c *serviceClient_SessionSdlV3DumpClient) Recv() (*SessionSdlV3Details, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *SessionSdlV3Details:
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
