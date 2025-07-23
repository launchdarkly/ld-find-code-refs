#!/bin/bash

set -euo pipefail

RELEASE_TAG="v${LD_RELEASE_VERSION}"

update_go() (
  sed -i "s/const Version =.*/const Version = \"${LD_RELEASE_VERSION}\"/g" internal/version/version.go
)

update_orb() (
  sed -i "s#launchdarkly/ld-find-code-refs@.*#launchdarkly/ld-find-code-refs@${LD_RELEASE_VERSION}#g" build/package/circleci/orb.yml
  sed -i "s#- image: launchdarkly/ld-find-code-refs:.*#- image: launchdarkly/ld-find-code-refs:${LD_RELEASE_VERSION}#g" build/package/circleci/orb.yml
)

update_gha() (
  sed -i "s#launchdarkly/find-code-references@v.*#launchdarkly/find-code-references@${RELEASE_TAG}#g" build/metadata/github-actions/README.md
  sed -i "s#launchdarkly/ld-find-code-refs-github-action:.*#launchdarkly/ld-find-code-refs-github-action:${LD_RELEASE_VERSION}#g" build/metadata/github-actions/Dockerfile
)

update_bitbucket() (
  sed -i "s#- pipe: launchdarkly/ld-find-code-refs-pipe.*#- pipe: launchdarkly/ld-find-code-refs-pipe:${LD_RELEASE_VERSION}#g" build/metadata/bitbucket/README.md
  sed -i "s#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:.*#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:${LD_RELEASE_VERSION}#g" build/metadata/bitbucket/pipe.yml
)

tag_exists() (
  git fetch --tags
  git rev-parse "${RELEASE_TAG}" >/dev/null 2>&1
)

update_go
update_orb
update_gha
update_bitbucket
