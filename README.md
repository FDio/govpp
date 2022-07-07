⚠️ The GoVPP project has changed URL :
- the import URL has moved to [go.fd.io/govpp](https://go.fd.io/govpp)
- the repository location has moved to [https://github.com/FDio/govpp](https://github.com/FDio/govpp).

The last version archived on [git.fd.io/govpp.git](https://git.fd.io/govpp) is `v0.5.0`.


# GoVPP

[![stable](https://img.shields.io/github/v/tag/fdio/govpp.svg?label=release&logo=github)](https://github.com/ligato/vpp-agent/releases/latest) [![PkgGoDev](https://pkg.go.dev/badge/go.fd.io/govpp)](https://pkg.go.dev/go.fd.io/govpp)

The GoVPP repository contains a Go client library for interacting with the VPP, 
generator of Go bindings for the VPP binary API and various other tooling for VPP.

## Overview

Here is a brief overview for the repository structure.

- [govpp](govpp.go) - the entry point for the GoVPP client
  - [adapter](adapter) - VPP binary & stats API interface
    - [mock](adapter/mock) - Mock adapter used for testing
    - [socketclient](adapter/socketclient) - Go implementation of VPP API client for unix socket
    - [statsclient](adapter/statsclient) - Go implementation of VPP Stats client for shared memory
    - [vppapiclient](adapter/vppapiclient) - CGo wrapper of vppapiclient library (DEPRECATED!)
  - [api](api) - GoVPP client API
  - [binapigen](binapigen) - library for generating code from VPP API
    - [vppapi](binapigen/vppapi) - VPP API parser
  - cmd/
    - [binapi-generator](cmd/binapi-generator) - VPP binary API generator
    - [vpp-proxy](cmd/vpp-proxy) - VPP proxy for remote access
  - [codec](codec) - handles encoding/decoding of generated messages into binary form
  - [core](core) - implementation of the GoVPP client
  - [docs](docs) - documentation
  - [examples](examples) - examples demonstrating GoVPP functionality
  - [internal](internal) - packages used internally by GoVPP
  - [proxy](proxy) - contains client/server implementation for proxy
  - [test](test) - integration tests, benchmarks and performance tests

## Quick Start

Below are some code examples showing GoVPP client interacting with VPP API.

### Using raw messages directly

Here is a code for low-level way to access the VPP API using the generated messages directly for sending/receiving.

```go
package main

import (
    "log"
    
	"go.fd.io/govpp"
	"go.fd.io/govpp/binapi/interfaces"
	"go.fd.io/govpp/binapi/vpe"
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

For a complete example see [simple-client](examples/simple-client).

### Using RPC service client

Here is a sample code for an effortless way to access the VPP API using a generated RPC service client for calling.

```go
package main

import (
    "context"
    "log"

	"go.fd.io/govpp"
	"go.fd.io/govpp/binapi/vpe"
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

For a complete example see [rpc-service](examples/rpc-service).

### More code examples

More examples can be found in [examples](examples) directory.

- [api-trace](api-trace) - trace sent/received messages
- [binapi-types](binapi-types) - using common types from generated code
- [multi-vpp](multi-vpp) - connect to multiple VPP instances
- [perf-bench](perf-bench) - very basic performance test for measuring throughput
- [rpc-service](rpc-service) - effortless way to call VPP API via RPC client
- [simple-client](simple-client) - send and receive VPP API messages using GoVPP API directly
- [stats-client](stats-client) - client for retrieving VPP stats data
- [stream-client](stream-client) - using new stream API to call VPP API

## Documentation

Further documentation can be found in [docs](docs) directory.

- [Adapters](docs/ADAPTERS.md) - detailed info about transport adapters and their implementations
- [Binapi Generator](docs/GENERATOR.md) - user guide for generating VPP binary API
