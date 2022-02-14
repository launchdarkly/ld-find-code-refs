#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

# Read username and password from the command line so we can debug this script. Defaults to releaser env variable.
BITBUCKET_USERNAME=${1:-"${LD_RELEASE_SECRETS_DIR}/bitbucket_username"}
BITBUCKET_TOKEN=${2:-"${LD_RELEASE_SECRETS_DIR}/bitbucket_token"}

rm -rf bitbucketMetadataUpdates
mkdir -p bitbucketMetadataUpdates
git clone git@bitbucket.org:launchdarkly/ld-find-code-refs-pipe.git bitbucketMetadataUpdates
cp build/metadata/bitbucket/* bitbucketMetadataUpdates/
cd bitbucketMetadataUpdates
git add -u
git commit -m "Release auto update version"
git remote set-url origin "https://${BITBUCKET_USERNAME}:${BITBUCKET_TOKEN}@bitbucket.org/launchdarkly/ld-find-code-refs-pipe.git"
git push origin master

cd .. && rm -rf bitbucketMetadataUpdates