# Quick Start Guide

Get up and running with OneMoney Go SDK in 5 minutes.

## Installation

```bash
go get github.com/1Money-Co/1money-go-sdk
```

## Basic Usage

```go
package main

import (
    "context"
    "log"
    "github.com/1Money-Co/1money-go-sdk/scp"
)

func main() {
    // Create client
    c := scp.NewClient(&scp.Config{
        AccessKey: "your-access-key",
        SecretKey: "your-secret-key",
        BaseURL:   "http://localhost:9000",
    })

    // Use service
    resp, err := c.Echo.Get(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Response: %+v", resp)
}
```

## Development Setup

### 1. Install Just (Task Runner)

```bash
# macOS
brew install just

# Linux
cargo install just

# Windows
scoop install just
```

### 2. Initialize Development Environment

```bash
just init
```

This installs all development tools:
- golangci-lint
- goimports
- license-eye
- gosec
- govulncheck

### 3. Common Commands

```bash
just            # Show all available commands
just build      # Build project
just test       # Run tests
just lint       # Run linter
just fmt        # Format code
just check      # Run all checks (format + lint + test)
```

## CLI Tool

### Build CLI

```bash
just build-cli
```

### Run CLI

```bash
./bin/onemoney-cli -access-key KEY -secret-key SECRET echo
```

## Creating a Custom Service

### 1. Generate Service Template

```bash
just new-service payment
```

### 2. Define Interface

Edit `scp/services/payment/payment.go`:

```go
package payment

import "context"

// Service defines payment operations
type Service interface {
    Create(ctx context.Context, amount float64) (*Payment, error)
    Get(ctx context.Context, id string) (*Payment, error)
}
```

### 3. Implement Service

```go
import "github.com/1Money-Co/1money-go-sdk/scp"

type serviceImpl struct {
    scp.BaseService
}

func NewService() Service {
    return &serviceImpl{}
}

func (s *serviceImpl) Create(ctx context.Context, amount float64) (*Payment, error) {
    // Implementation using s.Post(), s.Get(), etc.
}
```

### 4. Register Service

```go
paymentSvc := payment.NewService()
if svc, ok := paymentSvc.(scp.Service); ok {
    c.RegisterService("payment", svc)
}
```

## Testing

```bash
# Run all tests
just test

# Run tests with coverage
just test-coverage

# Run specific test
go test -v ./services/echo -run TestEcho
```

## Code Quality

```bash
# Format code
just fmt

# Run linter
just lint

# Fix issues automatically
just fix

# Run all checks
just check
```

## Pre-commit Checklist

Before committing code:

```bash
just pre-commit
```

This runs:
1. ✓ Format check
2. ✓ Lint
3. ✓ License headers
4. ✓ Tests

## Project Structure

```
1money-go-sdk/
├── scp/               # Main public SDK package
│   ├── client.go      # Client with pre-registered services
│   ├── service.go     # Service interface and BaseService
│   └── services/      # Business services
│       └── echo/      # Example service
├── internal/          # Internal packages (not exposed)
│   ├── auth/          # Authentication
│   └── transport/     # HTTP transport
├── cmd/               # CLI tool
├── docs/              # Documentation
├── .golangci.yml      # Linter config
├── .licenserc.yaml    # License config
├── Justfile           # Task runner
└── README.md
```

## Configuration Files

- **`.golangci.yml`** - Linter configuration
- **`.licenserc.yaml`** - License header configuration
- **`.editorconfig`** - Editor configuration
- **`Justfile`** - Task definitions

## Quick Reference

### Just Commands

| Command | Description |
|---------|-------------|
| `just build` | Build project |
| `just test` | Run tests |
| `just lint` | Run linter |
| `just fmt` | Format code |
| `just check` | Run all checks |
| `just fix` | Auto-fix issues |
| `just clean` | Clean artifacts |
| `just pre-commit` | Pre-commit checks |
| `just ci` | Simulate CI |

### Go Commands

| Command | Description |
|---------|-------------|
| `go test ./...` | Run all tests |
| `go test -race ./...` | Run with race detector |
| `go test -cover ./...` | Run with coverage |
| `go build ./...` | Build all packages |
| `go mod tidy` | Tidy dependencies |

## Documentation

- **[README.md](./README.md)** - Overview and features
- **[DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md)** - Extending the SDK
- **[docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)** - Development workflow
- **[docs/LICENSE_MANAGEMENT.md](./docs/LICENSE_MANAGEMENT.md)** - License headers

## Getting Help

- Check existing documentation
- Run `just` to see all available commands
- Open an issue on GitHub
- Read the code in `services/echo/` for examples

## Next Steps

1. ✓ Set up development environment
2. ✓ Run tests to verify setup
3. ✓ Build and run CLI tool
4. Create your first custom service
5. Write tests for your service
6. Run pre-commit checks
7. Commit your changes

Happy coding! 🚀
