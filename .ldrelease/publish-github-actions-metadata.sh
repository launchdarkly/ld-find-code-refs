#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

mkdir -p githubActionsMetadataUpdates
#git clone git@github.com:launchdarkly/find-code-refs.git githubActionsMetadataUpdates
git clone git@github.com:yusinto/find-code-references.git githubActionsMetadataUpdates
cp build/metadata/github-actions/* githubActionsMetadataUpdates
cd githubActionsMetadataUpdates
git add -u
git commit -m "Release auto update version"
git push origin master

rm -rf githubActionsMetadataUpdates