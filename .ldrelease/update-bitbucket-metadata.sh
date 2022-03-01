#!/bin/bash

set -ex
# Read from the command line so we can debug this script. Defaults to releaser env variable.
RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}

README=build/metadata/bitbucket/README.md
README_TEMP=${README}.tmp
sed "s#- pipe: launchdarkly/ld-find-code-refs-pipe.*#- pipe: launchdarkly/ld-find-code-refs-pipe:${RELEASE_VERSION}#g" ${README} > ${README_TEMP}
mv ${README_TEMP} ${README}

YML=build/metadata/bitbucket/pipe.yml
YML_TEMP=${YML}.tmp
sed "s#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:.*#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:v${RELEASE_VERSION}#g" ${YML} > ${YML_TEMP}
mv ${YML_TEMP} ${YML}