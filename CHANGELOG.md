# Change log

All notable changes to the ld-find-code-refs program will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [2.0.0] <PLACEHOLDER DATE>

ℹ️ This release includes breaking changes to the command line tool. If you experience errors or unexpected behavior after upgrading, be sure to read these changelog notes carefully to make adjustments for any breaking changes.

### Added

- Most command line flags can now be [specified in a YAML file](https://github.com/launchdarkly/ld-find-code-refs/blob/master/docs/CONFIGURATION.md#yaml) located in the `.launchdarkly/coderefs.yaml` subdirectory of your repository. [docs](https://github.com/launchdarkly/ld-find-code-refs/blob/master/docs/CONFIGURATION.md#yaml)
  - The following options cannot be specified in YAML, and must be set using the command line or as environment variables:
    - `--dir` / `LD_DIR`
    - `--accessToken` / `LD_ACCESS_TOKEN`
- All command line flags can now be specified as environment variables. [docs](https://github.com/launchdarkly/ld-find-code-refs/blob/master/docs/CONFIGURATION.md#environment-variables)
- When flags with no code references are detected, `ld-find-code-refs` will search Git commit history to detect when the last reference to a feature flag was removed. Use the `--lookback` command line flag to configure the number of commits you would like to search.  The lookback will start at the current commit and will review up to the last n commits to find the last reference of the flag.  The default is 10 commits. 
- Added support for scanning non-git repositories. Use the `--revision` flag to specify your repository version number.
- Added the `prune` sub-command to delete stale code reference data from LaunchDarkly manually by providing a list of branch names as arguments. example: `ld-find-code-refs prune [flags] "branch1" "branch2"`
- The GitHub actions wrapper now supports the `pull_request` event

### Fixed

- Exclude negations in `.ldignore` (lines beginning with an exlamation mark) now correctly include files.

### Changed

- Command line arguments names have been improved. Now, a flag specified with a single dash indicates a shorthand name, while 2 dashes indicate the longform name. Some existing configurations may be invalid, see `ld-find-code-refs --help` for details.
- The default delimiters (single quotes, double quotes and backticks) can now be disabled in the `coderefs.yaml` configuration. [docs](https://github.com/launchdarkly/ld-find-code-refs/blob/master/docs/CONFIGURATION.md#delimiters). Delimiters can no longer be specified using command line flags or environment variables. If you use additional delimiters, or would like to disable delimiters completely, use YAML configuration instead.

### Removed

- The `exclude` command-line option has been removed. Use the `.ldignore` file instead.
- `ld-find-code-refs` no longer requires the silver searcher (ag) as a runtime dependency.

## [1.5.1] - 2020-05-22

### Added

- Added support for specifying a custom default branch for the GitHub actions and Bitbucket pipes wrappers.

## [1.5.0] - 2020-05-11

### Added

- Added the ability to configure flag alias detection using a YAML configuration. See [the README](https://github.com/launchdarkly/ld-find-code-refs#configuring-aliases) for instructions.

### Fixed

- Improved logging around limitations.
- Fixed an edge case where false positives might be picked up for flag keys containing regular expression characters.

## [1.4.0] - 2020-03-16

### Added

- Added a `--ignoreServiceErrors` option to the CLI. If enabled, the scanner will terminate with exit code 0 when the LaunchDarkly API is unreachable or returns an unexpected response.

### Changed

- ld-find-code-refs now requires go1.13 to build.

## [1.3.1] - 2019-09-24

### Fixed

- Fixed a regression causing no references to be found when a relative path is supplied to `dir`

## [1.3.0] - 2019-09-19

### Added

- Added a `--outDir` option to the CLI. If provided, code references will be written to a csv file in `outDir`.
- Added a `--dryRun` option to the CLI. If provided, `ld-find-code-refs` will scan for code references without sending them to LaunchDarkly. May be used in conjunction with `--outDir` to output code references data to a csv file instead of sending data to LaunchDarkly.

### Fixed

- `ld-find-code-refs` now supports scanning repositories with a large number of flags using a pagination strategy. Thanks @cuzzasoft!
- Delimiters will now always be respected when searching for flags referenced in code. This fixes a bug causing references for certain flag keys to match against other flag keys that are substrings of the matched reference.

## [1.2.0] - 2019-08-13

### Added

- Added a `--branch` option to the CLI. This lets a branch name be manually specified when the repo is in a detached head state.
- Github actions v2 support: the github actions wrapper reads the branch name from `GITHUB_REF` and populates the `branch` option with it.

## [1.1.1] - 2019-04-11

### Fixed

- `ld-find-code-refs` will no longer exit with a fatal error when Git credentials have not been configured (required for branch cleanup). Instead, a warning will be logged.

## [1.1.0] - 2019-04-11

### Added

- `ld-find-code-refs` will now remove branches that no longer exist in the git remote from LaunchDarkly.

## [1.0.1] - 2019-03-12

### Changed

- Fixed a potential bug causing `.ldignore` paths to not be detected in some environments.
- When `.ldignore` is found, a debug message is logged.

## [1.0.0] - 2019-02-21

Official release

## [0.7.0] - 2019-02-15

### Added

- Added support for Windows. `ld-find-code-refs` releases will now contain a windows executable.
- Added a new option `-delimiters` (`-D` for short), which may be specified multiple times to specify delimiters used to match flag keys.

### Fixed

- The `dir` command line option was marked as optional, but is actually required. `ld-find-code-refs` will now recognize this option as required.
- `ld-find-code-refs` was performing extra steps to ignore directories for files in directories matched by patterns in `.ldignore`. This ignore process has been streamlined directly into the search so files in `.ldignore` are never scanned.

### Changed

- The command-line [docker image](https://hub.docker.com/r/launchdarkly/ld-find-code-refs) now specifies `ld-find-code-refs` as the entrypoint. See our [documentation](https://github.com/launchdarkly/ld-find-code-refs#docker) for instructions on running `ld-find-code-refs` via docker.
- `ld-find-code-refs` will now only match flag keys delimited by single-quotes, double-quotes, or backticks by default. To add more delimiters, use the `delimiters` command line option.

## [0.6.0] - 2019-02-11

### Added

- Added a new command line argument, `version`. If provided, the current `ld-find-code-refs` version number will be logged, and the scanner will exit with a return code of 0.
- The `debug` option is now available to the CircleCI orb.
- Added support for parsing `.ldignore` files specified in the root directory of the scanned repository. `.ldignore` may be used to specify a pattern (compatible with the `.gitignore` spec: https://git-scm.com/docs/gitignore#_pattern_format) for files to exclude from scanning.

### Changed

- The internal API for specifying the default git branch (`defaultBranch`) has been changed. The `defaultBranch` argument on earlier versions of `ld-find-code-refs` will no longer do anything.

### Fixed

- `ld-find-code-refs` will no longer error out if an unknown error occurs when scanning for code reference hunks within a file. Instead, an error will be logged.

## [0.5.0] - 2019-02-01

### Master

- Added support for parsing `.ldignore` files specified in the root directory of the scanned repository. `.ldignore` may be used to specify a pattern (compatible with the `.gitignore` spec: https://git-scm.com/docs/gitignore#_pattern_format) for files to exclude from scanning.

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
