<h1 align="center" style="border-bottom: none">
    <img alt="logo" src="./docs/govpp-logo.png"><br>GoVPP
</h1>

<p align="center">
	<a href="https://github.com/FDio/govpp/tags"><img src="https://img.shields.io/github/v/tag/fdio/govpp?label=latest&logo=github&sort=semver&color=blue" alt="Latest"></a>
	<a href="https://pkg.go.dev/go.fd.io/govpp"><img src="https://pkg.go.dev/badge/go.fd.io/govpp" alt="PkgGoDev"></a>
	<a href="https://github.com/FDio/govpp/actions/workflows/ci.yaml"><img src="https://github.com/FDio/govpp/actions/workflows/ci.yaml/badge.svg" alt="CI"></a>
	<a href="https://github.com/FDio/govpp/actions/workflows/test.yaml"><img src="https://github.com/FDio/govpp/actions/workflows/test.yaml/badge.svg" alt="Test"></a>
</p>

The GoVPP repository contains Go client libraries, code bindings generator and other toolings for VPP.

---

## Features

* 🆕 CLI app for interacting with VPP instance and development of VPP API (see [GoVPP CLI](https://github.com/FDio/govpp/blob/master/docs/GOVPP_CLI.md))
* 🆕 Extendable code generator supporting custom plugins (see [Enhanced Generator](https://github.com/FDio/govpp/discussions/94))
* 🆕 Generated RPC client code that handles all boilerplate (see [RPC Services](https://github.com/FDio/govpp/discussions/58))
* Simple VPP client API that is not dependent on any VPP API semantics (see [Stream API](https://github.com/FDio/govpp/discussions/43))
* Generator of Go bindings for VPP API schema (see [Binapi Generator](https://github.com/FDio/govpp/blob/master/docs/USER_GUIDE.md#binary-api-generator))
* Go client library for VPP binary API & Stats API (see [VPP API calls](https://github.com/FDio/govpp/blob/master/docs/USER_GUIDE.md#vpp-api-calls))
* Pure Go implementation of VPP binary API protocol (see [socketclient](https://github.com/FDio/govpp/blob/master/adapter/socketclient/socketclient.go))
* Efficient reader of VPP Stats data from shared memory (see [stats client example](https://github.com/FDio/govpp/tree/master/examples/stats-client))

## Quick Start

Here is a code sample of an effortless way for calling the VPP API services by using a generated RPC client.

> **Note**
> For extensive info about using generated RPC client , see [RPC Services](https://github.com/FDio/govpp/discussions/58)

```go
// Connect to VPP API socket
conn, err := govpp.Connect("/run/vpp/api.sock")
if err != nil {
  // handle err
}
defer conn.Disconnect()

// Initialize VPP API service client
client := vpe.NewServiceClient(conn)

// Call VPP API service method
reply, err := client.ShowVersion(context.Background(), &vpe.ShowVersion{})
if err != nil {
  // handle err
}
log.Print("Version: ", reply.Version)
```

See complete code for the example above: [examples/rpc-service](examples/rpc-service).

### Examples

For complete code examples demonstrating vrious GoVPP features, please refer to the [examples](examples) directory.

## Documentation

Refer to [User Guide](docs/USER_GUIDE.md) document for info about how to use GoVPP. 
If you run into any issues or need some help with debugging GoVPP, read our [Troubleshooting](docs/TROUBLESHOOTING.md) document.

Go reference docs are available at [pkg.go.dev](https://pkg.go.dev/go.fd.io/govpp). 

For other documentation refer to [docs](docs) directory.

## How to contribute?

Anyone insterested in GoVPP development is welcome to join our bi-weekly [📣 GoVPP Community Meeting](https://github.com/FDio/govpp/discussions/46), where we accept inputs from projects using GoVPP and have technical discussions about actual development.

- **Contribute code**: submit a [Pull Request](https://github.com/FDio/govpp/pulls)
- **Report bugs**: open an [Issue](https://github.com/FDio/govpp/issues)
- **Ask questions**: start a [Discussion](https://github.com/FDio/govpp/discussions)

## Repository Structure

Here is a brief overview of the repository structure.

- [govpp](govpp.go) - the entry point for the GoVPP client
  - [adapter](adapter) - VPP binary & stats API interface
    - [mock](adapter/mock) - Mock adapter used for testing
    - [socketclient](adapter/socketclient) - Go implementation of VPP API client for unix socket
    - [statsclient](adapter/statsclient) - Go implementation of VPP Stats client for shared memory
  - [api](api) - GoVPP client API
  - [binapi](binapi) - generated Go bindings for the latest VPP release
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
  
