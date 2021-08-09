# Note: These commands pertain to the development of ld-find-code-refs.
#       They are not intended for use by the end-users of this program.
SHELL=/bin/bash

init:
	pre-commit install

test: lint
	go test ./...

lint:
	pre-commit run -a --verbose golangci-lint

# Strip debug informatino from production builds
BUILD_FLAGS = -ldflags="-s -w"

compile-macos-binary:
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o out/ld-find-code-refs ./cmd/ld-find-code-refs

compile-windows-binary:
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o out/ld-find-code-refs.exe ./cmd/ld-find-code-refs

compile-linux-binary:
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o build/package/cmd/ld-find-code-refs ./cmd/ld-find-code-refs

compile-github-actions-binary:
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o build/package/github-actions/ld-find-code-refs-github-action ./build/package/github-actions

compile-bitbucket-pipelines-binary:
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o build/package/bitbucket-pipelines/ld-find-code-refs-bitbucket-pipeline ./build/package/bitbucket-pipelines

# Get the lines added to the most recent changelog update (minus the first 2 lines)
RELEASE_NOTES=<(GIT_EXTERNAL_DIFF='bash -c "diff --unchanged-line-format=\"\" $$2 $$5" || true' git log --ext-diff -1 --pretty= -p CHANGELOG.md)

echo-release-notes:
	@cat $(RELEASE_NOTES)

define publish_docker
	test $(1) || (echo "Please provide tag"; exit 1)
	docker build -t launchdarkly/$(3):$(1) build/package/$(4)
	docker push launchdarkly/$(3):$(1)
	# test $(2) && (echo "Not pushing latest tag for prerelease")
	test $(2) || docker tag launchdarkly/$(3):$(1) launchdarkly/$(3):latest
	test $(2) || docker push launchdarkly/$(3):latest
endef

validate-circle-orb:
	test $(TAG) || (echo "Please provide tag"; exit 1)
	circleci orb validate build/package/circleci/orb.yml || (echo "Unable to validate orb"; exit 1)

publish-dev-circle-orb: validate-circle-orb
	circleci orb publish build/package/circleci/orb.yml launchdarkly/ld-find-code-refs@dev:$(TAG)

publish-release-circle-orb: validate-circle-orb
	circleci orb publish build/package/circleci/orb.yml launchdarkly/ld-find-code-refs@$(TAG)

publish-all: publish-release-circle-orb

clean:
	rm -rf out/
	rm -f build/pacakge/cmd/ld-find-code-refs
	rm -f build/package/github-actions/ld-find-code-refs-github-action
	rm -f build/package/bitbucket-pipelines/ld-find-code-refs-bitbucket-pipeline

.PHONY: init test lint compile-github-actions-binary compile-macos-binary compile-linux-binary compile-windows-binary compile-bitbucket-pipelines-binary echo-release-notes publish-dev-circle-orb publish-release-circle-orb publish-all clean
