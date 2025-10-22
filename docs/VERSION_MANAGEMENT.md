# Version Management Guide

This SDK uses **git tags** as the source of truth for version management.

## Version Strategy

### Development vs Release

- **Development builds**: Use version from `version.go` (default: `0.0.0-dev`)
- **Release builds**: Automatically use version from git tag

### Version Sources (Priority Order)

1. **Git tag** (highest priority) - for release builds
2. **version.go** - for development builds

## Quick Start

### Creating a New Release

Use the integrated `just` command:

```bash
# Create tag v0.1.0 and update version.go
just version-tag 0.1.0

# Review changes
git diff version.go

# Commit the version change
git add version.go
git commit -m "chore: bump version to 0.1.0"

# Push tag to remote
git push origin v0.1.0
```

### Updating version.go from Existing Tag

If you already have git tags and want to sync `version.go`:

```bash
# Update version.go based on latest git tag
just version-update
```

## Detailed Workflows

### Workflow 1: Create New Release (Recommended)

```bash
# Step 1: Create version tag and update version.go
just version-tag 0.2.0

# Output:
# 🏷️  Creating git tag: v0.2.0
# 📝 Updating version.go...
# ✅ Tag created and version.go updated!

# Step 2: Commit version change
git add version.go
git commit -m "chore: bump version to 0.2.0"

# Step 3: Push both commit and tag
git push origin main
git push origin v0.2.0

# Step 4: Build release binaries
just build-release
```

### Workflow 2: Manual Tag Creation

```bash
# Step 1: Create git tag manually
git tag -a v0.2.0 -m "Release v0.2.0"

# Step 2: Update version.go to match tag
just version-update

# Step 3: Commit and push
git add version.go
git commit -m "chore: bump version to 0.2.0"
git push origin main
git push origin v0.2.0
```

### Workflow 3: Pre-release Versions

```bash
# Beta release
just version-tag 0.3.0-beta

# Release candidate
just version-tag 1.0.0-rc1

# Alpha release
just version-tag 0.2.0-alpha.1
```

## Version Format

Follows [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH[-PRERELEASE]
```

**Examples:**
- `0.1.0` - Standard release
- `1.0.0` - Major version
- `0.2.0-beta` - Beta release
- `1.0.0-rc1` - Release candidate
- `0.3.0-alpha.2` - Alpha release

**Git tags** should have `v` prefix:
- Git tag: `v0.1.0`
- Version in code: `0.1.0`

## How It Works

### Build-Time Version Detection

The `Justfile` automatically detects version:

```makefile
# Try to get version from git tag first
GIT_TAG := `git describe --tags --exact-match 2>/dev/null || echo ""`
VERSION_FROM_FILE := `grep -o 'Version = ".*"' version.go | cut -d'"' -f2`
VERSION := if GIT_TAG != "" { trim_start_match(GIT_TAG, "v") } else { VERSION_FROM_FILE }
```

**Behavior:**
1. If current commit has a tag (e.g., `v0.1.0`) → use `0.1.0`
2. Otherwise → use version from `version.go`

### Version Injection

During build, version is injected via `-ldflags`:

```bash
go build -ldflags="\
  -X main.version=0.1.0 \
  -X main.gitCommit=a1b2c3d \
  -X main.buildTime=2025-10-22_06:35:28"
```

## Commands Reference

### `just version`
Show current version information:
```bash
just version
# Output:
# Version:    0.1.0
# Git Tag:    v0.1.0
# Git Commit: a1b2c3d
# Build Time: 2025-10-22_06:35:28
```

### `just version-tag <version>`
Create new tag and update version.go:
```bash
just version-tag 0.2.0
```

Validates version format and:
- Creates annotated git tag `v<version>`
- Updates `version.go` with new version
- Shows next steps

### `just version-update`
Sync version.go with latest git tag:
```bash
just version-update
```

### `just build-cli`
Build CLI with version injection:
```bash
just build-cli
# Output:
# 🔨 Building CLI tool (v0.1.0)...
# ✅ Binary created at: bin/onemoney-cli
# 📦 Version: 0.1.0 (a1b2c3d)
```

### `just build-release`
Build release binaries for all platforms:
```bash
just build-release
```

## Checking Version

### In Code (SDK)

```go
import "github.com/1Money-Co/1money-go-sdk"

fmt.Println(onemoney.Version)
// Output: 0.1.0 (or 0.0.0-dev for development)
```

### In Code (Client)

```go
client, _ := scp.NewClient(&scp.Config{/*...*/})
fmt.Printf("SDK v%s\n", client.Version())
// Output: SDK v0.1.0
```

### CLI Tool

```bash
# Short version
onemoney-cli --version
# Output: 0.1.0 (commit: a1b2c3d)

# Detailed version
onemoney-cli version
# Output:
# Version:    0.1.0
# Git Commit: a1b2c3d
# Build Time: 2025-10-22_06:35:28
# Go Version: go1.25.2
```

### User-Agent Header

Every HTTP request includes:
```
User-Agent: OneMoney-Go-SDK/0.1.0 (Go/go1.25.2; darwin/arm64)
```

## Best Practices

### 1. Always Use Git Tags for Releases

```bash
# ✅ Good - use git tag
just version-tag 0.2.0

# ❌ Bad - manually edit version.go
vim version.go  # Don't do this for releases
```

### 2. Commit Version Changes

Always commit `version.go` after updating:
```bash
git add version.go
git commit -m "chore: bump version to 0.2.0"
```

### 3. Push Tags to Remote

```bash
# Push specific tag
git push origin v0.2.0

# Or push all tags
git push origin --tags
```

### 4. Use Semantic Versioning

- **MAJOR** (1.0.0): Incompatible API changes
- **MINOR** (0.2.0): New functionality, backwards compatible
- **PATCH** (0.1.1): Bug fixes, backwards compatible

### 5. Development Version

Keep `version.go` at `0.0.0-dev` for active development:
```go
const Version = "0.0.0-dev"
```

This makes it clear that the binary is from development, not a release.

## Troubleshooting

### No git tags found

```bash
$ just version-update
❌ No git tags found. Create a tag first with: git tag v0.1.0
```

**Solution:** Create your first tag:
```bash
just version-tag 0.1.0
```

### Version mismatch

If `version.go` doesn't match latest tag:
```bash
just version-update
```

### Build shows wrong version

Make sure you're on a tagged commit:
```bash
# Check current tag
git describe --tags --exact-match

# If not on a tag, it will use version.go
just version
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Important: fetch all tags

      - uses: extractions/setup-just@v1

      - name: Build release binaries
        run: just build-release

      - name: Show version
        run: just version
```

### GitLab CI Example

```yaml
release:
  stage: deploy
  only:
    - tags
  script:
    - just build-release
    - just version
```

## Migration from Manual Versioning

If you previously managed versions manually:

```bash
# 1. Create tag for current version
git tag -a v0.1.0 -m "Release v0.1.0"

# 2. Update version.go to match
just version-update

# 3. Commit
git add version.go
git commit -m "chore: migrate to git tag-based versioning"

# 4. Push
git push origin main v0.1.0
```

## Summary

**For Releases:**
```bash
just version-tag <version>  # Create tag + update version.go
git push origin v<version>   # Push to remote
just build-release          # Build release binaries
```

**For Development:**
- Keep `version.go` at `0.0.0-dev`
- Build with `just build-cli` (uses dev version)

**Version Appears In:**
- ✅ `onemoney.Version` constant
- ✅ `client.Version()` method
- ✅ CLI `--version` flag
- ✅ `User-Agent` HTTP header
- ✅ Build output

This ensures version consistency across your entire SDK! 🎉
