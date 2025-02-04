# bucketeer-code-refs

Command line program for generating Bucketeer feature flag code references.

This repository provides solutions for configuring [Bucketeer code references](https://bucketeer.io) with various systems out-of-the-box, as well as the ability to automate code reference discovery on your own infrastructure using the provided command line interface.

### Documentation quick links

- [Execution via CLI](#execution-via-cli)
  - [Prerequisites](#prerequisites)
  - [Installing](#installing)
    - [MacOS](#macOS)
    - [Linux](#linux)
    - [Windows](#windows)
    - [Docker](#docker)
- [Configuration](#cli-configuration)
  - [Required arguments](docs/CONFIGURATION.md#required-arguments)
  - [All arguments](docs/CONFIGURATION.md#command-line)
  - [Using environment variables](docs/CONFIGURATION.md#environment-variables)
  - [Using a YAML file](docs/CONFIGURATION.md#YAML)
  - [Delimiters](docs/CONFIGURATION.md#delimiters)
  - [Ignoring files and directories](docs/CONFIGURATION.md#ignoring-files-and-directories)

## Execution via CLI

The command line program may be run manually, and executed in an environment of your choosing. The program requires your `git` repo to be cloned locally, and the currently checked out branch will be scanned for code references.

We recommend incorporating `bucketeer-code-refs` into your CI/CD build process. `bucketeer-code-refs` should run whenever a commit is pushed to your repository.

### Prerequisites

If you are scanning a git repository, `bucketeer-code-refs` requires git (tested with version 2.21.0) to be installed on the system path.

### Installing

#### macOS

```bash
# TODO: Add installation instructions for macOS
```

#### Linux

```bash
# TODO: Add installation instructions for Linux
```

#### Windows

```bash
# TODO: Add installation instructions for Windows
```

#### Docker

```bash
# TODO: Add Docker instructions
```

### Configuration

`bucketeer-code-refs` provides a number of configuration options to customize how code references are generated and surfaced in your Bucketeer dashboard.

Required configuration:
- `API_KEY`: Your Bucketeer API key
- `BASE_URI`: Your Bucketeer API base URI
- `ENVIRONMENT_ID`: The Bucketeer environment ID
- `REPO_NAME`: The name of your repository
- `REPO_OWNER`: The owner of your repository
- `REPO_TYPE`: The type of repository (GITHUB, GITLAB, BITBUCKET, or CUSTOM)

Optional configuration:
- `BRANCH`: The git branch to scan (defaults to current branch)
- `REVISION`: The git commit SHA to scan (defaults to current commit)
- `DEBUG`: Enable debug logging
- `DRY_RUN`: Run without sending data to Bucketeer
- `IGNORE_SERVICE_ERRORS`: Continue execution even if there are API errors
- `OUT_DIR`: Directory to write CSV output files
- `USER_AGENT`: Custom user agent string for API requests

For detailed configuration options, please refer to the [configuration documentation](docs/CONFIGURATION.md).

