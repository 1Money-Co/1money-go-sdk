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

### 1. Ensure main branch is ready

```bash
git checkout main
git pull origin main
```

### 2. Verify all tests pass

```bash
just check
just test
```

### 3. Create and push a version tag

```bash
# For a new release
git tag v1.2.3
git push origin v1.2.3

# For pre-release versions
git tag v1.2.3-beta.1
git push origin v1.2.3-beta.1
```

### 4. Automated Release

When a tag matching `v*.*.*` is pushed, the release workflow will:

1. Run tests to verify the release is valid
2. Generate changelog from commits since the last tag
3. Create a GitHub Release with the changelog
4. Trigger pkg.go.dev to index the new version

### 5. Verify the release

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
