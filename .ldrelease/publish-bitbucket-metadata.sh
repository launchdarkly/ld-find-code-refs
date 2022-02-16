#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

# Read username and password from the command line so we can debug this script. Defaults to releaser env variable.
BITBUCKET_USERNAME=${1:-"$(cat ${LD_RELEASE_SECRETS_DIR}/bitbucket_username)"}
BITBUCKET_TOKEN=${2:-"$(cat ${LD_RELEASE_SECRETS_DIR}/bitbucket_token)"}

mkdir –m700 ~/.ssh
touch ~/.ssh/known_hosts
chmod 644 ~/.ssh/known_hosts
ssh-keyscan -t rsa bitbucket.org >> ~/.ssh/known_hosts
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

mkdir -p bitbucketMetadataUpdates
git clone "https://${BITBUCKET_USERNAME}:${BITBUCKET_TOKEN}@bitbucket.org/launchdarkly/ld-find-code-refs-pipe.git" bitbucketMetadataUpdates
cp build/metadata/bitbucket/* bitbucketMetadataUpdates/
cd bitbucketMetadataUpdates
git config --global user.email "yus@launchdarkly.com"
git config --global user.name "Yus Ngadiman"
git add -u
git commit -m "Release auto update version"
git remote add bb-origin "https://${BITBUCKET_USERNAME}:${BITBUCKET_TOKEN}@bitbucket.org/launchdarkly/ld-find-code-refs-pipe.git"

if [[ -z "${LD_RELEASE_DRY_RUN}" ]]; then
  echo "Live run: will publish pipe to bitbucket."
  git push bb-origin master
else
  echo "Dry run: will not publish pipe to bitbucket."
  git push bb-origin master --dry-run
fi

cd .. && rm -rf bitbucketMetadataUpdates