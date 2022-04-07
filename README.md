# ld-find-code-refs

Command line program for generating flag code references.

This repository provides solutions for configuring [LaunchDarkly code references](https://docs.launchdarkly.com/home/code/code-references) with various systems out-of-the-box, as well as the ability to automate code reference discovery on your own infrastructure using the provided command line interface.

### Documentation quick links

- [Feature guide](https://docs.launchdarkly.com/home/code/code-references)
- [Turn-key configuration options](#turn-key-configuration-options)
- [Execuation via CLI](#execution-via-cli)
  - [Prerequisites](#prerequisites)
  - [Installing](#installing)
    - [MacOS](#macOS)
    - [Linux](#linux)
    - [Windows](#windows)
    - [Docker](#docker)
- [Federal environments](#using-code-references-in-federal-environments)
- [Configuration](#cli-configuration)
  - [Required arguments](docs/CONFIGURATION.md#required-arguments)
  - [All arguments](docs/CONFIGURATION.md#command-line)
  - [Using environment variables](docs/CONFIGURATION.md#environment-variables)
  - [Using a YAML file](docs/CONFIGURATION.md#YAML)
  - [Aliases](docs/ALIASES.md)
  - [Delimiters](docs/CONFIGURATION.md#delimiters)
  - [Ignoring files and directories](docs/CONFIGURATION.md#ignoring-files-and-directories)
- [Searching for unused flags](#searching-for-unused-flags-extinctions)
- [Branch garbage collection](#branch-garbage-collection)

## Turn-key Configuration options

We provide turnkey support for common trigger mechanisms and CI/CD providers. You can also invoke the `ld-find-code-refs` utility from the command line, which can be run in any custom workflow you define, such as from a bash script or a cron job.

| System           | Status                                                                                |
| ---------------- | ------------------------------------------------------------------------------------- |
| GitHub Actions   | [Supported](https://docs.launchdarkly.com/home/code/github-actions)                   |
| CircleCI Orbs    | [Supported](https://docs.launchdarkly.com/home/code/circleci)                    |
| Bitbucket Pipes  | [Supported](https://docs.launchdarkly.com/home/code/bitbucket)         |
| GitLab CI        | [Supported](https://docs.launchdarkly.com/home/code/gitlab) |
| Manually via CLI | [Supported](https://docs.launchdarkly.com/home/code/custom-config)     |

## Execution via CLI

The command line program may be run manually, and executed in an environment of your choosing. The program requires your `git` repo to be cloned locally, and the currently checked out branch will be scanned for code references.

We recommend incorporating `ld-find-code-refs` into your CI/CD build process. `ld-find-code-refs` should run whenever a commit is pushed to your repository.

### Prerequisites

If you are scanning a git repository, `ld-find-code-refs` requires git (tested with version 2.21.0) to be installed on the system path.

All turn-key configuration methods (docker images used by services like CircleCI or GitHub actions) come with git preinstalled.

### Installing

#### macOS

```bash
brew tap launchdarkly/tap
brew install ld-find-code-refs
```

You can now run `ld-find-code-refs`.

#### Linux

We do not yet have repositories set up for our linux packages, but we do upload deb and rpm packages with our [github releases](https://github.com/launchdarkly/ld-find-code-refs/releases/latest).

##### Ubuntu

This shell script can be used to download and install `ld-find-code-refs` on Ubuntu.

```bash
wget -qO- https://api.github.com/repos/launchdarkly/ld-find-code-refs/releases/latest \
	| grep "browser_download_url" \
	| grep "amd64.deb" \
	| cut -d'"' -f4 \
	| wget -qi - -O ld-find-code-refs.amd64.deb

dpkg -i ld-find-code-refs.amd64.deb
```

#### Windows

A Windows executable of `ld-find-code-refs` is available on the [releases page](https://github.com/launchdarkly/ld-find-code-refs/releases/latest). 

#### Docker

`ld-find-code-refs` is available as a [docker image](https://hub.docker.com/r/launchdarkly/ld-find-code-refs). The image provides an entrypoint for `ld-find-code-refs`, to which command line arguments may be passed. If using the entrypoint, your repository to be scanned should be mounted as a volume. Otherwise, you may override the entrypoint and access `ld-find-code-refs` directly from the shell.

```bash
docker pull launchdarkly/ld-find-code-refs
docker run \
  -v /path/to/your/repo:/repo \
  launchdarkly/ld-find-code-refs \
  --dir="/repo"
```

#### Manual

Precompiled binaries for the latest release can be found [here](https://github.com/launchdarkly/ld-find-code-refs/releases/latest). Be sure to install the required [dependencies](#prerequisities) before running `ld-find-code-refs`.

### Using code references in federal environments

If you are using the FedRAMP compliant [LaunchDarkly federal instance](https://docs.launchdarkly.com/home/advanced/federal), the `ld-find-code-refs` binary should be compiled with FIPS 140-2 support by using a tool like [Go+BoringCrypto](https://github.com/golang/go/tree/dev.boringcrypto/misc/boring).

### CLI Configuration

`ld-find-code-refs` provides a number of configuration options to customize how code references are generated and surfaced in your LaunchDarkly dashboard. See [CONFIGURATION.md](docs/CONFIGURATION.md) for details on configuration, and [EXAMPLES.md](docs/EXAMPLES.md) for detailed sample configurations.

Configuration options include, but are not limited to:

<!-- Headers are used here to maintain historic section links -->
- ##### Ignoring files and directories
- ##### Searching for flag key aliases, such as keys stored in variables or evaluated in an SDK wrapper
- ##### Providing flag key delimiters to reduce false positives and false negatives
- ##### Customizing the amount of data stored and displayed by LaunchDarkly
- ##### Exporting code references as a CSV file

### Searching for unused flags (extinctions)

After scanning has completed, `ld-find-code-refs` will search the Git commit history for flags that have become extinct. A flag is considered extinct in a repository if there were code references for the flag at some point in time that were removed. This behavior can be configured to disable or control how many commits will be searched for extinct flags using the [lookback](docs/CONFIGURATION.md#command-line) argument. Extinct flags will be surfaced in the LaunchDarkly UI.

### Branch garbage collection

After scanning has completed, `ld-find-code-refs` will search for and prune code reference data for stale branches. A branch is considered stale if it has references in LaunchDarkly, but no longer exists on the Git remote. As a consequence of this behavior, any code references on local branches or branches belonging only to a remote other than the default one will be removed the next time `ld-find-code-refs` is run on a different branch.

Stale branches may also be removed manually with the `ld-find-code-refs prune` subcommand.

This operation requires your environment to be authenticated for remote access to your repository. Branch cleanup is not currently supported when running `ld-find-code-refs` with Bitbucket pipelines.

