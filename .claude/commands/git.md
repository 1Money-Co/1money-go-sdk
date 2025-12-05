# Git Operations Guide

Guide for performing git operations in this project. Covers common workflows, branch management, commit conventions, and PR creation.

## Branch Naming Conventions

```
{type}/{description}
```

| Type | Description |
|------|-------------|
| `feature/` | New features |
| `fix/` | Bug fixes |
| `refactor/` | Code refactoring |
| `docs/` | Documentation changes |
| `test/` | Adding or updating tests |
| `chore/` | Maintenance tasks |
| `release/` | Release branches |

Examples:
```bash
git checkout -b feature/add-auto-conversion-rules
git checkout -b fix/external-account-validation
git checkout -b test/improve-e2e-coverage
git checkout -b ryan/update-openapi  # Personal branch
```

## Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, no code change |
| `refactor` | Code change that neither fixes nor adds |
| `perf` | Performance improvement |
| `test` | Adding/updating tests |
| `chore` | Maintenance, dependencies |

### Examples

```bash
# Simple commit
git commit -m "feat(auto-conversion): add CreateRule endpoint"

# With scope and body
git commit -m "$(cat <<'EOF'
fix(external-accounts): handle nil bank routing number

The API allows nil routing numbers for international accounts.
Updated validation to skip routing number check when nil.

Fixes #123
EOF
)"

# Test addition
git commit -m "test(e2e): add edge cases for auto-conversion rules"

# Breaking change
git commit -m "$(cat <<'EOF'
feat(api)!: change ListTransactions response format

BREAKING CHANGE: Response now uses 'items' instead of 'list' field.
Migration: Update all consumers to use resp.Items instead of resp.List.
EOF
)"
```

## Common Workflows

### Start New Feature

```bash
# Ensure main is up to date
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/my-new-feature

# Make changes, then commit
git add .
git commit -m "feat(scope): description"

# Push branch
git push -u origin feature/my-new-feature
```

### Update Branch with Main

```bash
# Option 1: Merge (preserves history)
git checkout feature/my-branch
git fetch origin
git merge origin/main

# Option 2: Rebase (cleaner history)
git checkout feature/my-branch
git fetch origin
git rebase origin/main

# If conflicts, resolve then:
git add .
git rebase --continue
```

### Amend Last Commit

```bash
# Check you own the commit first
git log -1 --format='%an %ae'

# Amend (only if not pushed or you're the author)
git add .
git commit --amend -m "new message"

# Or keep same message
git commit --amend --no-edit
```

### Interactive Rebase (Squash Commits)

```bash
# Squash last N commits
git rebase -i HEAD~N

# In editor, change 'pick' to 'squash' or 's' for commits to combine
# Save and edit the combined commit message
```

### Stash Changes

```bash
# Stash current changes
git stash

# Stash with message
git stash push -m "WIP: feature description"

# List stashes
git stash list

# Apply and remove latest stash
git stash pop

# Apply specific stash
git stash apply stash@{1}

# Drop stash
git stash drop stash@{0}
```

### Undo Operations

```bash
# Undo last commit, keep changes staged
git reset --soft HEAD~1

# Undo last commit, keep changes unstaged
git reset HEAD~1

# Undo last commit, discard changes (DANGEROUS)
git reset --hard HEAD~1

# Undo changes to specific file
git checkout -- path/to/file

# Undo staged changes
git reset HEAD path/to/file
```

## Pull Request Workflow

### Before Creating PR

```bash
# Ensure branch is up to date
git fetch origin
git rebase origin/main

# Run tests
go test -v ./...

# Run linter (if configured)
golangci-lint run

# Check for uncommitted changes
git status
```

### Create PR with gh CLI

```bash
# Simple PR
gh pr create --title "feat: add new feature" --body "Description of changes"

# PR with full template
gh pr create --title "feat(auto-conversion): add CreateRule endpoint" --body "$(cat <<'EOF'
## Summary
- Added CreateRule method to auto-conversion-rules service
- Added request/response types
- Added e2e tests

## Test plan
- [ ] Run `go test -v ./tests/e2e/... -run TestAutoConversionRulesTestSuite`
- [ ] Verify rule creation in sandbox environment

## Related
- Closes #123
EOF
)"

# Create draft PR
gh pr create --draft --title "WIP: feature" --body "Work in progress"

# Create PR targeting specific base branch
gh pr create --base develop --title "feat: new feature"
```

### Review PR

```bash
# List open PRs
gh pr list

# View specific PR
gh pr view 123

# Check out PR locally
gh pr checkout 123

# View PR diff
gh pr diff 123

# Approve PR
gh pr review 123 --approve

# Request changes
gh pr review 123 --request-changes --body "Please fix X"

# Add comment
gh pr comment 123 --body "Looks good, just one question..."
```

### Merge PR

```bash
# Merge with merge commit
gh pr merge 123 --merge

# Squash and merge
gh pr merge 123 --squash

# Rebase and merge
gh pr merge 123 --rebase

# Auto-merge when checks pass
gh pr merge 123 --auto --squash
```

## Status and History

### View Status

```bash
# Current status
git status

# Short status
git status -s

# Show branch info
git branch -vv
```

### View History

```bash
# Recent commits
git log --oneline -10

# Commits with diff
git log -p -2

# Commits by author
git log --author="ryan" --oneline

# Commits in date range
git log --since="2024-01-01" --until="2024-02-01" --oneline

# Graph view
git log --oneline --graph --all

# Search commits by message
git log --grep="fix" --oneline
```

### View Changes

```bash
# Unstaged changes
git diff

# Staged changes
git diff --cached

# Changes between branches
git diff main..feature/my-branch

# Changes in specific file
git diff path/to/file

# Show file at specific commit
git show HEAD:path/to/file
```

## Branch Management

### List Branches

```bash
# Local branches
git branch

# Remote branches
git branch -r

# All branches
git branch -a

# Branches with last commit
git branch -v
```

### Delete Branches

```bash
# Delete local branch (safe)
git branch -d feature/old-branch

# Delete local branch (force)
git branch -D feature/old-branch

# Delete remote branch
git push origin --delete feature/old-branch

# Prune deleted remote branches
git fetch --prune
```

### Rename Branch

```bash
# Rename current branch
git branch -m new-name

# Rename specific branch
git branch -m old-name new-name
```

## Tags and Releases

### Create Tags

```bash
# Lightweight tag
git tag v1.0.0

# Annotated tag (recommended)
git tag -a v1.0.0 -m "Release version 1.0.0"

# Tag specific commit
git tag -a v1.0.0 abc1234 -m "Release version 1.0.0"

# Push tag
git push origin v1.0.0

# Push all tags
git push origin --tags
```

### List and Delete Tags

```bash
# List tags
git tag

# List tags with pattern
git tag -l "v1.*"

# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin --delete v1.0.0
```

## Troubleshooting

### Fix Detached HEAD

```bash
# Create branch from detached HEAD
git checkout -b my-branch

# Or return to branch
git checkout main
```

### Recover Deleted Branch

```bash
# Find the commit
git reflog

# Recreate branch
git checkout -b recovered-branch abc1234
```

### Fix Wrong Branch

```bash
# Move commits to correct branch
git stash
git checkout correct-branch
git stash pop

# Or cherry-pick specific commits
git checkout correct-branch
git cherry-pick abc1234
```

### Clean Working Directory

```bash
# Remove untracked files (dry run)
git clean -n

# Remove untracked files
git clean -f

# Remove untracked files and directories
git clean -fd

# Remove ignored files too
git clean -fdx
```

## Safety Guidelines

1. **Never force push to main/master** without explicit request
2. **Always check authorship** before amending commits: `git log -1 --format='%an %ae'`
3. **Never skip hooks** (`--no-verify`) unless explicitly requested
4. **Prefer rebase over merge** for feature branches to keep history clean
5. **Always pull before push** to avoid conflicts
6. **Use `git status`** before committing to verify staged changes
7. **Create backup branch** before risky operations: `git branch backup-branch`

## Quick Reference

| Task | Command |
|------|---------|
| Create branch | `git checkout -b branch-name` |
| Switch branch | `git checkout branch-name` |
| Stage all | `git add .` |
| Commit | `git commit -m "message"` |
| Push | `git push -u origin branch-name` |
| Pull | `git pull origin main` |
| Status | `git status` |
| Log | `git log --oneline -10` |
| Diff | `git diff` |
| Stash | `git stash` |
| Unstash | `git stash pop` |
| Reset file | `git checkout -- file` |
| Undo commit | `git reset --soft HEAD~1` |
