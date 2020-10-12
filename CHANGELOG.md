# Changelog

This file lists changes for the GoVPP releases.

<!-- TEMPLATE
### Fixes
-
### Features
-
### Other
-
-->

## 0.4.0 (in development)
> _NOT RELEASED YET_

### Binapi Generator
- generator code has been split into multiple packages:
  - [vppapi](binapigen/vppapi) - parses VPP API (`.api.json`) files
  - [binapigen](binapigen) - processes parsed VPP API and handles code generation
- previously required manual patches for generated code should no longer be needed
- many generated aliases were removed and referenced to `*_types` files for simpler reading
- any types imported from other VPP API (`*_types.api`) files are now automatically resolved for generated Go code
- marshal/unmarshal methods for memory client messages are now generated
- generated new helper methods for more convenient IP and MAC address conversion
- added option to generate HTTP handlers for RPC services
- RPC service code is now generated into a separated file (`*_rpc.ba.go`) in the same directory and uses low level
  stream API
- generated code now contains comment with information about versions of VPP and binapi-generator
- in addition to the file name, the binapi generator now accepts full path (including extension, e.g. `/usr/share/vpp/api/core/vpe.api.json`)
- dependency on `github.com/lunixbochs/struc` was removed

### Features
- [socketclient](adapter/socketclient) received a new method to add client name
- added list of compatible messages to `CompatibilityError`
- removed global binary API adapter - this change allows GoVPP to manage multiple VPP connections with different 
  sockets simultaneously
- added support for the stats v2. The statsclient adapter recognized the version automatically so the `StatsAPI`
  remained unchanged. In relation to this change, the legacy support (i.e. stat segment v0) for VPP <=19.04 was dropped.
- GoVPP now recognizes VPP state `NotResponding` which can be used to prevent disconnecting in case the VPP hangs
  or is overloaded
- added method `SetLogger` for setting the global logger

### Fixes
- `MsgCodec` will recover panic occurring during a message decoding
- calling `Unsubscibe` will close the notification channel
- GoVPP omits sending `sockclnt_delete` while cleaning up socket clients in order to remove VPP duplicate close
  complaints - VPP handles it itself
- fixed major bug causing GoVPP to not receive stats updates after VPP restart
- fixed name conflict in generated union field constructors
- size of unions composed of another unions is now calculated correctly
- fixed race condition in the VPP adapter mock

### Other
- improved log messages to provide more relevant info
- updated extras/libmemif to be compatible again
- default health check parameter was increased to 250 milliseconds (up from 100 milliseconds), and the default
  threshold was increased to 2 (up from 1)

#### Examples
- added more code samples of working with unions in [union example](examples/union-example)
- added profiling mode to [perf bench](examples/perf-bench) example
- improved [simple client](examples/simple-client) example to work properly even with multiple runs
- added [multi-vpp](examples/multi-vpp) example displaying management of two VPP instances from single
  application
- added [stream-client](examples/stream-client) example showing usage of the new stream API  

#### Dependencies
- updated `github.com/sirupsen/logrus` dep to `v1.6.0`
- updated `github.com/lunixbochs/struc` dep to `v0.0.0-20200521075829-a4cb8d33dbbe`

## 0.3.5
> _18 May 2020_

### Fixes
- statsclient: Fix stats data errors and panic for VPP 20.05

## 0.3.4
> _17 April 2020_

### Features
- binapi-generator: Format generated Go source code in-process

## 0.3.3
> _9 April 2020_

### Fixes
- proxy: Unexport methods that do not satisfy rpc to remove warning

## 0.3.2
> _20 March 2020_

### Fixes
- statsclient: Fix panic occurring with VPP 20.05-rc0 (master)

## 0.3.1
> _18 March 2020_

### Fixes
- Fix import path in examples/binapi

## 0.3.0
> _18 March 2020_

### Fixes
- binapi-generator: Fix parsing default meta parameter

### Features
- api: Improve compatibility checking with new error types:
  `adapter.UnknownMsgError` and `api.CompatibilityError`
- api: Added exported function `api.GetRegisteredMessageTypes()`
  for getting list of all registered message types
- binapi-generator: Support imports of common types from other packages
- binapi-generator: Generate `Reset()` method for messages
- binapi-generator: Compact generated methods

### Other
- deps: Update `github.com/bennyscetbun/jsongo` to `v1.1.0`
- regenerate examples/binapi for latest VPP from stable/2001

## 0.2.0
> _04 November 2019_

### Fixes
- fixed socketclient for 19.08
- fixed binapi compatibility with master (20.01-rc0)
- fixed panic during stat data conversion

### Features
- introduce proxy for remote access to stats and binapi
- optimizations for statclient

### Other
- migrate to Go modules
- print info for users when sockets are missing

## 0.1.0
> _03 July 2019_

The first release that introduces versioning for GoVPP.
