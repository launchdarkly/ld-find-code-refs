#!/bin/bash

set -euo pipefail

stage_artifacts() (
  local target=$1

  echo "$DOCKER_TOKEN" | sudo docker login --username "$DOCKER_USERNAME" --password-stdin

  # write homebrew key to temporary file for Goreleaser
  if [[ -n "${HOMEBREW_GH_TOKEN:-}" ]]; then
    HOMEBREW_KEY_PATH="/tmp/homebrew-tap-deploy-key"
    echo "$HOMEBREW_GH_TOKEN" > "$HOMEBREW_KEY_PATH"
    chmod 600 "$HOMEBREW_KEY_PATH"
    export HOMEBREW_KEY_PATH
  fi

  sudo PATH="$PATH" GITHUB_TOKEN="$GITHUB_TOKEN" HOMEBREW_KEY_PATH="$HOMEBREW_KEY_PATH" make "$target"

  mkdir -p "$ARTIFACT_DIRECTORY"
  cp ./dist/*.deb ./dist/*.rpm ./dist/*.tar.gz ./dist/*.txt "$ARTIFACT_DIRECTORY"
)
