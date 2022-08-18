// jumbo-frames is simple example how to send larger and larger jumbo packets with libmemif adapter. This is simple copy
// of raw-data but with sending larger packets, so for more information read its code and docs.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.fd.io/govpp/extras/libmemif"
)

const (
	Socket             = "/tmp/jumbo-frames-example"
	Secret             = "secret"
	ConnectionID       = 1
	NumQueues    uint8 = 3
)

var wg sync.WaitGroup
var stopCh chan struct{}

func OnConnect(memif *libmemif.Memif) (err error) {
	details, err := memif.GetDetails()
	if err != nil {
		fmt.Printf("libmemif.GetDetails() error: %v\n", err)
	}
	fmt.Printf("memif %s has been connected: %+v\n", memif.IfName, details)

	stopCh = make(chan struct{})
	var i uint8
	for i = 0; i < uint8(len(details.RxQueues)); i++ {
		wg.Add(1)
		go ReadAndPrintPackets(memif, i)
	}
	for i = 0; i < uint8(len(details.TxQueues)); i++ {
		wg.Add(1)
		go SendPackets(memif, i)
	}
	return nil
}

func OnDisconnect(memif *libmemif.Memif) (err error) {
	fmt.Printf("memif %s has been disconnected\n", memif.IfName)
	close(stopCh)
	wg.Wait()
	return nil
}

func ReadAndPrintPackets(memif *libmemif.Memif, queueID uint8) {
	defer wg.Done()

	interruptCh, err := memif.GetQueueInterruptChan(queueID)
	if err != nil {
		switch err {
		case libmemif.ErrQueueID:
			fmt.Printf("libmemif.Memif.GetQueueInterruptChan() complains about invalid queue id!?")
		default:
			fmt.Printf("libmemif.Memif.GetQueueInterruptChan() error: %v\n", err)
		}
		return
	}

	counter := 0
	for {
		select {
		case <-interruptCh:
			counter++
			for {
				packets, err := memif.RxBurst(queueID, 10)
				if err != nil {
					fmt.Printf("libmemif.Memif.RxBurst() error: %v\n", err)
				} else {
					if len(packets) == 0 {
						break
					}
					for _, packet := range packets {
						fmt.Printf("Received packet queue=%d: %v in burst %d\n", queueID, len(packet), counter)
					}
				}
			}
		case <-stopCh:
			return
		}
	}
}

func SendPackets(memif *libmemif.Memif, queueID uint8) {
	defer wg.Done()

	counter := 0
	for {
		select {
		case <-time.After(3 * time.Second):
			counter++
			packetMul := counter%100 + 1 // Limit max iterations to 100 to not go out of bounds
			packets := []libmemif.RawPacketData{
				make([]byte, 128*packetMul),
				make([]byte, 256*packetMul),
				make([]byte, 512*packetMul),
			}
			sent := 0
			for {
				count, err := memif.TxBurst(queueID, packets[sent:])
				if err != nil {
					fmt.Printf("libmemif.Memif.TxBurst() error: %v\n", err)
					break
				} else {
					fmt.Printf("libmemif.Memif.TxBurst() has sent %d packets in burst %v.\n", count, counter)
					sent += int(count)
					if sent == len(packets) {
						break
					}
				}
			}
		case <-stopCh:
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

	appName := "jumbo-frames" + appSuffix
	fmt.Println("Initializing libmemif as ", appName)
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
			AppName:        appName,
		},
		MemifShmSpecs: libmemif.MemifShmSpecs{
			NumRxQueues:  NumQueues,
			NumTxQueues:  NumQueues,
			BufferSize:   2048,
			Log2RingSize: 10,
		},
	}

	fmt.Printf("Callbacks: %+v\n", memifCallbacks)
	fmt.Printf("Config: %+v\n", memifConfig)

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
