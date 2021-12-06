#!/bin/bash

set -ex
CR_RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}

README_BB=build/metadata/bitbucket/README.md
README_BB_TEMP=${README_BB}.tmp
sed "s#- pipe: launchdarkly/ld-find-code-refs-pipe.*#- pipe: launchdarkly/ld-find-code-refs-pipe:${CR_RELEASE_VERSION}#g" ${README_BB} > ${README_BB_TEMP}
mv ${README_BB_TEMP} ${README_BB}

VERSION_PIPE_CONTAINER=build/metadata/bitbucket/pipe.yml
VERSION_PIPE_CONTAINER_TEMP=${VERSION_PIPE_CONTAINER}.tmp
sed "s#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:.*#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:${CR_RELEASE_VERSION}#g" ${VERSION_PIPE_CONTAINER} > ${VERSION_PIPE_CONTAINER_TEMP}
mv ${VERSION_PIPE_CONTAINER_TEMP} ${VERSION_PIPE_CONTAINER}