// Copyright (c) 2017 Cisco and/or its affiliates.
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

// Binary simple-client is an example VPP management application that exercises the
// govpp API on real-world use-cases.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/adapter/statsclient"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/binapi/vpe"
	"git.fd.io/govpp.git/core"
)

const (
	defaultSyncRequestCount  = 1000
	defaultAsyncRequestCount = 10000
)

func main() {
	// parse optional flags
	var sync bool
	var cnt int
	var sock, prof string
	flag.BoolVar(&sync, "sync", false, "run synchronous perf test")
	flag.StringVar(&sock, "api-socket", socketclient.DefaultSocketName, "Path to VPP API socket")
	flag.String("stats-socket", statsclient.DefaultSocketName, "Path to VPP stats socket")
	flag.IntVar(&cnt, "count", 0, "count of requests to be sent to VPP")
	flag.StringVar(&prof, "prof", "", "enable profiling mode [mem, cpu]")
	flag.Parse()

	if cnt == 0 {
		// no specific count defined - use defaults
		if sync {
			cnt = defaultSyncRequestCount
		} else {
			cnt = defaultAsyncRequestCount
		}
	}

	switch prof {
	case "mem":
		defer profile.Start(profile.MemProfile, profile.MemProfileRate(1)).Stop()
	case "cpu":
		defer profile.Start(profile.CPUProfile).Stop()
	case "":
	default:
		fmt.Printf("invalid profiling mode: %q\n", prof)
		flag.Usage()
		os.Exit(1)
	}

	a := socketclient.NewVppClient(sock)

	// connect to VPP
	conn, err := core.Connect(a)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	defer conn.Disconnect()

	// create an API channel
	ch, err := conn.NewAPIChannelBuffered(cnt, cnt)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	defer ch.Close()

	ch.SetReplyTimeout(time.Second * 2)

	// log only errors
	core.SetLogger(&logrus.Logger{Level: logrus.ErrorLevel})

	// run the test & measure the time
	start := time.Now()

	if sync {
		// run synchronous test
		syncTest(ch, cnt)
		//syncTest2(conn, cnt)
	} else {
		// run asynchronous test
		asyncTest(ch, cnt)
		//asyncTest2(conn, cnt)
	}

	elapsed := time.Since(start)
	fmt.Println("Test took:", elapsed)
	fmt.Printf("Requests per second: %.0f\n", float64(cnt)/elapsed.Seconds())

	time.Sleep(time.Second)
}

func syncTest(ch api.Channel, cnt int) {
	fmt.Printf("Running synchronous perf test with %d requests...\n", cnt)

	for i := 0; i < cnt; i++ {
		req := &vpe.ControlPing{}
		reply := &vpe.ControlPingReply{}

		if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
			log.Fatalln("Error in reply:", err)
		}
	}
}

func syncTest2(conn api.Connection, cnt int) {
	fmt.Printf("Running synchronous perf test with %d requests...\n", cnt)

	stream, err := conn.NewStream(context.Background())
	if err != nil {
		log.Fatalln("Error NewStream:", err)
	}
	for i := 0; i < cnt; i++ {
		if err := stream.SendMsg(&vpe.ControlPing{}); err != nil {
			log.Fatalln("Error SendMsg:", err)
		}
		if msg, err := stream.RecvMsg(); err != nil {
			log.Fatalln("Error RecvMsg:", err)
		} else if _, ok := msg.(*vpe.ControlPingReply); ok {
			// ok
		} else {
			log.Fatalf("unexpected reply: %v", msg.GetMessageName())
		}
	}
}

func asyncTest(ch api.Channel, cnt int) {
	fmt.Printf("Running asynchronous perf test with %d requests...\n", cnt)

	ctxChan := make(chan api.RequestCtx, cnt)

	go func() {
		for i := 0; i < cnt; i++ {
			ctxChan <- ch.SendRequest(&vpe.ControlPing{})
		}
		close(ctxChan)
		fmt.Printf("Sending asynchronous requests finished\n")
	}()

	for ctx := range ctxChan {
		reply := &vpe.ControlPingReply{}
		if err := ctx.ReceiveReply(reply); err != nil {
			log.Fatalln("Error in reply:", err)
		}
	}
}

func asyncTest2(conn api.Connection, cnt int) {
	fmt.Printf("Running asynchronous perf test with %d requests...\n", cnt)

	ctxChan := make(chan api.Stream, cnt)

	go func() {
		for i := 0; i < cnt; i++ {
			stream, err := conn.NewStream(context.Background())
			if err != nil {
				log.Fatalln("Error NewStream:", err)
			}
			if err := stream.SendMsg(&vpe.ControlPing{}); err != nil {
				log.Fatalln("Error SendMsg:", err)
			}
			ctxChan <- stream
		}
		close(ctxChan)
		fmt.Printf("Sending asynchronous requests finished\n")
	}()

	for ctx := range ctxChan {
		if msg, err := ctx.RecvMsg(); err != nil {
			log.Fatalln("Error RecvMsg:", err)
		} else if _, ok := msg.(*vpe.ControlPingReply); ok {
			// ok
		} else {
			log.Fatalf("unexpected reply: %v", msg.GetMessageName())
		}
		ctx.Close()
	}
}
