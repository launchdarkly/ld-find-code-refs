# Bitbucket Pipelines Pipe: LaunchDarkly Code references
Job for finding and sending feature flag code references to LaunchDarkly

## YAML Definition
Add the following snippet to the script section of your `bitbucket-pipelines.yml` file:

```yaml
- pipe: launchdarkly/ld-find-code-refs-pipe:2.6.1
  environment:
    LD_ACCESS_TOKEN: "<string>"
    LD_PROJ_KEY: "<string>" # Required unless using 'projects' block in configuration file then it must be omitted.
    # LD_REPO_NAME: "<string>" # Optional.
    # LD_CONTEXT_LINES: "<integer>" # Optional.
    # LD_BASE_URI: "<string>" # Optional.
    # LD_DEBUG: "<boolean>" # Optional.
    # LD_DEFAULT_BRANCH: "<string>" # Optional.
    # LD_DELIMITERS "<string>" # Optional.
    # LD_IGNORE_SERVICE_ERRORS "<boolean>" # Optional.
    # LD_LOOKBACK "<integer>" # Optional.
    # LD_ALLOW_TAGS "<boolean>" #Optional.
```

## Variables

See [the configuration documentation](https://github.com/launchdarkly/ld-find-code-refs/blob/main/docs/CONFIGURATION.md) for information about additional configuration options, such as configuration delimiters and ignoring files and directories.

| Variable                 | Usage |
| --------------------------- | ----- |
| LD_ACCESS_TOKEN (*)       | A LaunchDarkly personal access token with writer-level access, or access to the `code-reference-repository` [custom role](https://docs.launchdarkly.com/v2.0/docs/custom-roles) resource. Should be provided as a [secured repository variable](https://confluence.atlassian.com/bitbucket/variables-in-pipelines-794502608.html) to secure it. |
| LD_PROJ_KEY (*)   | A LaunchDarkly project key. The pipewill search this project for code references in this project. Cannot be combined with `projects` block in configuration file. |
| LD_REPO_NAME | The repository name. Defaults to the current Bitbucket repository. |
| LD_CONTEXT_LINES        | The number of context lines above and below a code reference for the flag parser to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the lines containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided. Default: `2` |
| LD_BASE_URI                 | Set the base URL of the LaunchDarkly server for this configuration. Defaults to https://app.launchdarkly.com |
| LD_DEBUG | Enables verbose debug logging. Default: `false`|
| LD_DEFAULT_BRANCH | The git default branch. The LaunchDarkly UI will default to display code references for this branch. Default: `main`. |
| LD_IGNORE_SERVICE_ERRORS | If enabled, the scanner will terminate with exit code 0 when the LaunchDarkly API is unreachable or returns an unexpected response. Default: `false` |
| LD_LOOKBACK | Sets the number of Git commits to search in history for whether a feature flag was removed from code. May be set to 0 to disabled this feature. Setting this option to a high value will increase search time. Defaults to 10 |

## Details
LaunchDarkly's Code references feature allows you to find source code references to your feature flags within LaunchDarkly. This makes it easy to determine which projects reference your feature flags, and makes cleanup and removal of technical debt easy. For more information, visit our [documentation](https://docs.launchdarkly.com/home/code/code-references). For documentation on the source code of this pipe, see the [source repo](https://github.com/launchdarkly/ld-find-code-refs).


## Prerequisites
A LaunchDarkly personal access token with writer-level access, or access to the `code-reference-repository` [custom role](https://docs.launchdarkly.com/home/members/custom-roles) resource.

## Examples
Minimal configuration:
```yaml
script:
  - pipe: launchdarkly/ld-find-code-refs-pipe:2.6.1
    environment:
      LD_ACCESS_TOKEN: $LD_ACCESS_TOKEN
      LD_PROJ_KEY: $LD_PROJ_KEY
```

Configuration sending 3 context lines to LaunchDarkly:
```yaml
script:
  - pipe: launchdarkly/ld-find-code-refs-pipe:2.6.1
    environment:
      LD_ACCESS_TOKEN: $LD_ACCESS_TOKEN
      LD_PROJ_KEY: $LD_PROJ_KEY
      LD_CONTEXT_LINES: "3"
```

## Support
If you'd like help with this pipe, or you have an issue or feature request, [submit a request](https://support.launchdarkly.com/hc/en-us/requests/new).

If you're reporting an issue, please include:

* the version of the pipe
* relevant logs and error messages
* steps to reproduce

## License
Copyright 2021 Catamorphic, Co.
Licensed under the Apache License, Version 2.0. See [LICENSE.txt](LICENSE.txt) file.
