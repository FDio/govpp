# Troubleshooting guide

## Table of contents

* [Enable debug logs](#enable-debug-logs)
  + [core](#core)
  + [msgid](#msgid)
  + [socketclient](#socketclient)
  + [msgtable](#msgtable)
  + [statsclient](#statsclient)
  + [proxy](#proxy)
  + [Binapi generator debug logs](#binapi-generator-debug-logs)
* [Connection problems](#connection-problems)
  + [Socket file does not exist](#socket-file-does-not-exist)
  + [Connection refused](#connection-refused)
  + [Permission denied](#permission-denied)
  + [The socket is not ready](#the-socket-is-not-ready)
* [Problems with sending/receiving messages](#problems-with-sendingreceiving-messages)
  + [Unknown message error](#unknown-message-error)
  + [VPP not replying to some messages](#vpp-not-replying-to-some-messages)
* [Message versioning, deprecated and in-progesss messages](#message-versioning-deprecated-and-in-progesss-messages)
  + [Message versioning](#message-versioning)
  + [Deprecated messages](#deprecated-messages)
  + [In-progress messages](#in-progress-messages)

---

## Enable debug logs

This section describes how to increase log verbosity to enable debug logs of various GoVPP components.

### core

Set the environment variable `DEBUG_GOVPP` to any non-empty value.

Example:
```
DEBUG_GOVPP=y ./bin/rpc-service
```

To verify that it works, check for the following line in the console output:
```
DEBU[0000] govpp: debug level enabled
```

### msgid

Assign value `msgid` to the variable `DEBUG_GOVPP`.

NOTE: To make msgid debugging work, you must also enable core debugging (in this case, it is enabled automatically, because `msgid` is a non-emtpy value).

Example:
```
DEBUG_GOVPP=msgid ./bin/rpc-service
```

It should output lines similar to this:
```
DEBU[0000] message "show_version" (show_version_51077d14) has ID: 1380
```

### socketclient

Assign value `socketclient` to the variable `DEBUG_GOVPP`.

Example:
```
DEBUG_GOVPP=socketclient ./bin/rpc-service
```

To verify that it works, check for the following line in the console output:
```
DEBU[0000] govpp: debug level enabled for socketclient
```

### msgtable

Assign value `msgtable` to the variable `DEBUG_GOVPP`.

NOTE: To make msgtable debugging work, you must also enable socketclient debugging.

NOTE: To assign multiple values to a variable, you can separate the values with a colon.

Example:
```
DEBUG_GOVPP=socketclient:msgtable ./bin/rpc-service
```

It should output lines similar to this:
```
DEBU[0000]  - 1380: "show_version_51077d14"              logger=govpp/socketclient
```

### statsclient

Set the environment variable `DEBUG_GOVPP_STATS` to any non-empty value.

NOTE: statsclient debugging only works if package `go.fd.io/govpp/adapter/statsclient` is imported.

Example:
```
DEBUG_GOVPP_STATS=y ./bin/vpp-proxy server
```

To verify that it works, check for the following line in the console output:
```
DEBU[0000] govpp/statsclient: enabled debug mode
```

### proxy

Set the environment variable `DEBUG_GOVPP_PROXY` to any non-empty value.

NOTE: proxy debugging only works if package `go.fd.io/govpp/proxy` is imported.

Example:
```
DEBUG_GOVPP_PROXY=y ./bin/vpp-proxy server
```

To verify that it works, check for the following line in the console output:
```
DEBU[0000] govpp/proxy: debug mode enabled
```

### Binapi generator debug logs

You can enable logs separately for vppapi parser and for binapi generator. The binapi generator has two verbosity levels.

To enable parser logs, assign value `parser` to the variable `DEBUG_GOVPP`.

To enable the first verbosity level for generator, pass the flag `-debug` when invoking the generator binary.

To enable the second verbosity level for generator, assign value `binapigen` to the variable `DEBUG_GOVPP`.

Examples:
```bash
# first verbosity level for generator
./bin/binapi-generator -debug
# maximum logs for all components
DEBUG_GOVPP=parser:binapigen ./bin/binapi-generator
```



## Connection problems

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

**NOTE**: The solutions apply only if you use the function `govpp.AsyncConnect`. If you instead use the function `govpp.Connect`, then GoVPP tries to connect immediately and only once, and if the socket does not exist, it immediately throws error. In this case, you have to ensure that the socket is ready before starting GoVPP.



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



## Message versioning, deprecated and in-progesss messages

### Message versioning

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
