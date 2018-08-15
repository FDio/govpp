# GoVPP

This set of packages provide the API for communication with VPP from Go. 

It consists of the following packages:
- [adapter](adapter/): adapter between GoVPP core and the VPP
- [api](api/api.go): API for communication with GoVPP core
- [binapi-generator](cmd/binapi-generator/): Generator for the VPP binary API definitions in JSON format to Go code
- [codec](codec/): handles encoding/decoding of generated messages into binary form
- [core](core/): main functionality of the GoVPP
- [examples](examples/): examples that use the GoVPP API in real use-cases of VPP management application
- [extras](extras/): contains Go implementation for libmemif library
- [govpp](govpp.go): provides the entry point to GoVPP functionality

The design with separated GoVPP [API package](api/api.go) and the GoVPP [core package](core/) enables 
plugin-based infrastructure, where one entity acts as a master responsible for talking with VPP and multiple 
entities act as clients that are using the master for the communication with VPP. 
The clients can be built into standalone shared libraries without the need 
of linking the GoVPP core and all its dependencies into them.

```
                                       +--------------+
    +--------------+                   |              |
    |              |                   |  App plugin  |
    |              |                   |              |
    |     App      |                   +--------------+
    |              |            +------+  GoVPP API   |
    |              |            |      +--------------+
    +--------------+     Go     |
    |              |  channels  |      +--------------+
    |  GoVPP core  +------------+      |              |
    |              |            |      |  App plugin  |
    +------+-------+            |      |              |
           |                    |      +--------------+
binary API |                    +------+  GoVPP API   |
 (shmem)   |                           +--------------+
           |
    +------+-------+
    |              |
    |  VPP process |    
    |              |
    +--------------+
```

## Quick Start

#### Code Generator

Generating Go bindings from the JSON files located in the `/usr/share/vpp/api/` directory 
into the Go packages that will be created inside of the `bin_api` directory:
```
binapi-generator --input-dir=/usr/share/vpp/api/ --output-dir=bin_api
```

#### Example Usage

Usage of the generated bindings:

```go
func main() {
	conn, _ := govpp.Connect()
	defer conn.Disconnect()

	ch, _ := conn.NewAPIChannel()
	defer ch.Close()
  
	req := &acl.ACLAddReplace{
		ACLIndex: ^uint32(0),
		Tag:      []byte("access list 1"),
		R: []acl.ACLRule{
			{
				IsPermit:       1,
				SrcIPAddr:      net.ParseIP("10.0.0.0").To4(),
				SrcIPPrefixLen: 8,
				DstIPAddr:      net.ParseIP("192.168.1.0").To4(),
				DstIPPrefixLen: 24,
				Proto:          6,
			},
			{
				IsPermit:       1,
				SrcIPAddr:      net.ParseIP("8.8.8.8").To4(),
				SrcIPPrefixLen: 32,
				DstIPAddr:      net.ParseIP("172.16.0.0").To4(),
				DstIPPrefixLen: 16,
				Proto:          6,
			},
		},
	}
	reply := &acl.ACLAddReplaceReply{}

	err := ch.SendRequest(req).ReceiveReply(reply)
}
```

The example above uses simple wrapper API over underlying go channels, see [example client](examples/cmd/simple-client/simple_client.go) 
for more examples, including the example on how to use the Go channels directly.

## Build & Installation Procedure

Govpp uses `vppapiclient` library from VPP codebase to communicate with VPP. To build GoVPP, vpp-dev package must be installed,
either [from packages](https://wiki.fd.io/view/VPP/Installing_VPP_binaries_from_packages) or 
[from sources](https://wiki.fd.io/view/VPP/Build,_install,_and_test_images#Build_A_VPP_Package).

To build & install `vpp-dev` from sources:

```
git clone https://gerrit.fd.io/r/vpp
cd vpp
make install-dep
make bootstrap
make pkg-deb
cd build-root
sudo dpkg -i vpp*.deb
```

To build & install all GoVPP binaries into your `$GOPATH`:

```
go get git.fd.io/govpp.git
cd $GOPATH/src/git.fd.io/govpp.git
make
make install
```

## Building Go bindings from VPP binary APIs

Once you have `binapi-generator` installed in your `$GOPATH`, you can use it to generate Go bindings from
VPP APis in JSON format. The JSON input can be specified as a single file (`input-file` argument), or
as a directory that will be scanned for all `.json` files (`input-dir`). The generated Go bindings will
be placed into `output-dir` (by default current working directory), where each Go package will be placed into 
a separated directory, e.g.:

```
binapi-generator --input-file=examples/bin_api/acl.api.json --output-dir=examples/bin_api
binapi-generator --input-dir=examples/bin_api --output-dir=examples/bin_api
```

In Go, [go generate](https://blog.golang.org/generate) tool can be leveraged to ease the code generation
process. It allows to specify generator instructions in any one of the regular (non-generated) `.go` files
that are dependent on generated code using special comments, e.g. the one from 
[example client](examples/cmd/simple-client/simple_client.go):

```go
//go:generate binapi-generator --input-dir=bin_api --output-dir=bin_api
```
