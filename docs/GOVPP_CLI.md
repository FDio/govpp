# GoVPP CLI

This document provides guide for the GoVPP CLI app.

---

## Installation

To install GoVPP CLI it is currently required to have Go installed on your system.

#### Prerequisites
 
- **Go** 1.19+ ([installation](https://go.dev/doc/install))

### Install with Go

To install from the latest GoVPP release, run:

```shell
go install go.fd.io/govpp/cmd/govpp@latest
```

To install the most recent development version, run: 

```shell
go install go.fd.io/govpp/cmd/govpp@master
```

## Getting Started

The GoVPP CLI is a powerful tool for managing your VPP API development, maintenance and integration. With features such as VPP API schema export, comparison, linting, breaking change detector, Go code bindings generation and running a proxy or HTTP service for the VPP API, the GoVPP CLI offers a comprehensive solution for the VPP API management. The GoVPP CLI is designed to integrate seamlessly with your existing workflow, so you can focus on what matters most: write _great VPP APIs_ and develop control-plane apps to use them. Whether you are working with a small, focused project or a large, complex system, the GoVPP CLI is the perfect choice. In the next few minutes, you will learn how to use the GoVPP CLI to easily execute VPP CLI commands, generate code, compare schemas, and serve VPP API as an HTTP service.

> **Note**
> We will assume that you have already installed the GoVPP CLI and the necessary dependencies in your `$PATH`. If you haven't, head on over to our installation guide first.

By the end of this Getting Started guide, you will have a strong understanding of the core components of the GoVPP CLI, including:

- Methods for selecting input source of VPP API
- Browse VPP API schema definition and show its contents
- Run linter check and export VPP API schema
- Compare multiple VPP API schemas and detect breaking changes
- Code generation of Go bindings for VPP API

## Before you begin

Let's check the version of GoVPP CLI you'll be using is up-to-date.

```sh
govpp --version
```

This will print the version of GoVPP CLI.

## Usage

The `govpp` command will print the usage help for the top-level commands and their subcommands by default.

```
   ______         _    _  _____   _____     govpp v0.8.0
  |  ____  _____   \  /  |_____] |_____]    user@machine (go1.20 linux/amd64)
  |_____| [_____]   \/   |       |          Mon Jul  3 12:13:24 CEST 2023
                                          
Usage:
  govpp [command]
  
Available Commands:
  cli         Send CLI via VPP API
  generate    Generate code
  help        Help about any command
  http        VPP API as HTTP service
  vppapi      Manage VPP API

Flags:
  -D, --debug             Enable debug mode
  -L, --loglevel string   Set logging level
      --color string      Color mode; auto/always/never
      --version           version for govpp

Use "govpp [command] --help" for more information about a command.
```

### Show VPP API contents

The `vppapi ls` command allows you to print the VPP API files and their specific contents
in various formats. This can be useful for debugging or for generating documentation.

Here's an example usage of the `vppapi ls` command:

```sh
# List VPP API files for default input
govpp vppapi ls
```

You can use the `--input` flag to specify the input source for the VPP API files.

```sh
# Use current directory as input source (local VPP repository)
govpp vppapi ls .
```

The default format prints the output data in a table, but you can also specify 
other output formats such as JSON, YAML or Go template using the `--format` flag.

```sh
# Print using common formats
govpp vppapi ls --format="json"
govpp vppapi ls --format="yaml"

# Print using a Go template
govpp vppapi ls --format='{{ printf "%+v" . }}'
```

You can use the `--show-contents`, `--show-messages`, `--show-raw`, and `--show-rpc` 
flags to show specific parts of the VPP API file(s).

```sh
# List VPP API contents
govpp vppapi ls --show-contents

# List VPP API messages
govpp vppapi ls --show-message

# List RPC services
govpp vppapi ls --show-rpc

# Print raw VPP API files
govpp vppapi ls --show-raw
```

You can also use the `--include-fields` and `--include-imported` flags to include 
message fields and imported types, respectively.

For more information on the available flags and options, use the `-h` or `--help` flag.

### Run linter for VPP API definitions

The `lint` command allows you to run linter checks for your VPP API files. This can help you catch issues early and ensure that your code follows best practices.

Here's an example usage of the `lint` command:

shell
```sh
govpp vppapi lint https://github.com/FDio/vpp.git
```

This command runs the linter checks on the `master` branch of official VPP repository and outputs any issues found.

You can use the `--input` flag to specify the input location for your VPP API files, such as a path to a VPP API directory or a local VPP repository.

You can use the `--help` flag to get more information about the available flags and options.

### Compare VPP API schemas

The `diff` command allows you to compare two VPP API schemas and lists the differences between them. This can be useful for detecting breaking changes between different versions of the API.

Here's an example usage of the `diff` command:

```sh
 govpp vppapi diff 'http://github.com/FDio/vpp.git#tag=v23.10' --against 'http://github.com/FDio/vpp.git#tag=v24.02'
```

This command compares the VPP API schema of version`v23.10` against the VPP API schema of veersion `v24.02` and lists the differences between them. 
The output shows related information details for each difference.

You can use the `--help` flag to get more information about the available flags and options.

> **Note**
> The `--against` flag is required and should point to an input source for the schema to compare against.

## Troubleshooting

If you run into any problems when executing some commands, you can use the `--debug` option to increase log verbosity to help when debugging the issue.
For the highest verbosity use `--log-level=trace`.


> **Warning**
> Be sure to open an issue on GitHub for any bug that you may encounter.
