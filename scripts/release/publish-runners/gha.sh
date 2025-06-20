#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

RELEASE_TAG="v${LD_RELEASE_VERSION}"
RELEASE_NOTES="$(make echo-release-notes)"

# All users of github action to reference major version tag
VERSION_MAJOR="${LD_RELEASE_VERSION%%\.*}"
RELEASE_TAG_MAJOR="v${VERSION_MAJOR}"

setup() (
  # install gh cli so we can create a release later https://github.com/cli/cli/blob/trunk/docs/install_linux.md
  curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
  sudo apt update
  sudo apt install gh

  # use gh cli to login to github and set up git credentials
  gh auth login
  echo "okay here"
  gh auth setup-git
  echo "not here prolly"

  # clone checkout commit and push all metadata changes to gha repo
  mkdir -p githubActionsMetadataUpdates
  gh repo clone launchdarkly/find-code-references githubActionsMetadataUpdates
  cp build/metadata/github-actions/* githubActionsMetadataUpdates
  cd githubActionsMetadataUpdates
  git config user.email "launchdarklyreleasebot@launchdarkly.com"
  git config user.name "LaunchDarklyReleaseBot"
  git add -u
  git commit -m "Release auto update version $LD_RELEASE_VERSION"
)

clean_up() (
  cd .. && rm -rf githubActionsMetadataUpdates
)

publish_gha() (
  setup

  echo "Live run: will publish action to github action marketplace."
  # tag the commit with the release version and create release
  git tag $RELEASE_TAG
  git push origin main --tags
  git tag -f $RELEASE_TAG_MAJOR
  git push -f origin $RELEASE_TAG_MAJOR
  gh release create $RELEASE_TAG --notes "$RELEASE_NOTES"

  clean_up
)

dry_run_gha() (
  setup

  echo "Dry run: will not publish action to github action marketplace."
  git push origin main --tags --dry-run

  clean_up
)
