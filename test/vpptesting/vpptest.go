//  Copyright (c) 2022 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpptesting

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
	"go.fd.io/govpp/binapi/vpe"

	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/adapter/statsclient"
	govppapi "go.fd.io/govpp/api"
	"go.fd.io/govpp/binapi/memclnt"
	"go.fd.io/govpp/binapi/vlib"
	govppcore "go.fd.io/govpp/core"
	"go.fd.io/govpp/test/vpptesting/vpplauncher"
)

var (
	runDir    = "/run/vpp"
	apiSock   = "/run/vpp/api.sock"
	pidFile   = "/run/vpp/vpp.pid"
	statsSock = "/run/vpp/stats.sock"
	//cliSock   = "/run/vpp/cli.sock"
)

const (
	vppConnectDelay      = time.Millisecond * 50
	vppConnectRetryDelay = time.Millisecond * 50
	vppConnectRetryNum   = 3
	vppStopDelay         = time.Millisecond * 50
	vppDisconnectTimeout = time.Millisecond * 50
	vppReplyTimeout      = time.Second * 1
)

type TestCtx struct {
	T          testing.TB
	Context    context.Context
	Cancel     context.CancelFunc
	vppCmd     *vpplauncher.VPP
	Conn       *govppcore.Connection
	statsConn  *govppcore.StatsConnection
	memclntRPC memclnt.RPCService
	vlibRPC    vlib.RPCService
	version    string
}

func SetupVPP(t testing.TB) (tc *TestCtx) {
	// check VPP is not running
	if err := vppAlreadyRunning(); err != nil {
		t.Fatalf("%v", err)
	}

	// remove files from previous run
	removeFile(apiSock)
	removeFile(statsSock)
	removeFile(pidFile)

	// ensure VPP runtime directory exists
	createDir(runDir)

	// start VPP process
	vppCmd, err := vpplauncher.LaunchVPP()
	if err != nil {
		t.Fatalf("starting VPP failed: %v", err)
	}

	defer func() {
		if t.Failed() {
			// if SetupVPP fails we need stop the VPP process
			if err := vppCmd.Stop(); err != nil {
				t.Errorf("stopping VPP failed: %v", err)
			}

		} else {
			// if not, we register a cleanup function to be called when the test completes
			t.Cleanup(tc.TeardownVPP)
		}
	}()

	time.Sleep(vppConnectDelay)

	// connect to binapi
	adapter := socketclient.NewVppClient(apiSock)
	if err := adapter.WaitReady(); err != nil {
		logrus.Warnf("WaitReady error: %v", err)
	}

	var conn *govppcore.Connection
	err = retry(vppConnectRetryNum, func() (err error) {
		govppcore.DefaultReplyTimeout = vppReplyTimeout
		conn, err = govppcore.Connect(adapter)
		return
	})
	if err != nil {
		t.Fatalf("connecting to VPP failed: %v", err)
	}

	memclntRPC := memclnt.NewServiceClient(conn)
	vlibRPC := vlib.NewServiceClient(conn)
	vpeRPC := vpe.NewServiceClient(conn)

	// send ping
	vpeInfo, err := memclntRPC.ControlPing(context.Background(), &memclnt.ControlPing{})
	if err != nil {
		t.Fatalf("getting vpp info failed: %v", err)
	}

	// compare PID
	vppPID := uint32(vppCmd.PID())
	if vpeInfo.VpePID != vppPID {
		t.Fatalf("expected VPP PID to be %v, got %v", vppPID, vpeInfo.VpePID)
	}

	// get version
	versionInfo, err := vpeRPC.ShowVersion(context.Background(), &vpe.ShowVersion{})
	if err != nil {
		t.Fatalf("getting vpp version failed: %v", err)
	}

	go func() {
		q := make(chan struct{})
		t.Cleanup(func() {
			close(q)
		})
		select {
		case <-q:
			// do no wait after test
		case exitErr := <-vppCmd.OnExit():
			if exitErr != nil {
				t.Errorf("VPP process exited with error: %v", exitErr)
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	return &TestCtx{
		T:          t,
		Context:    ctx,
		Cancel:     cancel,
		vppCmd:     vppCmd,
		Conn:       conn,
		memclntRPC: memclntRPC,
		vlibRPC:    vlibRPC,
		version:    versionInfo.Version,
	}
}

func vppAlreadyRunning() error {
	processes, err := ps.Processes()
	if err != nil {
		return fmt.Errorf("listing processes failed: %v", err)
	}
	for _, process := range processes {
		proc := process.Executable()
		switch proc {
		case "vpp", "vpp_main":
			return fmt.Errorf("VPP is already running (PID: %v)", process.Pid())
		}
	}
	return nil
}

func (ctx *TestCtx) VPPVersion() string {
	return ctx.version
}

func (ctx *TestCtx) StatsConn() *govppcore.StatsConnection {
	if ctx.statsConn != nil {
		return ctx.statsConn
	}
	statsClient := statsclient.NewStatsClient(statsSock)
	var err error
	ctx.statsConn, err = govppcore.ConnectStats(statsClient)
	if err != nil {
		ctx.T.Fatalf("connecting to stats failed: %v", err)
	}
	return ctx.statsConn
}

func (ctx *TestCtx) TeardownVPP() {
	// disconnect sometimes hangs
	disconnected := make(chan struct{})
	go func() {
		if ctx.statsConn != nil {
			ctx.statsConn.Disconnect()
			ctx.statsConn = nil
		}
		ctx.Conn.Disconnect()
		close(disconnected)
	}()
	// wait until disconnected or timeout
	select {
	case <-disconnected:
		time.Sleep(vppStopDelay)
	case <-time.After(vppDisconnectTimeout):
		logrus.Debugf("VPP disconnect timeout")
	}

	if err := ctx.vppCmd.Stop(); err != nil {
		ctx.T.Logf("stopping VPP failed: %v", err)
	} else {
		ctx.T.Logf("VPP stopped")
	}
}

func (ctx *TestCtx) MustCli(cmd ...string) {
	for _, c := range cmd {
		if _, err := ctx.RunCli(c); err != nil {
			ctx.T.Fatal(err)
		}
	}
}

func (ctx *TestCtx) RunCli(cmd string) (string, error) {
	ctx.T.Helper()
	logrus.Debugf("RunCli: %q", cmd)
	reply, err := ctx.vlibRPC.CliInband(context.Background(), &vlib.CliInband{
		Cmd: cmd,
	})
	if err != nil {
		return "", fmt.Errorf("CLI '%v' failed: %v", cmd, err)
	}
	if strings.TrimSpace(reply.Reply) != "" {
		logrus.Debugf(" cli reply: %s", reply.Reply)
	}
	if err := govppapi.RetvalToVPPApiError(reply.Retval); err != nil {
		return "", fmt.Errorf("CLI '%v' error: %v", cmd, err)
	}
	return reply.Reply, nil
}

func (ctx *TestCtx) RunCmd(cmd string, args ...string) (string, error) {
	ctx.T.Helper()
	stdout, stderr, err := ctx.execCmd(cmd, args...)
	if err != nil {
		return "", fmt.Errorf("command failed: %v\n%s", err, stderr)
	}
	return stdout, nil
}

func (ctx *TestCtx) execCmd(cmd string, args ...string) (string, string, error) {
	logrus.Debugf("exec: '%s %s'", cmd, strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if strings.TrimSpace(stdout.String()) != "" {
		logrus.Debugf(" stdout:\n%s", stdout.String())
	}
	if strings.TrimSpace(stderr.String()) != "" {
		logrus.Debugf(" stderr:\n%s", stderr.String())
	}
	return stdout.String(), stderr.String(), err
}

func retry(retries int, fn func() error) (err error) {
	for i := 1; true; i++ {
		if err = fn(); err == nil {
			logrus.Debugf("retry attempt #%d succeeded", i)
			return nil
		} else if i >= retries {
			break
		}
		logrus.Debugf("retry attempt #%d failed: %v, retrying in %v", i, err, vppConnectRetryDelay)
		time.Sleep(vppConnectRetryDelay)
	}
	return fmt.Errorf("%d retry attempts failed: %w", retries, err)
}

func createDir(dir string) {
	if err := os.Mkdir(dir, 0655); err == nil {
		logrus.Debugf("created dir %s", dir)
	} else if !os.IsExist(err) {
		logrus.Warnf("creating dir %s failed: %v", dir, err)
	}
}

func removeFile(path string) {
	if err := os.Remove(path); err == nil {
		logrus.Debugf("removed file %s", path)
	} else if !os.IsNotExist(err) {
		logrus.Warnf("removing file %s failed: %v", path, err)
	}
}
