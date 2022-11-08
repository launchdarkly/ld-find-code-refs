# Releasing

## Versioning
This project adheres to [Semantic Versioning](http://semver.org). Release version tags should be in the form `MAJOR.MINOR.PATCH`, with no leading v. When releasing, be sure to update the version number in [`version.go`](https://github.com/launchdarkly/ld-find-code-refs/blob/main/internal/version/version.go), and in the [CircleCI orb](https://github.com/launchdarkly/ld-find-code-refs/blob/main/build/package/circleci/orb.yml).

## Releaser

This project uses LD Releaser to publish changes to Docker, GitHub, BitBucket, and CircleCI
