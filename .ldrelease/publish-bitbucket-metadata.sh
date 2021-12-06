#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

mkdir -p bitbucketMetadataUpdates
git clone git@bitbucket.org:launchdarkly/ld-find-code-refs-pipe.git bitbucketMetadataUpdates
cp build/metadata/bitbucket/* bitbucketMetadataUpdates/
cd bitbucketMetadataUpdates
git add -u
git commit -m "Update version"
git push origin master
