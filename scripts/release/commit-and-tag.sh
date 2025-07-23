#!/bin/bash

set -euo pipefail

RELEASE_TAG="v${LD_RELEASE_VERSION}"

echo "Changes staged for release $RELEASE_TAG:"
git diff

if [[ "$DRY_RUN" == "true" ]]; then
  echo "Dry run mode: skipping commit, tag, and push"
else
  if tag_exists; then
    echo "Tag $RELEASE_TAG already exists. Aborting."
    exit 1
  fi

  git config user.name "LaunchDarklyReleaseBot"
  git config user.email "releasebot@launchdarkly.com"
  git add .
  git commit -m "Prepare release ${RELEASE_TAG}"
  git tag "${RELEASE_TAG}"
  git push origin HEAD
  git push origin "${RELEASE_TAG}"
fi

