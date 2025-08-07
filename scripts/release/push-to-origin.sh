#!/bin/bash

set -euo pipefail

release_tag="v$LD_RELEASE_VERSION"

tag_exists() (
  git ls-remote --tags git@github.com:launchdarkly/ld-find-code-refs.git "refs/tags/$release_tag" | grep -q "$release_tag"
)

push_to_origin() (
  if tag_exists; then
    echo "Tag $release_tag already exists. Aborting."
    return 0
  fi

  git push origin HEAD
  git push origin "$release_tag"
)

if [[ "$DRY_RUN" == "true" ]]; then
  git tag -d "$release_tag" # defensive
  git reset --hard HEAD^    # defensive
  echo "Dry run mode: skipping push"
else
  push_to_origin
fi

