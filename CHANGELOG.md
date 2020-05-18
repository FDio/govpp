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
