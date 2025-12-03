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

### 1. Create a Pull Request

All changes must go through a PR to merge into `main`:

```bash
# Create feature branch
git checkout -b feat/your-feature

# Make changes and commit
git add .
git commit -m "feat: your feature description"

# Push and create PR
git push origin feat/your-feature
```

Then create a Pull Request on GitHub and wait for review/approval.

### 2. Merge PR and verify tests

After the PR is merged to `main`, ensure all CI checks pass.

### 3. Create a release tag on GitHub

1. Go to the repository's [Releases page](https://github.com/1Money-Co/1money-go-sdk/releases)
2. Click **"Draft a new release"**
3. Click **"Choose a tag"** and create a new tag (e.g., `v1.2.3`)
4. Set the target branch to `main`
5. Click **"Publish release"** (release notes will be auto-generated)

The release workflow will automatically:

1. Extract version from tag
2. Run tests
3. Update `version.go` with the version number
4. Commit and update the tag to include the version change
5. Generate changelog
6. Update the GitHub Release with changelog
7. Trigger pkg.go.dev indexing

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
