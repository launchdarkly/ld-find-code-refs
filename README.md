# git-flag-parser

Command line program for generating flag code references. Possible command line arguments:
| name | default | description | required |
| --- | --- | --- | --- |
apiKey | n/a | LaunchDarkly personal access token with write-level access. | yes |
dir | n/a | Path to git repo. | yes |
projKey | n/a | LaunchDarkly project key. | yes |
pushTime | n/a | The time the push was initiated formatted as a unix milliseconds timestamp | yes |
repoName | n/a | Git repo name. Will be displayed in LaunchDarkly. | yes |
baseUri | "https://app.launchdarkly.com" | LaunchDarkly base URI. | no |
contextLines | -1 | The number of context lines. If < 0, no source code will be sent to LaunchDarkly. If 0, only the lines containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. | no |
cloneEndpoint | n/a | If provided, will clone the repo from this endpoint to the provided dir. If authentication is required, this endpoint should be authenticated. Supports the https protocol for git cloning. Example: https://username:password@github.com/username/repository.git | no |
exclude | "" | Exclude any files with code references that match this regex pattern | no |
defaultBranch | "master" | The git default branch. The LaunchDarkly UI will default to this branch. | no |
logLevel | "WARN" | PANIC\|FATAL\|ERROR\|WARN\|INFO\|DEBUG\|TRACE | no |
repoHead | "master" | The HEAD or ref to retrieve code references from. | no |
repoOwner | "" | Git repo owner/org. | no |
