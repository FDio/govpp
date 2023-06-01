<h1 align="center" style="border-bottom: none">
    <img alt="logo" src="./docs/govpp-logo.png"><br>GoVPP
</h1>

<p align="center">
	<a href="https://github.com/FDio/govpp/actions/workflows/ci.yaml"><img src="https://github.com/FDio/govpp/actions/workflows/ci.yaml/badge.svg" alt="CI"></a>
	<a href="https://github.com/FDio/govpp/tags"><img src="https://img.shields.io/github/v/tag/fdio/govpp?label=latest&logo=github&sort=semver&color=blue" alt="Latest"></a>
	<a href="https://pkg.go.dev/go.fd.io/govpp"><img src="https://pkg.go.dev/badge/go.fd.io/govpp" alt="PkgGoDev"></a>
</p>

The GoVPP repository contains Go client libraries, code bindings generator and other toolings for VPP.

Here is a brief summary of features provided by GoVPP:

* Generator of Go bindings for VPP API
* Go client library for VPP binary API & Stats API
* Extendable code generator supporting custom plugins
* Pure Go implementation of VPP binary API protocol
* Efficient reader of VPP Stats data from shared memory
* Simple client API that does not rely on VPP API semantics
* Generated RPC client code that handles all boilerplate
* ..and much more!

---

## Quick Start

Here is a code sample of an effortless way for calling the VPP API by using a generated RPC client.

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

See complete example in [examples/rpc-service](examples/rpc-service).

### Examples

For all code examples demonstrating GoVPP features, please refer to the [examples](examples) directory.

## Documentation

Refer to [User Guide](docs/USER_GUIDE.md) document for all the basics. If you run into issues or just need help debugging read our [Troubleshooting](docs/TROUBLESHOOTING.md) document.

Go reference is available at https://pkg.go.dev/go.fd.io/govpp. More documentation can be found under [docs](docs) directory.

## How to contribute?

- Contribute code by submitting a [Pull Request](https://github.com/FDio/govpp/pulls).
- Report bugs by opening an [Issue](https://github.com/FDio/govpp/issues).
- Ask questions & open discussions by starting a [Discussion](https://github.com/FDio/govpp/discussions).

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
  
