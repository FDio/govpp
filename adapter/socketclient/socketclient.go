// Copyright (c) 2019 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package socketclient

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"

	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/binapi/memclnt"
	"go.fd.io/govpp/codec"
)

const (
	// DefaultSocketName is default VPP API socket file path.
	DefaultSocketName = "/run/vpp/api.sock"
	// DefaultClientName is used for identifying client in socket registration
	DefaultClientName = "govppsock"
)

var (

	// DefaultConnectTimeout is default timeout for connecting
	DefaultConnectTimeout = time.Second * 3
	// DefaultDisconnectTimeout is default timeout for discconnecting
	DefaultDisconnectTimeout = time.Millisecond * 100
	// MaxWaitReady defines maximum duration of waiting for socket file
	MaxWaitReady = time.Second * 3
)

var (
	debug       = strings.Contains(os.Getenv("DEBUG_GOVPP"), "socketclient")
	debugMsgIds = strings.Contains(os.Getenv("DEBUG_GOVPP"), "msgtable")

	log logrus.FieldLogger
)

// SetLogger sets global logger.
func SetLogger(logger logrus.FieldLogger) {
	log = logger
}

func init() {
	logger := logrus.New()
	if debug {
		logger.Level = logrus.DebugLevel
		logger.Debug("govpp: debug level enabled for socketclient")
	}
	log = logger.WithField("logger", "govpp/socketclient")
}

type Client struct {
	socketPath string
	clientName string

	conn   *net.UnixConn
	reader *bufio.Reader
	writer *bufio.Writer

	connectTimeout    time.Duration
	disconnectTimeout time.Duration

	msgCallback  adapter.MsgCallback
	clientIndex  uint32
	msgTable     map[string]uint16
	msgTableMu   sync.RWMutex
	sockDelMsgId uint16
	writeMu      sync.Mutex

	headerPool *sync.Pool

	quit chan struct{}
	wg   sync.WaitGroup
}

// NewVppClient returns a new Client using socket.
// If socket is empty string DefaultSocketName is used.
func NewVppClient(socket string) *Client {
	if socket == "" {
		socket = DefaultSocketName
	}
	return &Client{
		socketPath:        socket,
		clientName:        DefaultClientName,
		connectTimeout:    DefaultConnectTimeout,
		disconnectTimeout: DefaultDisconnectTimeout,
		headerPool: &sync.Pool{New: func() interface{} {
			x := make([]byte, 16)
			return &x
		}},
		msgCallback: func(msgID uint16, data []byte) {
			log.Debugf("no callback set, dropping message: ID=%v len=%d", msgID, len(data))
		},
	}
}

// SetClientName sets a client name used for identification.
func (c *Client) SetClientName(name string) {
	c.clientName = name
}

// SetConnectTimeout sets timeout used during connecting.
func (c *Client) SetConnectTimeout(t time.Duration) {
	c.connectTimeout = t
}

// SetDisconnectTimeout sets timeout used during disconnecting.
func (c *Client) SetDisconnectTimeout(t time.Duration) {
	c.disconnectTimeout = t
}

// SetMsgCallback sets the callback for incoming messages.
func (c *Client) SetMsgCallback(cb adapter.MsgCallback) {
	log.Debug("SetMsgCallback")
	c.msgCallback = cb
}

// WaitReady checks if the socket file exists and if it does not exist waits for
// it for the duration defined by MaxWaitReady.
func (c *Client) WaitReady() error {
	socketDir, _ := filepath.Split(c.socketPath)
	dirChain := strings.Split(filepath.ToSlash(filepath.Clean(socketDir)), "/")

	dir := "/"
	for _, dirElem := range dirChain {
		dir = filepath.Join(dir, dirElem)
		if err := waitForDir(dir); err != nil {
			return err
		}
		log.Debugf("dir ready: %v", dir)
	}

	// check if socket already exists
	if _, err := os.Stat(c.socketPath); err == nil {
		return nil // socket exists, we are ready
	} else if !errors.Is(err, fs.ErrNotExist) {
		log.Debugf("error is: %+v", err)
		return err // some other error occurred
	}

	log.Debugf("waiting for file: %v", c.socketPath)

	// socket does not exist, watch for it
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Debugf("failed to close file watcher: %v", err)
		}
	}()

	// start directory watcher
	d := filepath.Dir(c.socketPath)
	if err := watcher.Add(d); err != nil {
		log.Debugf("watcher add(%v) error: %v", d, err)
		return err
	}

	timeout := time.NewTimer(MaxWaitReady)
	for {
		select {
		case <-timeout.C:
			log.Debugf("watcher timeout after: %v", MaxWaitReady)
			return fmt.Errorf("timeout waiting (%s) for socket file: %s", MaxWaitReady, c.socketPath)

		case e := <-watcher.Errors:
			log.Debugf("watcher error: %+v", e)
			return e

		case ev := <-watcher.Events:
			log.Debugf("watcher event: %+v", ev)
			if ev.Name == c.socketPath && (ev.Op&fsnotify.Create) == fsnotify.Create {
				// socket created, we are ready
				return nil
			}
		}
	}
}

func waitForDir(dir string) error {
	// check if dir already exists
	if _, err := os.Stat(dir); err == nil {
		return nil // dir exists, we are ready
	} else if !errors.Is(err, fs.ErrNotExist) {
		log.Debugf("error is: %+v", err)
		return err // some other error occurred
	}

	log.Debugf("waiting for dir: %v", dir)

	// dir does not exist, watch for it
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Debugf("failed to close file watcher: %v", err)
		}
	}()

	// start watching directory
	d := filepath.Dir(dir)
	if err := watcher.Add(d); err != nil {
		log.Debugf("watcher add (%v) error: %v", d, err)
		return err
	}

	timeout := time.NewTimer(MaxWaitReady)
	for {
		select {
		case <-timeout.C:
			log.Debugf("watcher timeout after: %v", MaxWaitReady)
			return fmt.Errorf("timeout waiting (%s) for directory: %s", MaxWaitReady, dir)

		case e := <-watcher.Errors:
			log.Debugf("watcher error: %+v", e)
			return e

		case ev := <-watcher.Events:
			log.Debugf("watcher event: %+v", ev)
			if ev.Name == dir && (ev.Op&fsnotify.Create) == fsnotify.Create {
				// socket created, we are ready
				return nil
			}
		}
	}
}

func (c *Client) Connect() error {
	// check if socket exists
	if _, err := os.Stat(c.socketPath); os.IsNotExist(err) {
		return fmt.Errorf("VPP API socket file %s does not exist", c.socketPath)
	} else if err != nil {
		return fmt.Errorf("VPP API socket error: %v", err)
	}

	if err := c.connect(c.socketPath); err != nil {
		return err
	}

	if err := c.open(); err != nil {
		_ = c.disconnect()
		return err
	}

	c.quit = make(chan struct{})
	c.wg.Add(1)
	go c.readerLoop()

	return nil
}

func (c *Client) Disconnect() error {
	if c.conn == nil {
		return nil
	}
	log.Debugf("Disconnecting..")

	close(c.quit)

	if err := c.conn.CloseRead(); err != nil {
		log.Debugf("closing readMsg failed: %v", err)
	}

	// wait for readerLoop to return
	c.wg.Wait()

	// Don't bother sending a vl_api_sockclnt_delete_t message,
	// just close the socket.
	if err := c.disconnect(); err != nil {
		return err
	}

	return nil
}

const defaultBufferSize = 4096

func (c *Client) connect(sockAddr string) error {
	addr := &net.UnixAddr{Name: sockAddr, Net: "unix"}

	log.Debugf("Connecting to: %v", c.socketPath)

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		// we try different type of socket for backwards compatbility with VPP<=19.04
		if strings.Contains(err.Error(), "wrong type for socket") {
			addr.Net = "unixpacket"
			log.Debugf("%s, retrying connect with type unixpacket", err)
			conn, err = net.DialUnix("unixpacket", nil, addr)
		}
		if err != nil {
			log.Debugf("Connecting to socket %s failed: %s", addr, err)
			return err
		}
	}

	c.conn = conn
	log.Debugf("Connected to socket (local addr: %v)", c.conn.LocalAddr().(*net.UnixAddr))

	c.reader = bufio.NewReaderSize(c.conn, defaultBufferSize)
	c.writer = bufio.NewWriterSize(c.conn, defaultBufferSize)

	return nil
}

func (c *Client) disconnect() error {
	log.Debugf("Closing socket")

	// cleanup msg table
	c.setMsgTable(make(map[string]uint16), 0)

	if err := c.conn.Close(); err != nil {
		log.Debugln("Closing socket failed:", err)
		return err
	}
	return nil
}

const (
	sockCreateMsgId  = 15 // hard-coded sockclnt_create message ID
	createMsgContext = byte(123)
	deleteMsgContext = byte(124)
)

func (c *Client) open() error {
	var msgCodec = codec.DefaultCodec

	// Request socket client create
	req := &memclnt.SockclntCreate{
		Name: c.clientName,
	}
	msg, err := msgCodec.EncodeMsg(req, sockCreateMsgId)
	if err != nil {
		log.Debugln("Encode  error:", err)
		return err
	}
	// set non-0 context
	msg[5] = createMsgContext

	if err := c.writeMsg(msg); err != nil {
		log.Debugln("Write error: ", err)
		return err
	}
	msgReply, err := c.readMsgTimeout(nil, c.connectTimeout)
	if err != nil {
		log.Println("Read error:", err)
		return err
	}

	reply := new(memclnt.SockclntCreateReply)
	if err := msgCodec.DecodeMsg(msgReply, reply); err != nil {
		log.Println("Decoding sockclnt_create_reply failed:", err)
		return err
	} else if reply.Response != 0 {
		return fmt.Errorf("sockclnt_create_reply: response error (%d)", reply.Response)
	}

	log.Debugf("SockclntCreateReply: Response=%v Index=%v Count=%v",
		reply.Response, reply.Index, reply.Count)

	c.clientIndex = reply.Index
	msgTable := make(map[string]uint16, reply.Count)
	var sockDelMsgId uint16
	for _, x := range reply.MessageTable {
		msgName := strings.Split(x.Name, "\x00")[0]
		name := strings.TrimSuffix(msgName, "\x13")
		msgTable[name] = x.Index
		if strings.HasPrefix(name, "sockclnt_delete_") {
			sockDelMsgId = x.Index
		}
		if debugMsgIds {
			log.Debugf(" - %4d: %q", x.Index, name)
		}
	}
	c.setMsgTable(msgTable, sockDelMsgId)

	return nil
}

func (c *Client) setMsgTable(msgTable map[string]uint16, sockDelMsgId uint16) {
	c.msgTableMu.Lock()
	defer c.msgTableMu.Unlock()

	c.msgTable = msgTable
	c.sockDelMsgId = sockDelMsgId
}

func (c *Client) GetMsgID(msgName string, msgCrc string) (uint16, error) {
	c.msgTableMu.RLock()
	defer c.msgTableMu.RUnlock()

	if msgID, ok := c.msgTable[msgName+"_"+msgCrc]; ok {
		return msgID, nil
	}
	return 0, &adapter.UnknownMsgError{
		MsgName: msgName,
		MsgCrc:  msgCrc,
	}
}

func (c *Client) SendMsg(context uint32, data []byte) error {
	if len(data) < 10 {
		return fmt.Errorf("invalid message data, length must be at least 10 bytes")
	}
	setMsgRequestHeader(data, c.clientIndex, context)

	if debug {
		log.Debugf("sendMsg (%d) context=%v client=%d: % 02X", len(data), context, c.clientIndex, data)
	}

	if err := c.writeMsg(data); err != nil {
		log.Debugln("writeMsg error: ", err)
		return err
	}

	return nil
}

// setMsgRequestHeader sets client index and context in the message request header
//
// Message request has following structure:
//
//	type msgRequestHeader struct {
//	    MsgID       uint16
//	    ClientIndex uint32
//	    Context     uint32
//	}
func setMsgRequestHeader(data []byte, clientIndex, context uint32) {
	// message ID is already set
	binary.BigEndian.PutUint32(data[2:6], clientIndex)
	binary.BigEndian.PutUint32(data[6:10], context)
}

func (c *Client) writeMsg(msg []byte) error {
	// we lock to prevent mixing multiple message writes
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	header, ok := c.headerPool.Get().(*[]byte)
	if !ok {
		return fmt.Errorf("failed to get header from pool")
	}
	err := writeMsgHeader(c.writer, *header, len(msg))
	if err != nil {
		return err
	}
	c.headerPool.Put(header)

	if err := writeMsgData(c.writer, msg, c.writer.Size()); err != nil {
		return err
	}

	if err := c.writer.Flush(); err != nil {
		return err
	}

	log.Debugf(" -- writeMsg done")

	return nil
}

func writeMsgHeader(w io.Writer, header []byte, dataLen int) error {
	binary.BigEndian.PutUint32(header[8:12], uint32(dataLen))

	n, err := w.Write(header)
	if err != nil {
		return err
	}
	if debug {
		log.Debugf(" - header sent (%d/%d): % 0X", n, len(header), header)
	}

	return nil
}

func writeMsgData(w io.Writer, msg []byte, writerSize int) error {
	for i := 0; i <= len(msg)/writerSize; i++ {
		x := i*writerSize + writerSize
		if x > len(msg) {
			x = len(msg)
		}
		if debug {
			log.Debugf(" - x=%v i=%v len=%v mod=%v", x, i, len(msg), len(msg)/writerSize)
		}
		n, err := w.Write(msg[i*writerSize : x])
		if err != nil {
			return err
		}
		if debug {
			log.Debugf(" - data sent x=%d (%d/%d): % 0X", x, n, len(msg), msg)
		}
	}
	return nil
}

func (c *Client) readerLoop() {
	defer c.wg.Done()
	defer log.Debugf("reader loop done")

	var buf [8192]byte

	for {
		select {
		case <-c.quit:
			return
		default:
		}

		msg, err := c.readMsg(buf[:])
		if err != nil {
			if isClosedError(err) {
				return
			}
			log.Debugf("readMsg error: %v", err)
			continue
		}

		msgID, context := getMsgReplyHeader(msg)
		if debug {
			log.Debugf("recvMsg (%d) msgID=%d context=%v", len(msg), msgID, context)
		}

		c.msgCallback(msgID, msg)
	}
}

// getMsgReplyHeader gets message ID and context from the message reply header
//
// Message reply has the following structure:
//
//	type msgReplyHeader struct {
//	    MsgID       uint16
//	    Context     uint32
//	}
func getMsgReplyHeader(msg []byte) (msgID uint16, context uint32) {
	msgID = binary.BigEndian.Uint16(msg[0:2])
	context = binary.BigEndian.Uint32(msg[2:6])
	return
}

func (c *Client) readMsgTimeout(buf []byte, timeout time.Duration) ([]byte, error) {
	// set read deadline
	readDeadline := time.Now().Add(timeout)
	if err := c.conn.SetReadDeadline(readDeadline); err != nil {
		return nil, err
	}

	// read message
	msgReply, err := c.readMsg(buf)
	if err != nil {
		return nil, err
	}

	// reset read deadline
	if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
		return nil, err
	}

	return msgReply, nil
}

func (c *Client) readMsg(buf []byte) ([]byte, error) {
	log.Debug("reading msg..")

	header, ok := c.headerPool.Get().(*[]byte)
	if !ok {
		return nil, fmt.Errorf("failed to get header from pool")
	}
	msgLen, err := readMsgHeader(c.reader, *header)
	if err != nil {
		return nil, err
	}
	c.headerPool.Put(header)

	msg, err := readMsgData(c.reader, buf, msgLen)
	if err != nil {
		return nil, err
	}

	log.Debugf(" -- readMsg done (buffered: %d)", c.reader.Buffered())

	return msg, nil
}

func readMsgHeader(r io.Reader, header []byte) (int, error) {
	n, err := io.ReadAtLeast(r, header, 16)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		log.Debugln("zero bytes header")
		return 0, nil
	} else if n != 16 {
		log.Debugf("invalid header (%d bytes): % 0X", n, header[:n])
		return 0, fmt.Errorf("invalid header (expected 16 bytes, got %d)", n)
	}

	dataLen := binary.BigEndian.Uint32(header[8:12])

	return int(dataLen), nil
}

func readMsgData(r io.Reader, buf []byte, dataLen int) ([]byte, error) {
	var msg []byte
	if buf == nil || len(buf) < dataLen {
		msg = make([]byte, dataLen)
	} else {
		msg = buf[0:dataLen]
	}

	n, err := r.Read(msg)
	if err != nil {
		return nil, err
	}
	if debug {
		log.Debugf(" - read data (%d bytes): % 0X", n, msg[:n])
	}

	if dataLen > n {
		remain := dataLen - n
		log.Debugf("continue reading remaining %d bytes", remain)
		view := msg[n:]

		for remain > 0 {
			nbytes, err := r.Read(view)
			if err != nil {
				return nil, err
			} else if nbytes == 0 {
				return nil, fmt.Errorf("zero nbytes")
			}

			remain -= nbytes
			log.Debugf("another data received: %d bytes (remain: %d)", nbytes, remain)

			view = view[nbytes:]
		}
	}

	return msg, nil
}

func isClosedError(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	}
	return strings.HasSuffix(err.Error(), "use of closed network connection")
}
