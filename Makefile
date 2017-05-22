build:
	@cd cmd/binapi-generator && go build -v
	@cd examples && go build -v

test:
	@cd cmd/binapi-generator && go test -cover .
	@cd api && go test -cover ./...
	@cd core && go test -cover .

install:
	@cd cmd/binapi-generator && go install -v

clean:
	@rm cmd/binapi-generator/binapi-generator
	@rm examples/examples

generate:
	@cd core && go generate ./...
	@cd examples && go generate ./...

lint:
	@golint ./... | grep -v vendor | grep -v bin_api || true

.PHONY: build test install clean generate
