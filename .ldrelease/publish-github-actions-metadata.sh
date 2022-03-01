#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

# Read from the command line so we can debug this script. Defaults to releaser env variable.
RELEASE_VERSION=${1:-$LD_RELEASE_VERSION}
RELEASE_TAG=${1:-$LD_RELEASE_TAG}
GITHUB_TOKEN=${2:-"${LD_RELEASE_SECRETS_DIR}/github_token"}
RELEASE_NOTES="$(make echo-release-notes)"

# install gh cli so we can create a release later https://github.com/cli/cli/blob/trunk/docs/install_linux.md
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install gh

# use gh cli to login to github and set up git credentials
gh auth login --with-token < $GITHUB_TOKEN
gh auth setup-git

# clone checkout commit and push all metadata changes to gha repo
rm -rf githubActionsMetadataUpdates
mkdir -p githubActionsMetadataUpdates
gh repo clone launchdarkly/find-code-references githubActionsMetadataUpdates
cp build/metadata/github-actions/* githubActionsMetadataUpdates
cd githubActionsMetadataUpdates
git add -u
git commit -m "Release auto update version"

if [[ -z "${LD_RELEASE_DRY_RUN}" ]]; then
  echo "Live run: will publish action to github action marketplace."

  # tag the commit with the release version and create release
  git tag $RELEASE_TAG
  git push origin master --tags
  gh release create $RELEASE_TAG --notes "$RELEASE_NOTES"
else
  echo "Dry run: will not publish action to github action marketplace."
  git push origin master --tags --dry-run
fi

# clean up
cd .. && rm -rf githubActionsMetadataUpdates