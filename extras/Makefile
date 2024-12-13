GO ?= GO111MODULE=on go

build: extras

extras:
	@echo "=> building extras"
	cd libmemif/examples/gopacket && $(GO) build -v
	cd libmemif/examples/icmp-responder && $(GO) build -v
	cd libmemif/examples/jumbo-frames && $(GO) build -v
	cd libmemif/examples/raw-data && $(GO) build -v
	cd gomemif/examples/bridge && $(GO) build -v
	cd gomemif/examples/icmp_responder_cb && $(GO) build -v
	cd gomemif/examples/icmp_responder_poll && $(GO) build -v

clean:
	@echo "=> cleaning extras"
	go clean -v ./...

test-extras: extras
	## Run extras integration tests
	@echo "# running extras integration tests"
	VPP_REPO=$(VPP_REPO) ./test/run_integration.sh

.PHONY: build extras clean
