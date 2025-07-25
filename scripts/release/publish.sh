#!/bin/bash

set -euo pipefail

source $(dirname $0)/stage-artifacts.sh
stage-artifacts publish

# make bitbucket and github known hosts to push successfully
mkdir -m700 ~/.ssh
touch ~/.ssh/known_hosts
chmod 644 ~/.ssh/known_hosts
ssh-keyscan -t rsa bitbucket.org >> ~/.ssh/known_hosts
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

for script in $(dirname $0)/targets/*.sh; do
  source $script
done

publish_gha
publish_circleci
publish_bitbucket
