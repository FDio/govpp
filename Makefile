VERSION ?= $(shell git describe --always --tags --dirty)

all: test build examples

install:
	@echo "=> installing ${VERSION}"
	go install ./cmd/binapi-generator

build:
	@echo "=> building ${VERSION}"
	cd cmd/binapi-generator && go build -v

examples:
	@echo "=> building examples"
	cd examples/cmd/simple-client && go build -v
	cd examples/cmd/stats-client && go build -v
	cd examples/cmd/perf-bench && go build -v

test:
	@echo "=> testing"
	go test -cover ./cmd/...
	go test -cover ./core ./api ./codec

extras:
	@echo "=> building extras"
	cd extras/libmemif/examples/gopacket && go build -v
	cd extras/libmemif/examples/icmp-responder && go build -v
	cd extras/libmemif/examples/jumbo-frames && go build -v
	cd extras/libmemif/examples/raw-data && go build -v

clean:
	@echo "=> cleaning"
	rm -f cmd/binapi-generator/binapi-generator
	rm -f examples/cmd/perf-bench/perf-bench
	rm -f examples/cmd/simple-client/simple-client
	rm -f examples/cmd/stats-client/stats-client
	rm -f extras/libmemif/examples/gopacket/gopacket
	rm -f extras/libmemif/examples/icmp-responder/icmp-responder
	rm -f extras/libmemif/examples/jumbo-frames/jumbo-frames
	rm -f extras/libmemif/examples/raw-data/raw-data

generate: install
	@echo "=> generating code"
	cd examples && go generate ./...

lint:
	@echo "=> running linter"
	@golint ./... | grep -v vendor | grep -v bin_api || true

.PHONY: all \
	install build examples test \
	extras clean generate lint
