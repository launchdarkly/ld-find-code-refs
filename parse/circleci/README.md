# CircleCI

The flag parser can be used with CircleCI via a reusable [CircleCI orb](https://circleci.com/docs/2.0/orb-intro/) to automate population of code references in LaunchDarkly. CircleCI orbs require a Circle workflow version of 2.1 or greater. If you're using an earlier version, consider manually using the flag parser [binary or docker image](https://github.com/launchdarkly/git-flag-parser/tree/master/README.md#execution-via-cli) to create your own workflow job.

# Setup
Create a [LaunchDarkly personal access token](https://docs.launchdarkly.com/docs/api-access-tokens) with writer-level access, or access to the `code-references` [custom role](https://docs.launchdarkly.com/v2.0/docs/custom-roles) resource. Store this newly created access token as an [environment variable](https://circleci.com/docs/2.0/env-vars/#setting-an-environment-variable-in-a-project) titled `LD_ACCESS_TOKEN` in your CircleCI project settings.

Here's an example minimal configuration, using LaunchDarkly's Orb:

```yaml
version: 2.1

orbs:
  launchdarkly: launchdarkly/git-flag-parser@dev:0.0.1

workflows:
  main:
    jobs:
      - launchdarkly/find-code-references:
          proj_key: default # your LaunchDarkly project key
          repo_type: github # can be 'github', 'bitbucket', or 'custom'
          repo_url: https://github.com/launchdarkly/SupportService # used to generate links to your repository in the LaunchDarkly webapp
```

Documentation for all configuration options may be found here: https://circleci.com/orbs/registry/orb/launchdarkly/git-flag-parser
