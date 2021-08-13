This is an example of generation using the binapi-generator.

## How to run

````
export VPP_DIR=~/vpp
go generate -v .
````

## What happens under the hood.

The binapi-generator finds the VPP repository passed in the CLI,
it runs `make json-api-files` there. Then consumes the produced
json files, and outputs the api bindings to `./impl`
