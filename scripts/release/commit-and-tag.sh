#!/bin/bash

set -euo pipefail

RELEASE_TAG="v${LD_RELEASE_VERSION}"

tag_exists() (
  git fetch --tags
  git rev-parse "${RELEASE_TAG}" >/dev/null 2>&1
)

update_changelog() (
  local ts=$(date +"%Y-%m-%d")
  local changelog_entry=$(cat << EOF
## [$LD_RELEASE_VERSION] - $ts
$CHANGELOG_ENTRY
EOF

  # insert the new changelog entry (followed by empty line) after line 4
  # of CHANGELOG.md
  sed -i "4r /dev/stdin" CHANGELOG.md <<< "$changelog_entry"$'\n'
)

release_summary() (
  echo
  echo "Changes staged for release $RELEASE_TAG:"
  git diff

  echo
  echo "Updated CHANGELOG:"
  cat CHANGELOG.md
  echo
)

commit_and_tag() (
  if tag_exists; then
    echo "Tag $RELEASE_TAG already exists. Aborting."
    exit 1
  fi

  update_changelog
  release_summary

  git config user.name "LaunchDarklyReleaseBot"
  git config user.email "releasebot@launchdarkly.com"
  git add .
  git commit -m "Prepare release ${RELEASE_TAG}"
  git tag "${RELEASE_TAG}"
  git push origin HEAD
  git push origin "${RELEASE_TAG}"
)

if [[ "$DRY_RUN" == "true" ]]; then
  update_changelog
  release_summary
  echo "Dry run mode: skipping commit, tag, and push"
else
  commit_and_tag
fi

