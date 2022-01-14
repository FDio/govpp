# GoVPP

[![stable](https://img.shields.io/github/v/tag/fdio/govpp.svg?label=release&logo=github)](https://github.com/ligato/vpp-agent/releases/latest) [![PkgGoDev](https://pkg.go.dev/badge/git.fd.io/govpp.git)](https://pkg.go.dev/git.fd.io/govpp.git)

The GoVPP repository contains a Go client library for interacting with the VPP and also Go code generator for the VPP API.

## Overview

- [govpp](govpp.go) - the entry point for the GoVPP client
  - [adapter](adapter) - VPP binary & stats API interface
      - [mock](adapter/mock) - Mock adapter used for testing
      - [socketclient](adapter/socketclient) - Go implementation of VPP API client for unix socket
      - [statsclient](adapter/statsclient) - Go implementation of VPP Stats client for shared memory
      - [vppapiclient](adapter/vppapiclient) - CGo wrapper of vppapiclient library (DEPRECATED!)
  - [api](api) - GoVPP client API
  - [binapigen](binapigen) - library for generating code from VPP API
      - [vppapi](binapigen/vppapi) - VPP API parser
  - [cmd](cmd)
      - [binapi-generator](cmd/binapi-generator) - VPP binary API generator
      - [vpp-proxy](cmd/vpp-proxy) - VPP proxy for remote access
  - [codec](codec) - handles encoding/decoding of generated messages into binary form
  - [core](core) - implementation of the GoVPP client
  - [docs](docs) - documentation
  - [examples](examples) - examples demonstrating GoVPP functionality
  - [internal](internal) - packages used internally by GoVPP
  - [proxy](proxy) - contains client/server implementation for proxy
  - [test](test) - integration tests, benchmarks and performance tests

## Install binapi-generator

### Prerequisites

- Go 1.13+ ([download]((https://golang.org/dl)))

### Install via Go toolchain

```shell
# Latest version (most recent tag)
go install git.fd.io/govpp.git/cmd/binapi-generator@latest

# Development version (master branch)
go install git.fd.io/govpp.git/cmd/binapi-generator@master
```

NOTE: Using `go install` to install programs is only supported in Go 1.16+ ([more info](https://go.dev/doc/go1.16#go-command)). For Go 1.15 or older, use `go get` instead of `go install`.

### Install from source

```sh
# Clone repository anywhere
git clone https://gerrit.fd.io/r/govpp.git
cd govpp
make install-generator
```

NOTE: There is no need to setup or clone inside `GOPATH` for Go 1.13+ ([more info](https://go.dev/doc/go1.13#modules)) 
and you can simply clone the repository _anywhere_ you want. 

## Examples

The [examples/](examples/) directory contains several examples for GoVPP

- [api-trace](examples/api-trace) - trace sent/received messages
- [binapi-types](examples/binapi-types) - using common types from generated code
- [multi-vpp](examples/multi-vpp) - connect to multiple VPP instances
- [perf-bench](examples/perf-bench) - very basic performance test for measuring throughput
- [rpc-service](examples/rpc-service) - effortless way to call VPP API via RPC client
- [simple-client](examples/simple-client) - send and receive VPP API messages using GoVPP API directly
- [stats-client](examples/stats-client) - client for retrieving VPP stats data
- [stream-client](examples/stream-client) - using new stream API to call VPP API

### Documentation

Further documentation can be found in [docs/](docs/) directory.

- [Adapters](docs/ADAPTERS.md) - detailed info about transport adapters and their implementations
- [Binapi Generator](docs/GENERATOR.md) - user guide for generating VPP binary API

## Quick Start

### Using raw messages directly

Here is a sample code for low-level way to access the VPP API using the generated messages directly for sending/receiving.

```go
package main

import (
    "log"
    
	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/binapi/interfaces"
	"git.fd.io/govpp.git/binapi/vpe"
)

func main() {
	// Connect to VPP
	conn, _ := govpp.Connect("/run/vpp/api.sock")
	defer conn.Disconnect()

	// Open channel
	ch, _ := conn.NewAPIChannel()
	defer ch.Close()

	// Prepare messages
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	// Send the request
	err := ch.SendRequest(req).ReceiveReply(reply)
	if err != nil {
        log.Fatal("ERROR: ", err)
	}

    log.Print("Version: ", reply.Version)
}
```

For more extensive example of using raw VPP API see [simple-client](examples/simple-client).

### Using RPC service client

Here is a sample code for an effortless way to access the VPP API using a generated RPC service client for calling.

```go
package main

import (
    "context"
    "log"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/binapi/vpe"
)

func main() {
	// Connect to VPP API socket
	conn, _ := govpp.Connect("/run/vpp/api.sock")
	defer conn.Disconnect()

	// Init vpe service client
    client := vpe.NewServiceClient(conn)

	reply, err := client.ShowVersion(context.Background(), &vpe.ShowVersion{})
	if err != nil {
		log.Fatal("ERROR: ", err)
	}

	log.Print("Version: ", reply.Version)
}
```

For more extensive example of using RPC service client see [rpc-service](examples/rpc-service).
