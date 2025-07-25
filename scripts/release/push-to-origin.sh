#!/bin/bash

set -euo pipefail

release_tag="v${LD_RELEASE_VERSION}"

tag_exists() (
  git fetch --tags
  git rev-parse "${release_tag}" >/dev/null 2>&1
)

push_to_origin() (
  if tag_exists; then
    echo "Tag $release_tag already exists. Aborting."
    exit 1
  fi

  git push origin HEAD
  git push origin "${release_tag}"
)

if [[ "$DRY_RUN" == "true" ]]; then
  git tag -d "$release_tag" # defensive
  git reset --hard HEAD^    # defensive
  echo "Dry run mode: skipping push"
else
  push_to_origin
fi

