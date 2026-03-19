BINARY_NAME=healthcheck
MODULE=github.com/enriquevaldivia1988/api-health-cli
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test install lint clean fmt vet

## build: Compile the binary
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

## test: Run all tests with race detection
test:
	go test -race -cover ./...

## install: Install the binary to $GOPATH/bin
install:
	go install $(LDFLAGS) .

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format all Go files
fmt:
	gofmt -s -w .

## vet: Run go vet
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/
	go clean

## tidy: Clean up go.mod and go.sum
tidy:
	go mod tidy

## help: Show this help message
help:
	@echo "Usage:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
