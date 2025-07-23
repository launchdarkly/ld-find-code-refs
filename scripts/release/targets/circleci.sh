#!/bin/bash

# Run this in publish step after all version information have been updated.
set -euo pipefail

CIRCLECI_CLI_HOST="https://circleci.com"

install_circleci() (
  # Install the CircleCI CLI tool.
  # https://github.com/CircleCI-Public/circleci-cli
  # Dependencies: curl, cut
  # The version to install and the binary location can be passed in via VERSION and DESTDIR respectively.

  # GitHub's URL for the latest release, will redirect.
  local GITHUB_BASE_URL="https://github.com/CircleCI-Public/circleci-cli"
  local LATEST_URL="${GITHUB_BASE_URL}/releases/latest/"
  local DESTDIR="${DESTDIR:-/usr/local/bin}"
  local VERSION=$(curl -sLI -o /dev/null -w '%{url_effective}' "$LATEST_URL" | cut -d "v" -f 2)

  # Run the script in a temporary directory that we know is empty.
  local SCRATCH=$(mktemp -d || mktemp -d -t 'tmp')
  cd "$SCRATCH"

  error() (
    echo "An error occured installing the tool."
    echo "The contents of the directory $SCRATCH have been left in place to help to debug the issue."
  )

  trap error ERR

  case "$(uname)" in
    Linux)
      OS='linux'
    ;;
    Darwin)
      OS='darwin'
    ;;
    *)
      echo "This operating system is not supported."
      exit 1
    ;;
  esac

  local RELEASE_URL="${GITHUB_BASE_URL}/releases/download/v${VERSION}/circleci-cli_${VERSION}_${OS}_amd64.tar.gz"

  # Download & unpack the release tarball.
  curl -sL --retry 3 "${RELEASE_URL}" | tar zx --strip 1
  install circleci "$DESTDIR"
  command -v circleci

  # Delete the working directory when the install was successful.
  rm -r "$SCRATCH"
)

validate_circleci_orb_config() (
  circleci orb validate build/package/circleci/orb.yml || (echo "Unable to validate orb"; exit 1)
)

publish_circleci() (
  install_circleci
  validate_circleci_orb_config

  if circleci orb list | grep launchdarkly/ld-find-code-refs@$LD_RELEASE_VERSION; then
    echo "Version exists; skipping publishing CircleCI Orb"
  else
    echo "Live run: will publish orb to production circleci repo."
    circleci orb publish build/package/circleci/orb.yml launchdarkly/ld-find-code-refs@$LD_RELEASE_VERSION --token $CIRCLECI_CLI_TOKEN --host $CIRCLECI_CLI_HOST
  fi
)

dry_run_circleci() (
  install_circleci
  validate_circleci_orb_config

  echo "Dry run: will not publish orb to production. Will publish to circleci dev repo instead."
  circleci orb publish build/package/circleci/orb.yml launchdarkly/ld-find-code-refs@dev:$LD_RELEASE_VERSION --token $CIRCLECI_CLI_TOKEN --host $CIRCLECI_CLI_HOST
)
