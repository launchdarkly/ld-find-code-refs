#!/bin/bash

# The "products-for-release" makefile target does a goreleaser build but doesn't push to DockerHub
$(dirname $0)/run-publish-target.sh products-for-release

# Copy the Docker image that goreleaser just built into the artifacts - we only do
# this in a dry run, because in a real release the image will be available from
# DockerHub anyway so there's no point in attaching it to the release.
BASE_CODEREFS=ld-find-code-refs
GH_CODEREFS=ld-find-code-refs-github-action
BB_CODEREFS=ld-find-code-refs-bitbucket-pipeline
sudo docker save launchdarkly/${BASE_CODEREFS}:${LD_RELEASE_VERSION} | gzip >${LD_RELEASE_ARTIFACTS_DIR}/${BASE_CODEREFS}.tar.gz
sudo docker save launchdarkly/${GH_CODEREFS}:${LD_RELEASE_VERSION} | gzip >${LD_RELEASE_ARTIFACTS_DIR}/${GH_CODEREFS}.tar.gz
sudo docker save launchdarkly/${BB_CODEREFS}:${LD_RELEASE_VERSION} | gzip >${LD_RELEASE_ARTIFACTS_DIR}/${BB_CODEREFS}.tar.gz