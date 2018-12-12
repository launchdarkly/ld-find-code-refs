# git-flag-parser

Command line program for generating flag code references.

This repository provides solutions for configuring [LaunchDarkly code references](https://docs.launchdarkly.com) <!-- TODO: Real docs link --> with various systems out-of-the-box, as well as the ability to automate code reference discovery on your own infrastructure using the provided command line interface.

## Configuration options
| System | Status |
|---------------------|---------------------------------------------------------------------------------------------------------|
| GitHub Actions | [Supported](https://github.com/launchdarkly/git-flag-parser/tree/master/parse/github-actions/README.md) |
| BitBucket Pipelines | Planned |
| CircleCI Workflows | Planned |
| AWS Lambda jobs | Planned |
| Manually via CLI | [Supported](https://github.com/launchdarkly/git-flag-parser/tree/master/README.md#execution-via-cli) |


## Execution via CLI
<!-- TODO: Link to latest binary / dockerfile when released -->
The command line program may be run manually, and executed in an environment of your choosing. The following options are available to the program:

| Option | Description | Default | Required |
|---------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------|------------------------------------|
| `accessToken` | LaunchDarkly [personal access token](https://docs.launchdarkly.com/docs/api-access-tokens) with writer-level access, or access to the `code-references` [custom role](https://docs.launchdarkly.com/v2.0/docs/custom-roles) resource | n/a | yes |
| `baseUri` | Set the base URL of the LaunchDarkly server for this configuration. Only necessary if using a private instance of LaunchDarkly. | "https://app.launchdarkly.com" | no |
| `cloneEndpoint` | If provided, will clone the repo from this endpoint to the provided `dir`. If authentication is required, this endpoint should be authenticated. Supports the https protocol for git cloning. Example: https://username:password@github.com/username/repository.git | n/a | no |
| `contextLines` | The number of context lines to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the line containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided. | -1 | no |
| `defaultBranch` | The git default branch. The LaunchDarkly UI will default to display code references for this branch. | "master" | no |
| `dir` | Path to existing checkout of the git repo. If a cloneEndpoint is provided, this option is not required. |  | only if `cloneEndpoint` is not set |
| `exclude` | A regular expression defining the files and directories which the flag parser should exclude. |  | no |
| `projKey` | A LaunchDarkly project key. |  | yes |
| `pushTime` | The time the `git push` was initiated, formatted as a unix millis timestamp. Used by the LaunchDarkly API to correctly order updates. |  | yes |
| `repoHead` | The HEAD or ref to retrieve code references from. Should be provided if the `git push` was initiated on a non-master branch. | "master" | no |
| `repoName` | Git repo name. Will be displayed in LaunchDarkly |  | yes |
| `repoType` | The repo service provider. Used to correctly categorize repositories in the LaunchDarkly UI. Acceptable values: github\|bitbucket\|custom | "custom" | no |
| `repoUrl` | The display url for the repository. If provided for a github or bitbucket repository, LaunchDarkly will attempt to automatically generate source code links. | "" | no |
| `branchUrlTemplate` | If provided, LaunchDarkly will attempt to generate links to your Git service provider per branch. Example: 'https://github.com/launchdarkly/git-flag-parser/tree/${branchName}'. Allowed template variables: branchName, sha. If branchUrlTemplate is not provided, but repoUrl is provided, LaunchDarkly will automatically generates links for github and bitbucket repo types. | "" | no |
| `hunkUrlTemplate` | If provided, LaunchDarkly will attempt to generate links to your Git service provider per code reference. Example: 'https://github.com/launchdarkly/git-flag-parser/blob/${sha}/${filePath}#L${lineNumber}'. Allowed template variables: sha, filePath, lineNumber. If hunkUrlTemplate is not provided, but repoUrl is provided, LaunchDarkly will automatically generate links for github and bitbucket repo types. | "" | no |
