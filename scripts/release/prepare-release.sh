#!/bin/bash

set -euo pipefail

release_tag="v${LD_RELEASE_VERSION}"

update_go() (
  sed -i "s/const Version =.*/const Version = \"${LD_RELEASE_VERSION}\"/g" internal/version/version.go
)

update_orb() (
  sed -i "s#launchdarkly/ld-find-code-refs@.*#launchdarkly/ld-find-code-refs@${LD_RELEASE_VERSION}#g" build/package/circleci/orb.yml
  sed -i "s#- image: launchdarkly/ld-find-code-refs:.*#- image: launchdarkly/ld-find-code-refs:${LD_RELEASE_VERSION}#g" build/package/circleci/orb.yml
)

update_gha() (
  sed -i "s#launchdarkly/find-code-references@v.*#launchdarkly/find-code-references@${release_tag}#g" build/metadata/github-actions/README.md
  sed -i "s#launchdarkly/ld-find-code-refs-github-action:.*#launchdarkly/ld-find-code-refs-github-action:${LD_RELEASE_VERSION}#g" build/metadata/github-actions/Dockerfile
)

update_bitbucket() (
  sed -i "s#- pipe: launchdarkly/ld-find-code-refs-pipe.*#- pipe: launchdarkly/ld-find-code-refs-pipe:${LD_RELEASE_VERSION}#g" build/metadata/bitbucket/README.md
  sed -i "s#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:.*#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:${LD_RELEASE_VERSION}#g" build/metadata/bitbucket/pipe.yml
)

update_changelog() (
  local ts=$(date +"%Y-%m-%d")
  # multiline strings don't seem to be supported for GHA inputs, so for now we
  # require that the changelog include \n characters for new lines and then we
  # just expand them here
  # TODO improve this
  local changelog_content=$(printf "%b" "$CHANGELOG_ENTRY")
  local changelog_entry=$(printf "## [%s] - %s\n%s\n" "$LD_RELEASE_VERSION" "$ts" "$changelog_content")

  # insert the new changelog entry (followed by empty line) after line 4
  # of CHANGELOG.md
  sed -i "4r /dev/stdin" CHANGELOG.md <<< "$changelog_entry"$'\n'
)

update_go
update_orb
update_gha
update_bitbucket
update_changelog

git config user.name "LaunchDarklyReleaseBot"
git config user.email "releasebot@launchdarkly.com"
git add .
git commit -m "Prepare release ${release_tag}"
