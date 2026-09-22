GOLANGCI_LINT ?= golangci-lint

.PHONY: all build test lint fmt cover integration

all: build lint test

build:
	go build ./...
	go vet ./...

test:
	go test -race ./...

lint:
	$(GOLANGCI_LINT) run

fmt:
	$(GOLANGCI_LINT) fmt

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

# Read-only calls against the real services; needs ITMO_REFRESH_TOKEN.
integration:
	go test -tags integration -count=1 ./...
