# git-flag-parser

Command line program for generating flag code references. Possible command line arguments:
TODO: TURN THIS INTO A TABLE

-accessToken string
LaunchDarkly personal access token with write-level access.

-baseUri string
LaunchDarkly base URI. (default "https://app.launchdarkly.com")

-cloneEndpoint string
If provided, will clone the repo from this endpoint to the provided dir. If authentication is required, this endpoint should be authenticated. Supports the https protocol for git cloning. Example: https://username:password@github.com/username/repository.git

-contextLines int
The number of context lines to send to LaunchDarkly. If < 0, no source code will be sent to LaunchDarkly. If 0, only the lines containing flag references will be sent. If > 0, will send that number of context lines above and below the flag reference. A maximum of 5 context lines may be provided. (default -1)

-defaultBranch string
The git default branch. The LaunchDarkly UI will default to this branch. (default "master")

-dir string
Path to existing checkout of the git repo. If a cloneEndpoint is provided, this option is not required.

-exclude string
Exclude any files with code references that match this regex pattern

-projKey string
LaunchDarkly project key.

-pushTime uint
The time the push was initiated formatted as a unix millis timestamp.

-repoHead string
The HEAD or ref to retrieve code references from. (default "master")

-repoName string
Git repo name. Will be displayed in LaunchDarkly.

-repoOwner string
Git repo owner/org.

-repoType string
github|custom (default "custom")
