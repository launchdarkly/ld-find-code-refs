#!/bin/bash

set -euo pipefail

SCRIPT="$(cd "$(dirname "$0")" && pwd)/push-to-origin.sh"
root="$(cd "$(dirname "$0")/../.." && pwd)"
workdir=$(mktemp -d "$root/.push-to-origin-test.XXXXXX")
trap 'rm -rf "$workdir"' EXIT

git init --bare --initial-branch=main "$workdir/remote.git" >/dev/null
git clone "$workdir/remote.git" "$workdir/local" >/dev/null 2>&1
cd "$workdir/local"

git config user.email "test@example.com"
git config user.name "test"

echo init > file
git add file
git commit --no-verify -m init >/dev/null
git push -u origin main >/dev/null
main_sha=$(git rev-parse origin/main)

echo bump > file
git add file
git commit --no-verify -m bump >/dev/null
git tag v2.99.0
bump_sha=$(git rev-parse HEAD)

export LD_RELEASE_VERSION=2.99.0
export DRY_RUN=false
"$SCRIPT"

remote_tag=$(git ls-remote --tags origin "refs/tags/v2.99.0" | awk '{print $1}')
if [[ "$remote_tag" != "$bump_sha" ]]; then
  echo "expected origin tag v2.99.0 -> $bump_sha, got ${remote_tag:-empty}" >&2
  exit 1
fi

remote_main=$(git rev-parse origin/main)
if [[ "$remote_main" != "$main_sha" ]]; then
  echo "origin/main moved; expected $main_sha, got $remote_main" >&2
  exit 1
fi

"$SCRIPT"

echo "push-to-origin_test.sh ok"
