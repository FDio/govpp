# GoVPP CLI

This document provides guide for the GoVPP CLI app.


## Installation of GoVPP CLI

### Prerequisites
 
- Go 1.19+

### Install

To install GoVPP CLI, run:

```shell
# Latest release
go install go.fd.io/govpp/cmd/govpp@latest

# Development version
go install go.fd.io/govpp/cmd/govpp@master
```

## Getting Started with the GoVPP CLI

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

## Usage

```
 _________     ___    _________________
 __  ____/_______ |  / /__  __ \__  __ \
 _  / __ _  __ \_ | / /__  /_/ /_  /_/ / 
 / /_/ / / /_/ /_ |/ / _  ____/_  ____/  
 \____/  \____/_____/  /_/     /_/      

Usage:
  govpp [command]

Available Commands:
  cli         Send VPP CLI command
  diff        Compare two schemas
  export      Export files to output location
  generate    Generate code bindings
  http        Serve VPP API as HTTP service
  lint        Lint VPP API files
  vppapi      Print VPP API

Flags:
  -D, --debug             Enable debug mode
  -L, --loglevel string   Set logging level
      --color string      Color mode; auto/always/never
  -v, --version           version for govpp

Use "govpp [command] --help" for more information about a command.
```

### Print VPP API contents

The `vppapi` command allows you to print the VPP API files and their specific contents
in various formats. This can be useful for debugging or for generating documentation.

Here's an example usage of the `vppapi` command:

```sh
# List VPP API files for installed VPP
govpp vppapi
```

You can use the `--input` flag to specify the input source for the VPP API files.

```sh
# Use current directory as input source (local VPP repository)
govpp vppapi --input="."
```

The default format prints the output data in a table, but you can also specify 
other output formats such as JSON, YAML or Go template using the `--format` flag.

```sh
# Print in common formats
govpp vppapi --format="json"
govpp vppapi --format="yaml"

# Print as a Go template
govpp vppapi --format='{{ printf "%+v" . }}'
```

You can use the `--show-contents`, `--show-messages`, `--show-raw`, and `--show-rpc` 
flags to show specific parts of the VPP API file(s).

```sh
# Print VPP API file contents
govpp vppapi --show-contents

# Print VPP API messages
govpp vppapi --show-message

# Print VPP API services
govpp vppapi --show-rpc

# Print raw VPP API file
govpp vppapi --show-raw
```

You can also use the `--include-fields` and `--include-imported` flags to include 
message fields and imported types, respectively.

For more information on the available flags and options, use the `-h` or `--help` flag.

### Lint VPP API files

The `lint` command allows you to run linter checks for your VPP API files. This can help you catch issues early and ensure that your code follows best practices.

Here's an example usage of the `lint` command:

shell
```sh
govpp lint --input="https://github.com/FDio/vpp.git"
```

This command runs the linter checks on the `master` branch of official VPP repository and outputs any issues found.

You can use the `--input` flag to specify the input location for your VPP API files, such as a path to a VPP API directory or a local VPP repository.

You can use the `--help` flag to get more information about the available flags and options.

### Compare VPP API schemas

The `diff` command allows you to compare two VPP API schemas and lists the differences between them. This can be useful for detecting breaking changes between different versions of the API.

Here's an example usage of the `diff` command:

```sh
govpp diff ./vppapi2210 --against ./vppapi2302
```

This command compares the VPP API schema from `vppapi2210` directory against the VPP API schema in `vppapi2302` and lists the differences between them. The output shows related information details for each difference.

You can use the `--help` flag to get more information about the available flags and options.

> **Note**
> The `--against` flag is required and should point to an input source for the schema to compare against.

## Troubleshooting

If you run into any problems when executing some commands, you can use the `--debug` option to increase log verbosity to help when debugging the issue.
For the highest verbosity use `--log-level=trace`.


> **Warning**
> Be sure to open an issue on GitHub for any bug that you may encounter.
