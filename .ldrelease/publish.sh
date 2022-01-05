#!/bin/bash

# the "publish" makefile target pushes the image to Docker
$(dirname $0)/run-publish-target.sh publish

# TODO: publish to github actions, bitbucket and circleci marketplaces
#$(dirname $0)/publish-github-actions-metadata.sh
#$(dirname $0)/publish-bitbucket-metadata.sh
#$(dirname $0)/publish-circleci.sh