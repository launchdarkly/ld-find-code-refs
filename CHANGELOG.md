# Change log

All notable changes to the ld-find-code-refs program will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## Master
### Added
- The `debug` option is now available to the CircleCI orb.

### Changed
- The command-line [docker image](https://hub.docker.com/r/launchdarkly/ld-find-code-refs) now specifies `ld-find-code-refs` as the entrypoint. The docker image may now be run as `docker run launchdarkly/ld-find-code-refs -accessToken="api-xxx"...`

## [0.5.0] - 2019-02-01
### Added
- Generate deb and rpm packages when releasing artifacts.

### Changed
- Automate Homebrew releases
- Added word boundaries to flag key regexes.
  - This should reduce false positives. E.g. for flag key `cool-feature` we will no longer match `verycool-features`.

## [0.4.0] - 2019-01-30

### Added
- Added support for relative paths to CLI `-dir` parameter.
- Added a new command line argument, `debug`, which enables verbose debug logging.
- `ld-find-code-refs` will now exit early if required dependencies are not installed on the system PATH.

### Changed
- Renamed `parse` package to `coderefs`. The `Parse()` method in the aformentioned package is now `Scan()`.

### Fixed
- `ld-find-code-refs` will no longer erroneously make PATCH API requests to LaunchDarkly when url template parameters have not been configured.


## [0.3.0] - 2019-01-23

### Added
- Added openssh as a dependency for the command-line docker image.

### Changed
- The default for `contextLines` is now 2. To disable sending source code to LaunchDarkly, set the `contextLines` argument to `-1`.
- Improved logging to provide more detailed summaries of actions performed by the scanner.

### Fixed
- Fixed a bug in the CircleCI orb config causing `contextLines` to be a string parameter, instead of an integer.

### Removed
- Removed the `repoHead` parameter. `ld-find-code-refs` now only supports scanning repositories already checked out to the desired branch.
- Removed an unnecessary dependency on openssh in Dockerfiles.

## [0.2.1] - 2019-01-17
### Fixed
- Fix a bug causing an error to be returned when a repository connection to LaunchDarkly does not initially exist on execution.

### Removed
- Removed the `cloneEndpoint` command line argument. `ld-find-code-refs` now only supports scanning existing repository clones.

## [0.2.0] - 2019-01-16
### Fixed
- Use case-sensitive `ag` search so we don't get false positives that look like flag keys but have different casing.

### Changed
- This project has been renamed to `ld-find-code-refs`.
- Logging has been overhauled.
- Project layout has been updated to comply with https://github.com/golang-standards/project-layout.
- `updateSequenceId` is now an optional parameter. If not provided, data will always be updated. If provided, data will only be updated if the existing `updateSequenceId` is less than the new `updateSequenceId`.
- Payload limits have been implemented
  - Flags with keys shorter than 3 characters are no longer supported.
  - Lines are truncated after 500 characters.
  - Search is terminated after 5,000 files are matched.
  - Search is terminated after 5,000 hunks are generated.
  - Number of hunks per file is limited to 1,000.
  - A file can only have 500 hunked lines per flag.
- Use `launchdarkly` docker hub namespace instead of `ldactions`.

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
