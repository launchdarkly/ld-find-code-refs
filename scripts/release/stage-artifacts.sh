#!/bin/bash

set -euo pipefail

stage_artifacts() (
  local target=$1

  echo "$DOCKER_TOKEN" | sudo docker login --username "$DOCKER_USERNAME" --password-stdin

  sudo PATH="$PATH" GITHUB_TOKEN="$GITHUB_TOKEN" HOMEBREW_GH_TOKEN="$HOMEBREW_GH_TOKEN" make "$target"

  mkdir -p "$ARTIFACT_DIRECTORY"
  cp ./dist/*.deb ./dist/*.rpm ./dist/*.tar.gz ./dist/*.txt "$ARTIFACT_DIRECTORY"
)
