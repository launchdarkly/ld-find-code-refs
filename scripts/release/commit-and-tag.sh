#!/bin/bash

set -euo pipefail

release_tag="v${LD_RELEASE_VERSION}"

tag_exists() (
  git fetch --tags
  git rev-parse "${release_tag}" >/dev/null 2>&1
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

release_summary() (
  echo "Changes staged for release $release_tag:"
  git diff

  echo "Updated CHANGELOG:"
  cat CHANGELOG.md
)

commit_and_tag() (
  if tag_exists; then
    echo "Tag $release_tag already exists. Aborting."
    exit 1
  fi

  update_changelog
  release_summary

  git config user.name "LaunchDarklyReleaseBot"
  git config user.email "releasebot@launchdarkly.com"
  git add .
  git commit -m "Prepare release ${release_tag}"
  git tag "${release_tag}"
  git push origin HEAD
  git push origin "${release_tag}"
)

if [[ "$DRY_RUN" == "true" ]]; then
  update_changelog
  release_summary
  echo "Dry run mode: skipping commit, tag, and push"
else
  commit_and_tag
fi

