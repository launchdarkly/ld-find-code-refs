GOLANGCI_VERSION=v1.12.3

SHELL=/bin/bash

test: lint
	go test ./...

lint:
	./bin/golangci-lint run ./...

init:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s $(GOLANGCI_VERSION)