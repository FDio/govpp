# GoVPP user guide

* [Binary API generator](#binary-api-generator)
    * [Installation](#Installation)
        * [Using the Go toolchain](#using-the--go-toolchain)
        * [From the source](#from-the-source)
    * [Plugins](#plugins)
    * [Options](#options)
* [VPP API calls](#vpp-api-calls)
    * [Connection](#connection)
        * [Synchronous](#synchronous-connect)
        * [Asynchronous](#asynchronous-connect)
    * [Sending API messages](#sending-api-messages)
        * [Channel](#channel)
        * [Stream client](#stream-client)
* [The RPC service](#the-rpc-service)
* [VPP stats](#vpp-stats)
    * [Low-level API connection](#low-level-stats-api-connection)
    * [Low-level API usage](#low-level-stats-api-usage)

## Binary API generator

The binary API generator's purpose is to create Go bindings out of the VPP `.json` binary API definitions.

### Installation

Requires Go 1.18 or higher ([download](https://golang.org/dl))

#### Using the  Go toolchain

Latest version (most recent tag):

```
go install go.fd.io/govpp/cmd/binapi-generator@latest
```

Development version (master branch):

```
go install go.fd.io/govpp/cmd/binapi-generator@master
```

#### From the source

Clone the repository:

```
git clone https://github.com/FDio/govpp
```

Call the following make taget inside:

```
make install-generator
```

# Generate API bindings

At first, you need VPP JSON API bindings to build go bindings from.
Build the VPP locally, or download
from `packagecloud.io` ([learn more](https://fd.io/docs/vpp/master/gettingstarted/installing)).

JSON API files are located in `/usr/share/vpp/api/`

```
$ make generate-binapi 
# installing binapi-generator v0.8.0-alpha-6-g2054a76-dirty
# generating binapi
binapi-generator -input-dir=/usr/share/vpp/api -output-dir=. -gen=rpc
INFO[0000] resolved VPP version from installed package: 23.06-rc0~65-g2ddb2fdaa
INFO[0000] Generating 236 files                         
binapi-generator -input-file=/usr/share/vpp/api/core/vpe.api.json -output-dir=. -gen=http
INFO[0000] resolved VPP version from installed package: 23.06-rc0~65-g2ddb2fdaa
INFO[0000] Generating 3 files   
```

Generated files can be found in `/binapi` package.

### Plugins

The binary API generator supports custom plugins generating additional files.

Available plugins:

- `http` generates HTTP handlers
- `rpc` generates RPC services

### Options

- `version` prints the generator version, for example, `govpp v0.8.0-dev`
- `gen` allows injecting generator plugin (`rpc` is the default)
- `import-prefix` generates files with custom import prefix
- `input-dir` sets the custom input directory instead of the default
- `output-dir` sets the custom output directory
- `debug` prints some additional logs

## VPP API calls

GoVPP uses adapters to talk to VPP sockets.

* **socketclient** connects to the VPP `api.sock` and supports all VPP binary APIs
* **statsclient** connects to the VPP `stats.sock` and reads VPP stats

### Connection

Two connection types to the binary API socket exist - synchronous and asynchronous. GoVpp can connect to more VPPs by
creating multiple connections.

#### Synchronous connect

The synchronous connect creates a new adapter instance. The call blocks until the connection is established.

```
conn, err := govpp.Connect(socketPath)
if err != nil {
    // handle the error
}
defer conn.Disconnect()
```

#### Asynchronous connect

The main difference between the synchronous and asynchronous connection is that the asynchronous connection does not
block until the connection is ready. Instead, the caller receives a channel to watch connection events.

```
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

You might notice additional parameters `attemptNum` and `interval` defining the number of reconnect attempts, and an
interval (in seconds) between those attempts.

### Sending API messages

Each binary API message in the Go-generated API is a data structure. The caller can send API messages either using
the `Channel`, or the `Stream client`.

Messages can be requests or responses. The request might expect just one response (request), or more than one
response (multirequest). It is possible to determine the message type out of its name.

* *_Requests_* have no special suffix for the request, or `Dump` or `Get` for the multirequest.
* *_Responses_* have a `Reply` suffix for the request or `Details` for multirequest.

#### Channel

The `Channel` is the main communication unit between the caller and the VPP. After the successful connection, the
channel is simply created from the connection object.

```
ch, err := conn.NewAPIChannel()
if err != nil {
    // handle error
}
```

The new channel starts watching caller requests immediately.

The channel allows direct compatibility check of messages from any generated Go binary API file.

```
if err := ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
    // handle error
}
```

A single request is done by calling the asynchronous request/reply on a channel. The request returns the request context
allowing to receive the reply.

```
req := &interfaces.CreateLoopback{} // fill with data
reply := &interfaces.CreateLoopbackReply{}
if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
    // handler error
}
```

The multirequest expects more than one response. Its context contains information about the last item in the request
list.

```
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

The `Stream` is an alternative "low-level" API offering better control of the communication process but requires the
user to handle the message type and order.

The `Stream` client is created from the connection instance (as the channel is):

```
stream, err := conn.NewStream(context.Background(), options...)
if err != nil {
   // handle error
}
```

The stream can be customized with the following options:

* `WithRequestSize` sets the size of the request channel buffer
* `WithReplySize` sets the size of the reply channel buffer
* `WithReplyTimeout` sets the reply timeout

The single request procedure requires the user to convert the generic reply (the `Message` interface type) to the proper
reply type.

```
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

```
req := &interfaces.CreateLoopback{} // fill with data
reply, err := stream.RecvMsg()
err := c.conn.Invoke(context.Background(), req, reply)
if err != nil {
    // handle error
}
```

The multirequest message must be followed up by the control ping request. The loop collecting replies must watch for the
control ping reply signalizing the end.

```
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

```
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

## The RPC service

Before using the RPC service, the VPP binary API binding must be generated with the `rpc` plugin to create RPC files.

Each RPC service file defines the `RPCService` interface with methods available for the given file.

Create the new RPC service client (requires the connection instance)

```
nat64RpcClient := nat64.NewServiceClient(conn)
vpeRpcClient := vpe.NewServiceClient(conn)
```

Simple request:

```
c := memif.NewServiceClient(conn)
req := &memif.MemifCreate{} // fill with data
reply, err := c.MemifCreate(context.Background(), &req)
if err != nil {
    // handle error
}
```

Multirequest (the end of the multirequest is determined by checking the `EOF` error):

```
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

```
statsClient = statsclient.NewStatsClient(socket, options...)
```

Supported options to customize the connection parameters.

* `SetSocketRetryPeriod` specifies the custom retry period while waiting for the VPP socket
* `SetSocketRetryTimeout` specifies the custom retry timeout while waiting for the VPP socket

The stats connection can be done synchronously or asynchronously (as for the binary API socket).

The synchronous connection:

```
statsConn, err = core.ConnectStats(statsClient)
if err != nil {
    // handle error
}
defer c.Disconnect()
```

The asynchronous connection:

```
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

The stats connection implements the `StatsProvider` interface and gives access to various stats getters.

```
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

```
statsClient = statsclient.NewStatsClient(socket, options...)
statsConn, err = core.ConnectStats(statsClient) // or core.AsyncConnectStats(statsClient, attemptNum, interval)
    // handle error
}
defer statsConn.Disconnect()
```

If you do not need the `statsConn` instance and just use the low-level API, you must connect anyway. Either by ignoring
the instance, or calling the connect method directly.

```
statsClient = statsclient.NewStatsClient(socket, options...)
err = statsClient.Connect()
    // handle error
}
defer statsClient.Disconnect()
```

### Low-level stats API usage

List all indexed stats directories. Use patterns to filter the output. The `StatIdentifier` is a simple name to the
index value.

```
list, err = client.ListStats(patterns...)
if err != nil {
    // handle error
}
```

Return values of stats directories. Use patterns to filter the output. The `StatEntry` is the combination of stat
identifier, type, data, and the symlink flag.

```
dump, err = client.DumpStats(patterns...)
if err != nil {
    // handle error
}
```

Prepare the stat directory. Returns the directory instance with accessible stat entries. Use patterns or indexes as a
filter.

```
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

```
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

