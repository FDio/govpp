# GoVPP User Guide

This section contains reference documentation for working with GoVPP, generating VPP binary API bindings, calling VPP API, retrieving VPP Stats data and other GoVPP features.

---

**Table of Contents**

* [Binary API Generator](#binary-api-generator)
    * [Installation](#installation)
    * [Using Generator](#using-generator)
    * [Generator Plugins](#generator-plugins)
* [VPP API calls](#vpp-api-calls)
    * [Connection](#connection)
        * [Synchronous](#synchronous-connect)
        * [Asynchronous](#asynchronous-connect)
    * [Sending API messages](#sending-api-messages)
        * [Channel](#channel)
        * [Stream client](#stream-client)
* [The HTTP service](#http-service)
* [The RPC service](#rpc-client)
* [VPP stats](#vpp-stats)
    * [Low-level API connection](#low-level-stats-api-connection)
    * [Low-level API usage](#low-level-stats-api-usage)

## Binary API Generator

The binary API generator's purpose is to generate Go code bindings for the VPP API. It uses VPP API definitions from `*.api.json` files as source and generates a Go package for each VPP API file, where each package contains Go types for objects from the VPP API file (messages, types, enums..).

### Installation

#### Prerequisites

- Go 1.18+ is required ([download](https://golang.org/dl))

Install binapi-generator into `$GOPATH/bin` (defaults to: `$HOME/go/bin`) using the Go toolchain:

```sh
# Latest release
go install go.fd.io/govpp/cmd/binapi-generator@latest

# Specific version
go install go.fd.io/govpp/cmd/binapi-generator@v0.7.0

# Development branch
go install go.fd.io/govpp/cmd/binapi-generator@master
```

> **Note**
> Development branch might not have stable behavior.

Print the version for the installed binapi-generator:

```sh
$ binapi-generator -version
govpp v0.8.0-dev
```

### Using Generator

#### Source of VPP API files

To generate Go code bindings, you need to provide source of VPP API files in JSON format (`*.api.json`). There are several ways to get it:

1. Download and install VPP from `packagecloud.io` ([learn more](https://fd.io/docs/vpp/master/gettingstarted/installing)).
2. Clone the VPP repository (`git clone https://github.com/FDio/vpp.git`) and run `make json-api-files`. 
   The VPP API files will be in: `./vpp/build-root/install-vpp-native/vpp/share/vpp/api/`.
3. If the VPP is already installed in your system, the default location for the VPP API files is: `/usr/share/vpp/api/`

#### Generating Go Bindings

If the VPP JSON API definitions are in the default directory `/usr/share/vpp/api`, call:

```sh
binapi-generator --input "/usr/share/vpp/api"
```

Generated Go code bindings will be under `./binapi` directory. 

> **Note**
> The Go package path of generated files will be resolved automatically from the module path in `go.mod` file.
> To set the Go package path manually, use option `--import-prefix=PREFIX`.

If you have a Go project (called `myproject` in `$HOME/myproject/vppbinapi`) with GoVPP as a dependency with both, the
binary API generator and the VPP connection client, generated bindings are required to be inside the project (for
example in `$HOME/myproject/vppbinapi`)

To modify the VPP API input directory, use `-input-dir` option. To modify the output directory, use
the `-output` option.

```sh
binapi-generator --input-dir="/path/to/vpp/api" --output-dir="/path/to/generated/bindings"
```

This option might require setting the correct`--import-prefix` as well.

#### Options

The list of binapi-generator optional arguments in form `binapi-generator [OPTIONS] ARGS`

- `binapi-generator -version` prints the generator version, for example, `govpp v0.8.0-dev`
- `binapi-generator -gen=http` injects the `http` plugin to generate additional files.
  More plugins are separated with comma, i.e. `-gen-http,rpc`. Note that `rpc` plugin is used by default
- `binapi-generator -import-prefix=/desired/prefix/path` sets the go package name to be used in the generated bindings (
  the string used in imports)
- `binapi-generator -input-dir=/vpp/api/input/dir` sets the custom input directory instead of the default
- `binapi-generator -output-dir=/bin/api/output/dir` sets the custom output directory. It may or may not match
  the `-import-prefx`, based on go.mod.
- `binapi-generator -debug` prints some additional logs

### Generator Plugins

The binary API generator supports extending its functionality with plugins that can generate additional files.

Available built-in plugins:

- `http` generates HTTP handlers (more information in the [HTTP service part](#http-service))
- `rpc` generates RPC services (more information in the [RPC service part](#rpc-client))

## VPP Startup

Define the minimal `startup.conf`:

```sh
unix {interactive}
socksvr { socket-name /var/run/vpp/api.sock }
```

Start with `startup.conf` if the VPP was installed from sources:

```sh
<vpp_repo_dir>/build-root/install-vpp_debug-native/vpp/bin/vpp -c /tmp/startup.conf
```

If it was installed from the package:

```sh
vpp -c /tmp/startup.conf
```

## VPP API calls

GoVPP client uses its own (pure Go) implementation to handle the low-level parts of the VPP API communication.

### Connection

Two connection types to the binary API socket are exist - synchronous and asynchronous. GoVPP can connect to more VPPs
by creating multiple connections (see the [multi-vpp example](../examples/multi-vpp/README.md))

#### Synchronous Connect

The synchronous connecting creates a new adapter instance. The call blocks until the connection is established.

```go
conn, err := govpp.Connect(socketPath)
if err != nil {
  // handle the error
}
defer conn.Disconnect()
```

#### Asynchronous Connect

The main difference between the synchronous and asynchronous connection is that the asynchronous connection does not
block until the connection is ready. Instead, the caller receives a channel to watch connection events. Asynchronous
connection defines additional parameters `attemptNum` and `interval` defining the number of reconnecting attempts, and
an interval (in seconds) between those attempts.

```go
conn, connEv, err := govpp.AsyncConnect(socketPath, attemptNum, interval)
if err != nil {
  // handle error
}
defer conn.Disconnect()

// wait for the connection event
e := <-connEv
if e.State != core.Connected {
  // handle error in e.Error
}
```

### Sending API messages

Each binary API message in the Go-generated API is a data structure. The caller can send API messages either using
a `Channel` (legacy method) or a `Stream`. 
In-depth discussion about differences between `Channel` and `Stream` can be found https://github.com/FDio/govpp/discussions/43.

Messages can be requests, replies or events. The request might expect just one response (request), or more than one
response (multirequest). It is possible to determine the message type out of its name.

> **Note**
> The naming might not be consistent across the VPP API.

* *_Requests_* have no special suffix for the request, or `Dump` or `Get` for the multirequest.
* *_Responses_* have a `Reply` suffix for the request or `Details` for multirequest.

#### Stream client

The `Stream` is the new and preferred way to call VPP API. It provides a "low-level" API to allow complete control of
the communication process. This means that it requires the caller to handle all aspects of VPP API semantics, e.g. send
control ping after dump request, check the message type and value of `Retval` field in responses.

*_Note: handling of the VPP API semantics mentioned above is only required when using `Stream` directly. Users should
use generated RPC clients which handle all of this automatically._*

New `Stream` can be created by calling `Connection`'s method `NewStream`:

```go
stream, err := conn.NewStream(context.Background(), options...)
if err != nil {
  // handle error
}
```

The `NewStream` method also accepts following options:

* `WithRequestSize(size int)` sets the size of the request channel buffer
* `WithReplySize(size int)` sets the size of the reply channel buffer
* `WithReplyTimeout(timeout time.Duration)` sets the reply timeout

The single request procedure requires the user to convert the generic reply (the `Message` interface type) to the proper
reply type.

```go
req := &interfaces.CreateLoopback{
   // fill with data
}
if err := stream.SendMsg(req); err != nil {
  // handle error
}
replyMsg, err := stream.RecvMsg()
if err != nil {
  // handle error
}
reply := replyMsg.(*interfaces.CreateLoopbackReply)
```

The simpler way is to use the `Invoke()` method, which does the same procedure as above.

```go
req := &interfaces.CreateLoopback{
   // fill with data
}
var reply interfaces.CreateLoopbackReply
err := c.conn.Invoke(context.Background(), req, &reply)
if err != nil {
  // handle error
}
```

The multirequest message must be followed up by the control ping request. The loop collecting replies must watch for the
control ping reply signalizing the end.

```go
   if err := stream.SendMsg(&interfaces.SwInterfaceDump{}); err != nil {
     // handle error
   }
   if err := stream.SendMsg(&memclnt.ControlPing{}); err != nil {
     // handle error
   }

Loop:
   for {
      reply, err := stream.RecvMsg()
      if err != nil {
         // handle error
      }
      switch reply.(type) {
         case *interfaces.SwInterfaceDetails:
            // handle the message
         case *memclnt.ControlPingReply:
            break Loop
         default:
            // unexpected message type
      }
   }
```

There is another type of the multirequest, and this one needs to be handled differently. It can be easily identified
with the `Get` suffix and the **cursor** parameter. This type of multirequest is read in batches. Each batch has a
number of data replies, and if the reader should continue with another batch, the cursor points to the next index.

The user has to expect two reply types, the `Details` containing the data, and the `Reply` marking the end of the batch,
and the new cursor value.

The initial batch always starts with zero.

```go
cursor uint32 = 0
for {
   if cursor == ^uint32(0) { // cursor pointing to this value marks the end of the multirequest
      return
   }
   if err := stream.SendMsg(&pnat.PnatBindingsGet{
      Cursor: cursor,
   }); err != nil {
      // handle error
   }
   // new batch loop
   for {
      msg, err := stream.RecvMsg()
      if err != nil {
         // handle error
      }
      switch reply := msg.(type) {
         case *pnat.PnatBindingsDetails:
            // handle the data
         case *pnat.PnatBindingsGetReply:
            cursor = reply.Cursor // the new cursor value
            // handle error
            break
         default:
            // unexpected reply type
      }
   }
}
```

#### Channel

> **Warning**
> The Channel are planned to be deprecated in the future.

The `Channel` is the main communication unit between the caller and the VPP. After the successful connection, the
channel is simply created from the connection object.

```go
ch, err := conn.NewAPIChannel()
if err != nil {
   // handle error
}
```

The new channel starts watching caller requests immediately.

The channel can do a compatibility check for all the messages from any generated VPP API file.

```go
if err := ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
   // handle error
}
```

A single request is done by calling the asynchronous request/reply on a channel. The request returns the request context
allowing to receive the reply.

```go
req := &interfaces.CreateLoopback{} // fill with data
reply := &interfaces.CreateLoopbackReply{}
if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
   // handler error
}
```

The multirequest expects more than one response. Its context contains information about the last item in the request
list.

```go
req := &interfaces.SwInterfaceDump{} // fill with data
reqCtx := ch.SendMultiRequest(req)
for {
   reply := &interfaces.SwInterfaceDetails{}
   stop, err := reqCtx.ReceiveReply(reply)
   if err != nil {
      // handle error
   }
   if stop {
      break
   }
   // handle the reply before the next iteration
}
```

> **Note**
> The multirequest message suffixed with `Get` needs a different type of handling using the stream client.

## HTTP Service

Before using the HTTP service, the VPP binary API binding must be generated with the `http` plugin to create HTTP files,
and the `rpc` plugin to create RPC files. RPC is used inside HTTP handler functions.

Create the new HTTP handler and listener (requires the connection instance, and the RPC service client)

```go
vpeRpcClient := vpe.NewServiceClient(conn)
vpeHttpClient := vpe.HTTPHandler(vpeRpcClient)
if err := http.ListenAndServe(":8000", vpeHttpClient); err != nil {
   // handle error
}
```

With the http server running, you can use `curl` command to call VPP API via HTTP request:

```sh
$ curl http://localhost:8000/show_version
{
  "program": "vpe",
  "version": "23.06-rc0~71-g182d2b466",
  "build_date": "2023-02-20T07:52:13",
  "build_directory": "/home/go/src/git.fd.io/vpp"
}
```

## RPC Client

The RPC client is a generated client implementation by generator plugin `rpc` in separate file named `*.rpc.ba` for each
generated file (package) of the VPP binary API.

Each generated VPP binary API file (package) contains RPC service interface `RPCService` and RPC client `serviceClient`
that implements the RPC service interface. For each service RPC method defined in VPP API file, the RPC service
interface defines a method and the RPC client implements it according to the service definition of the VPP binary API
file.

To create the new RPC service client, use function `NewServiceClient(conn api.Connection)`, which accepts the connection
instance.

```go
nat64RpcClient := nat64.NewServiceClient(conn)
vpeRpcClient := vpe.NewServiceClient(conn)
```

Simple request:

```go
c := memif.NewServiceClient(conn)
req := &memif.MemifCreate{} // fill with data
reply, err := c.MemifCreate(context.Background(), req)
if err != nil {
   // handle error
}
```

Multirequest (the end of the multirequest is determined by checking the `EOF` error):

```go
c := interfaces.NewServiceClient(conn)
stream, err := c.SwInterfaceDump(context.Background(), &interfaces.SwInterfaceDump{})
if err != nil {
   // handle error
}
for {
   iface, err := stream.Recv()
   if err == io.EOF {
      break
   }
   if err != nil {
      // handle error
   }
}
```

## VPP Stats

The *_statsclient_* adapter connects to the VPP `stats.sock`.

```go
statsClient = statsclient.NewStatsClient(socket, options...)
```

Supported options to customize the connection parameters.

* `SetSocketRetryPeriod` - specifies a **time period between retries**
* `SetSocketRetryTimeout` - specifies a **timeout duration for retry**

The stats connection can be done synchronously or asynchronously (as for the binary API socket).

The synchronous connection:

```go
statsConn, err = core.ConnectStats(statsClient)
if err != nil {
   // handle error
}
defer c.Disconnect()
```

The asynchronous connection:

```go
statsConn, statsChan, err = core.AsyncConnectStats(statsClient, attemptNum, interval)
if err != nil {
   // handle error
}
e := <-statsChan
if e.State != core.Connected {
   // handle error
}
defer statsConn.Disconnect()
```

The asynchronous connection comes with additional parameters `attemptNum` and `interval` defining the number of
reconnection attempts, and an interval (in seconds) between those attempts.

The `stats` connection implements the `StatsProvider` interface and gives access to various stats getters.

```go
systemStats := new(api.SystemStats)
if err := statsConn.GetSystemStats(systemStats); err != nil {
   // handle error
}
nodeStats := new(api.NodeStats)
if err := statsConn.GetNodeStats(nodeStats); err != nil {
   // handle error
}
...
```

### Low-level stats API connection

The `StatsProvider` is considered the "high-level" API with easy access to VPP stats, but it might be restricted in
certain cases. Using the stats client directly is the solution. The stats client implements the `StatsAPI`, the kind of
low-level API with broader options for managing stats.

In any case, make sure the stats client is connected.
The example above created the new stats client and used the GoVPP core method to connect, giving access to both,
the `statsClient` (low-level API) and the `statsConn` (high-level API).

```go
statsClient = statsclient.NewStatsClient(socket, options...)
statsConn, err = core.ConnectStats(statsClient) // or core.AsyncConnectStats(statsClient, attemptNum, interval)
   // handle error
}
defer statsConn.Disconnect()
```

If you do not need the `statsConn` instance and just use the low-level API, you must connect anyway. Either by ignoring
the instance, or calling the connect method directly.

```go
statsClient = statsclient.NewStatsClient(socket, options...)
err = statsClient.Connect()
// handle error
}
defer statsClient.Disconnect()
```

### Low-level stats API usage

List all indexed stats directories. Use patterns to filter the output. The `StatIdentifier` is a simple name to the
index value.

```go
list, err = client.ListStats(patterns...)
if err != nil {
   // handle error
}
```

Return values of stats directories. Use patterns to filter the output. The `StatEntry` is the combination of stat
identifier, type, data, and the symlink flag.

```go
dump, err = client.DumpStats(patterns...)
if err != nil {
   // handle error
}
```

Prepare the stat directory. Returns the directory instance with accessible stat entries. Use patterns or indexes as a
filter.

```go
// pattern filter
dir, err := client.PrepareDir(patterns...)
if err != nil {
   // handle error
}

// index filter
dir, err := client.PrepareDirOnIndex(patterns...)
if err != nil {
   // handle error
}
```

Update the previously prepared directory. Suitable for polling VPP stats.

```go
if err := client.UpdateDir(dir); err != nil {
   if err == adapter.ErrStatsDirStale { // if the directory is stale, re-load it
      if dir, err = client.PrepareDir(patterns...); err != nil {
         // handle error
      }
      continue
   }
   // handle error
}
```
