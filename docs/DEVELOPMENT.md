# Development Guide

This guide will help you set up your development environment and understand the development workflow.

## Prerequisites

### Required Tools

- **Go 1.21+** - [Install Go](https://golang.org/dl/)
- **Just** - Task runner (alternative to Make)
  ```bash
  # macOS
  brew install just

  # Linux
  cargo install just
  # or
  wget -qO - 'https://proget.makedeb.org/debian-feeds/prebuilt-mpr.pub' | gpg --dearmor | sudo tee /usr/share/keyrings/prebuilt-mpr-archive-keyring.gpg 1> /dev/null
  echo "deb [signed-by=/usr/share/keyrings/prebuilt-mpr-archive-keyring.gpg] https://proget.makedeb.org prebuilt-mpr $(lsb_release -cs)" | sudo tee /etc/apt/sources.list.d/prebuilt-mpr.list
  sudo apt update
  sudo apt install just

  # Windows (via scoop)
  scoop install just
  ```

### Recommended Tools

- **golangci-lint** - Linter aggregator
- **goimports** - Auto-format imports
- **license-eye** - License header management
- **gosec** - Security scanner
- **govulncheck** - Vulnerability checker
- **watchexec** - File watcher for development

## Quick Setup

Initialize your development environment with all tools:

```bash
just init
```

This will install:
- golangci-lint
- goimports
- license-eye
- gosec
- govulncheck

## Daily Workflow

### Show All Available Commands

```bash
just
# or
just --list
```

### Building

```bash
# Build the project
just build

# Build CLI tool
just build-cli

# Build release binaries (all platforms)
just build-release
```

### Testing

```bash
# Run all tests
just test

# Run tests with coverage
just test-coverage

# Run benchmarks
just bench
```

### Code Quality

```bash
# Format code
just fmt

# Check formatting
just fmt-check

# Run linter
just lint

# Fix linting issues automatically
just lint-fix

# Run all checks (format + lint + test)
just check

# Fix everything automatically
just fix
```

### License Management

```bash
# Check license headers
just license-check

# Add missing license headers
just license-fix
```

### Security

```bash
# Run security audit
just security

# Check for vulnerabilities
just vuln
```

### Development

```bash
# Run in development mode (auto-reload on changes)
just dev

# Run example
just example

# Run CLI tool
just run-cli YOUR_ACCESS_KEY YOUR_SECRET_KEY
```

### Maintenance

```bash
# Clean build artifacts
just clean

# Tidy dependencies
just tidy

# Update dependencies
just update

# Show project statistics
just stats

# Show dependencies
just deps

# Check for outdated dependencies
just deps-outdated
```

### Pre-commit Checks

Before committing, run:

```bash
just pre-commit
```

This will:
1. Format code
2. Run linter
3. Check license headers
4. Run tests

### CI Simulation

To simulate CI pipeline locally:

```bash
just ci
```

## Project Structure

```
1money-go-sdk/
├── internal/              # Internal packages (not exposed)
│   ├── auth/             # Authentication & signing
│   └── transport/        # HTTP transport
├── client/               # Public client API
│   ├── client.go
│   └── service.go
├── services/             # Business service modules
│   └── echo/            # Example service
├── cmd/                  # CLI application
│   └── main.go
├── docs/                 # Documentation
├── bin/                  # Compiled binaries (gitignored)
├── .golangci.yml        # Linter configuration
├── .licenserc.yaml      # License header configuration
├── Justfile             # Task runner recipes
└── go.mod
```

## Creating a New Service

Use the template generator:

```bash
just new-service payment
```

This creates a new service skeleton at `services/payment/payment.go` with:
- Interface definition
- Implementation struct
- Constructor function

Then implement your service methods following the interface pattern.

## Code Style Guidelines

### Formatting

- Use `gofmt` for formatting (automatically applied by `just fmt`)
- Use `goimports` for import management
- Follow [Google Go Style Guide](https://google.github.io/styleguide/go/)

### Naming Conventions

- **Packages**: lowercase, single word
- **Interfaces**: noun or adjective (e.g., `Service`, `Reader`)
- **Structs**: PascalCase
- **Functions**: camelCase for unexported, PascalCase for exported
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE

### Documentation

- Every exported symbol must have a doc comment
- Doc comments start with the symbol name
- Use complete sentences

Example:
```go
// Service provides payment processing functionality.
//
// It handles payment creation, retrieval, and cancellation
// through the OneMoney API.
type Service interface {
    // Create creates a new payment transaction.
    Create(ctx context.Context, req *CreateRequest) (*Payment, error)
}
```

## Testing

### Writing Tests

- Test files: `*_test.go`
- Test functions: `func TestXxx(t *testing.T)`
- Use table-driven tests for multiple cases
- Use subtests with `t.Run()`

Example:
```go
func TestService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   *CreateRequest
        want    *Payment
        wantErr bool
    }{
        {
            name:    "valid request",
            input:   &CreateRequest{Amount: 100},
            want:    &Payment{ID: "123", Amount: 100},
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Running Specific Tests

```bash
# Run specific package
go test ./services/echo

# Run specific test
go test -run TestEcho ./services/echo

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...
```

## Linting

The project uses `golangci-lint` with configuration in `.golangci.yml`.

### Enabled Linters

- **errcheck** - Check error returns
- **govet** - Vet examines Go source code
- **staticcheck** - Static analysis
- **gosimple** - Simplify code
- **stylecheck** - Style checks
- **gofmt** - Check formatting
- **goimports** - Check imports
- **gosec** - Security issues
- And more...

### Running Linter

```bash
# Run all linters
just lint

# Run specific linter
golangci-lint run --disable-all --enable=errcheck

# Run with auto-fix
just lint-fix
```

## Git Workflow

### Before Committing

Always run pre-commit checks:

```bash
just pre-commit
```

### Commit Message Format

Follow conventional commits:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting changes
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance

Example:
```
feat(auth): add support for OAuth2 authentication

Implement OAuth2 authentication flow with token refresh.
This allows users to authenticate using OAuth2 providers.

Closes #123
```

## Debugging

### Using Delve

Install Delve:
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Debug tests:
```bash
dlv test ./services/echo
```

Debug application:
```bash
dlv debug cmd/main.go -- -access-key KEY -secret-key SECRET echo
```

### VS Code Configuration

Add to `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug CLI",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd",
            "args": ["-access-key", "KEY", "-secret-key", "SECRET", "echo"]
        }
    ]
}
```

## Performance Profiling

### CPU Profiling

```bash
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof
```

### Benchmarking

```bash
# Run benchmarks
just bench

# Compare benchmarks
go test -bench=. -count=10 ./... > old.txt
# Make changes
go test -bench=. -count=10 ./... > new.txt
benchstat old.txt new.txt
```

## Continuous Integration

The project uses the following checks in CI:

1. **Format Check**: `just fmt-check`
2. **Linting**: `just lint`
3. **License Check**: `just license-check`
4. **Tests**: `just test`
5. **Security**: `just security` (optional)
6. **Vulnerability**: `just vuln` (optional)

Simulate CI locally:
```bash
just ci
```

## Troubleshooting

### Linter Issues

If linter takes too long:
```bash
golangci-lint run --timeout 10m
```

Clear cache:
```bash
golangci-lint cache clean
```

### Module Issues

If dependencies are broken:
```bash
just clean
just tidy
go mod download
```

### Test Cache Issues

Clear test cache:
```bash
go clean -testcache
```

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Just Manual](https://just.systems/man/en/)
- [golangci-lint Docs](https://golangci-lint.run/)

## Getting Help

- Check [README.md](../README.md) for basic usage
- Read [DEVELOPER_GUIDE.md](../DEVELOPER_GUIDE.md) for extending the SDK
- Open an issue on GitHub
- Check existing issues and discussions
