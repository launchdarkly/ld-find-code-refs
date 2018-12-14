GOLANGCI_VERSION=v1.12.3

SHELL=/bin/bash

init:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s $(GOLANGCI_VERSION)

test: lint
	go test ./...

lint:
	./bin/golangci-lint run ./...

compile-github-actions-binary:
	cd parse/github-actions && GOARCH=amd64 go build -o out/github-actions-flag-parser

# Get the lines added to the most recent changelog update (minus the first 2 lines)
RELEASE_NOTES=<(GIT_EXTERNAL_DIFF='bash -c "diff --unchanged-line-format=\"\" $$2 $$5" || true' git log --ext-diff -1 --pretty= -p CHANGELOG.md)

echo-release-notes:
	@cat $(RELEASE_NOTES)

.PHONY: init test lint compile-github-actions-binary echo-release-notes
