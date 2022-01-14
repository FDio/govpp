# Binapi Generator

## Install the binary API generator

```sh
# Install binapi generator
make install-generator
```

> NOTE: This installs `binapi-generator` to `$GOPATH/bin` directory, ensure 
> it is in your `$PATH` before running the command.

## Install vpp binary artifacts

Build locally, or download from packagecloud. Read more: https://fd.io/docs/vpp/master/gettingstarted/installing

## Generate binapi (Go bindings)

Generating Go bindings for VPP binary API from the JSON files
installed with the vpp binary artifacts - located in `/usr/share/vpp/api/`.

```sh
make generate-binapi
INFO[0000] found 110 files in API dir "/usr/share/vpp/api"
INFO[0000] Generating 203 files
```

The generated files will be generated under `binapi` directory.

## Generate VPP binary API code (Go bindings)

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
process. It allows specifying generator instructions in any one of the regular (non-generated) `.go` files
that are dependent on generated code using special comments:

```go
//go:generate binapi-generator --input-dir=bin_api --output-dir=bin_api
```

## Tracking down generated go code for a specific binary API

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
