#!/bin/bash

# This is expected to run after update-bitbucket.sh has updated the verison information.
set -ev

mkdir -p metadataUpdates
git clone git@bitbucket.org:launchdarkly/ld-find-code-refs-pipe.git metadataUpdates/bitbucket
cp build/metadata/bitbucket/* metadataUpdates/bitbucket/
cd metadataUpdates/bitbucket
git add -u
git commit -m "Update version"
git push origin master
