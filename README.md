# ld-find-code-refs

Command line program for generating flag code references.

This repository provides solutions for configuring [LaunchDarkly code references](hhttps://docs.launchdarkly.com/v2.0/docs/git-code-references) with various systems out-of-the-box, as well as the ability to automate code reference discovery on your own infrastructure using the provided command line interface.

### Documentation quick links

- [Feature guide](https://docs.launchdarkly.com/docs/git-code-references)
- [Turn-key configuration options](#configuration-options)
- [Execuation via CLI](#execution-via-cli)
- [Prerequisites](#prerequisites)
- [Installing](#installing)
- [Examples](#examples)
- [Required arguments](#required-arguments)
- [Optional arguments](#optional-arguments)
- [Ignoring files and directories](#ignoring-files-and-directories)
- [Branch garbage collection](#branch-garbage-collection)

## Configuration options

We provide turnkey support for common trigger mechanisms and CI / CD providers. You can also invoke the ld-find-code-refs utility from the command line, which can be run in any custom workflow you define (e.g. from a bash script, or a cron job).

| System           | Status                                                                            |
| ---------------- | --------------------------------------------------------------------------------- |
| GitHub Actions   | [Supported](https://docs.launchdarkly.com/v2.0/docs/github-actions)               |
| CircleCI Orbs    | [Supported](https://docs.launchdarkly.com/v2.0/docs/circleci-orbs)                |
| Bitbucket Pipes  | [Supported](https://docs.launchdarkly.com/v2.0/docs/bitbucket-pipes-coderefs)     |
| Manually via CLI | [Supported](https://docs.launchdarkly.com/v2.0/docs/custom-configuration-via-cli) |

## Execution via CLI

The command line program may be run manually, and executed in an environment of your choosing. The program requires your `git` repo to be cloned locally, and the currently checked out branch will be scanned for code references.

We recommend ingraining `ld-find-code-refs` into your CI/CD build process. `ld-find-code-refs` should run whenever a commit is pushed to your repository.

### Prerequisites

`ld-find-code-refs` has two dependencies, which need to be installed in the system path:

| Dependency | Version Tested |
| ---------- | -------------- |
| git        | 2.21.0         |
| ag         | 2.2.0          |

All turn-key configuration methods (docker images used by services like CircleCI or Github actions) come with these dependencies preinstalled.

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

This shell script can be used to download and install `ag` and `ld-find-code-refs` on Ubuntu.

```bash
apt-get install silversearcher-ag

wget -qO- https://api.github.com/repos/launchdarkly/ld-find-code-refs/releases/latest \
	| grep "browser_download_url" \
	| grep "amd64.deb" \
	| cut -d'"' -f4 \
	| wget -qi - -O ld-find-code-refs.amd64.deb

dpkg -i ld-find-code-refs.amd64.deb
```

#### Windows

A Windows executable of `ld-find-code-refs` is available on the [releases page](https://github.com/launchdarkly/ld-find-code-refs/releases/latest). The following Chocolatey command may be used to install the required dependency, `ag`. If you do not have Chocolatey installed, see `ag`'s documentation for [installation instructions](https://github.com/ggreer/the_silver_searcher#windows).

```powershell
choco install ag
```

#### Docker

`ld-find-code-refs` is available as a [docker image](https://hub.docker.com/r/launchdarkly/ld-find-code-refs). The image provides an entrypoint for `ld-find-code-refs`, to which command line arguments may be passed. If using the entrypoint, your git repository to be scanned should be mounted as a volume. Otherwise, you may override the entrypoint and access `ld-find-code-refs` directly from the shell.

```bash
docker pull launchdarkly/ld-find-code-refs
docker run \
  -v /Users/arnold/Documents/projects/launchdarkly/support-service:/repo \
  launchdarkly/ld-find-code-refs \
  -dir="/repo"
```

#### Manual

Precompiled binaries for the latest release can be found [here](https://github.com/launchdarkly/ld-find-code-refs/releases/latest). Be sure to install the required [dependencies](#prerequisities) before running `ld-find-code-refs`

### Examples

The section provides examples of various `bash` commands to execute `ld-find-code-refs` (when installed in the system PATH) with various configurations. We recommend reading through the following examples to gain an understanding of common configurations, as well as the detailed sections below documenting advanced configuration options.

Minimal configuration:

```bash
ld-find-code-refs \
  -accessToken=$YOUR_LAUNCHDARKLY_ACCESS_TOKEN \ # example: api-xxxx
  -projKey=$YOUR_LAUNCHDARKLY_PROJECT_KEY \ # example: my-project
  -repoName=$YOUR_REPOSITORY_NAME \ # example: my-repo
  -dir="/path/to/git/repo"
```

Configuration with [context lines](https://docs.launchdarkly.com/v2.0/docs/git-code-references#section-adding-context-lines) provided:

```bash
ld-find-code-refs \
  -accessToken="$YOUR_LAUNCHDARKLY_ACCESS_TOKEN" \
  -projKey="$YOUR_LAUNCHDARKLY_PROJECT_KEY" \
  -repoName="$YOUR_REPOSITORY_NAME" \
  -dir="/path/to/git/repo" \
  -contextLines=3 # can be up to 5. If < 0, no source code will be sent to LD
```

The above configuration, with the `vendor/` directory and all `css` files ignored by the scanner. The `exclude` parameter may be configuration as any regular expression that matches the files and directories you'd like to ignore for your repository:

```bash
ld-find-code-refs \
  -accessToken="$YOUR_LAUNCHDARKLY_ACCESS_TOKEN" \
  -projKey="$YOUR_LAUNCHDARKLY_PROJECT_KEY" \
  -repoName="$YOUR_REPOSITORY_NAME" \
  -dir="/path/to/git/repo" \
  -contextLines=3
  -exclude="vendor/|\.css" # replace with desired regex pattern
```

A configuration with the the `repoType` set to GitHub, and the `repuUrl` set to a GitHub URL. We recommend configuring these parameters so LaunchDarkly is able to generate reference links to your source code:

```bash
ld-find-code-refs \
  -accessToken="$YOUR_LAUNCHDARKLY_ACCESS_TOKEN" \
  -projKey="$YOUR_LAUNCHDARKLY_PROJECT_KEY" \
  -repoName="$YOUR_REPOSITORY_NAME" \
  -dir="/path/to/git/repo" \
  -contextLines=3
  -repoType="github"
  -repoUrl="$YOUR_REPOSITORY_URL" # example: https://github.com/launchdarkly/ld-find-code-refs
```

The above configuration with left and right carets specified as flag key delimiters:

```bash
ld-find-code-refs \
  -accessToken="$YOUR_LAUNCHDARKLY_ACCESS_TOKEN" \
  -projKey="$YOUR_LAUNCHDARKLY_PROJECT_KEY" \
  -repoName="$YOUR_REPOSITORY_NAME" \
  -dir="/path/to/git/repo" \
  -contextLines=3
  -repoType="github"
  -repoUrl="$YOUR_REPOSITORY_URL" # example: https://github.com/launchdarkly/ld-find-code-refs
  -D="<"
  -D=">"
```

### Required arguments

A number of command-line arguments are available to the code ref finder, some optional, and some required. Command line arguments may be passed to the program in any order.

| Option        | Description                                                                                                                                                                                                                                    |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `accessToken` | LaunchDarkly [personal access token](https://docs.launchdarkly.com/docs/api-access-tokens) with writer-level access, or access to the `code-reference-repository` [custom role](https://docs.launchdarkly.com/v2.0/docs/custom-roles) resource |
| `dir`         | Path to existing checkout of the git repo. The currently checked out branch will be scanned for code references.                                                                                                                               |
| `projKey`     | A LaunchDarkly project key.                                                                                                                                                                                                                    |
| `repoName`    | Git repo name. Will be displayed in LaunchDarkly. Repo names must only contain letters, numbers, '.', '\_' or '-'."                                                                                                                            |

### Optional arguments

Although these arguments are optional, a (\*) indicates a recommended parameter that adds great value if configured.

| Option              | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                              | Default                        |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------ |
| `baseUri`           | Set the base URL of the LaunchDarkly server for this configuration. Only necessary if using a private instance of LaunchDarkly.                                                                                                                                                                                                                                                                                                                                          | `https://app.launchdarkly.com` |
| `contextLines` (\*) | The number of context lines to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the line containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided.                                                                                                                                                                  | `2`                            |
| `debug`             | Enables verbose debug logging.                                                                                                                                                                                                                                                                                                                                                                                                                                           | `false`                        |
| `defaultBranch`     | The git default branch. The LaunchDarkly UI will default to display code references for this branch.                                                                                                                                                                                                                                                                                                                                                                     | `master`                       |
| `delimiters` or `D` | Specifies additional delimiters used to match flag keys. Must be a non-control ASCII character. If more than one character is provided in `delimiters`, each character will be treated as a separate delimiter. Will only match flag keys with surrounded by any of the specified delimeters. This option may also be specified multiple times for multiple delimiters. By default, only flags delimited by single-quotes, double-quotes, and backticks will be matched. | `` [" ' `] ``                  |
| `exclude` (\*)      | A regular expression (PCRE) defining the files and directories which the flag finder should exclude. Partial matches are allowed. Examples: `vendor/`, `\.css`, `vendor/\|\.css`                                                                                                                                                                                                                                                                                         |                                |
| `updateSequenceId`  | An integer representing the order number of code reference updates. Used to version updates across concurrent executions of the program. If not provided, data will always be updated. If provided, data will only be updated if the existing `updateSequenceId` is less than the new `updateSequenceId`. Examples: the time a `git push` was initiated, CI build number, the current unix timestamp.                                                                    |                                |
| `repoType` (\*)     | The repo service provider. Used to generate repository links in the LaunchDarkly UI. Acceptable values: github\|bitbucket\|custom                                                                                                                                                                                                                                                                                                                                        | `custom`                       |
| `repoUrl` (\*)      | The display url for the repository. If provided for a github or bitbucket repository, LaunchDarkly will attempt to automatically generate source code links. Example: `https://github.com/launchdarkly/ld-find-code-refs`                                                                                                                                                                                                                                                |                                |
| `commitUrlTemplate` | If provided, LaunchDarkly will attempt to generate links to your Git service provider per commit. Example: `https://github.com/launchdarkly/ld-find-code-refs/commit/${sha}`. Allowed template variables: `branchName`, `sha`. If `commitUrlTemplate` is not provided, but `repoUrl` is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each commit.                                                         |                                |
| `hunkUrlTemplate`   | If provided, LaunchDarkly will attempt to generate links to your Git service provider per code reference. Example: `https://github.com/launchdarkly/ld-find-code-refs/blob/${sha}/${filePath}#L${lineNumber}`. Allowed template variables: `sha`, `filePath`, `lineNumber`. If `hunkUrlTemplate` is not provided, but `repoUrl` is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each code reference.      |                                |
| `version`           | If provided, the current `ld-find-code-refs` version number will be logged, and the scanner will exit with a return code of 0.                                                                                                                                                                                                                                                                                                                                           | `false`                        |

### Ignoring files and directories

`ld-find-code-refs` provides multiple methods for ignoring files and directories:

1. All dotfiles and patterns in `.gitignore`, `.hgignore`, and `.ignore` will be excluded by default.
2. Provide a `.ldignore` file in the root directory of your Git repository. All patterns specified in `.ldignore` file will be excluded by the scanner. Patterns must follow the `.gitignore` format as specified here: https://git-scm.com/docs/gitignore#_pattern_format
3. The `exclude` command line option (see above section) may be used to specify a single regular expression for the exclude pattern.

If both `.ldignore` and the `exclude` argument are provided, `ld-find-code-refs` will test against both for file exclusion. Do note that `.ldignore` expects shell glob patterns, while the `exclude` option expects a PCRE-compliant regular expression.

### Branch garbage collection

After scanning has completed, `ld-find-code-refs` will search for and delete any stale branches. A branch is considered stale if it has references in LaunchDarkly, but no longer exists on the Git remote. As a consequence of this behavior, any code references on local branches or branches belonging only to a remote other than the default one will be removed the next time `ld-find-code-refs` is run on a different branch.
