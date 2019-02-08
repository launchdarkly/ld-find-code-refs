# ld-find-code-refs

Command line program for generating flag code references.

This repository provides solutions for configuring [LaunchDarkly code references](hhttps://docs.launchdarkly.com/v2.0/docs/git-code-references) with various systems out-of-the-box, as well as the ability to automate code reference discovery on your own infrastructure using the provided command line interface.

## Configuration options

We provide turnkey support for common trigger mechanisms and CI / CD providers. You can also invoke the ld-find-code-refs utility from the command line, which can be run in any custom workflow you define (e.g. from a bash script, or a cron job).

| System | Status |
|---------------------|---------------------------------------------------------------------------------------------------------|
| GitHub Actions | [Supported](https://docs.launchdarkly.com/v2.0/docs/github-actions) |
| CircleCI Orbs | [Supported](https://docs.launchdarkly.com/v2.0/docs/circleci-orbs) |
| BitBucket Pipelines | [Supported](https://docs.launchdarkly.com/v2.0/docs/bitbucket-pipelines-coderefs)
| Manually via CLI | [Supported](https://docs.launchdarkly.com/v2.0/docs/custom-configuration-via-cli) |
| AWS Lambda jobs | Planned |


## Execution via CLI

The command line program may be run manually, and executed in an environment of your choosing. The program requires your `git` repo to be cloned locally, and the currently checked out branch will be scanned for code references.

### Installing

#### macOS

```shell
brew tap launchdarkly/tap
brew install ld-find-code-refs
```

You can now run `ld-find-code-refs`.

#### Linux
We do not yet have repositories set up for our linux packages, but we do upload deb and rpm packages with our github releases.

##### Ubuntu
This shell script can be used to download and install `ag` and `ld-find-code-refs` on Ubuntu.

```shell
apt-get install silversearcher-ag

wget -qO- https://api.github.com/repos/launchdarkly/ld-find-code-refs/releases/latest \
	| grep "browser_download_url" \
	| grep "amd64.deb" \
	| cut -d'"' -f4 \
	| wget -qi - -O ld-find-code-refs.amd64.deb

dpkg -i ld-find-code-refs.amd64.deb
```

### Windows
A Windows executable of `ld-find-code-refs` is not currently available. If you'd like to test `ld-find-code-refs` on a Windows machine, we recommend using the [docker image](https://github.com/launchdarkly/ld-find-code-refs#docker). Windows 10 users may use the [Ubuntu subsystem for Windows](https://docs.microsoft.com/en-us/windows/wsl/install-win10) as a stopgap.

### Docker
`ld-find-code-refs` is available as a [docker image](https://hub.docker.com/r/launchdarkly/ld-find-code-refs). The command line program is installed in the system path of this docker image.
<!-- TODO: update with entrypoint execution when available -->

```shell
docker pull launchdarkly/ld-find-code-refs
docker run -it launchdarkly/ld-find-code-refs
```

#### Manual

Precompiled binaries for the latest release can be found [here](https://github.com/launchdarkly/ld-find-code-refs/releases/latest).

The `ld-find-code-refs` program requires [Git](https://git-scm.org) and [The Silver Searcher](https://github.com/ggreer/the_silver_searcher#installing) to be installed as a dependency, so make sure these dependencies have been installed and added to your system path before running `ld-find-code-refs`.

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

### Required arguments

A number of command-line arguments are available to the code ref finder, some optional, and some required. Command line arguments may be passed to the program in any order.

| Option | Description |
|-|-|
| `accessToken` | LaunchDarkly [personal access token](https://docs.launchdarkly.com/docs/api-access-tokens) with writer-level access, or access to the `code-reference-repository` [custom role](https://docs.launchdarkly.com/v2.0/docs/custom-roles) resource |
| `dir` | Path to existing checkout of the git repo. The currently checked out branch will be scanned for code references. |
| `projKey` | A LaunchDarkly project key. |
| `repoName` | Git repo name. Will be displayed in LaunchDarkly. Repo names must only contain letters, numbers, '.', '_' or '-'." |

### Optional arguments

Although these arguments are optional, a (*) indicates a recommended parameter that adds great value if configured.

| Option | Description | Default |
|-|-|-|
| `baseUri` | Set the base URL of the LaunchDarkly server for this configuration. Only necessary if using a private instance of LaunchDarkly. | `https://app.launchdarkly.com` |
| `contextLines` (*) | The number of context lines to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the line containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided. | `2` |
| `debug` | Enables verbose debug logging. | `false` |
| `defaultBranch` | The git default branch. The LaunchDarkly UI will default to display code references for this branch. | `master` |
| `exclude` (*) | A regular expression (PCRE) defining the files and directories which the flag finder should exclude. Partial matches are allowed. Examples: `vendor/`, `\.css`, `vendor/\|\.css` | |
| `updateSequenceId` | An integer representing the order number of code reference updates. Used to version updates across concurrent executions of the program. If not provided, data will always be updated. If provided, data will only be updated if the existing `updateSequenceId` is less than the new `updateSequenceId`. Examples: the time a `git push` was initiated, CI build number, the current unix timestamp. | |
| `repoType` (*) | The repo service provider. Used to generate repository links in the LaunchDarkly UI. Acceptable values: github\|bitbucket\|custom | `custom` |
| `repoUrl` (*) | The display url for the repository. If provided for a github or bitbucket repository, LaunchDarkly will attempt to automatically generate source code links. Example: `https://github.com/launchdarkly/ld-find-code-refs` | |
| `commitUrlTemplate` | If provided, LaunchDarkly will attempt to generate links to your Git service provider per commit. Example: `https://github.com/launchdarkly/ld-find-code-refs/commit/${sha}`. Allowed template variables: `branchName`, `sha`. If `commitUrlTemplate` is not provided, but `repoUrl` is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each commit. | |
| `hunkUrlTemplate` | If provided, LaunchDarkly will attempt to generate links to your Git service provider per code reference. Example: `https://github.com/launchdarkly/ld-find-code-refs/blob/${sha}/${filePath}#L${lineNumber}`. Allowed template variables: `sha`, `filePath`, `lineNumber`. If `hunkUrlTemplate` is not provided, but `repoUrl` is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each code reference.  | |
