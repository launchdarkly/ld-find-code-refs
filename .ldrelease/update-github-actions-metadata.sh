#!/bin/bash

set -ex

# Read from the command line so we can debug this script. Defaults to releaser env variable.
RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}

README=build/metadata/github-actions/README.md
README_TEMP=${README}.tmp
sed "s#launchdarkly/find-code-references@v.*#launchdarkly/find-code-references@v${RELEASE_VERSION}#g" ${README} > ${README_TEMP}
mv ${README_TEMP} ${README}

Dockerfile=build/metadata/github-actions/Dockerfile
Dockerfile_TEMP=${Dockerfile}.tmp
sed "s#launchdarkly/ld-find-code-refs-github-action:.*#launchdarkly/ld-find-code-refs-github-action:${RELEASE_VERSION}#g" ${Dockerfile} > ${Dockerfile_TEMP}
mv ${Dockerfile_TEMP} ${Dockerfile}