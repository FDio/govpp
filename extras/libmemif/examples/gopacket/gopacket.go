// gopacket is simple example how to use alternative method of using gopacket calls and packet handles
// to implement simple use-case that is shown in raw-data. For more informations check raw-data example.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"git.fd.io/govpp.git/extras/libmemif"
	"io"
)

const (
	Socket             = "/tmp/gopacket-example"
	Secret             = "secret"
	ConnectionID       = 1
	NumQueues    uint8 = 3
)

var stopCh chan struct{}

func OnConnect(memif *libmemif.Memif) (err error) {
	details, err := memif.GetDetails()

	if err != nil {
		fmt.Printf("libmemif.GetDetails() error: %v\n", err)
		return
	}

	fmt.Printf("memif %s has been connected: %+v\n", memif.IfName, details)

	stopCh = make(chan struct{})

	for _, queue := range details.RxQueues {
		ch, err := memif.GetQueueInterruptChan(queue.QueueID)
		if err != nil {
			continue
		}

		go ReadPackets(memif.NewPacketHandle(queue.QueueID, 3), ch)
	}

	for _, queue := range details.TxQueues {
		go SendPackets(memif.NewPacketHandle(queue.QueueID, 3))
	}

	return nil
}

func OnDisconnect(memif *libmemif.Memif) (err error) {
	fmt.Printf("memif %s has been disconnected\n", memif.IfName)
	close(stopCh)
	return nil
}

func ReadPackets(handle libmemif.MemifPacketHandle, interruptCh <-chan struct{}) {
	for {
		select {
		case <-interruptCh:
		read:
			for {
				data, _, err := handle.ReadPacketData()

				switch err {
				case io.EOF:
					break read
				case nil:
					fmt.Printf("Received packet %v\n", string(data[:]))
				default:
					fmt.Printf("Got error while reading packet %v\n", err)
				}
			}
		case <-stopCh:
			handle.Close()
			return
		}
	}
}

func SendPackets(handle libmemif.MemifPacketHandle) {
	counter := 0

	for {
		select {
		case <-time.After(3 * time.Second):
			counter++

			// Prepare fake packets.
			packets := [][]byte{
				[]byte("Packet #1 in burst number " + strconv.Itoa(counter)),
				[]byte("Packet #2 in burst number " + strconv.Itoa(counter)),
				[]byte("Packet #3 in burst number " + strconv.Itoa(counter)),
			}

		write:
			for _, data := range packets {
				err := handle.WritePacketData(data)

				switch err {
				case io.EOF:
					break write
				case nil:
					fmt.Printf("Sent packet %v\n", string(data[:]))
				default:
					fmt.Printf("Got error while sending packet %v\n", err)
				}
			}
		case <-stopCh:
			handle.Close()
			return
		}
	}
}

func main() {
	var isMaster = true
	var appSuffix string
	if len(os.Args) > 1 && (os.Args[1] == "--slave" || os.Args[1] == "-slave") {
		isMaster = false
		appSuffix = "-slave"
	}

	appName := "gopacket" + appSuffix
	err := libmemif.Init(appName)
	if err != nil {
		fmt.Printf("libmemif.Init() error: %v\n", err)
		return
	}

	defer libmemif.Cleanup()

	memifCallbacks := &libmemif.MemifCallbacks{
		OnConnect:    OnConnect,
		OnDisconnect: OnDisconnect,
	}

	memifConfig := &libmemif.MemifConfig{
		MemifMeta: libmemif.MemifMeta{
			IfName:         "memif1",
			ConnID:         ConnectionID,
			SocketFilename: Socket,
			Secret:         Secret,
			IsMaster:       isMaster,
			Mode:           libmemif.IfModeEthernet,
		},
		MemifShmSpecs: libmemif.MemifShmSpecs{
			NumRxQueues: NumQueues,
			NumTxQueues: NumQueues,
		},
	}

	memif, err := libmemif.CreateInterface(memifConfig, memifCallbacks)
	if err != nil {
		fmt.Printf("libmemif.CreateInterface() error: %v\n", err)
		return
	}

	defer memif.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
