#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

sudo .ldrelease/install-circleci.sh

# Read 2 arguments from the command line so we can debug this script.
# Argument 1 is the release version. Defaults to releaser env variable.
RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}

# Argument 2 is the circleci token. Defaults releaser circleci secret.
CIRCLECI_CLI_TOKEN=${2:-"${LD_RELEASE_SECRETS_DIR}/circleci_token"}
CIRCLECI_CLI_HOST="https://circleci.com"

# Validate circleci orb config. No publishing done on this step.
circleci orb validate build/package/circleci/orb.yml || (echo "Unable to validate orb"; exit 1)

if [[ -z "${LD_RELEASE_DRY_RUN}" ]]; then
  echo "Live run: will publish orb to production circleci repo."
  circleci orb publish build/package/circleci/orb.yml launchdarkly/ld-find-code-refs@$RELEASE_VERSION --token $CIRCLECI_CLI_TOKEN --host $CIRCLECI_CLI_HOST
else
  echo "Dry run: will not publish orb to production. Will publish to circleci dev repo instead."
  circleci orb publish build/package/circleci/orb.yml launchdarkly/ld-find-code-refs@dev:$RELEASE_VERSION --token $CIRCLECI_CLI_TOKEN --host $CIRCLECI_CLI_HOST
fi