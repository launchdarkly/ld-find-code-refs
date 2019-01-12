# git-flag-parser

*Note:* [Code references](https://docs.launchdarkly.com/v2.0/docs/git-code-references) is currently a beta LaunchDarkly feature. If you'd like to join the beta, please email beta@launchdarkly.com.

Command line program for generating flag code references.

This repository provides solutions for configuring [LaunchDarkly code references](hhttps://docs.launchdarkly.com/v2.0/docs/git-code-references) with various systems out-of-the-box, as well as the ability to automate code reference discovery on your own infrastructure using the provided command line interface.

## Configuration options

| System | Status |
|---------------------|---------------------------------------------------------------------------------------------------------|
| GitHub Actions | [Supported](https://docs.launchdarkly.com/v2.0/docs/github-actions) |
| CircleCI Orbs | [Supported](https://docs.launchdarkly.com/v2.0/docs/circleci-orbs) |
| BitBucket Pipelines | [Supported](https://docs.launchdarkly.com/v2.0/docs/bitbucket-pipelines-1)
| Manually via CLI | [Supported](https://docs.launchdarkly.com/v2.0/docs/custom-configuration-via-cli) |
| AWS Lambda jobs | Planned |


## Execution via CLI

The command line program may be run manually, and executed in an environment of your choosing. Downloads for the latest release can be found [here](https://github.com/launchdarkly/git-flag-parser/releases/latest). Additionally, a docker image containing the git flag parser is available on the docker registry as [`ldactions/git-flag-parser`](https://hub.docker.com/r/ldactions/git-flag-parser)

The following options are available to the program:

| Option | Description | Default | Required |
|---------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------|------------------------------------|
| `accessToken` | LaunchDarkly [personal access token](https://docs.launchdarkly.com/docs/api-access-tokens) with writer-level access, or access to the `code-references` [custom role](https://docs.launchdarkly.com/v2.0/docs/custom-roles) resource | n/a | yes |
| `baseUri` | Set the base URL of the LaunchDarkly server for this configuration. Only necessary if using a private instance of LaunchDarkly. | `https://app.launchdarkly.com` | no |
| `cloneEndpoint` | If provided, will clone the repo from this endpoint to the provided `dir`. If authentication is required, this endpoint should be authenticated. Supports the https protocol for git cloning. Example: `https://username:password@github.com/username/repository.git` | n/a | no |
| `contextLines` | The number of context lines to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the line containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided. | -1 | no |
| `defaultBranch` | The git default branch. The LaunchDarkly UI will default to display code references for this branch. | "master" | no |
| `dir` | Path to existing checkout of the git repo. If a cloneEndpoint is provided, this option is not required. |  | only if `cloneEndpoint` is not set |
| `exclude` | A regular expression defining the files and directories which the flag parser should exclude. |  | no |
| `projKey` | A LaunchDarkly project key. |  | yes |
| `updateSequenceId` | An integer representing the order number of code reference updates. Used to version updates across concurrent executions of the flag parser. If not provided, data will always be updated. If provided, data will only be updated if the existing `updateSequenceId` is less than the new `updateSequenceId`. Examples: the time a `git push` was initiated, CI build number, the current unix timestamp. |  | no |
| `repoHead` | The HEAD or ref to retrieve code references from. Should be provided if the `git push` was initiated on a non-master branch. | "master" | no |
| `repoName` | Git repo name. Will be displayed in LaunchDarkly |  | yes |
| `repoType` | The repo service provider. Used to correctly categorize repositories in the LaunchDarkly UI. Acceptable values: github\|bitbucket\|custom | "custom" | no |
| `repoUrl` | The display url for the repository. If provided for a github or bitbucket repository, LaunchDarkly will attempt to automatically generate source code links. Example: `https://github.com/launchdarkly/git-flag-parser` | "" | no |
| `commitUrlTemplate` | If provided, LaunchDarkly will attempt to generate links to your Git service provider per commit. Example: `https://github.com/launchdarkly/git-flag-parser/commit/${sha}`. Allowed template variables: `branchName`, `sha`. If `commitUrlTemplate` is not provided, but `repoUrl` is provided, LaunchDarkly will automatically generate links for github or bitbucket repo types. | "" | no |
| `hunkUrlTemplate` | If provided, LaunchDarkly will attempt to generate links to your Git service provider per code reference. Example: `https://github.com/launchdarkly/git-flag-parser/blob/${sha}/${filePath}#L${lineNumber}`. Allowed template variables: `sha`, `filePath`, `lineNumber`. If `hunkUrlTemplate` is not provided, but `repoUrl` is provided, LaunchDarkly will automatically generate links for github or bitbucket repo types. | "" | no |

## Testing

Set up your development environment by installing Go and running `make init` to install the linter. To lint and run tests, run `make test`.
