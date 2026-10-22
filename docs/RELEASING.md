# Releasing

## Versioning
This project adheres to [Semantic Versioning](http://semver.org). Releases are tagged `vMAJOR.MINOR.PATCH`.

Versions come from [release-please](https://github.com/googleapis/release-please), which reads [Conventional Commit](https://www.conventionalcommits.org) PR titles (enforced by `lint-pr-title.yml`): `fix:` bumps the patch version, `feat:` the minor version, and `feat!:` or a `BREAKING CHANGE:` footer the major version. `chore:`, `docs:`, `ci:` and changes under `.github/` do not trigger a release.

## Cutting a release

1. Merge PRs to `main` as usual. release-please keeps a `chore(main): release X.Y.Z` PR open that bumps the version and adds the `CHANGELOG.md` entry.
2. Review the release PR. Edit `CHANGELOG.md` on its branch if the generated notes need more context.
3. Merge the release PR. `.github/workflows/release-please.yml` then:
   - creates a draft GitHub release and pushes the `vX.Y.Z` tag
   - runs goreleaser once at the tag: pushes the `ld-find-code-refs`, `ld-find-code-refs-github-action` and `ld-find-code-refs-bitbucket-pipeline` Docker images, updates the Homebrew formula, and uploads the same archives, packages and `checksums.txt` to the release
   - publishes the release
   - pushes a `release-X.Y.Z` branch to [launchdarkly/find-code-references](https://github.com/launchdarkly/find-code-references) with the synced `build/metadata/github-actions/` files; the job summary links to open its PR
4. Merge the find-code-references PR, then tag it `vX.Y.Z`, move `v2` to the same commit, and create its release.

Every file that carries the version is annotated with `x-release-please-version` and listed under `extra-files` in `release-please-config.json`. Add any new one there.

## Dry runs and recovery

- `make release-snapshot` builds every artifact locally without publishing (needs goreleaser v2 and Docker).
- **Manual Publish** (`.github/workflows/manual-publish.yml`) builds an existing tag. Leave `dry_run` checked to build without pushing anything, or uncheck it to re-publish a release whose publish job failed.

## Not automated

The Bitbucket pipe (`build/metadata/bitbucket/`) and CircleCI orb (`build/package/circleci/orb.yml`) are not published by this workflow.
