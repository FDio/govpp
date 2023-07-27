# Troubleshooting Guide

This document provides guidance for troubleshooting issues or any debugging for GoVPP.

---

**Table of contents**

* [Debug Logging](#debug-logging)
  + [Enable Debug](#enable-debug)
  + [Debug Components](#debug-components)
    - [Debug Message IDs](#debug-message-ids)
    - [Debug Socket Client](#debug-socket-client)
    - [Debug Binapi Generator](#debug-binapi-generator)
    - [Debug Stats Client](#debug-stats-client)
    - [Debug Proxy](#debug-proxy)
* [Connection Problems](#connection-problems)
  + [Socket file does not exist](#socket-file-does-not-exist)
  + [Connection refused](#connection-refused)
  + [Permission denied](#permission-denied)
  + [The socket is not ready](#the-socket-is-not-ready)
* [Problems with sending/receiving messages](#problems-with-sendingreceiving-messages)
  + [Unknown message error](#unknown-message-error)
  + [VPP not replying to some messages](#vpp-not-replying-to-some-messages)
* [Message versioning](#message-versioning)
  + [Deprecated messages](#deprecated-messages)
  + [In-progress messages](#in-progress-messages)

---

## Debug Logging

This section describes ways to enable debug logs for GoVPP and its components.
Most of the debug logs is controlled via the environment variable `DEBUG_GOVPP`.

### Enable Debug

To enable debug logs for GoVPP core, set the environment variable `DEBUG_GOVPP` to a **non-empty value**.

```sh
DEBUG_GOVPP=y ./app
```

The following line will be printed to _stderr_ output during program initialization:

```
DEBU[0000] govpp: debug level enabled
```

### Debug Components

The environment variable `DEBUG_GOVPP` can be used also to enable debug logs for other GoVPP components separately. 

> **Note**
> Multiple components can be enabled silmutaneously, by separating values with comma, e.g. `DEBUG_GOVPP=socketclient,msgtable`

#### Debug Message IDs

To enable debugging of message IDs, set `DEBUG_GOVPP=msgid`.

*log output sample:*
```
DEBU[0000] message "show_version" (show_version_51077d14) has ID: 1380
```

#### Debug Socket Client

To debug `socketclient` package, set `DEBUG_GOVPP=socketclient`.

*log output sample:*
```
DEBU[0000] govpp: debug level enabled for socketclient
```

Additionally, to debug the message table retrieved upon connecting to socket, set `DEBUG_GOVPP=socketclient,msgtable`.

*log output sample:*
```
DEBU[0000]  - 1380: "show_version_51077d14"              logger=govpp/socketclient
```

#### Debug Binapi Generator

To enable the basic debug logs in the binapi generator, pass the flag `-debug` to the program.

Additionaly, to enable debug logs for the _parser_, assign value `DEBUG_GOVPP=parser` and
to enable debug logs for the _generator_, set `DEBUG_GOVPP=binapigen`.

Examples:
```bash
# Enable debug logs
./bin/binapi-generator -debug

# Maximum verbosity
DEBUG_GOVPP=parser,binapigen ./binapi-generator
```

#### Debug Stats Client

To enable debug logs for `statsclient` package, set `DEBUG_GOVPP_STATS` to a non-empty value.

```sh
DEBUG_GOVPP_STATS=y ./stats-client
```

To verify that it works, check for the following line in the console output:
```
DEBU[0000] govpp/statsclient: enabled debug mode
```

#### Debug Proxy

To enable debug logs for `proxy` package, set `DEBUG_GOVPP_PROXY` to a non-empty value.

```sh
DEBUG_GOVPP_PROXY=y ./vpp-proxy server
```

To verify that it works, check for the following line in the console output:

```
DEBU[0000] govpp/proxy: debug mode enabled
```

## Connection Problems

This section describes various problems when GoVPP tries to connect to VPP.

### Socket file does not exist

**Symptoms**: You see the following error:

```
ERROR: connecting to VPP failed: VPP API socket file /run/vpp/api.sock does not exist
```

**Probable cause**: VPP is not running or it is configured to have the socket at non-default path.

**Possible solutions**:

- Check if the VPP is running. It may have exited unexpectedly.

- Run the VPP with `sudo`. It may need root permissions to create the socket.

- Configure the socket path in VPP configuration. Find the file `startup.conf` (default location: `/etc/vpp/startup.conf`) or create it if it does not exist and add there the following:
  ```
  socksvr {
      socket-name /path/to/api.sock
  }
  ```
  (change the path to the actual path, the default is `/run/vpp/api.sock`). Then run the VPP with the `-c` flag and specify the path to the configuration file. For example, with default locations: `/usr/bin/vpp -c /etc/vpp/startup.conf`.

  - If you built VPP from source, you can apply the configuration using `make` with `STARTUP_CONF` environment variable. For example: `STARTUP_CONF=/etc/vpp/startup.conf make run`.

  - If VPP outputs a line similar to this: `clib_socket_init: bind (fd 9, '/home/user/vpp/api.sock'): No such file or directory`, it likely means that the directory for the socket does not exist. If this happens, create the directory manually and restart VPP.

- In GoVPP code that connects to the VPP, use the actual socket path configured by VPP:
  ```go
  govpp.Connect("/path/to/api.sock")
  ```
  (change the path to the actual path).
  
  - In GoVPP examples, you can usually set the path with the flag `-socket` or similar (use `--help` to see the actual flag). For example: `./bin/rpc-service -sock /run/vpp/api.sock`.

- If you are running VPP inside Docker container, mount the socket directory as a Docker volume so it can be used from the host. If you use configuration file, also mount the file. For example: `docker run -it -v /run/vpp:/run/vpp -v /home/user/vpp/startup.conf:/etc/vpp/startup.conf ligato/vpp-base:22.10`

- Increase the number of connection retries or timeout. See the subsection *The socket is not ready*.

### Connection refused

**Symptoms**: You see the following error:

```
ERROR: connecting to VPP failed: dial unix /run/vpp/api.sock: connect: connection refused
```

**Probable cause**: VPP exited or its configuration changed and you are trying to connect to the now obsolete socket.

**Possible solutions**: See the solutions in the previous subsection.

### Permission denied

**Symptoms**: You see the following error:

```
ERROR: connecting to VPP failed: dial unix /run/vpp/api.sock: connect: permission denied
```

**Probable cause**: The socket is owned by the `root` group.

**Possible solutions**:

- Configure the VPP to create the socket as current user's group. Add the following to the file `startup.conf`:
  ```
  unix {
  	gid <ID>
  }
  ```
  (replace the `<ID>` with actual group ID; you can see the ID with `id -g`).

  - If you are running VPP inside Docker container, run the container as current user's group (but keep the user as root), for example `docker run -it -v /run/vpp:/run/vpp -u 0:$(id -g) ligato/vpp-base:22.10`. In this case, do NOT use the `unix { gid ... }` in `startup.conf`.
- Run GoVPP with `sudo`.

- Manually change permissions or ownership of the socket file.

### The socket is not ready

**Example scenario**: User has built a Docker image with a VPP and a custom binary that uses GoVPP. When he runs it, the GoVPP binary starts up instantly, but it takes a long time for the VPP to apply initial configuration and start up. The GoVPP binary tries to connect before the VPP is ready, so it fails to connect and exits. When the user inspects it, he finds out that the VPP is running and the GoVPP binary is not, and may mistakenly think that the GoVPP binary is faulty. But actually the VPP was just not ready in time.

**Possible solutions**:
- Increase the number of connection retries (the second argument, `attempts`, of the `govpp.AsyncConnect` function).
- Increase the interval between reconnects (the third argument, `interval`, of the `govpp.AsyncConnect` function).
- Increase the timeout for waiting for the socket existence (variable `MaxWaitReady` in the file `adapter/socketclient/socketclient.go`).

> **Note**
> The solutions apply only if you use the function `govpp.AsyncConnect`. If you instead use the function `govpp.Connect`, then GoVPP tries to connect immediately and only once, and if the socket does not exist, it immediately throws error. In this case, you have to ensure that the socket is ready before starting GoVPP.

## Problems with sending/receiving messages

### Unknown message error

**Symptoms**: When you try to send/receive a message, you see `unknown message: ...` error.

**Cause**: Incompatibility of messages between VPP binary API and GoVPP's generated Go bindings for binary API (binapi).

**Solution 1**: Regenerate binapi. Use make target `generate-binapi` (for VPP installed via package manager) or `gen-binapi-local` (for locally built VPP) or `gen-binapi-docker` (independent of local VPP). For more information about the generator, see [the generator documentation](GENERATOR.md). After generating binapi, rebuild GoVPP.

**Solution 2**: Use VPP version that is compatible with GoVPP's binapi. To see for which VPP version the binapi was generated, look at the beginning of any `.ba.go` file (not `_rpc.ba.go`).

### VPP not replying to some messages

If VPP is not replying to a message, it is likely because one of the following reasons:
- The VPP has frozen or crashed.
- The VPP correctly processed the request and possibly applied the requested configuration (you may see the configured items in the VPP), but it did not send reply. This could be either because of a bug in VPP or because the message contains some illegal/unusual combination of parameters that VPP can not handle properly.

To solve it, you can try the following:
- Check if the VPP is frozen or crashed.
- Try sending the message with different values.

## Message versioning

When a GoVPP's message is updated, the message itself is not changed. Instead, a new version of the message is added that has a `v2` (or `v3`, `v4`, ...) suffix to its name. The old version of the message is then marked as deprecated.

### Deprecated messages

If a message is marked as deprecated, it should no longer be used. The message works fine for now, but it may be removed in the future. When a message is deprecated, there is likely available a new version of the message or an another message with a similar behaviour which should be used instead. Also, if some message is not deprecated, but there already exists a new version of the message, the new version should be used, as the older version may soon become deprecated.

For example, here is an excerpt from the file `binapi/vhost_user/vhost_user.ba.go` (as of 31-Oct-2022):

```go
// CreateVhostUserIf defines message 'create_vhost_user_if'.
// Deprecated: the message will be removed in the future versions
type CreateVhostUserIf struct {
	// ...
}

// ...

// CreateVhostUserIfReply defines message 'create_vhost_user_if_reply'.
// Deprecated: the message will be removed in the future versions
type CreateVhostUserIfReply struct {
	// ...
}

// ...

// CreateVhostUserIfV2 defines message 'create_vhost_user_if_v2'.
type CreateVhostUserIfV2 struct {
	// ...
}

// ...

// CreateVhostUserIfV2Reply defines message 'create_vhost_user_if_v2_reply'.
type CreateVhostUserIfV2Reply struct {
	// ...
}
```

The messages `CreateVhostUserIf` and `CreateVhostUserIfReply` are deprecated and should not be used. Instead, there are new versions of the messages: `CreateVhostUserIfV2` and `CreateVhostUserIfV2Reply`, which should be used.

### In-progress messages

When a new message is added, usually it is initially experimental and is marked as in-progress. In this case, an older message should be preferrably used. Later, when the message will be considered stable, the in-progress label will be removed, the message will become recommended to use and the old message may become deprecated.

For example, here is an excerpt from the file `binapi/ip/ip.ba.go` (as of 31-Oct-2022):

```go
// IPRouteAddDel defines message 'ip_route_add_del'.
type IPRouteAddDel struct {
	// ...
}

// ...

// IPRouteAddDelReply defines message 'ip_route_add_del_reply'.
type IPRouteAddDelReply struct {
	// ...
}

// ...

// IPRouteAddDelV2 defines message 'ip_route_add_del_v2'.
// InProgress: the message form may change in the future versions
type IPRouteAddDelV2 struct {
	// ...
}

// ...

// IPRouteAddDelV2Reply defines message 'ip_route_add_del_v2_reply'.
// InProgress: the message form may change in the future versions
type IPRouteAddDelV2Reply struct {
	// ...
}
```

In this case, the messages `IPRouteAddDel` and `IPRouteAddDelReply` should be used; the messages `IPRouteAddDelV2` and `IPRouteAddDelV2Reply` are in-progress.
