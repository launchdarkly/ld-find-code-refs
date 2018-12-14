# Releasing

This project uses [goreleaser](https://goreleaser.com/) to generate github releases and push docker images. Releases are automated via CircleCI. To generate a new release, simply tag the commit you want to release and push the tag. If the tag ends in -rc(\d+), the github release will be marked as "Pre-release." If you'd like to see how release notes are generated, see the .circleci/config.yml publish job.
**Note:** Pre-releases still get the docker `latest` tag.

Make sure you update the changelog before generating a release.
