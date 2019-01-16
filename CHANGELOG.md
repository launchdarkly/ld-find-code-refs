# Change log

All notable changes to the ld-find-code-refs program will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

### Changed
- `updateSequenceId` is now an optional parameter. If not provided, data will always be updated. If provided, data will only be updated if the existing `updateSequenceId` is less than the new `updateSequenceId`.
- Flags with keys shorter than 3 characters are no longer supported.
- lines are truncated after 500 characters.
- project layout modified to follow patterns from https://github.com/golang-standards/project-layout
- use `launchdarkly` docker hub namespace

## [0.1.0] - 2019-01-02
### Changed
- `pushTime` CLI arg renamed to `updateSequenceId`. Its type has been changed from timestamp to integer.
  - Note: this is not considered a breaking change as the CLI args are still in flux. After the 1.0 release arg changes will be considered breaking.

### Fixed
- Upserting repos no longer fails on non-existent repos

## [0.0.1] - 2018-12-14
### Added
- Automated release pipeline for github releases and docker images
- Changelog
