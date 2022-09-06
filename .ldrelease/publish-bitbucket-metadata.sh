#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

# Read username and password from the command line so we can debug this script. Defaults to releaser env variable.
BITBUCKET_USERNAME=${1:-"$(cat ${LD_RELEASE_SECRETS_DIR}/bitbucket_username)"}
BITBUCKET_TOKEN=${2:-"$(cat ${LD_RELEASE_SECRETS_DIR}/bitbucket_token)"}

RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}

mkdir -p bitbucketMetadataUpdates
git clone "https://${BITBUCKET_USERNAME}:${BITBUCKET_TOKEN}@bitbucket.org/launchdarkly/ld-find-code-refs-pipe.git" bitbucketMetadataUpdates
cp build/metadata/bitbucket/* bitbucketMetadataUpdates/
cd bitbucketMetadataUpdates
git config user.email "launchdarklyreleasebot@launchdarkly.com"
git config user.name "LaunchDarklyReleaseBot"
git add -u
git commit -m "Release auto update version"
git remote add bb-origin "https://${BITBUCKET_USERNAME}:${BITBUCKET_TOKEN}@bitbucket.org/launchdarkly/ld-find-code-refs-pipe.git"

if [[ -z "${LD_RELEASE_DRY_RUN}" ]]; then
  echo "Live run: will publish pipe to bitbucket."
  git tag $RELEASE_VERSION
  git push bb-origin master --tags
else
  echo "Dry run: will not publish pipe to bitbucket."
  git push bb-origin master --tags --dry-run
fi

cd .. && rm -rf bitbucketMetadataUpdates
