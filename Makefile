SHELL := /usr/bin/env bash -o pipefail

PROJECT := govpp

VERSION      ?= $(shell git describe --always --tags --dirty --match='v*')
COMMIT       ?= $(shell git rev-parse HEAD)
BUILD_STAMP  ?= $(shell git log -1 --format='%ct')
BUILD_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

BUILD_HOST ?= $(shell hostname)
BUILD_USER ?= $(shell id -un)

GOVPP_PKG := go.fd.io/govpp/internal

VPP_API_DIR ?= ${VPP_DIR}/build-root/install-vpp-native/vpp/share/vpp/api

VERSION_PKG := $(GOVPP_PKG)/version
LDFLAGS = \
	-X $(VERSION_PKG).version=$(VERSION) \
	-X $(VERSION_PKG).commit=$(COMMIT) \
	-X $(VERSION_PKG).branch=$(BUILD_BRANCH) \
	-X $(VERSION_PKG).buildStamp=$(BUILD_STAMP) \
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

# Package cloud repo for VPP.
VPP_REPO		  ?= release
# VPP Docker image to use for api generation (gen-binapi-docker)
VPP_IMG 	      ?= ligato/vpp-base:25.10-release
# Local VPP directory used for binary api generation (gen-binapi-from-code)
VPP_DIR           ?=
# Target directory for generated go api bindings
BINAPI_DIR	      ?= ./binapi
# Binapi generator path
BINAPI_GENERATOR  = ./bin/binapi-generator

.DEFAULT_GOAL = help

export CGO_ENABLED=0

check-%:
	@: $(if $(value $*),,$(error $* is undefined))

bin:
	@mkdir -p bin

.PHONY: build
build: cmd examples ## Build all

.PHONY: cmd
cmd: bin ## Build commands
	@go build ${GO_BUILD_ARGS} -o bin ./cmd/...

.PHONY: binapi-generator
binapi-generator: bin ## Build only the binapi generator
	@go build ${GO_BUILD_ARGS} -o bin ./cmd/binapi-generator/

.PHONY: examples
examples: bin ## Build examples
	@go build ${GO_BUILD_ARGS} -o bin ./examples/...

.PHONY: test
test: ## Run unit tests
	@echo "# running tests"
	go test -tags="${GO_BUILD_TAGS}" ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "# running integration tests"
	VPP_REPO=$(VPP_REPO) ./test/run_integration.sh

.PHONY: lint ## Run code linter
lint:
	@golangci-lint run
	@echo "Done"

.PHONY: install
install: install-generator install-proxy ## Install all

.PHONY: install-generator
install-generator: ## Install binapi-generator
	@echo "# installing binapi-generator ${VERSION}"
	@go install ${GO_BUILD_ARGS} ./cmd/binapi-generator

.PHONY: install-proxy
install-proxy: ## Install vpp-proxy
	@echo "# installing vpp-proxy ${VERSION}"
	@go install ${GO_BUILD_ARGS} ./cmd/vpp-proxy

.PHONY: install-goreleaser
install-goreleaser: ## Install goreleaser
	@echo "# installing goreleaser"
	@go install github.com/goreleaser/goreleaser/v2@latest

.PHONY: release-snapshot
release-snapshot: ## Release snapshot
	@goreleaser release --clean --snapshot

.PHONY: generate
generate: generate-binapi ## Generate all

.PHONY: generate-binapi
generate-binapi: install-generator ## Generate binapi code
	@echo "# generating binapi"
	@go generate -x "$(BINAPI_DIR)"

.PHONY: gen-binapi-local
gen-binapi-local: binapi-generator check-VPP_DIR ## Generate binapi code (using locally cloned VPP)
	@make -C ${VPP_DIR} json-api-files
	@find $(BINAPI_DIR)/*/*.ba.go -delete || true
	@find $(BINAPI_DIR)/* -type d -delete
	@./bin/binapi-generator --input=$(VPP_API_DIR) --output-dir=$(BINAPI_DIR) --gen=rpc
	@./bin/binapi-generator --input=$(VPP_API_DIR)/core --output-dir=$(BINAPI_DIR) --gen=http vpe
	@sed -i 's@$(VPP_API_DIR)@/usr/share/vpp/api@g' $(BINAPI_DIR)/*/*.ba.go

.PHONY: gen-binapi-docker
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

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


