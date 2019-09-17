SHELL = /bin/bash

GO ?= GO111MODULE=on go
GOVPP_PKG := $(shell go list)

VERSION ?= $(shell git describe --always --tags --dirty)
COMMIT ?= $(shell git rev-parse HEAD)
BUILD_STAMP ?= $(shell git log -1 --format="%ct")
BUILD_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
BUILD_HOST ?= $(shell hostname)
BUILD_USER ?= $(shell id -un)

VPP_VERSION	= $(shell dpkg-query -f '\${Version}' -W vpp)

VPP_IMG 	?= ligato/vpp-base:latest
BINAPI_DIR	?= ./examples/binapi

LDFLAGS = -w -s \
	-X ${GOVPP_PKG}/version.version=$(VERSION) \
	-X ${GOVPP_PKG}/version.commitHash=$(COMMIT) \
	-X ${GOVPP_PKG}/version.buildStamp=$(BUILD_STAMP) \
	-X ${GOVPP_PKG}/version.buildBranch=$(BUILD_BRANCH) \
	-X ${GOVPP_PKG}/version.buildUser=$(BUILD_USER) \
	-X ${GOVPP_PKG}/version.buildHost=$(BUILD_HOST)

GO_BUILD_ARGS = -ldflags "${LDFLAGS}"
ifeq ($(V),1)
GO_BUILD_ARGS += -v
endif
ifneq ($(GO_BUILD_TAGS),)
GO_BUILD_ARGS += -tags="${GO_BUILD_TAGS}"
endif

all: test build examples

install:
	@echo "=> installing binapi-generator ${VERSION}"
	$(GO) install ${GO_BUILD_ARGS} ./cmd/binapi-generator

build:
	@echo "=> building binapi-generator ${VERSION}"
	cd cmd/binapi-generator && $(GO) build ${GO_BUILD_ARGS}

examples:
	@echo "=> building examples"
	cd examples/perf-bench && $(GO) build ${GO_BUILD_ARGS} -v
	cd examples/rpc-service && $(GO) build ${GO_BUILD_ARGS} -v
	cd examples/simple-client && $(GO) build ${GO_BUILD_ARGS} -v
	cd examples/stats-client && $(GO) build ${GO_BUILD_ARGS} -v
	cd examples/union-example && $(GO) build ${GO_BUILD_ARGS} -v

clean:
	@echo "=> cleaning"
	go clean -v ./cmd/...
	go clean -v ./examples/...

test:
	@echo "=> running tests"
	$(GO) test ${GO_BUILD_ARGS} ./cmd/...
	$(GO) test ${GO_BUILD_ARGS} ./ ./api ./adapter ./codec ./core

test-integration:
	@echo "=> running integration tests"
	$(GO) test ${GO_BUILD_ARGS} ./test/integration

lint:
	@echo "=> running linter"
	@golint ./... | grep -v vendor | grep -v /binapi/ || true

gen-binapi-docker: install
	@echo "=> generating binapi in docker image ${VPP_IMG}"
	$(eval cmds := $(shell go generate -n $(BINAPI_DIR) 2>&1 | tr "\n" ";"))
	docker run -t --rm \
		-v "$(shell which gofmt):/usr/local/bin/gofmt:ro" \
		-v "$(shell which binapi-generator):/usr/local/bin/binapi-generator:ro" \
		-v "$(shell pwd):/govpp" -w /govpp \
		-u "$(shell id -u):$(shell id -g)" \
		"${VPP_IMG}" \
	  sh -xc "cd $(BINAPI_DIR) && $(cmds)"

generate-binapi: install
	@echo "=> generating binapi VPP $(VPP_VERSION)"
	$(GO) generate -x "$(BINAPI_DIR)"

generate:
	@echo "=> generating code"
	$(GO) generate -x ./...

extras:
	@make -C extras


.PHONY: all \
	install build examples clean test test-integration lint \
	generate generate-binapi gen-binapi-docker \
	extras
