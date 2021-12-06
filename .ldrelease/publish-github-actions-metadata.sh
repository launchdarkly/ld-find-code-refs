#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

# Read from the command line so we can debug this script. Defaults to releaser env variable.
RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}

rm -rf githubActionsMetadataUpdates
mkdir -p githubActionsMetadataUpdates
git clone git@github.com:launchdarkly/find-code-references.git githubActionsMetadataUpdates
cp build/metadata/github-actions/* githubActionsMetadataUpdates
cd githubActionsMetadataUpdates
git add -u
git commit -m "Release auto update version"

# TODO: how do we reconcile versions between ld-find-code-refs and github-action? The github action version is v13?
git tag v$RELEASE_VERSION
git push origin master --tags

#TODO: create github release

cd .. && rm -rf githubActionsMetadataUpdates