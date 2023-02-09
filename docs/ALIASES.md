
# Feature flag aliases

Aliases may be generated to find indirect references to feature flags, such as flag keys stored in variables or wrapped SDK code. Once aliases are generated, `ld-find-code-refs` will automatically scan for aliases in addition to flag keys, and surface them in the LaunchDarkly dashboard.

To generate aliases for your flag keys, you may use any combination of the patterns described below. Configuration types may be used in conjunction and defined more than once for comprehensive alias coverage.

Alias patterns are defined using a YAML file stored in your repository at `.launchdarkly/coderefs.yaml`.

## Alias scope

⚠️ Aliases are not aware of scope. So, adding aliases may introduce false positives. For the best results, we recommend not reusing aliases across multiple feature flags

Don't do this:
```
var featureFlag = 'first-flag-key';

...

var featureFlag = 'second-flag-key';
```

Do this instead:
```
var firstFeatureFlag = 'first-flag-key'

...

var secondFeatureFlag = 'second-flag-key'
```

## Configuring aliases

### Hardcoded map of flag keys to aliases

Aliases can be hardcoded using the `literal` type. This is intended to be used for testing aliasing functionality.

Example hardcoding aliases for a couple flags:

```yaml
aliases:
  - type: literal
    flags:
      my-flag:
        - myFlag
        - isMyFlagOn
      my-other-flag:
        - other.flag.alias
```

### Flag keys transposed to common casing conventions

Aliases can be generated using any of the following common naming conventions. For more robust patterns, see the other available options below this section.

Example flag key: `AnyKind.of_key`

| Type             | After             |
|------------------|-------------------|
| `camelcase`      | `anyKind.ofKey`   |
| `pascalcase`     | `AnyKind.OfKey`   |
| `snakecase`      | `any_kind.of_key` |
| `uppersnakecase` | `ANY_KIND.OF_KEY` |
| `kebabcase`      | `any-kind.of-key` |
| `dotcase`        | `any.kind.of.key` |

Example generating aliases in camelCase and PascalCase:

```yaml
aliases:
  - type: camelcase
  - type: pascalcase
```

### Search files for a specific pattern

You can specify a number of files (`paths`) using [glob patterns](https://en.wikipedia.org/wiki/Glob_(programming)) to search. To achieve the best performance, be as specific as possible with your path globs to minimize the number of files searched for aliases.

You must also specify at least one regular expression (`pattern`) containing a capture group to match aliases. The pattern must contain the the text `FLAG_KEY`, which will be interpolated with flag keys.

Example matching all variable names storing flag keys of the form `var ENABLE_WIDGETS = "enable-widgets"` in .go files do not end with `_test`:

```yaml
aliases:
  - type: filepattern
    paths:
      - '*[!_test].go'
    patterns: 
      - '(\w+) = "FLAG_KEY"'
```

### Execute a command script

For more control over your aliases, you can write a script to generate aliases. The script will receive a flag key as standard input. `ld-find-code-refs` expects a valid JSON array of flag keys output to standard output.

Here's an example of a bash script which returns the the flag key as it's own alias:

```yaml
aliases:
  - type: command
    command: ./launchdarkly/launchdarklyAlias.sh # must be a valid shell command.
    timeout: 5 # seconds
```

Contents of `./launchdarkly/launchdarklyAlias.sh`:

```sh
#! /bin/sh
read flagKey <&0; echo "[\"$flagKey\"]"
```
