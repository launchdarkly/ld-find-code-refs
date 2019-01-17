# ld-find-code-refs

Command line program for generating flag code references.

This repository provides solutions for configuring [LaunchDarkly code references](hhttps://docs.launchdarkly.com/v2.0/docs/git-code-references) with various systems out-of-the-box, as well as the ability to automate code reference discovery on your own infrastructure using the provided command line interface.

## Configuration options

| System | Status |
|---------------------|---------------------------------------------------------------------------------------------------------|
| GitHub Actions | [Supported](https://docs.launchdarkly.com/v2.0/docs/github-actions) |
| CircleCI Orbs | [Supported](https://docs.launchdarkly.com/v2.0/docs/circleci-orbs) |
| BitBucket Pipelines | [Supported](https://docs.launchdarkly.com/v2.0/docs/bitbucket-pipelines-coderefs)
| Manually via CLI | [Supported](https://docs.launchdarkly.com/v2.0/docs/custom-configuration-via-cli) |
| AWS Lambda jobs | Planned |


## Execution via CLI

The command line program may be run manually, and executed in an environment of your choosing.

### Installing
#### MacOS

```shell
brew tap launchdarkly/tap
brew install ld-find-code-refs
```

You can now run `ld-find-code-refs`.

#### Manual

Precompiled binaries for the latest release can be found [here](https://github.com/launchdarkly/ld-find-code-refs/releases/latest).

The `ld-find-code-refs` program requires [The Silver Searcher](https://github.com/ggreer/the_silver_searcher#installing) to be installed as a dependency, so make sure it has been installed and added to your system path before running the ld-find-code-refs.


### Required arguments

A number of command-line arguments are available to the code ref finder, some optional, and some required. Command line arguments may be passed to the program in any order.

| Option | Description |
|-|-|
| `accessToken` | LaunchDarkly [personal access token](https://docs.launchdarkly.com/docs/api-access-tokens) with writer-level access, or access to the `code-reference-repository` [custom role](https://docs.launchdarkly.com/v2.0/docs/custom-roles) resource |
| `dir` | **Absolute** path to existing checkout of the git repo. Relative paths are not currently supported. |
| `projKey` | A LaunchDarkly project key. |
| `repoName` | Git repo name. Will be displayed in LaunchDarkly. Repo names must only contain letters, numbers, '.', '_' or '-'." |

Here's an example shell invocation of the ld-find-code-refs for one of LaunchDarkly's demo repositories, with the binary located in the current directory, and a minimal configuration:

```shell
./ld-find-code-refs \
  -accessToken="$YOUR_LAUNCHDARKLY_ACCESS_TOKEN" \
  -dir="/path/to/git/repo" \
  -repoName="my-repository" \
  -projKey="default"
```

### Optional arguments

| Option | Description | Default |
|-|-|-|
| `baseUri` | Set the base URL of the LaunchDarkly server for this configuration. Only necessary if using a private instance of LaunchDarkly. | `https://app.launchdarkly.com` |
| `contextLines` | The number of context lines to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the line containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided. | -1 |
| `defaultBranch` | The git default branch. The LaunchDarkly UI will default to display code references for this branch. | "master" |
| `exclude` | A regular expression (PCRE) defining the files and directories which the flag finder should exclude. Partial matches are allowed. Examples: `vendor/`, `vendor/.*` | "" |
| `updateSequenceId` | An integer representing the order number of code reference updates. Used to version updates across concurrent executions of the program. If not provided, data will always be updated. If provided, data will only be updated if the existing `updateSequenceId` is less than the new `updateSequenceId`. Examples: the time a `git push` was initiated, CI build number, the current unix timestamp. | n/a |
| `repoHead` | The branch to scan for code references. Should be provided if the `git push` was initiated on a non-master branch. | "master" | no |
| `repoType` | The repo service provider. Used to generate repository links in the LaunchDarkly UI. Acceptable values: github\|bitbucket\|custom | "custom" |
| `repoUrl` | The display url for the repository. If provided for a github or bitbucket repository, LaunchDarkly will attempt to automatically generate source code links. Example: `https://github.com/launchdarkly/ld-find-code-refs` | "" |
| `commitUrlTemplate` | If provided, LaunchDarkly will attempt to generate links to your Git service provider per commit. Example: `https://github.com/launchdarkly/ld-find-code-refs/commit/${sha}`. Allowed template variables: `branchName`, `sha`. If `commitUrlTemplate` is not provided, but `repoUrl` is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each commit. | "" |
| `hunkUrlTemplate` | If provided, LaunchDarkly will attempt to generate links to your Git service provider per code reference. Example: `https://github.com/launchdarkly/ld-find-code-refs/blob/${sha}/${filePath}#L${lineNumber}`. Allowed template variables: `sha`, `filePath`, `lineNumber`. If `hunkUrlTemplate` is not provided, but `repoUrl` is provided and `repoType` is not custom, LaunchDarkly will automatically generate links to the repository for each code reference.  | "" |

Here's an example shell invocation of the `ld-find-code-refs` for one of LaunchDarkly's demo repositories, with the binary located in the current directory, with all options configured:

```shell
./ld-find-code-refs \
  -accessToken="$YOUR_LAUNCHDARKLY_ACCESS_TOKEN" \
  -dir="/path/to/git/repo" \
  -repoName="SupportService" \
  -projKey="default" \
  -baseUri="https://app.launchdarkly.com" \
  -contextLines=2 \
  -defaultBranch="master" \
  -exclude="vendor/" \
  -repoHead="master" \
  -repoType="github" \
  -repuUrl="https://github.com/launchdarkly/SupportService" \
  -updateSequenceId="$(date +%s)" \
  -commitUrlTemplate="https://github.com/launchdarkly/ld-find-code-refs/commit/${sha}"
  -hunkUrlTemplate="https://github.com/launchdarkly/ld-find-code-refs/blob/${sha}/${filePath}#L${lineNumber}"
```
