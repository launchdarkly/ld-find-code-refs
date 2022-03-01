#!/bin/bash

set -ex

# Read from the command line so we can debug this script. Defaults to releaser env variable.
RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}
RELEASE_TAG="v${RELEASE_VERSION}"

README=build/metadata/github-actions/README.md
README_TEMP=${README}.tmp
sed "s#launchdarkly/find-code-references@v.*#launchdarkly/find-code-references@${RELEASE_TAG}#g" ${README} > ${README_TEMP}
mv ${README_TEMP} ${README}

DOCKERFILE=build/metadata/github-actions/Dockerfile
DOCKERFILE_TEMP=${DOCKERFILE}.tmp
sed "s#launchdarkly/ld-find-code-refs-github-action:.*#launchdarkly/ld-find-code-refs-github-action:${RELEASE_VERSION}#g" ${DOCKERFILE} > ${DOCKERFILE_TEMP}
mv ${DOCKERFILE_TEMP} ${DOCKERFILE}