#!/bin/bash

# Run this in publish step after all version information have been updated.
set -euo pipefail

setup_bitbucket() (
  rm -rf bitbucketMetadataUpdates
  mkdir -p bitbucketMetadataUpdates
  git clone "https://${BITBUCKET_USERNAME}:${BITBUCKET_TOKEN}@bitbucket.org/launchdarkly/ld-find-code-refs-pipe.git" bitbucketMetadataUpdates
  cp build/metadata/bitbucket/* bitbucketMetadataUpdates/
  cp CHANGELOG.md bitbucketMetadataUpdates/
  cd bitbucketMetadataUpdates
  git config user.email "launchdarklyreleasebot@launchdarkly.com"
  git config user.name "LaunchDarklyReleaseBot"
  git add -u
  git commit -m "Release auto update version $LD_RELEASE_VERSION"
  git remote add bb-origin "https://${BITBUCKET_USERNAME}:${BITBUCKET_TOKEN}@bitbucket.org/launchdarkly/ld-find-code-refs-pipe.git"
)

clean_up_bitbucket() (
  cd .. && rm -rf bitbucketMetadataUpdates
)

tag_exists() (
  git ls-remote --tags https://bitbucket.org/launchdarkly/ld-find-code-refs-pipe.git "refs/tags/v$LD_RELEASE_VERSION" | grep -q "v$LD_RELEASE_VERSION"
)

publish_bitbucket() (
  if tag_exists; then
    echo "Version exists; skipping publishing BitBucket Pipe"
    return 0
  fi

  setup_bitbucket

  echo "Live run: will publish pipe to bitbucket."

  cd bitbucketMetadataUpdates
  git tag "$LD_RELEASE_VERSION"
  git push bb-origin master --tags

  clean_up_bitbucket
)

dry_run_bitbucket() (
  if tag_exists; then
    echo "Version exists; skipping push dry-run BitBucket Pipe"
    return 0
  fi

  setup_bitbucket

  echo "Dry run: will not publish pipe to bitbucket."
  cd bitbucketMetadataUpdates
  git push bb-origin master --tags --dry-run

  clean_up_bitbucket
)
