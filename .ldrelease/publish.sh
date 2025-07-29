#!/bin/bash

echo "brew deploy key is set: " && echo $LAUNCHDARKLY_HOMEBREW_TAP_DEPLOY_KEY != ""
# the "publish" makefile target pushes the image to Docker
$(dirname $0)/run-publish-target.sh publish

# make bitbucket and github known hosts to push successfully
mkdir –m700 ~/.ssh
touch ~/.ssh/known_hosts
chmod 644 ~/.ssh/known_hosts
ssh-keyscan -t rsa bitbucket.org >> ~/.ssh/known_hosts
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

# publish to github actions, bitbucket and circleci marketplaces
$(dirname $0)/publish-bitbucket-metadata.sh
$(dirname $0)/publish-circleci.sh
$(dirname $0)/publish-github-actions-metadata.sh
