# Releasing

## Github Releases

This project uses [goreleaser](https://goreleaser.com/) to generate github releases. Releases are automated via CircleCI. To generate a new release, simply tag the commit you want to release and push the tag. If the tag ends in -rc(.+), the github release will be marked as "Pre-release." If you'd like to see how release notes are generated, see the .circleci/config.yml publish job.

Make sure you update the changelog before generating a release.

## Docker Hub

To push a new image version to Docker hub, run `make publish-docker TAG=$VERSION`, where `$VERSION` is the version you want to release. This will compile the github-actions binary, build a new image with your version tagged, also point latest at that tag and push both latest and $VERSION to docker hub.
