#!/bin/bash

set -euo pipefail

release_tag="v$LD_RELEASE_VERSION"

tag_exists() (
  git ls-remote --tags origin "refs/tags/${release_tag}" | grep -q "refs/tags/${release_tag}$"
)

push_tag() (
  if tag_exists; then
    echo "Tag $release_tag already exists on origin. Skipping tag push."
    return 0
  fi

  git push origin "$release_tag"
)

if [[ "$DRY_RUN" == "true" ]]; then
  git tag -d "$release_tag" # defensive
  git reset --hard HEAD^    # defensive
  echo "Dry run mode: skipping push"
else
  push_tag
fi
