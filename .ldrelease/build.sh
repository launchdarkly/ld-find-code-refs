#!/bin/bash

set -ue

if [[ $LD_RELEASE_VERSION == v* ]]; then
  echo "Remove v prefix from version: $LD_RELEASE_VERSION"
  exit 1
fi

echo "Release tag: $LD_RELEASE_TAG"

make build