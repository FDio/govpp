VERSION ?= $(shell git describe --always --tags --dirty)

BINAPI_DIR ?= $(shell cd examples/bin_api && pwd)
VPP_VERSION := $(shell apt-cache show vpp | grep Version: | cut -d' ' -f2-)

all: test build examples

install:
	@echo "=> installing binapi generator ${VERSION}"
	go install ./cmd/binapi-generator

build:
	@echo "=> building binapi generator ${VERSION}"
	cd cmd/binapi-generator && go build -v

examples:
	@echo "=> building examples"
	cd examples/simple-client && go build -v
	cd examples/stats-api && go build -v
	cd examples/perf-bench && go build -v
	cd examples/union-example && go build -v

test:
	@echo "=> running tests"
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
	rm -f examples/perf-bench/perf-bench
	rm -f examples/simple-client/simple-client
	rm -f examples/stats-api/stats-api
	rm -f examples/union-example/union-example
	rm -f extras/libmemif/examples/gopacket/gopacket
	rm -f extras/libmemif/examples/icmp-responder/icmp-responder
	rm -f extras/libmemif/examples/jumbo-frames/jumbo-frames
	rm -f extras/libmemif/examples/raw-data/raw-data

generate-binapi:
	@echo "=> generating binapi"
	@go generate "${BINAPI_DIR}"

generate: install
	@echo "=> generating code"
	cd examples && go generate ./...

update-vppapi:
	@echo "=> updating API JSON files using installed VPP ${VPP_VERSION}"
	@cd ${BINAPI_DIR} && find . -type f -name '*.api.json' -exec cp /usr/share/vpp/api/'{}' '{}' \;
	@echo ${VPP_VERSION} > ${BINAPI_DIR}/VPP_VERSION

lint:
	@echo "=> running linter"
	@golint ./... | grep -v vendor | grep -v bin_api || true

.PHONY: all \
	install build examples test \
	extras clean generate lint
