# OneMoney Go SDK - Task Runner
# Inspired by RustFS project style

# ========================================================================================
# Environment Variables
# ========================================================================================

GO := env("GO", "go")
GOLANGCI_LINT := env("GOLANGCI_LINT", "golangci-lint")
HAWKEYE := env("HAWKEYE", "hawkeye")
GOSEC := env("GOSEC", "gosec")
GOVULNCHECK := env("GOVULNCHECK", "govulncheck")
GOIMPORTS := env("GOIMPORTS", "goimports")

BIN_DIR := env("BIN_DIR", "bin")
CLI_NAME := env("CLI_NAME", "onemoney-cli")
MODULE_NAME := env("MODULE_NAME", "github.com/1Money-Co/1money-go-sdk")

# Version information (for build-time injection)
# Try to get version from git tag first, fallback to version.go
GIT_TAG := trim(`git describe --tags --exact-match 2>/dev/null || echo ""`)
VERSION_FROM_FILE := trim(`grep -o 'Version = ".*"' version.go | cut -d'"' -f2 2>/dev/null || echo "dev"`)
VERSION := if GIT_TAG != "" { trim_start_match(GIT_TAG, "v") } else { VERSION_FROM_FILE }
GIT_COMMIT := trim(`git rev-parse --short HEAD 2>/dev/null || echo "unknown"`)
BUILD_TIME := trim(`date -u '+%Y-%m-%d_%H:%M:%S'`)

# Build flags for version injection (using lowercase variable names to match cmd/version.go)
LDFLAGS := "-s -w -X main.version=" + VERSION + " -X main.gitCommit=" + GIT_COMMIT + " -X main.buildTime=" + BUILD_TIME

# ========================================================================================
# Help
# ========================================================================================

[group("📒 Help")]
[private]
default:
    @just --list --list-heading '🚀 OneMoney Go SDK justfile manual page:\n'

[doc("show help")]
[group("📒 Help")]
help: default

[doc("show version information")]
[group("📒 Help")]
version:
    @echo "Version:    {{ VERSION }}"
    @echo "Git Tag:    {{ GIT_TAG }}"
    @echo "Git Commit: {{ GIT_COMMIT }}"
    @echo "Build Time: {{ BUILD_TIME }}"

[doc("update version.go based on latest git tag")]
[group("📒 Help")]
version-update:
    #!/usr/bin/env bash
    set -euo pipefail
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -z "$LATEST_TAG" ]; then
        echo "❌ No git tags found. Create a tag first with: git tag v0.1.0"
        exit 1
    fi
    VERSION="${LATEST_TAG#v}"
    echo "📝 Updating version.go to: $VERSION (from tag: $LATEST_TAG)"
    sed -i.bak "s/const Version = \".*\"/const Version = \"$VERSION\"/" version.go
    rm version.go.bak
    echo "✅ version.go updated to version $VERSION"
    echo "💡 Don't forget to commit this change!"

[doc("create a new git tag and update version.go")]
[group("📒 Help")]
version-tag VERSION_NUM:
    #!/usr/bin/env bash
    set -euo pipefail
    if [[ ! "{{ VERSION_NUM }}" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
        echo "❌ Invalid version format. Use: x.y.z or x.y.z-suffix"
        echo "   Examples: 0.1.0, 1.0.0-beta, 2.1.3-rc1"
        exit 1
    fi
    TAG="v{{ VERSION_NUM }}"
    echo "🏷️  Creating git tag: $TAG"
    git tag -a "$TAG" -m "Release $TAG"
    echo "📝 Updating version.go..."
    sed -i.bak "s/const Version = \".*\"/const Version = \"{{ VERSION_NUM }}\"/" version.go
    rm version.go.bak
    echo "✅ Tag created and version.go updated!"
    echo ""
    echo "Next steps:"
    echo "  1. Review changes: git diff version.go"
    echo "  2. Commit: git add version.go && git commit -m 'chore: bump version to {{ VERSION_NUM }}'"
    echo "  3. Push tag: git push origin $TAG"

# ========================================================================================
# Code Quality
# ========================================================================================

[doc("run `go fmt` and `goimports` to format code")]
[group("👆 Code Quality")]
fmt: hawkeye-fix
    @echo "🔧 Formatting code..."
    gofmt -w -s .
    {{ GOIMPORTS }} -w -local {{ MODULE_NAME }} .
    @echo "✅ Code formatted!"

[doc("check code formatting")]
[group("👆 Code Quality")]
fmt-check:
    @echo "📝 Checking code formatting..."
    @test -z "$(gofmt -l .)" || (echo "❌ Code is not formatted. Run 'just fmt'" && exit 1)
    @echo "✅ Code formatting is correct!"

[doc("run `golangci-lint`")]
[group("👆 Code Quality")]
lint:
    @echo "🔍 Running linter..."
    {{ GOLANGCI_LINT }} run --timeout 5m
    @echo "✅ Linter checks passed!"

alias l := lint

[doc("run `golangci-lint` with auto-fix")]
[group("👆 Code Quality")]
lint-fix:
    @echo "🔧 Running linter with auto-fix..."
    {{ GOLANGCI_LINT }} run --fix --timeout 5m
    @echo "✅ Auto-fix completed!"

[doc("run `fmt` and `lint-fix` at once")]
[group("👆 Code Quality")]
fix: fmt lint-fix
    @echo "✅ All fixes applied!"

[doc("run `fmt-check`, `lint`, and `test` at once")]
[group("👆 Code Quality")]
check: fmt-check lint test
    @echo "✅ All quality checks passed!"

[doc("verify code quality (alias for check)")]
[group("👆 Code Quality")]
verify: check

# ========================================================================================
# Testing
# ========================================================================================

[doc("run unit tests only")]
[group("🧪 Testing")]
test:
    @echo "🧪 Running unit tests..."
    {{ GO }} test -v -race -cover ./...
    @echo "✅ Unit tests passed!"

[doc("run integration tests (requires API credentials)")]
[group("🧪 Testing")]
test-integration:
    @echo "🌐 Running integration tests..."
    @echo "📝 Loading credentials from .env file..."
    INTEGRATION_TEST=true {{ GO }} test -v -race ./scp/...
    @echo "✅ Integration tests passed!"

[doc("run all tests (unit + integration)")]
[group("🧪 Testing")]
test-all:
    @echo "🎯 Running all tests..."
    @just test
    @just test-integration
    @echo "✅ All tests passed!"

[doc("run tests with coverage report")]
[group("🧪 Testing")]
test-coverage:
    @echo "📊 Running tests with coverage..."
    {{ GO }} test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    {{ GO }} tool cover -html=coverage.out -o coverage.html
    @echo "✅ Coverage report generated: coverage.html"

[doc("run integration tests with coverage")]
[group("🧪 Testing")]
test-integration-coverage:
    @echo "📊 Running integration tests with coverage..."
    INTEGRATION_TEST=true {{ GO }} test -v -race -coverprofile=coverage-integration.out -covermode=atomic ./scp/...
    {{ GO }} tool cover -html=coverage-integration.out -o coverage-integration.html
    @echo "✅ Integration coverage report generated: coverage-integration.html"

[doc("run benchmarks")]
[group("🧪 Testing")]
bench:
    @echo "⚡ Running benchmarks..."
    {{ GO }} test -bench=. -benchmem ./...
    @echo "✅ Benchmarks completed!"

# ========================================================================================
# Build
# ========================================================================================

[doc("build the project")]
[group("🔨 Build")]
build:
    @echo "🔨 Building project..."
    {{ GO }} build -v ./...
    @echo "✅ Build completed!"

[doc("build CLI tool with version information")]
[group("🔨 Build")]
build-cli:
    @echo "🔨 Building CLI tool (v{{ VERSION }})..."
    mkdir -p {{ BIN_DIR }}
    cd cmd && {{ GO }} build -o ../{{ BIN_DIR }}/{{ CLI_NAME }} -ldflags="{{ LDFLAGS }}" .
    @echo "✅ Binary created at: {{ BIN_DIR }}/{{ CLI_NAME }}"
    @echo "📦 Version: {{ VERSION }} ({{ GIT_COMMIT }})"

[group("🔨 Build")]
[private]
build-platform os arch ext="":
    @echo "🔨 Building {{ CLI_NAME }} for {{ os }}/{{ arch }}..."
    cd cmd && CGO_ENABLED=0 GOOS={{ os }} GOARCH={{ arch }} {{ GO }} build -ldflags="{{ LDFLAGS }}" -o ../{{ BIN_DIR }}/{{ CLI_NAME }}-{{ os }}-{{ arch }}{{ ext }} .

[doc("build release binaries for all platforms")]
[group("🔨 Build")]
build-release:
    @echo "🏗️ Building release binaries (v{{ VERSION }}) for all platforms..."
    mkdir -p {{ BIN_DIR }}
    just build-platform linux amd64
    just build-platform darwin amd64
    just build-platform darwin arm64
    just build-platform windows amd64 .exe
    @echo "✅ Release binaries (v{{ VERSION }}) created in {{ BIN_DIR }}/"
    @ls -lh {{ BIN_DIR }}/

[doc("install CLI tool globally")]
[group("🔨 Build")]
install:
    @echo "📦 Installing CLI tool..."
    cd cmd && {{ GO }} install
    @echo "✅ CLI tool installed!"

# ========================================================================================
# License Management
# ========================================================================================

[doc("check license headers")]
[group("📝 License")]
hawkeye: hawkeye-check

[group("📝 License")]
[private]
hawkeye-check:
    @echo "📝 Checking license headers with hawkeye..."
    @command -v {{ HAWKEYE }} >/dev/null 2>&1 || (echo "❌ hawkeye not found. Run 'just init' to install" && exit 1)
    {{ HAWKEYE }} check
    @echo "✅ License headers are correct!"

[doc("fix license headers")]
[group("📝 License")]
hawkeye-fix:
    @echo "🔧 Fixing license headers with hawkeye..."
    @command -v {{ HAWKEYE }} >/dev/null 2>&1 || (echo "❌ hawkeye not found. Run 'just init' to install" && exit 1)
    {{ HAWKEYE }} format
    @echo "✅ License headers fixed!"

# ========================================================================================
# Security
# ========================================================================================

[doc("run security audit with gosec")]
[group("🔒 Security")]
security:
    @echo "🔒 Running security audit..."
    @command -v {{ GOSEC }} >/dev/null 2>&1 || (echo "❌ gosec not found. Run 'just init' to install" && exit 1)
    {{ GOSEC }} -fmt=json -out=security-report.json ./...
    @echo "✅ Security report generated: security-report.json"

[doc("check for vulnerabilities")]
[group("🔒 Security")]
vuln:
    @echo "🛡️ Checking for vulnerabilities..."
    @command -v {{ GOVULNCHECK }} >/dev/null 2>&1 || (echo "❌ govulncheck not found. Run 'just init' to install" && exit 1)
    {{ GOVULNCHECK }} ./...
    @echo "✅ Vulnerability check completed!"

# ========================================================================================
# Maintenance
# ========================================================================================

[doc("clean build artifacts and caches")]
[group("🧹 Maintenance")]
clean:
    @echo "🧹 Cleaning build artifacts..."
    {{ GO }} clean -cache -testcache -modcache
    rm -rf {{ BIN_DIR }}/
    rm -rf coverage*.out coverage*.html
    rm -f {{ CLI_NAME }}
    rm -f security-report.json
    @echo "✅ Cleaned!"

[doc("tidy and verify dependencies")]
[group("🧹 Maintenance")]
tidy:
    @echo "🧹 Tidying dependencies..."
    {{ GO }} mod tidy
    {{ GO }} mod verify
    @echo "✅ Dependencies tidied!"

[doc("update all dependencies")]
[group("🧹 Maintenance")]
update:
    @echo "⬆️ Updating dependencies..."
    {{ GO }} get -u ./...
    {{ GO }} mod tidy
    @echo "✅ Dependencies updated!"

# ========================================================================================
# Development
# ========================================================================================

[doc("initialize development environment")]
[group("🚀 Development")]
init:
    @echo "🚀 Initializing development environment..."
    @echo "📦 Installing development tools..."
    {{ GO }} install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    {{ GO }} install golang.org/x/tools/cmd/goimports@latest
    {{ GO }} install github.com/securego/gosec/v2/cmd/gosec@latest
    {{ GO }} install golang.org/x/vuln/cmd/govulncheck@latest
    {{ GO }} install golang.org/x/tools/cmd/goimports@latest
    cargo install hawkeye
    @echo "📥 Downloading dependencies..."
    {{ GO }} mod download
    @echo "✅ Development environment ready!"

[doc("run all pre-commit checks")]
[group("🚀 Development")]
pre-commit: fmt lint hawkeye test
    @echo "✅ All pre-commit checks passed!"

[doc("simulate CI pipeline")]
[group("🚀 Development")]
ci: clean tidy fmt-check lint hawkeye test
    @echo "✅ CI pipeline simulation passed!"

[doc("watch for changes and run tests")]
[group("🚀 Development")]
dev:
    @echo "👀 Starting development mode..."
    @echo "📡 Watching for changes..."
    @command -v watchexec >/dev/null 2>&1 || (echo "❌ watchexec not found. Install: brew install watchexec" && exit 1)
    watchexec -e go -r -- just test

# ========================================================================================
# Information & Statistics
# ========================================================================================

[doc("show project statistics")]
[group("📊 Info")]
stats:
    @echo "📊 Project Statistics:"
    @echo "===================="
    @echo "Go files:"
    @find . -name "*.go" ! -path "./vendor/*" | wc -l
    @echo "Lines of code:"
    @find . -name "*.go" ! -path "./vendor/*" -exec cat {} \; | wc -l
    @echo "Test files:"
    @find . -name "*_test.go" ! -path "./vendor/*" | wc -l
    @echo "Packages:"
    @{{ GO }} list ./... | wc -l

[doc("show project dependencies")]
[group("📊 Info")]
deps:
    @echo "📦 Project Dependencies:"
    @echo "===================="
    {{ GO }} list -m all

[doc("check for outdated dependencies")]
[group("📊 Info")]
deps-outdated:
    @echo "🔍 Checking for outdated dependencies..."
    {{ GO }} list -u -m all

# ========================================================================================
# Tools & Utilities
# ========================================================================================

[doc("create a new service from template")]
[group("🛠️ Tools")]
new-service name:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "🔧 Creating new service: {{name}}"
    mkdir -p scp/services/{{name}}
    printf '%s\n' \
        '// Package {{name}} provides {{name}} service functionality.' \
        'package {{name}}' \
        '' \
        'import (' \
        '    "context"' \
        '    "github.com/1Money-Co/1money-go-sdk/scp"' \
        ')' \
        '' \
        '// Service defines the {{name}} service interface.' \
        '// All supported operations are visible here.' \
        'type Service interface {' \
        '    // Add your methods here' \
        '}' \
        '' \
        '// serviceImpl is the concrete implementation (private).' \
        'type serviceImpl struct {' \
        '    scp.BaseService' \
        '}' \
        '' \
        '// NewService creates a new {{name}} service instance.' \
        '// Returns interface type, not implementation.' \
        'func NewService() Service {' \
        '    return &serviceImpl{}' \
        '}' \
        > scp/services/{{name}}/{{name}}.go
    echo "✅ Service template created at scp/services/{{name}}/{{name}}.go"

[doc("run CLI tool with parameters")]
[group("🛠️ Tools")]
run-cli access-key secret-key:
    @echo "🚀 Running CLI tool..."
    {{ GO }} run cmd/main.go -access-key {{ access-key }} -secret-key {{ secret-key }} echo

[doc("run example code")]
[group("🛠️ Tools")]
example:
    @echo "🚀 Running example..."
    {{ GO }} run main_new.go

[doc("generate API documentation")]
[group("🛠️ Tools")]
docs:
    @echo "📚 Generating documentation..."
    godoc -http=:6060 &
    @echo "✅ Documentation server started at http://localhost:6060"
    @echo "Press Ctrl+C to stop"
