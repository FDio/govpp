SHELL := /usr/bin/env bash -o pipefail

VERSION ?= $(shell git describe --always --tags --dirty)
COMMIT ?= $(shell git rev-parse HEAD)
BUILD_STAMP ?= $(shell git log -1 --format="%ct")
BUILD_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

BUILD_HOST ?= $(shell hostname)
BUILD_USER ?= $(shell id -un)

GO ?= go

GOVPP_PKG := $(shell go list)
LDFLAGS = \
	-X ${GOVPP_PKG}/version.version=$(VERSION) \
	-X ${GOVPP_PKG}/version.commitHash=$(COMMIT) \
	-X ${GOVPP_PKG}/version.buildStamp=$(BUILD_STAMP) \
	-X ${GOVPP_PKG}/version.buildBranch=$(BUILD_BRANCH) \
	-X ${GOVPP_PKG}/version.buildUser=$(BUILD_USER) \
	-X ${GOVPP_PKG}/version.buildHost=$(BUILD_HOST)
ifeq ($(NOSTRIP),)
LDFLAGS += -w -s
endif

GO_BUILD_TAGS ?= novpp

GO_BUILD_ARGS = -ldflags "${LDFLAGS}"
ifneq ($(GO_BUILD_TAGS),)
GO_BUILD_ARGS += -tags="${GO_BUILD_TAGS}"
endif
ifneq ($(GO_NOTRIM),0)
GO_BUILD_ARGS += -trimpath
endif
ifeq ($(V),1)
GO_BUILD_ARGS += -v
endif

VPP_VERSION	= $(shell dpkg-query -f '\${Version}' -W vpp)

VPP_IMG 	?= ligato/vpp-base:latest
BINAPI_DIR	?= ./examples/binapi

help:
	@echo "List of make targets:"
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT = help

bin:
	mkdir -p bin

build: ## Build all
	@echo "# building ${VERSION}"
	$(GO) build ${GO_BUILD_ARGS} ./...

cmd: bin ## Build commands
	$(GO) build ${GO_BUILD_ARGS} -o bin ./cmd/...

examples: bin ## Build examples
	$(GO) build ${GO_BUILD_ARGS} -o bin ./examples/...

clean: ## Clean all
	@echo "# cleaning"
	$(GO) clean -v ./...

test: ## Run unit tests
	@echo "# running tests"
	$(GO) test -tags="${GO_BUILD_TAGS}" ./...

integration: ## Run integration tests
	@echo "# running integration tests"
	$(GO) test -tags="integration ${GO_BUILD_TAGS}" ./test/integration

lint: ## Run code linter
	@echo "# running linter"
	@golint ./...

install-generator: ## Install binapi-generator
	@echo "# installing binapi-generator ${VERSION}"
	$(GO) install ${GO_BUILD_ARGS} ./cmd/binapi-generator

generate: ## Generate code
	@echo "# generating code"
	$(GO) generate -x ./...

generate-binapi: install-generator
	@echo "# generating binapi VPP $(VPP_VERSION)"
	$(GO) generate -x "$(BINAPI_DIR)"

gen-binapi-docker: install-generator
	@echo "# generating binapi in docker image ${VPP_IMG}"
	$(eval cmds := $(shell go generate -n $(BINAPI_DIR) 2>&1 | tr "\n" ";"))
	docker run -t --rm \
		-v "$(shell which gofmt):/usr/local/bin/gofmt:ro" \
		-v "$(shell which binapi-generator):/usr/local/bin/binapi-generator:ro" \
		-v "$(shell pwd):/govpp" -w /govpp \
		-u "$(shell id -u):$(shell id -g)" \
		"${VPP_IMG}" \
	  sh -xc "cd $(BINAPI_DIR) && $(cmds)"

extras:
	@make -C extras


.PHONY: help \
    build cmd examples clean \
	lint test integration \
	install-generator generate generate-binapi gen-binapi-docker \
	extras

