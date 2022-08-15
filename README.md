### ⚠️The GoVPP project has been migrated from Gerrit to GitHub

#### What changed?
- the Go module path was changed from `git.fd.io/govpp.git` => `go.fd.io/govpp`
  - the final release for the old import path is [v0.5.0](https://pkg.go.dev/git.fd.io/govpp.git@v0.5.0) 
  - new module path can be imported with [v0.6.0-alpha](https://pkg.go.dev/go.fd.io/govpp@v0.6.0-alpha)
- repository is now located at [https://github.com/FDio/govpp](https://github.com/FDio/govpp)
  - any new contributions should be created as [pull requests](https://github.com/FDio/govpp/pulls)
  - any new issues should be tracked under [Issues](https://github.com/FDio/govpp/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc)

---

# GoVPP

[![stable](https://img.shields.io/github/v/tag/fdio/govpp.svg?label=release&logo=github)](https://github.com/FDio/govpp/releases/latest) [![PkgGoDev](https://pkg.go.dev/badge/go.fd.io/govpp)](https://pkg.go.dev/go.fd.io/govpp)

The GoVPP repository contains a Go client library for interacting with the VPP, 
generator of Go bindings for the VPP binary API and various other tooling for VPP.

## Overview

Here is a brief overview for the repository structure.

- [govpp](govpp.go) - the entry point for the GoVPP client
  - [adapter](adapter) - VPP binary & stats API interface
    - [mock](adapter/mock) - Mock adapter used for testing
    - [socketclient](adapter/socketclient) - Go implementation of VPP API client for unix socket
    - [statsclient](adapter/statsclient) - Go implementation of VPP Stats client for shared memory
  - [api](api) - GoVPP client API
  - [binapigen](binapigen) - library for generating code from VPP API
    - [vppapi](binapigen/vppapi) - VPP API parser
  - [cmd](cmd)
    - [binapi-generator](cmd/binapi-generator) - VPP binary API generator
    - [vpp-proxy](cmd/vpp-proxy) - VPP proxy for remote access
  - [codec](codec) - handles encoding/decoding of generated messages into binary form
  - [core](core) - implementation of the GoVPP client
  - [docs](docs) - user & developer documentation
  - [examples](examples) - examples demonstrating GoVPP functionality
  - [proxy](proxy) - contains client/server implementation for proxy
  - [test](test) - integration tests, benchmarks and performance tests

## Quick Start

Below are some code examples showing GoVPP client interacting with VPP API.

### Using RPC service client

Here is a sample code for an effortless way for calling the VPP API by using a generated RPC service client.

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
	conn, err := govpp.Connect("/run/vpp/api.sock")
    if err != nil {
      // handle err
    }
	defer conn.Disconnect()

	// Init vpe service client
    client := vpe.NewServiceClient(conn)

	reply, err := client.ShowVersion(context.Background(), &vpe.ShowVersion{})
    if err != nil {
      // handle err
    }

	log.Print("Version: ", reply.Version)
}
```

For a complete example see [rpc-service](examples/rpc-service).

### Using raw messages directly

Here is a code for low-level way to access the VPP API using the generated messages directly for sending/receiving.

```go
package main

import (
    "log"
    
	"go.fd.io/govpp"
	"go.fd.io/govpp/binapi/vpe"
)

func main() {
	// Connect to the VPP API socket
	conn, err := govpp.Connect("/run/vpp/api.sock")
    if err != nil {
        // handle err
    }
	defer conn.Disconnect()

	// Open a new channel
	ch, err := conn.NewAPIChannel()
    if err != nil {
      // handle err
    }
	defer ch.Close()

	// Prepare messages
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	// Send the request
	if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
        // handle err
	}

    log.Print("Version: ", reply.Version)
}
```

For a complete example see [simple-client](examples/simple-client).

### More code examples

More examples can be found in [examples](examples) directory.

- [api-trace](examples/api-trace) - trace sent/received messages
- [binapi-types](examples/binapi-types) - using common types from generated code
- [multi-vpp](examples/multi-vpp) - connect to multiple VPP instances
- [perf-bench](examples/perf-bench) - very basic performance test for measuring throughput
- [rpc-service](examples/rpc-service) - effortless way to call VPP API via RPC client
- [simple-client](examples/simple-client) - send and receive VPP API messages using GoVPP API directly
- [stats-client](examples/stats-client) - client for retrieving VPP stats data
- [stream-client](examples/stream-client) - using new stream API to call VPP API

## Documentation

Further documentation can be found in [docs](docs) directory.
