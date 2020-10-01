# GoVPP

The GoVPP projects provides the API for communication with VPP from Go.

It consists of the following packages:
- [adapter](adapter/): adapter between GoVPP core and the VPP binary/stats API
- [api](api/): API for communication with GoVPP core
- [cmd/binapi-generator](cmd/binapi-generator/): generator for the VPP binary API definitions in JSON format to Go code
- [codec](codec/): handles encoding/decoding of generated messages into binary form
- [core](core/): essential functionality of the GoVPP
- [examples](examples/): examples that use the GoVPP API in real use-cases of VPP management application
- [extras](extras/): contains Go implementation for libmemif library
- [govpp](govpp.go): provides the entry point to GoVPP functionality

The design with separated GoVPP [API package](api/) and the GoVPP [core package](core/) enables
plugin-based infrastructure, where one entity acts as a master responsible for talking with VPP and multiple
entities act as clients that are using the master for the communication with VPP.
The clients can be built into standalone shared libraries without the need
of linking the GoVPP core and all its dependencies into them.

```
                                       +--------------+
    +--------------+                   |              |
    |              |                   |    plugin    |
    |              |                   |              |
    |     App      |                   +--------------+
    |              |            +------+  GoVPP API   |
    |              |            |      +--------------+
    +--------------+   GoVPP    |
    |              |  channels  |      +--------------+
    |  GoVPP core  +------------+      |              |
    |              |            |      |    plugin    |
    +------+-------+            |      |              |
           |                    |      +--------------+
           |                    +------+  GoVPP API   |
           | binary API                +--------------+
           |
    +------+-------+
    |              |
    |  VPP process |
    |              |
    +--------------+
```

## Prerequisites

- [Go 1.13](https://golang.org/dl)

## Quick Start

Make sure that $GOPATH, $GOROOT, and $PATH are set. If you cloned the
govpp repo manually, you'll probably regret it.

Instead:

```
    go get git.fd.io/govpp.git
```

### Build and install the govpp binary api generator

```
    $ cd $GOPATH/src/git.fd.io/govpp.git/cmd/binapi-generator
    $ go install
    $ binapi-generator -version
    govpp v0.4.0-dev  # or some such
```
### Install vpp binary artifacts (including the "vpp-dev" package)

Build locally, or download from packagecloud. Beyond the scope of
README.md.

### Generate binapi (Go bindings)

Generating Go bindings for VPP binary API from the JSON files
installed with the vpp binary artifacts.

```
    $ cd $GOPATH/src/git.fd.io/govpp.git
    $ binapi-generator --output-dir=binapi
    INFO[0000] found 110 files in API dir "/usr/share/vpp/api"
    INFO[0000] Generating 203 files
```

The golang binding files land here: $GOPATH/src/git.fd.io/govpp.git/binapi

#### Example Usage

Here's a simple sample program which prints vpp version info, and
creates a loopback interface.

```go
package main

import (
	"fmt"
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
		fmt.Errorf("SendRequest: %w\n", err)
	}

	fmt.Printf("Program: %s\nVersion: %s\nBuildDate: %s\n",
		reply.Program, reply.Version, reply.BuildDate)

	loop_create := &interfaces.CreateLoopback{}
	loop_create_reply := &interfaces.CreateLoopbackReply{}

	err = ch.SendRequest(loop_create).ReceiveReply(loop_create_reply)

	if err != nil {
		fmt.Errorf("create_loopback: %w\n", err)
	}

	fmt.Printf("create_loopback: sw_if_index %d",
		int(loop_create_reply.SwIfIndex))
}
```

The example above uses GoVPP API to communicate over underlying go channels,
see [example client](examples/simple-client/simple_client.go)
for more examples, including the example on how to use the Go channels directly.

### Tracking down generated go code for a specific binary API

Golang uses capitalization to indicate exported names, so you'll have
to divide through by binapi-generator transformations. Example:

```
    define create_loopback  -> type CreateLoopback struct ...
      vpp binapi definition      govpp exported type definition
```
The droids you're looking for will be in a file named
<something>.ba.go.  Suggest:

```
    find git.fd.io/govpp/binapi -name "*.ba.go" | xargs grep -n GoTypeName
```

Look at the indicated <something>.ba.go file, deduce the package name
and import it. See the example above.

## Build & Install

### Using pure Go adapters (recommended)

GoVPP now supports pure Go implementation for VPP binary API. This does
not depend on CGo or any VPP library and can be easily compiled.

There are two packages providing pure Go implementations in GoVPP:
- [`socketclient`](adapter/socketclient) - for VPP binary API (via unix socket)
- [`statsclient`](adapter/statsclient) - for VPP stats API (via shared memory)

### Using vppapiclient library wrapper (requires CGo)

GoVPP also provides vppapiclient package which actually uses
`vppapiclient.so` library from VPP codebase to communicate with VPP API.
To build GoVPP, `vpp-dev` package must be installed,
either [from packages][from-packages] or [from sources][from-sources].

To build & install `vpp-dev` from sources:

```sh
git clone https://gerrit.fd.io/r/vpp
cd vpp
make install-dep
make pkg-deb
cd build-root
sudo dpkg -i vpp*.deb
```

To build & install GoVPP:

```sh
go get -u git.fd.io/govpp.git
cd $GOPATH/src/git.fd.io/govpp.git
make test
make install
```

## Generating Go bindings with binapi-generator

Once you have `binapi-generator` installed, you can use it to generate Go bindings for VPP binary API
using VPP APIs in JSON format. The JSON input can be specified as a single file (`input-file` argument), or
as a directory that will be scanned for all `.json` files (`input-dir`). The generated Go bindings will
be placed into `output-dir` (by default current working directory), where each Go package will be placed into
a separated directory, e.g.:

```sh
binapi-generator --input-file=acl.api.json --output-dir=binapi
binapi-generator --input-dir=/usr/share/vpp/api/core --output-dir=binapi
```

In Go, [go generate](https://blog.golang.org/generate) tool can be leveraged to ease the code generation
process. It allows to specify generator instructions in any one of the regular (non-generated) `.go` files
that are dependent on generated code using special comments, e.g. the one from
[example client](examples/simple-client/simple_client.go):

```go
//go:generate binapi-generator --input-dir=bin_api --output-dir=bin_api
```

[from-packages]: https://wiki.fd.io/view/VPP/Installing_VPP_binaries_from_packages
[from-sources]: https://wiki.fd.io/view/VPP/Build,_install,_and_test_images#Build_A_VPP_Package
