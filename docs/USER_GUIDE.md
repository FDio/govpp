# GoVPP User Guide

### Table of Contents

* [Binary API generator](#binary-api-generator)
    * [Installation](#installation)
    * [Plugins](#plugins)
    * [Options](#options)
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

## Binary API generator

The binary API generator's purpose is to create Go bindings out of the VPP `.json` binary API definitions.

### Installation

Requires Go 1.18 or higher ([download](https://golang.org/dl))

Install using the Go toolchain:

```
// Latest version (most recent tag):
go install go.fd.io/govpp/cmd/binapi-generator@latest

// Specific version
go install go.fd.io/govpp/cmd/binapi-generator@v0.8.0

// Development version (the master branch):
go install go.fd.io/govpp/cmd/binapi-generator@master
```

Check the binapi generator is installed:

```
$ binapi-generator -version
govpp v0.8.0-dev
```

# Source of VPP API input files

At first, you need VPP JSON API bindings to build go bindings from. There are several ways to get it.

1. Download from `packagecloud.io` ([learn more](https://fd.io/docs/vpp/master/gettingstarted/installing)).
2. Clone the VPP repository `git clone https://github.com/FDio/vpp.git` and run `make json-api-files`. Built JSON files
   can be found in `/vpp/build-root/install-vpp-native/vpp/share/vpp/api/`.
3. If the VPP is already built, the default JSON API path is `/usr/share/vpp/api/`

# Generate VPP API bindings

If the VPP JSON API definitions are in the default directory `/usr/share/vpp/api`, call:

```
binapi-generator
```

Binding will be created in the current directory in `/binapi` package.

To modify the VPP JSON API input directory, use `-input-dir` option. To modify the output directory, use
the `-output-dir` option.

```
binapi-generator -input-dir=/path/to/vpp/api -output-dir=/path/to/generated/bindings
```

For the full list of binary API generator plugins and options, see sections below.

### Plugins

The binary API generator supports custom plugins generating additional files.

Optional built-in plugins:

- `http` generates HTTP handlers (more information in the [HTTP service part](#http-service))
- `rpc` generates RPC services (more information in the [RPC service part](#rpc-client))

### Options

The list of binapi-generator optional arguments in form `binapi-generator [OPTIONS] ARGS`

- `binapi-generator -version` prints the generator version, for example, `govpp v0.8.0-dev`
- `binapi-generator -gen=http` injects the `http` plugin to generate additional files.
  More plugins are separated with comma, i.e. `-gen-http,rpc`. Note that `rpc` plugin is used by default
- `binapi-generator -import-prefix=/desired/prefix/path` generates a custom prefix for imports in generated files
- `binapi-generator -input-dir=/vpp/api/input/dir` sets the custom input directory instead of the default
- `binapi-generator -output-dir=/bin/api/output/dir` sets the custom output directory
- `binapi-generator -debug` prints some additional logs

## VPP API calls

GoVPP uses adapters to talk to VPP sockets.

### Connection

Two connection types to the binary API socket are exist - synchronous and asynchronous. GoVpp can connect to more VPPs
by creating multiple connections (see the [multi-vpp example](../examples/multi-vpp/README.md))

#### Synchronous connect

The synchronous connecting creates a new adapter instance. The call blocks until the connection is established.

```go
conn, err := govpp.Connect(socketPath)
if err != nil {
// handle the error
}
defer conn.Disconnect()
```

#### Asynchronous connect

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
the `Channel` (legacy method), or the `Stream client` (preferred). In-depth discussion about differences
between `Channel` and `Stream` can be found [here](https://github.com/FDio/govpp/discussions/43).

Messages can be requests or responses. The request might expect just one response (request), or more than one
response (multirequest). It is possible to determine the message type out of its name.

*_Note: the naming might not be consistent across the VPP API_*

* *_Requests_* have no special suffix for the request, or `Dump` or `Get` for the multirequest.
* *_Responses_* have a `Reply` suffix for the request or `Details` for multirequest.

#### Channel

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

*_Note: the multirequest message suffixed with `Get` needs a different type of handling using the stream client._*

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

* `WithRequestSize` sets the size of the request channel buffer
* `WithReplySize` sets the size of the reply channel buffer
* `WithReplyTimeout` sets the reply timeout

The single request procedure requires the user to convert the generic reply (the `Message` interface type) to the proper
reply type.

```go
req := &interfaces.CreateLoopback{} // fill with data
if err := stream.SendMsg(req); err != nil {
// handle error
}
reply, err := stream.RecvMsg()
if err != nil {
// handle error
}
replyMsg := reply.(*interfaces.CreateLoopbackReply)
```

The simpler way is to use the `Invoke()` method, which does the same procedure as above.

```go
req := &interfaces.CreateLoopback{} // fill with data
reply, err := stream.RecvMsg()
err := c.conn.Invoke(context.Background(), req, reply)
if err != nil {
// handle error
}
```

The multirequest message must be followed up by the control ping request. The loop collecting replies must watch for the
control ping reply signalizing the end.

```go
req := &interfaces.SwInterfaceDump{}
if err := stream.SendMsg(req); err != nil {
// handle error
}
cpReq := &memclnt.ControlPing{}
if err := stream.SendMsg(cpReq); err != nil {
// handle error
}

Loop:
for {
msg, err := stream.RecvMsg()
if err != nil {
// handle error
}
switch msg.(type) {
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
req := &pnat.PnatBindingsGet{
Cursor: cursor,
}
if err := stream.SendMsg(req); err != nil {
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
})
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

With the http server running, use `curl` command with the given pattern to make the VPP API call

```
$ curl http://localhost:8000/show_version
{
  "program": "vpe",
  "version": "23.06-rc0~71-g182d2b466",
  "build_date": "2023-02-20T07:52:13",
  "build_directory": "/home/go/src/git.fd.io/vpp"
}
```

## RPC Client

The RPC client is a client implementation generated by generator plugin `rpc` in separate file named `*.rpc.ba` for each
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
reply, err := c.MemifCreate(context.Background(), &req)
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

* `SetSocketRetryPeriod` specifies the custom retry period while waiting for the VPP socket
* `SetSocketRetryTimeout` specifies the custom retry timeout while waiting for the VPP socket

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

