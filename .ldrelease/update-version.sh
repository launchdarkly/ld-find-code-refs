#!/bin/bash

set -e

VERSION_GO=internal/version/version.go
VERSION_GO_TEMP=${VERSION_GO}.tmp
sed "s/const Version =.*/const Version = \"${LD_RELEASE_VERSION}\"/g" ${VERSION_GO} > ${VERSION_GO_TEMP}
mv ${VERSION_GO_TEMP} ${VERSION_GO}

VERSION_ORB=build/package/circleci/orb.yml
VERSION_ORB_TEMP=${VERSION_ORB}.tmp
sed -i "s#launchdarkly: launchdarkly/ld-find-code-refs@.*#launchdarkly: launchdarkly/ld-find-code-refs@${LD_RELEASE_VERSION}#g" ${VERSION_ORB}
sed -i "s#- image: launchdarkly/ld-find-code-refs:.*#- image: launchdarkly/ld-find-code-refs:${LD_RELEASE_VERSION}#g" ${VERSION_ORB}