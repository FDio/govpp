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

## 0.13.0

> _13 November 2025_

### Changes

- Add VPP 25.10 (#318)
- Update ControlPing/Reply in core (#291)
- buffered events and notifications channel (#310)
- optimize Unsubscribe with the latest element in msgSubscriptions (#321)

### Fixes
- Upgrade tablewriter and fix API changes (#298)
- Fix stats connection panic (#309)
- statsclient: fix race between reconnect() and access (#305)
- do not build debug logger entries until necessary (#301)

## 0.12.0

> _14 May 2025_

### Changes

- Add VPP 25.02 (#282)
- Add binapi benchmarks (#269)
- Docker image and release improvements (#242)
- Enhance logging and optimize some usage (#281)

### Fixes

- Send `vl_api_sockclnt_delete_t` message on disconnect (#248)

## 0.11.0

> _19 September 2024_

### Changes

- Minor changes (#223)
- Add well-known reply timeout error (#234)

### GoMemif

- Gomemif improvemets (#228)
- Improves gomemif library (#216)

### CI

- Add Release workflow with goreleaser (#202)
- Add Login to ghcr step to release Workflow (#212)
- Extras integration testing (#233)
- lint: check markdown files (#238)

## 0.10.0

> _03 April 2024_

### Changes

- Add missing VPPApiError (#183)
- Trace refactored (#116)

### Fixes

- Fix stream timer leak (#182)
- Memory leaks in govpp core (#184)

## 0.9.0

> _04 January 2024_

### Changes

- Update run_integration.sh (#151, #154)
- Add VPP 23.10 to CI (#175)
- Improvements GoVPP CLI (#156)

### Fixes

- Fix running integration tests outside of GitHub (#168)
- Add check for stats vector length (#172)

### GoMemif 

- gomemif: Packet sent on Tx queue is received back on Rx queue of sender (#165)

## 0.8.0

> _18 July 2023_

### Changes

- Add VPP 23.06 to CI  (#136)
- Improvements for GoVPP CLI (#135)
- Skip running CI tests for docs updates (#137)
- binapigen: initial support for counters and paths (#121)
- Refactor resolving VPP API input  (#130)
- Update Dockerfile.integration (#134)
- Add README for examples (#128)
- Invalidate msgTable map during reconnect (#127)
- Add more GoVPP CLI features (#117)
- Update README.md (#115)
- Generate message comments (#109)
- Add User Guide (#110)
- Enhancements for binapigen (#93)
- Update Go and test if binapi is up-to-date (#105)
- Create RELEASE document (#89)

### Fixes

- Fix returning message reply on retval errors (#147)
- Fix memory leak for timers (#138)
- Fix channel pool (#131)
- Fix race in statsclient during reconnect (#126)
- Fix memory leak with reply timers (#124)
- Fix for Dockerfile smell DL3008 (#123)
- Fix disconnect for AsyncConnect case (#106)
- Fix binapi generation if old binapi files are not present (#73)

### GoMemif

- Fix memif abstract socket support (#119)
- Handle EINTR error (#99)


## 0.7.0

> _29 November 2022_

### Changes

- Switch to using a generic pool for channels in Connection (#39)
- feat: Disable default reply timeout (#45)
- gomemif: migrate from vpp repository (#65)
- Run golangci-lint in CI (#69)
- Add GoVPP logo to README (#79)
- Introduce VPP integration tests (#75)
- docs: Add troubleshooting guide (#70)
- Add support of watching events to Connection (#80)


### Fixes

- Fix format of package comment (#44)
- Fixes for staticcheck (#27)
- Prevent data race on msgMapByPath map (#50)
- Fix panic on the pool put/get operations (#54)
- Fix endless loop in Reset (#57)
- Fix generate RPC client for stream with reply (#72)
- Fix endless Stream's reply waiting on vpp disconnect (#77)
- Fix WaitReady if directory does not exist (#84)

### Changes

- Added GitHub CI
- Remove deprecated vppapiclient
- Remove unused Travis CI
- Remove gerrit remains

## 0.6.0

> _09 September 2022_

### Changes

- Added GitHub CI
- Remove deprecated vppapiclient
- Remove unused Travis CI
- Remove gerrit remains

### Fixes

- Call munmap when failed to check stat segment version (#34)

## 0.5.0

> _28 July 2022_

### Features

- Also generate APIName/APIVersion/CrcVersion constants for the types packages
- Stat segment client fixes & improvements

### Fixes

- Fix go 1.18 support
- Fixed data race in core.Connection.Close()
- Fix channel ID overlap

## 0.4.0

> _17 January 2022_

### Binapi Generator

- the generator code has been split into multiple packages:
    - [vppapi](binapigen/vppapi) - parses VPP API (`.api.json`) files
    - [binapigen](binapigen) - processes parsed VPP API files and handles code generation
- added support for VPP enumflags type
- previously required manual patches for generated code should no longer be needed
- many generated aliases were removed and referenced to `*_types` files for simpler reading
- any types imported from other VPP API (`*_types.api`) files are now automatically resolved for generated Go code
- marshal/unmarshal methods for memory client messages are now generated
- generated new helper methods for more convenient IP and MAC address conversion
- RPC service code is now generated into a separated file (`*_rpc.ba.go`) in the same directory and uses a new low level
  stream API
- added option to generate HTTP handlers for RPC services
- generated code now contains comments with information about versions of VPP and binapi-generator
- in addition to the file name, the binapi generator now accepts full path (including extension,
  e.g. `/usr/share/vpp/api/core/vpe.api.json`)
- dependency on `github.com/lunixbochs/struc` was removed
- generated helper methods for `vpe_types.Timestamp`
- generated API messages comment may contain additional information about the given API development state (in progress,
  deprecated)
- type `[]bool` is now known to the generator
- enhanced VPP version resolution - the generator is more reliable to evaluate installed VPP version

### Features

- [socketclient](adapter/socketclient) was optimized and received a new method to add client name
- added new API [stream](api/api.go) for low-level access to VPP API
    - the `Stream` API uses the same default values as the `Channel` API
    - it supports the same functional options (request size, reply size, reply timeout)
- [statsclient](adapter/statsclient) supports additional options (retry period, retry timeout)
- the compatibility check error now contains a list of compatible and incompatible messages wrapped
  as `CompatibilityError`
- removed global binary API adapter. This change allows GoVPP to manage multiple VPP connections with different sockets
  simultaneously
- added support for the stats v2. The `statsclient` adapter recognizes the version automatically so the `StatsAPI`
  remains unchanged. In relation to this change, the legacy support (i.e. stat segment v0) for VPP <=19.04 was dropped.
- GoVPP now recognizes VPP state `NotResponding` which can be used to prevent disconnecting in case the VPP hangs or is
  overloaded
- added method `SetLogger()` for setting the global logger
- `StatsAPI` has a new method `GetMemory()` retrieving values related to the statseg memory heap
- the stats socket allows an option to connect asynchronously in the same manner as the API socket connection.
  Use `AsyncConnectStats()` to receive `ConnectionEvent` notifications
- Instead of a cumulative value, the `statsclient` error counter now provides a value per worker
- `ListStats(patterns ...string)` returns `StatIdentifier` containing the message name and the ID (instead of the name
  only)
- added support for VPP stat symlinks
- GoVPP received its own [API trace](api/trace.go). See the [example](examples/api-trace) for more details.

### Fixes

- `MsgCodec` will recover panic occurring during a message decoding
- calling `Unsubscibe` will close the notification channel
- GoVPP omits to send `sockclnt_delete` while cleaning up socket clients in order to remove VPP duplicated close
  complaints - VPP handles it itself
- fixed major bug causing GoVPP to not receive stats updates after VPP restart
- fixed name conflict in generated union field constructors
- the size of unions composed of other unions is now calculated correctly
- fixed race condition in the VPP adapter mock
- fixed crash caused by the return value of uint kind
- fixed encoding/decoding of float64
- fixed the stats reconnect procedure which occasionally failed re-enable the connection.
- fixed occasional panic during disconnect
- `statsclient` wait for socket procedure works properly
- fixed memory leak in health check

### Other

- improved log messages to provide more relevant info
- updated extras/libmemif to be compatible with the latest VPP changes
- default health check parameter was increased to 250 milliseconds (up from 100 milliseconds), and the default threshold
  was increased to 2 (up from 1)
- improved decoding of message context (message evaluation uses also the package path, not just the message ID)
- `statsclient` now recognizes between empty stat directories and does not treat them as unknown
- proxy client was updated to VPP 20.05

#### Examples

- added more code samples of working with unions in [binapi-types](examples/binapi-types)
- added profiling mode to [perf bench](examples/perf-bench) example
- improved [simple client](examples/simple-client) example to work properly even with multiple runs
- added [multi-vpp](examples/multi-vpp) example displaying management of two VPP instances from single application
- added [stream-client](examples/stream-client) example showing usage of the new stream API
- [simple client](examples/simple-client) and [binapi-types](examples/binapi-types) examples show usage of
  the `vpe_types.Timestamp`
- added [API trace example](examples/api-trace)

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
