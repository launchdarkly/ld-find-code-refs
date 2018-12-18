GOLANGCI_VERSION=v1.12.3

SHELL=/bin/bash

init:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s $(GOLANGCI_VERSION)

test: lint
	go test ./...

lint:
	./bin/golangci-lint run ./...

compile-github-actions-binary:
	cd parse/github-actions && GOOS=linux GOARCH=amd64 go build -o github-actions-flag-parser

compile-binary:
	cd parse/cmd && GOOS=linux GOARCH=amd64 go build -o git-flag-parser

# Get the lines added to the most recent changelog update (minus the first 2 lines)
RELEASE_NOTES=<(GIT_EXTERNAL_DIFF='bash -c "diff --unchanged-line-format=\"\" $$2 $$5" || true' git log --ext-diff -1 --pretty= -p CHANGELOG.md)

echo-release-notes:
	@cat $(RELEASE_NOTES)

define publish_docker
	test $(1) || (echo "Please provide tag"; exit 1)
	docker build -t ldactions/$(2):$(1) parse/$(3)
	docker tag ldactions/$(2):$(1) ldactions/$(2):latest
	docker push ldactions/$(2):$(1)
	docker push ldactions/$(2):latest
endef

publish-cli-docker: compile-binary
	$(call publish_docker,$(TAG),git-flag-parser,cmd)

publish-github-actions-docker: compile-github-actions-binary
	$(call publish_docker,$(TAG),git-flag-parser-gh-action,github-actions)

validate-circle-orb:
	test $(NAMESPACE) || (echo "Please provide namespace"; exit 1) # Remove once we always publish to launchdarkly namespace
	test $(TAG) || (echo "Please provide tag"; exit 1)
	circleci orb validate parse/circleci/orb.yml || (echo "Unable to validate orb"; exit 1)

publish-dev-circle-orb: validate-circle-orb
	# TODO: rename orb name (cr) to something more user-friendly
	circleci orb publish parse/circleci/orb.yml $(NAMESPACE)/cr@dev:$(TAG)

publish-release-circle-orb: validate-circle-orb
	circleci orb publish parse/circleci/orb.yml $(NAMESPACE)/cr@$(TAG)

.PHONY: init test lint compile-github-actions-binary compile-binary echo-release-notes publish-cli-docker publish-github-actions-docker publish-dev-circle-orb publish-release-circle-orb
