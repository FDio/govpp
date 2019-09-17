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
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/adapter/statsclient"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/examples/binapi/vpe"
)

const (
	defaultSyncRequestCount  = 1000
	defaultAsyncRequestCount = 10000
)

func main() {
	// parse optional flags
	var sync, prof bool
	var cnt int
	var sock string
	flag.BoolVar(&sync, "sync", false, "run synchronous perf test")
	flag.StringVar(&sock, "socket", socketclient.DefaultSocketName, "Path to VPP API socket")
	flag.String("socket", statsclient.DefaultSocketName, "Path to VPP stats socket")
	flag.IntVar(&cnt, "count", 0, "count of requests to be sent to VPP")
	flag.BoolVar(&prof, "prof", false, "generate profile data")
	flag.Parse()

	if cnt == 0 {
		// no specific count defined - use defaults
		if sync {
			cnt = defaultSyncRequestCount
		} else {
			cnt = defaultAsyncRequestCount
		}
	}

	if prof {
		defer profile.Start().Stop()
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
	} else {
		// run asynchronous test
		asyncTest(ch, cnt)
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
