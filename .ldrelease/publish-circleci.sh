#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

sudo apt install curl
curl -fLSs https://raw.githubusercontent.com/CircleCI-Public/circleci-cli/master/install.sh | sudo bash

# Read 2 arguments from the command line so we can debug this script.
# Argument 1 is the release version. Defaults to releaser env variable.
RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}

# Argument 2 is the circleci token. Defaults releaser circleci secret.
CIRCLECI_CLI_TOKEN=${2:-"${LD_RELEASE_SECRETS_DIR}/circleci_token"}
CIRCLECI_CLI_HOST="https://circleci.com"

# Validate circleci orb config. No publishing done on this step.
circleci orb validate build/package/circleci/orb.yml || (echo "Unable to validate orb"; exit 1)

# dev publish only, not production
circleci orb publish build/package/circleci/orb.yml launchdarkly/ld-find-code-refs@dev:$RELEASE_VERSION --token $CIRCLECI_CLI_TOKEN --host $CIRCLECI_CLI_HOST

# TODO: uncomment this once we have the circleci token in SSM.
#circleci orb publish build/package/circleci/orb.yml launchdarkly/ld-find-code-refs@$RELEASE_VERSION --token $CIRCLECI_CLI_TOKEN --host $CIRCLECI_CLI_HOST