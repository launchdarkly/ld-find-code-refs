#!/bin/bash

set -e

RELEASE_TAG="v${LD_RELEASE_VERSION}"

update_go() (
  VERSION_GO=internal/version/version.go
  sed -i "s/const Version =.*/const Version = \"${LD_RELEASE_VERSION}\"/g" ${VERSION_GO}
)

update_orb() (
  VERSION_ORB=build/package/circleci/orb.yml
  sed -i "s#launchdarkly: launchdarkly/ld-find-code-refs@.*#launchdarkly: launchdarkly/ld-find-code-refs@${LD_RELEASE_VERSION}#g" ${VERSION_ORB}
  sed -i "s#- image: launchdarkly/ld-find-code-refs:.*#- image: launchdarkly/ld-find-code-refs:${LD_RELEASE_VERSION}#g" ${VERSION_ORB}
)

update_gha() (
  README=build/metadata/github-actions/README.md
  sed -i "s#launchdarkly/find-code-references@v.*#launchdarkly/find-code-references@${RELEASE_TAG}#g" ${README}

  DOCKERFILE=build/metadata/github-actions/Dockerfile
  sed -i "s#launchdarkly/ld-find-code-refs-github-action:.*#launchdarkly/ld-find-code-refs-github-action:${LD_RELEASE_VERSION}#g" ${DOCKERFILE}
)

update_bitbucket() (
  README=build/metadata/bitbucket/README.md
  sed -i "s#- pipe: launchdarkly/ld-find-code-refs-pipe.*#- pipe: launchdarkly/ld-find-code-refs-pipe:${LD_RELEASE_VERSION}#g" ${README}

  YML=build/metadata/bitbucket/pipe.yml
  sed -i "s#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:.*#image: launchdarkly/ld-find-code-refs-bitbucket-pipeline:${LD_RELEASE_VERSION}#g" ${YML}
)

update_go
update_orb
update_gha
update_bitbucket
