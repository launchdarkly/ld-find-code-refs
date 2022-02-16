#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

# Read from the command line so we can debug this script. Defaults to releaser env variable.
RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}
GITHUB_TOKEN=${2:-"$(cat ${LD_RELEASE_SECRETS_DIR}/github_token)"}
RELEASE_NOTES="$(make echo-release-notes)"

# install gh cli so we can create a release later https://github.com/cli/cli/blob/trunk/docs/install_linux.md
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install gh

gh auth login --with-token < $GITHUB_TOKEN

# clone checkout commit and push all metadata changes to gha repo
rm -rf githubActionsMetadataUpdates
mkdir -p githubActionsMetadataUpdates
git clone git@github.com:launchdarkly/find-code-references.git githubActionsMetadataUpdates
cp build/metadata/github-actions/* githubActionsMetadataUpdates
cd githubActionsMetadataUpdates
git add -u
git commit -m "Release auto update version"

# tag the commit with the release version
# GOTCHA: The gha version was non-semver but now we are restarting it to follow the core ld-find-code-refs semvers.
git tag v$RELEASE_VERSION

if [[ -z "${LD_RELEASE_DRY_RUN}" ]]; then
  echo "Live run: will publish action to github action marketplace."
  git push origin master --tags

  # create a github release with release notes
  gh release create v$RELEASE_VERSION --notes "$RELEASE_NOTES"
else
  echo "Dry run: will not publish action to github action marketplace."
  git push origin master --tags --dry-run
fi

# clean up
cd .. && rm -rf githubActionsMetadataUpdates