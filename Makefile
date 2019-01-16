# Note: These commands pertain to the development of git-flag-parser.
#       They are not intended for use by the end-users of this program.
GOLANGCI_VERSION=v1.12.3

SHELL=/bin/bash

init:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s $(GOLANGCI_VERSION)

test: lint
	go test ./...

lint:
	./bin/golangci-lint run ./...

compile-binary:
	GOOS=linux GOARCH=amd64 go build -o build/package/cmd/git-flag-parser ./cmd/git-flag-parser

compile-github-actions-binary:
	GOOS=linux GOARCH=amd64 go build -o out/github-actions-flag-parser ./build/package/github-actions

compile-bitbucket-pipelines-binary:
	GOOS=linux GOARCH=amd64 go build -o out/bitbucket-pipelines-flag-parser ./build/package/bitbucket-pipelines

# Get the lines added to the most recent changelog update (minus the first 2 lines)
RELEASE_NOTES=<(GIT_EXTERNAL_DIFF='bash -c "diff --unchanged-line-format=\"\" $$2 $$5" || true' git log --ext-diff -1 --pretty= -p CHANGELOG.md)

echo-release-notes:
	@cat $(RELEASE_NOTES)

define publish_docker
	test $(1) || (echo "Please provide tag"; exit 1)
	docker build -t ldactions/$(2):$(1) build/package/$(3)
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
	circleci orb validate build/package/circleci/orb.yml || (echo "Unable to validate orb"; exit 1)

publish-dev-circle-orb: validate-circle-orb
	circleci orb publish build/package/circleci/orb.yml launchdarkly/git-flag-parser@dev:$(TAG)

publish-release-circle-orb: validate-circle-orb
	circleci orb publish build/package/circleci/orb.yml launchdarkly/git-flag-parser@$(TAG)

publish-all: publish-cli-docker publish-github-actions-docker publish-bitbucket-pipelines-docker publish-release-circle-orb

clean:
	rm -f build/pacakge/cmd/git-flag-parser
	rm -f build/pacakge/github-actions/github-actions-flag-parser
	rm -f build/pacakge/bitbucket-pipelines-flag-parser/bitbucket-pipelines-flag-parser

.PHONY: init test lint compile-github-actions-binary compile-binary compile-bitbucket-pipelines-binary echo-release-notes publish-cli-docker publish-github-actions-docker publish-bitbucket-pipelines-docker publish-dev-circle-orb publish-release-circle-orb publish-all clean
