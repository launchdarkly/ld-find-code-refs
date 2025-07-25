#!/bin/bash

set -euo pipefail

release_tag="v${LD_RELEASE_VERSION}"

tag_exists() (
  git fetch --tags
  git rev-parse "${release_tag}" >/dev/null 2>&1
)

create_tag() (
  if tag_exists; then
    echo "Tag $release_tag already exists. Aborting."
    exit 1
  fi

  git tag "${release_tag}"
  git push origin HEAD
  git push origin "${release_tag}"
)

if [[ "$DRY_RUN" == "true" ]]; then
  git reset --hard HEAD^
  echo "Dry run mode: skipping tag and push"
else
  create_tag
fi

