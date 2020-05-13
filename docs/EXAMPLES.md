# Examples

The section provides examples of various `bash` commands to execute `ld-find-code-refs` (when installed in the system PATH) with various configurations. We recommend reading through the following examples to gain an understanding of common configurations. For more information on advanced configuration, see [CONFIGURATION.md](CONFIGURATION.md)

## Basic configuration

```bash
ld-find-code-refs \
  --accessToken=$YOUR_LAUNCHDARKLY_ACCESS_TOKEN \ # example: api-xxxx
  --projKey=$YOUR_LAUNCHDARKLY_PROJECT_KEY \ # example: my-project
  --repoName=$YOUR_REPOSITORY_NAME \ # example: my-repo
  --dir="/path/to/git/repo"
```

## Configuration with context lines

https://docs.launchdarkly.com/integrations/git-code-references#configuring-context-lines

```bash
ld-find-code-refs \
  --accessToken="$YOUR_LAUNCHDARKLY_ACCESS_TOKEN" \
  --projKey="$YOUR_LAUNCHDARKLY_PROJECT_KEY" \
  --repoName="$YOUR_REPOSITORY_NAME" \
  --dir="/path/to/git/repo" \
  --contextLines=3 # can be up to 5. If < 0, no source code will be sent to LD
```

## Configuration with repository metadata

A configuration with the the `repoType` set to GitHub, and the `repoUrl` set to a GitHub URL. We recommend configuring these parameters so LaunchDarkly is able to generate reference links to your source code:

```bash
ld-find-code-refs \
  --accessToken="$YOUR_LAUNCHDARKLY_ACCESS_TOKEN" \
  --projKey="$YOUR_LAUNCHDARKLY_PROJECT_KEY" \
  --repoName="$YOUR_REPOSITORY_NAME" \
  --dir="/path/to/git/repo" \
  --contextLines=3
  --repoType="github"
  --repoUrl="$YOUR_REPOSITORY_URL" # example: https://github.com/launchdarkly/ld-find-code-refs
```
