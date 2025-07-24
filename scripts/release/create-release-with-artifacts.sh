#!/bin/bash

set -euo pipefail

TARGET=$1

echo "$DOCKER_TOKEN" | sudo docker login --username "$DOCKER_USERNAME" --password-stdin

sudo PATH="$PATH" GITHUB_TOKEN="$GITHUB_TOKEN" make "$TARGET"

mkdir -p "$ARTIFACT_DIRECTORY"
cp ./dist/*.deb ./dist/*.rpm ./dist/*.tar.gz ./dist/*.txt "$ARTIFACT_DIRECTORY"
