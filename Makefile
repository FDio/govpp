build:
	@cd binapi_generator && go build -v

test:
	@cd binapi_generator && go test -cover .
	@cd api && go test -cover ./...
	@cd core && go test -cover .

install:
	@cd binapi_generator && go install -v

clean:
	@rm binapi_generator/binapi_generator

.PHONY: build test install clean
