SHELL := /usr/bin/env bash -o pipefail

PROJECT := govpp

VERSION      ?= $(shell git describe --always --tags --dirty)
COMMIT       ?= $(shell git rev-parse HEAD)
BUILD_STAMP  ?= $(shell git log -1 --format='%ct')
BUILD_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

BUILD_HOST ?= $(shell hostname)
BUILD_USER ?= $(shell id -un)

GOVPP_PKG := git.fd.io/govpp.git

VERSION_PKG := $(GOVPP_PKG)/internal/version
LDFLAGS = \
	-X $(VERSION_PKG).version=$(VERSION) \
	-X $(VERSION_PKG).commitHash=$(COMMIT) \
	-X $(VERSION_PKG).buildStamp=$(BUILD_STAMP) \
	-X $(VERSION_PKG).buildBranch=$(BUILD_BRANCH) \
	-X $(VERSION_PKG).buildUser=$(BUILD_USER) \
	-X $(VERSION_PKG).buildHost=$(BUILD_HOST)

ifeq ($(NOSTRIP),)
LDFLAGS += -w -s
endif

GO_BUILD_TAGS ?= novpp

GO_BUILD_ARGS = -ldflags "$(LDFLAGS)"
ifneq ($(GO_BUILD_TAGS),)
GO_BUILD_ARGS += -tags="${GO_BUILD_TAGS}"
endif
ifneq ($(GO_NOTRIM),0)
GO_BUILD_ARGS += -trimpath
endif
ifeq ($(V),1)
GO_BUILD_ARGS += -v
endif

# VPP Docker image to use for api generation (gen-binapi-docker)
VPP_IMG 	      ?= ligato/vpp-base:latest
# Local VPP directory used for binary api generation (gen-binapi-from-code)
VPP_DIR           ?=
# Target directory for generated go api bindings
BINAPI_DIR	      ?= ./binapi
# Binapi generator path
BINAPI_GENERATOR  = ./bin/binapi-generator

.DEFAULT_GOAL = help

bin:
	mkdir -p bin

build: ## Build all
	@echo "# building ${VERSION}"
	go build ${GO_BUILD_ARGS} ./...

cmd: bin ## Build commands
	go build ${GO_BUILD_ARGS} -o bin ./cmd/...

examples: bin ## Build examples
	go build ${GO_BUILD_ARGS} -o bin ./examples/...

test: ## Run unit tests
	@echo "# running tests"
	go test -tags="${GO_BUILD_TAGS}" ./...

test-integration: ## Run integration tests
	@echo "# running integration tests"
	go test -tags="integration ${GO_BUILD_TAGS}" ./test/integration

lint: ## Run code linter
	@echo "# running linter"
	@golint ./...

install: install-generator install-proxy ## Install all

install-generator: ## Install binapi-generator
	@echo "# installing binapi-generator ${VERSION}"
	@go install ${GO_BUILD_ARGS} ./cmd/binapi-generator

install-proxy: ## Install vpp-proxy
	@echo "# installing vpp-proxy ${VERSION}"
	go install ${GO_BUILD_ARGS} ./cmd/vpp-proxy

generate: generate-binapi ## Generate all

generate-binapi: install-generator ## Generate binapi code
	@echo "# generating binapi"
	go generate -x "$(BINAPI_DIR)"

gen-binapi-from-code: cmd-binapi-generator
	$(eval VPP_API_DIR := ${VPP_DIR}/build-root/install-vpp-native/vpp/share/vpp/api/)
	@echo "Generating vpp API.json and go bindings"
	@echo "Vpp Directory ${VPP_DIR}"
	@echo "Vpp API files ${VPP_API_DIR}"
	@echo "Go bindings   ${BINAPI_DIR}"
	@cd ${VPP_DIR} && make json-api-files
	@${BINAPI_GENERATOR} \
		--input-dir=${VPP_API_DIR} \
	    --output-dir=${BINAPI_DIR} \
	    --gen rpc,rest \
	    --no-source-path-info

gen-binapi-docker: install-generator ## Generate binapi code (using Docker)
	@echo "# generating binapi in docker image ${VPP_IMG}"
	$(eval cmds := $(shell go generate -n $(BINAPI_DIR) 2>&1 | tr "\n" ";"))
	docker run -t --rm \
		-e DEBUG_GOVPP \
		-v "$(shell which binapi-generator):/usr/local/bin/binapi-generator:ro" \
		-v "$(shell pwd):/govpp" \
		-w /govpp \
		-u "$(shell id -u):$(shell id -g)" \
		"${VPP_IMG}" \
	  sh -ec "cd $(BINAPI_DIR) && $(cmds)"

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: help \
    build cmd examples clean \
	lint test integration \
	install install-generator install-proxy \
	generate generate-binapi gen-binapi-docker \
	extras

