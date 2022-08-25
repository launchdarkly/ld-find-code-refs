# Releasing

## Versioning
This project adheres to [Semantic Versioning](http://semver.org). Release version tags should be in the form `MAJOR.MINOR.PATCH`, with no leading v. When releasing, be sure to update the version number in [`version.go`](https://github.com/launchdarkly/ld-find-code-refs/blob/main/internal/version/version.go), and in the [CircleCI orb](https://github.com/launchdarkly/ld-find-code-refs/blob/main/build/package/circleci/orb.yml).

## GitHub Releases

This project uses [goreleaser](https://goreleaser.com/) to generate GitHub releases. Releases are automated via CircleCI. To generate a new release, simply tag the commit you want to release and push the tag. If the tag ends in -rc(.+), the github release will be marked as "Pre-release." If you'd like to see how release notes are generated, see the .circleci/config.yml publish job.

Make sure you update the changelog before generating a release.

Once a release has been completed, update the [BitBucket pipelines](https://bitbucket.org/launchdarkly/ld-find-code-refs-pipe) repo with the new version number, and push a tag containing the version number along with your commit. Example release commit: https://bitbucket.org/launchdarkly/ld-find-code-refs-pipe/commits/0b1e920c7322cd495f4fc1a09d339342d32606e4

## Docker Hub

To push a new image version to Docker hub, run `make publish-cli-docker TAG=$VERSION` or `make publish-github-actions-docker TAG=$VERSION`, where `$VERSION` is the version you want to release. This will compile the ld-find-code-refs binary for either the base command line code ref finder or the github actions specialized finder, build a new image with your version tagged, and also point latest at that tag and push both latest and $VERSION to docker hub.

## CircleCI Orb Registry

To publish the CircleCI Orb to the Orb registry, you'll need the [CircleCI CLI](https://circleci.com/docs/2.0/local-cli/) installed, and will need to run `circleci setup` and authenticate with a circle token that has github owner status.

Run `make publish-dev-circle-orb TAG=$VERSION` or `make-publish-release-circle-orb TAG=$VERSION` to publish the orb to the orb registry, where `$VERSION` is the version you want to release. Running `publish-dev-circle-orb` will publish a development-tagged (e.g. `dev:0.0.1`) orb, which can be overwritten, and `publish-release-circle-orb` will publish a release version of the orb, which is immutable. Both dev and release orbs are open to the public, but development orbs are not visible in the list of registered orbs on Circle's [website](https://circleci.com/orbs/registry/?showAll=true).

## Beta builds

To push a beta build, set the `PRERELEASE=true` environment variable before running a release task. e.g. `make publish-all TAG=1.0.0-beta1`. Note: to publish a beta circle ci orb, run `make publish-dev-circle-orb TAG=$VERSION`
