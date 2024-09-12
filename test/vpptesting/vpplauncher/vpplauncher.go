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

package vpplauncher

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultVppBinary = "/usr/bin/vpp"
)

const (
	defaultRuntimeDir  = "/run/vpp"
	defaultAPISocket   = "/run/vpp/api.sock"
	defaultCLISocket   = "/run/vpp/cli.sock"
	defaultStatsSocket = "/run/vpp/stats.sock"
	defaultPIDFile     = "/run/vpp/vpp.pid"
)

/*const (
	vppBuildDirRelease   = "build-root/build-vpp-native/vpp/"
	vppBuildDirDebug     = "build-root/build-vpp_debug-native/vpp/"
	vppInstallDirRelease = "build-root/install-vpp-native/vpp/"
	vppInstallDirDebug   = "build-root/install-vpp_debug-native/vpp/"
)*/

const defaultStartupConf = `unix {
	nodaemon
	log /tmp/vpp.log
	cli-listen /run/vpp/cli.sock
	cli-no-pager
	cli-prompt vpp-test
	full-coredump
	coredump-size unlimited
	nosyslog
	gid vpp
	pidfile /run/vpp/vpp.pid
}
api-trace {
	on
}
api-segment {
	gid vpp
}
logging {
	size 1024
	default-log-level debug
}
socksvr {
    socket-name /run/vpp/api.sock
}
statseg {
    socket-name /run/vpp/stats.sock
}
plugins {
    plugin dpdk_plugin.so { disable }
}`

func LaunchVPP() (*VPP, error) {
	opts := Options{}
	vpp, err := NewVPP(opts)
	if err != nil {
		return nil, err
	}
	if vpp.CheckRunning() {
		return nil, fmt.Errorf("already running")
	}
	if err := vpp.Start(); err != nil {
		return nil, err
	}
	return vpp, vpp.WaitStarted()
}

type Options struct {
	Path    string // path to executable
	Config  string // startup config contents
	Pidfile string // path to pidfile
}

func (opts *Options) fillDefaults() {
	if opts.Path == "" {
		opts.Path = defaultVppBinary
	}
	if opts.Config == "" {
		opts.Config = defaultStartupConf
		opts.Pidfile = defaultPIDFile
	}
}

type VPP struct {
	opts Options

	cmd            *exec.Cmd
	stdout, stderr bytes.Buffer
	exit           chan struct{}
	exitErr        error
}

func NewVPP(opts Options) (*VPP, error) {
	p := &VPP{
		opts: opts,
	}
	p.opts.fillDefaults()
	return p, nil
}

func (p *VPP) CheckRunning() bool {
	if p.opts.Pidfile == "" {
		return false
	}
	if _, err := os.Stat(p.opts.Pidfile); err != nil {
		return false
	}
	b, err := os.ReadFile(p.opts.Pidfile)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(string(b))
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Pid == pid
}

const (
	checkPidfileDelay = time.Millisecond * 10
	checkPidfileRetry = 5000
)

func (p *VPP) WaitStarted() error {
	if p.opts.Pidfile == "" {
		return fmt.Errorf("no pidfile defined")
	}
	ch := make(chan error)
	go func() {
		for i := 0; i < checkPidfileRetry; i++ {
			_, err := os.Stat(p.opts.Pidfile)
			if os.IsNotExist(err) {
				time.Sleep(checkPidfileDelay)
				continue
			}
			ch <- err
			return
		}
		ch <- fmt.Errorf("timeout waiting for pidfile")
	}()
	select {
	case err := <-p.OnExit():
		return err
	case err := <-ch:
		return err
	}
}

func (p *VPP) Start() error {
	if p.cmd != nil {
		return fmt.Errorf("already started")
	}

	// start process
	p.cmd = exec.Command(p.opts.Path, p.opts.Config)
	p.cmd.Stderr = &p.stderr
	p.cmd.Stdout = &p.stdout
	p.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,            // set process group ID
		Pdeathsig: syscall.SIGKILL, // send kill to VPP process on exit
	}
	if err := p.cmd.Start(); err != nil {
		return err
	}

	// wait for exit in goroutine
	p.exit = make(chan struct{})
	go func() {
		err := p.cmd.Wait()
		if err != nil {
			var exiterr *exec.ExitError
			if errors.As(err, &exiterr) {
				if strings.Contains(exiterr.Error(), "core dumped") {
					err = fmt.Errorf("VPP crashed (%w) stderr: %s", exiterr, p.stderr.Bytes())
				} else {
					err = fmt.Errorf("VPP exited (%w) stderr: %s", exiterr, p.stderr.Bytes())
				}
			}
		}
		p.exitErr = err
		close(p.exit)
	}()

	return nil
}

func (p *VPP) PID() int {
	if p.cmd == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

func (p *VPP) OnExit() <-chan error {
	ch := make(chan error, 1)
	if p.cmd == nil {
		ch <- fmt.Errorf("not running")
		close(ch)
	} else {
		go func() {
			<-p.exit
			ch <- p.exitErr
			close(ch)
		}()
	}
	return ch
}

const (
	vppExitTimeout = time.Millisecond * 100
)

func (p *VPP) Stop() error {
	if p.cmd == nil {
		return fmt.Errorf("not running")
	}
	defer func() {
		p.cmd = nil
	}()
	if p.exitErr != nil {
		return nil
	}
	// send signal to stop
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("sending TERM signal failed: %w", err)
	}
	// wait until exit or timeout
	select {
	case <-p.exit:
		return p.exitErr
	case <-time.After(vppExitTimeout):
		return p.kill()
	}
}

func (p *VPP) kill() error {
	if err := p.cmd.Process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("sending KILL signal VPP failed: %w", err)
	}
	return nil
}
