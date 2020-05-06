package options

// Option values, set to their default
var (
	// Required
	AccessToken string
	Dir         string
	ProjKey     string
	RepoName    string

	// Optional
	BaseUri       = "https://app.launchdarkly.com"
	DefaultBranch = "master"
	Delimiters    []string
	RepoType      = "custom"
	OutDir        string
	RepoUrl       string

	ContextLines     = 2
	UpdateSequenceId = -1

	Debug               bool
	DryRun              bool
	IgnoreServiceErrors bool

	// Optional, but may be inferred
	Branch            string
	CommitUrlTemplate string
	HunkUrlTemplate   string
)

type option struct {
	v         interface{}
	name      string
	short     string
	usage     string
	required  bool
	directory bool
}

var options = []option{
	{
		v:        &AccessToken,
		name:     "accessToken",
		short:    "t",
		usage:    "LaunchDarkly personal access token with write-level access.",
		required: true,
	},
	{
		v:         &Dir,
		name:      "dir",
		short:     "d",
		usage:     "Path to existing checkout of the git repo.",
		required:  true,
		directory: true,
	},
	{
		v:     &RepoName,
		name:  "repoName",
		short: "r",
		usage: `Git repo name. Will be displayed in LaunchDarkly. Case insensitive.
Repo names must only contain letters, numbers, '.', '_' or '-'."`,
		required: true,
	},
	{
		v:        &ProjKey,
		name:     "projKey",
		short:    "p",
		usage:    `LaunchDarkly project key.`,
		required: true,
	},
	{
		v:     &BaseUri,
		name:  "baseUri",
		short: "U",
		usage: "LaunchDarkly base URI.",
	},
	{
		v:     &DefaultBranch,
		name:  "defaultBranch",
		short: "B",
		usage: `The git default branch. The LaunchDarkly UI will default to this branch.
If not provided, will fallback to 'master'.`,
	},
	{
		v:     &Delimiters,
		name:  "delimiters",
		short: "D",
		usage: `Specifies delimiters used to match flag keys. Must be a non-control ASCII
character. If more than one character is provided, each character will be treated
as a separate delimiter. Will only match flag keys with surrounded by any of the
specified delimiters. This option may also be specified multiple times for multiple
delimiters. By default, no delimiters are set.`,
	},
	{
		v:     &OutDir,
		name:  "outDir",
		short: "o",
		usage: `If provided, will output a csv file containing all code references for
the project to this directory.`,
		directory: true,
	},
	{
		v:     &RepoType,
		name:  "repoType",
		short: "T",
		usage: `The repo service provider. Used to correctly categorize repositories in the
LaunchDarkly UI. Aceptable values: github|bitbucket|custom.`,
	},
	{
		v:     &RepoUrl,
		name:  "repoUrl",
		short: "u",
		usage: `The display url for the repository. If provided for a github or
bitbucket repository, LaunchDarkly will attempt to automatically generate source code links.`,
	},
	{
		v:     &ContextLines,
		name:  "contextLines",
		short: "c",
		usage: `The number of context lines to send to LaunchDarkly. If < 0, no
source code will be sent to LaunchDarkly. If 0, only the lines containing
flag references will be sent. If > 0, will send that number of context
lines above and below the flag reference. A maximum of 5 context lines
may be provided.`,
	},
	{
		v:     &UpdateSequenceId,
		name:  "updateSequenceId",
		short: "s",
		usage: `An integer representing the order number of code reference updates.
Used to version updates across concurrent executions of the flag finder.
If not provided, data will always be updated. If provided, data will
only be updated if the existing "updateSequenceId" is less than the new
"updateSequenceId". Examples: the time a "git push" was initiated, CI
build number, the current unix timestamp.`,
	},
	{
		v:     &Debug,
		name:  "debug",
		usage: "Enables verbose debug logging",
	},
	{
		v:    &DryRun,
		name: "dryRun",
		usage: `If enabled, the scanner will run without sending code references to
LaunchDarkly. Combine with the outDir option to output code references to a CSV.`,
	},
	{
		v:     &IgnoreServiceErrors,
		name:  "ignoreServiceErrors",
		short: "i",
		usage: `If enabled, the scanner will terminate with exit code 0 when the
LaunchDarkly API is unreachable or returns an unexpected response.`,
	},
	{
		v:     &Branch,
		name:  "branch",
		short: "b",
		usage: `The currently checked out git branch. If not provided, branch
name will be auto-detected. Provide this option when using CI systems that
leave the repository in a detached HEAD state.`,
	},
	{
		v:    &CommitUrlTemplate,
		name: "commitUrlTemplate",
		usage: `If provided, LaunchDarkly will attempt to generate links to
your Git service provider per commit.
Example: https://github.com/launchdarkly/ld-find-code-refs/commit/${sha}.
Allowed template variables: 'branchName', 'sha'. If commitUrlTemplate is
not provided, but repoUrl is provided and repoType is not custom,
LaunchDarkly will automatically generate links to the repository for each commit.`,
	},
	{
		v:    &HunkUrlTemplate,
		name: "hunkUrlTemplate",
		usage: `If provided, LaunchDarkly will attempt to generate links to 
your Git service provider per code reference. 
Example: https://github.com/launchdarkly/ld-find-code-refs/blob/${sha}/${filePath}#L${lineNumber}.
Allowed template variables: 'sha', 'filePath', 'lineNumber'. If hunkUrlTemplate is not provided, 
but repoUrl is provided and repoType is not custom, LaunchDarkly will automatically generate
links to the repository for each code reference.`,
	},
}
