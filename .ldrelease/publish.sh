#!/bin/bash

# the "publish" makefile target pushes the image to Docker
$(dirname $0)/run-publish-target.sh publish

# TODO: publish to github actions, bitbucket and circleci marketplaces
#./publish-github-actions-metadata.sh
#./publish-bitbucket-metadata.sh
#./publish-circleci.sh