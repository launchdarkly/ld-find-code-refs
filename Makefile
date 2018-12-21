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

compile-bitbucket-pipelines-binary:
	cd parse/bitbucket-pipelines && GOOS=linux GOARCH=amd64 go build -o bitbucket-pipelines-flag-parser

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

publish-bitbucket-pipelines-docker: compile-bitbucket-pipelines-binary
	$(call publish_docker,$(TAG),git-flag-parser-bb-pipeline,bitbucket-pipelines)

validate-circle-orb:
	test $(TAG) || (echo "Please provide tag"; exit 1)
	circleci orb validate parse/circleci/orb.yml || (echo "Unable to validate orb"; exit 1)

publish-dev-circle-orb: validate-circle-orb
	circleci orb publish parse/circleci/orb.yml launchdarkly/git-flag-parser@dev:$(TAG)

publish-release-circle-orb: validate-circle-orb
	circleci orb publish parse/circleci/orb.yml launchdarkly/git-flag-parser@$(TAG)

.PHONY: init test lint compile-github-actions-binary compile-binary compile-bitbucket-pipelines-binary echo-release-notes publish-cli-docker publish-github-actions-docker publish-bitbucket-pipelines-docker publish-dev-circle-orb publish-release-circle-orb
