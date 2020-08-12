package statsclient

import (
	"context"
	"io"
	"net"
	"os"
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"git.fd.io/govpp.git/adapter"
)

type statsSeg2State int32

const (
	statsSeg2New statsSeg2State = iota
	statsSeg2Connecting
	statsSeg2Connected
	statsSeg2Closed
)

// StatsSeg2 maps stats segment from VPP, MT-safe
type StatsSeg2 struct {
	sockAddr string
	mutex    sync.Mutex
	shmem    []byte
	state    int32
	nclients int
	cancel   func()
}

// NewStatsSeg2 creates a new, initially disconnected, stats segment
func NewStatsSeg2(sockAddr string) *StatsSeg2 {
	return &StatsSeg2{sockAddr: sockAddr}
}

func (seg *StatsSeg2) Connect() error {
	return <-seg.ConnectAsync()
}

func (seg *StatsSeg2) ConnectAsync() <-chan error {
	res := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	seg.mutex.Lock()
	if seg.getState() != statsSeg2New {
		seg.mutex.Unlock()
		res <- adapter.ErrStatsDisconnected
		close(res)
		return res
	}
	seg.setState(statsSeg2Connecting)
	seg.cancel = cancel
	seg.mutex.Unlock()
	go func() {
		res <- seg.doConnect(ctx)
		close(res)
		cancel()
	}()
	return res
}

func (seg *StatsSeg2) doConnect(ctx context.Context) error {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "unixpacket", seg.sockAddr)
	if err != nil {
		return err
	}
	seg.mutex.Lock()
	seg.cancel = func() { conn.Close() }
	if seg.getState() != statsSeg2Connecting {
		seg.cancel()
	}
	seg.mutex.Unlock()
	oob := make([]byte, syscall.CmsgSpace(4)) // space for 1 fd
	_, noob, _, _, err := conn.(*net.UnixConn).ReadMsgUnix(nil, oob)
	conn.Close()
	if err != nil && err.(*net.OpError).Err != io.EOF {
		return err
	}
	var msgs []syscall.SocketControlMessage
	msgs, _ = syscall.ParseSocketControlMessage(oob[:noob])
	if len(msgs) == 0 {
		return adapter.ErrStatsBadServerReply
	}
	fds, _ := syscall.ParseUnixRights(&msgs[0])
	if len(fds) != 1 {
		panic("borked")
	}
	file := os.NewFile(uintptr(fds[0]), "")
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}
	seg.mutex.Lock()
	if seg.getState() == statsSeg2Connecting {
		seg.shmem, err = syscall.Mmap(
			int(file.Fd()), 0, int(info.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
		if err == nil {
			seg.setState(statsSeg2Connected)
		}
	} else {
		err = adapter.ErrStatsDisconnected
	}
	seg.mutex.Unlock()
	file.Close()
	return err
}

func (seg *StatsSeg2) Disconnect() {
	if seg == nil {
		return
	}
	seg.mutex.Lock()
	seg.setState(statsSeg2Closed)
	if seg.cancel != nil {
		seg.cancel()
		seg.cancel = nil
	}
	if seg.nclients == 0 {
		seg.unmapLocked()
	}
	seg.mutex.Unlock()
}

func (seg *StatsSeg2) getState() statsSeg2State {
	return statsSeg2State(atomic.LoadInt32(&seg.state))
}

func (seg *StatsSeg2) setState(state statsSeg2State) {
	atomic.StoreInt32(&seg.state, int32(state))
}

func (seg *StatsSeg2) unmapLocked() {
	if seg.shmem == nil {
		return
	}
	syscall.Munmap(seg.shmem)
	seg.shmem = nil
}

// StatsClient2 provides access into a stats segment. not MT-safe.
// It is safe to disconnect a stats segment while there are open
// clients - shared memory remians mapped until the last client is gone
type StatsClient2 struct {
	seg *StatsSeg2

	statSegment
}

// NewClient makes a new client
func (seg *StatsSeg2) NewClient() (*StatsClient2, error) {
	c := &StatsClient2{seg: seg}
	seg.mutex.Lock()
	if seg.getState() != statsSeg2Connected {
		seg.mutex.Unlock()
		return nil, adapter.ErrStatsDisconnected
	}
	shmem := seg.shmem
	seg.nclients++
	seg.mutex.Unlock()

	err := c.load(shmem)
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func (c *StatsClient2) Close() {
	if c != nil && c.seg != nil {
		c.seg.mutex.Lock()
		c.seg.nclients--
		if c.seg.nclients == 0 && c.seg.getState() == statsSeg2Closed {
			c.seg.unmapLocked()
		}
		c.seg.mutex.Unlock()
		c.seg = nil
		c.sharedHeader = nil
	}
}

type DumpOption func(*dump)

func DumpLimit(n uint32) DumpOption {
	return func(d *dump) { d.limit = n }
}

type dump struct {
	limit uint32
}

func (c *StatsClient2) DumpStats(
	ctx context.Context,
	re *regexp.Regexp,
	options ...DumpOption,
) ([]adapter.StatEntry, error) {
	d := dump{limit: ^uint32(0)}
	for _, opt := range options {
		opt(&d)
	}
	for {
		entries, err := c.dumpStats(d, re)
		if err == nil {
			return entries, nil
		}
		if err != adapter.ErrStatsDataBusy || c.seg.getState() != statsSeg2Connected ||
			ctx.Err() != nil {
			return nil, err
		}
		runtime.Gosched()
	}
}

func (c *StatsClient2) dumpStats(
	d dump,
	re *regexp.Regexp,
) ([]adapter.StatEntry, error) {
	var entries []adapter.StatEntry

	sa := c.accessStart()
	if sa.epoch == 0 {
		return nil, adapter.ErrStatsAccessFailed
	}

	dirVector := c.getStatDirVector()
	vecLen := uint32(vectorLen(dirVector))

	for i := uint32(0); i < vecLen; i++ {
		dirEntry := c.getStatDirIndex(dirVector, i)

		var name []byte
		for n := 0; n < len(dirEntry.name); n++ {
			if dirEntry.name[n] == 0 {
				name = dirEntry.name[:n]
				break
			}
		}
		if len(name) == 0 || re != nil && !re.Match(name) {
			continue
		}

		entry := adapter.StatEntry{
			Name: append([]byte(nil), name...),
			Type: adapter.StatType(dirEntry.directoryType),
			Data: c.copyEntryDataLimit(dirEntry, d.limit),
		}
		entries = append(entries, entry)
	}

	if !c.accessEnd(&sa) {
		return nil, adapter.ErrStatsDataBusy
	}

	return entries, nil
}
