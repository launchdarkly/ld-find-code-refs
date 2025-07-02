#!/bin/bash

echo ${DOCKER_TOKEN} | sudo docker login --username ${DOCKER_USERNAME} --password-stdin

sudo PATH=${PATH} GITHUB_TOKEN=${GITHUB_TOKEN} make products-for-release

mkdir -p ${ARTIFACT_DIRECTORY}

# Copy the Docker image that goreleaser just built into the artifacts - we only do
# this in a dry run, because in a real release the image will be available from
# DockerHub anyway so there's no point in attaching it to the release.
BASE_CODEREFS=ld-find-code-refs
GH_CODEREFS=ld-find-code-refs-github-action
BB_CODEREFS=ld-find-code-refs-bitbucket-pipeline
sudo docker save launchdarkly/${BASE_CODEREFS}:latest | gzip >${ARTIFACT_DIRECTORY}/${BASE_CODEREFS}.tar.gz
sudo docker save launchdarkly/${GH_CODEREFS}:latest | gzip >${ARTIFACT_DIRECTORY}/${GH_CODEREFS}.tar.gz
sudo docker save launchdarkly/${BB_CODEREFS}:latest | gzip >${ARTIFACT_DIRECTORY}/${BB_CODEREFS}.tar.gz

for script in $(dirname $0)/publish-runners/*.sh; do
  source $script
done

dry_run_gha
dry_run_circleci
dry_run_bitbucket
