# Releasing

This document describes the release process for the 1Money Go SDK.

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

## Commit Convention

Use [Conventional Commits](https://www.conventionalcommits.org/) for meaningful changelog generation:

```
feat: add new customer API endpoint
fix: resolve timeout issue in HTTP client
docs: update installation instructions
refactor: simplify error handling
chore: update dependencies
```

## Release Process

### 1. Tag the commit you want to release

Create and push a semantic version tag that points at `main`:

```bash
git checkout main
git pull origin main
git tag -a v1.2.3 -m "v1.2.3"
git push origin v1.2.3
```

> 💡 Pre-release tags such as `v1.2.3-rc.1` are supported.

### 2. Review the automated release PR

Pushing the tag triggers the **Prepare release PR** workflow which:

1. Updates `version.go` to match the tag.
2. Regenerates `CHANGELOG.md` via `git-cliff`.
3. Opens a pull request from `release/v1.2.3` (or similar) back into `main`.

Review the generated changes, make any necessary edits, and merge the PR like any other change. This is required because branch protection blocks direct pushes to `main`.

### 3. Automatic GitHub Release

Once the release PR merges into `main`, the **Publish GitHub release** job will:

1. Re-tag `v1.2.3` so it points at the merged commit.
2. Generate release notes from the merged history.
3. Publish the GitHub Release with the changelog body.
4. Ping `pkg.go.dev` so the new version is indexed.

No additional manual steps are needed—just monitor the workflow run to ensure it succeeds.

### 4. Verify the release

- Check the [GitHub Releases](https://github.com/1Money-Co/1money-go-sdk/releases) page
- Verify the package is available on [pkg.go.dev](https://pkg.go.dev/github.com/1Money-Co/1money-go-sdk)

## Pre-release Versions

Tags containing `-alpha`, `-beta`, or `-rc` will be marked as pre-releases on GitHub.

Examples:
- `v1.0.0-alpha.1`
- `v1.0.0-beta.2`
- `v1.0.0-rc.1`

## Local Changelog Preview

To preview the changelog locally, install [git-cliff](https://git-cliff.org/) and run:

```bash
# Install git-cliff
brew install git-cliff  # macOS
# or
cargo install git-cliff

# Generate changelog
git cliff --unreleased
```
