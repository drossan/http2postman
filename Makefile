.PHONY: build test test-coverage lint fmt vet check clean

BINARY_NAME=http2postman
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS=-ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'"

## Build

build:
	go build $(LDFLAGS) -mod=vendor -o bin/$(BINARY_NAME) .

## Test

test:
	go test -mod=vendor ./... -v

test-coverage:
	go test -mod=vendor ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@echo "---"
	@echo "HTML report: go tool cover -html=coverage.out"

## Code Quality

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet -mod=vendor ./...

check: fmt vet test

## Clean

clean:
	rm -rf bin/ coverage.out
