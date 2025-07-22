#!/bin/bash

# Run this in publish step after all version information have been updated.
set -ev

RELEASE_TAG="v${LD_RELEASE_VERSION}"
RELEASE_NOTES="$(make echo-release-notes)"

# All users of github action to reference major version tag
VERSION_MAJOR="${LD_RELEASE_VERSION%%\.*}"
RELEASE_TAG_MAJOR="v${VERSION_MAJOR}"

setup_gha() (
  # install gh cli so we can create a release later https://github.com/cli/cli/blob/trunk/docs/install_linux.md
  curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
  sudo apt update
  sudo apt install gh

  gh auth setup-git

  # clone checkout commit and push all metadata changes to gha repo
  mkdir -p githubActionsMetadataUpdates
  gh repo clone launchdarkly/find-code-references githubActionsMetadataUpdates
  cp build/metadata/github-actions/* githubActionsMetadataUpdates
  cd githubActionsMetadataUpdates
  git config user.email "launchdarklyreleasebot@launchdarkly.com"
  git config user.name "LaunchDarklyReleaseBot"
  git branch -vv
  git add -u
  git commit -m "Release auto update version $LD_RELEASE_VERSION"
  pwd
)

clean_up_gha() (
  cd .. && rm -rf githubActionsMetadataUpdates
)

publish_gha() (
  setup_gha

  if git ls-remote --tags origin "refs/tags/v$VERSION" | grep -q "v$VERSION"; then
    echo "Version exists; skipping publishing GHA"
  else
    echo "Live run: will publish action to github action marketplace."
    # tag the commit with the release version and create release
    git tag $RELEASE_TAG
    git push origin main --tags
    git tag -f $RELEASE_TAG_MAJOR
    git push -f origin $RELEASE_TAG_MAJOR
    gh release create $RELEASE_TAG --notes "$RELEASE_NOTES"
  fi

  clean_up
)

dry_run_gha() (
  setup

  echo "Dry run: will not publish action to github action marketplace."
  cd githubActionsMetadataUpdates
  git show-ref
  git push origin main --tags --dry-run

  clean_up_gha
)
