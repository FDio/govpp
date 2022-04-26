This is an example of generation using wrappergen.

## vpplink
vpplink/ is a provider of an api for which binapi specific implementations need to be generated.

vpplink/api contains that api

vpplink/cmd contains a generator binary which can be used to generate an implementation given a binapi package.

vpplink/cmd/templates contains the templates for the generated impl

the generator binary incorporates the templates into itself

## consumer
The consumer/ directory contains an example consumer.  It has a simple gen.go file, that when run with go generate
will create consumer/impl generated from the specified binapi in the go.gen as an implementation of the vpplink/api

