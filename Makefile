build:
	@cd binapi_generator && go build -v

test:
	@go test -cover $(glide novendor)

install:
	@cd binapi_generator && go install -v

clean:
	@rm binapi_generator/binapi_generator

.PHONY: build test install clean
