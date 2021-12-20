#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

# Read from the command line so we can debug this script. Defaults to releaser env variable.
RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}
RELEASE_NOTES="$(make echo-release-notes)"
GITHUB_TOKEN="${LD_RELEASE_SECRETS_DIR}/github_token"

# install gh cli so we can create a release later
brew install gh
gh auth login --with-token < $GITHUB_TOKEN

# clone checkout commit and push all metadata changes to gha repo
rm -rf githubActionsMetadataUpdates
mkdir -p githubActionsMetadataUpdates
git clone git@github.com:launchdarkly/find-code-references.git githubActionsMetadataUpdates
cp build/metadata/github-actions/* githubActionsMetadataUpdates
cd githubActionsMetadataUpdates
git add -u
git commit -m "Release auto update version"

# tag the commit with the release version
# GOTCHA: The gha version was non-semver but now we are restarting it to follow the core ld-find-code-refs semvers.
git tag v$RELEASE_VERSION
git push origin master --tags

# create a github release with release notes
gh release create v$RELEASE_VERSION --notes "$RELEASE_NOTES"

# clean up
cd .. && rm -rf githubActionsMetadataUpdates