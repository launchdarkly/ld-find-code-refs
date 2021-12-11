#!/bin/bash

# the "publish" makefile target pushes the image to Docker
$(dirname $0)/run-publish-target.sh publish

# TODO: publish metadata changes to github actions and bitbucket marketplaces
#./publish-github-actions-metadata.sh
#./publish-bitbucket-metadata.sh