# OneMoney Go SDK

A Go SDK for interacting with the OneMoney API, featuring automatic HMAC-SHA256 request signing, interface-based service architecture, and a clean, idiomatic interface.

## Features

- ✅ Automatic HMAC-SHA256 request signing
- ✅ **Interface-based service design** - Clear contracts, easy mocking
- ✅ **Extensible architecture** - Easy to add custom business modules
- ✅ Clean and idiomatic Go API
- ✅ Support for GET, POST, PUT, DELETE, PATCH methods
- ✅ Flexible configuration with functional options
- ✅ Context-aware requests
- ✅ Type-safe response parsing
- ✅ Comprehensive error handling
- ✅ Well-organized package structure

## Installation

```bash
go get github.com/1Money-Co/1money-go-sdk
```

## Quick Start

**👉 New to the SDK? Check out [QUICKSTART.md](./QUICKSTART.md) for a 5-minute guide!**

```go
package main

import (
    "context"
    "log"

    "github.com/1Money-Co/1money-go-sdk/scp"
    "github.com/1Money-Co/1money-go-sdk/scp/services/echo"
)

func main() {
    // Create client with pre-registered services
    c := scp.NewClient(&scp.Config{
        AccessKey: "your-access-key",
        SecretKey: "your-secret-key",
        BaseURL:   "http://localhost:9000",
    })

    ctx := context.Background()

    // Use service with type-safe interface
    resp, err := c.Echo.Get(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // POST request
    resp, err = c.Echo.Post(ctx, &echo.Request{
        Message: "hello",
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Architecture

### Interface-Based Services

Each service is defined as an interface, making capabilities immediately visible:

```go
// Service interface clearly shows all available operations
type Service interface {
    Get(ctx context.Context) (*Response, error)
    Post(ctx context.Context, req *Request) (*Response, error)
}

// Usage - IDE shows all available methods
resp, err := c.Echo.Get(ctx)
resp, err := c.Echo.Post(ctx, req)
```

### Package Structure

```
1money-go-sdk/
├── scp/                 # Main public SDK package
│   ├── client.go        # Client with pre-registered services
│   ├── service.go       # Service interface and BaseService
│   └── services/        # Business service modules
│       └── echo/        # Example echo service
│           └── echo.go  # Interface + implementation
├── internal/            # Internal packages (not exposed to users)
│   ├── auth/            # Authentication and signature generation
│   │   └── signer.go
│   └── transport/       # HTTP transport layer
│       └── transport.go
├── cmd/                 # CLI tool
│   └── main.go          # Command line interface for testing
├── example_scp.go       # Usage example
├── DEVELOPER_GUIDE.md   # Guide for extending the SDK
└── README.md
```

## Configuration Options

### WithBaseURL

Set the API base URL (default: "http://localhost:9000"):

```go
c := scp.NewClient(&scp.Config{
    AccessKey: "your-access-key",
    SecretKey: "your-secret-key",
}, scp.WithBaseURL("https://api.example.com"))
```

### WithTimeout

Set request timeout (default: 30 seconds):

```go
c := scp.NewClient(&scp.Config{
    AccessKey: "your-access-key",
    SecretKey: "your-secret-key",
}, scp.WithTimeout(60*time.Second))
```

### WithHTTPClient

Provide a custom HTTP client:

```go
customClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:    100,
        IdleConnTimeout: 90 * time.Second,
    },
}

c := scp.NewClient(&scp.Config{
    AccessKey: "your-access-key",
    SecretKey: "your-secret-key",
}, scp.WithHTTPClient(customClient))
```

## Creating Custom Services

### Step 1: Define Interface

```go
package payment

import "context"

// Service defines all supported operations
type Service interface {
    Create(ctx context.Context, req *CreateRequest) (*Payment, error)
    Get(ctx context.Context, id string) (*Payment, error)
    Cancel(ctx context.Context, id string) error
}
```

### Step 2: Implement Service

```go
import "github.com/1Money-Co/1money-go-sdk/scp"

// Private implementation
type serviceImpl struct {
    scp.BaseService
}

// Constructor returns interface
func NewService() Service {
    return &serviceImpl{}
}

func (s *serviceImpl) Create(ctx context.Context, req *CreateRequest) (*Payment, error) {
    body, _ := json.Marshal(req)
    resp, err := s.Post(ctx, "/openapi/payments", body)
    // Handle response...
    return &payment, nil
}
```

### Step 3: Register and Use

```go
// Create and register
paymentSvc := payment.NewService()
if svc, ok := paymentSvc.(scp.Service); ok {
    c.RegisterService("payment", svc)
}

// Use service
payment, err := paymentSvc.Create(ctx, req)
```

See [DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md) for complete examples and the [services/echo/echo.go](./scp/services/echo/echo.go) reference implementation.

## Error Handling

All errors are wrapped with context for better debugging:

```go
resp, err := c.Echo.Get(ctx)
if err != nil {
    // Error includes context about what failed
    log.Printf("Request failed: %v", err)
    return
}
```

## Testing

Services are interface-based, making them easy to mock:

```go
type mockEchoService struct{}

func (m *mockEchoService) Get(ctx context.Context) (*echo.Response, error) {
    return &echo.Response{Message: "mocked"}, nil
}

func (m *mockEchoService) Post(ctx context.Context, req *echo.Request) (*echo.Response, error) {
    return &echo.Response{Message: req.Message}, nil
}

// Use in tests
mockSvc := &mockEchoService{}
result, err := myFunction(mockSvc)
```

## Built-in Services

### Echo Service

Simple echo service for testing:

```go
// GET request
resp, err := c.Echo.Get(ctx)

// POST request
resp, err := c.Echo.Post(ctx, &echo.Request{
    Message: "hello",
})
```

## CLI Tool

A command-line interface tool is provided for testing and development.

### Installation

```bash
go install github.com/1Money-Co/1money-go-sdk/cmd@latest
```

Or build locally:

```bash
cd cmd
go build -o onemoney-cli
```

### Usage

```bash
# Set credentials via environment variables (recommended)
export ONEMONEY_ACCESS_KEY="your-key"
export ONEMONEY_SECRET_KEY="your-secret"

# Echo GET request
./onemoney-cli echo

# Echo POST request with custom message
./onemoney-cli echo post -m "Hello World"

# Custom GET request
./onemoney-cli request --path /openapi/users

# Custom POST request with JSON data
./onemoney-cli request \
  --method POST \
  --path /openapi/users \
  --data '{"name":"John"}' \
  --pretty

# Using short flags
./onemoney-cli -k KEY -s SECRET -p echo
```

### CLI Commands

- **`echo`** - Test echo service
  - `echo get` - Send GET request
  - `echo post` - Send POST request with message
- **`request`** - Make custom HTTP requests (aliases: `req`, `r`)
  - Supports GET, POST, PUT, DELETE methods

### Global Flags

| Flag | Short | Env Var | Description |
|------|-------|---------|-------------|
| `--access-key` | `-k` | `ONEMONEY_ACCESS_KEY` | API access key (required) |
| `--secret-key` | `-s` | `ONEMONEY_SECRET_KEY` | API secret key (required) |
| `--base-url` | `-u` | `ONEMONEY_BASE_URL` | API base URL |
| `--timeout` | `-t` | - | Request timeout (default: 30s) |
| `--pretty` | `-p` | - | Pretty print JSON output |

See [cmd/README.md](./cmd/README.md) for detailed CLI documentation.

## Development

### Quick Start

The project uses [Just](https://just.systems/) as a task runner (similar to Make but more modern).

```bash
# Install Just
brew install just  # macOS
# or visit https://just.systems/ for other platforms

# Initialize development environment
just init

# Show all available commands
just

# Run checks (format, lint, test)
just check

# Build CLI tool
just build-cli
```

### Common Tasks

```bash
just build          # Build the project
just test           # Run tests
just lint           # Run linter
just fmt            # Format code
just check          # Run all checks
just pre-commit     # Pre-commit checks
```

See [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) for detailed development guide.

## Documentation

- **[README.md](./README.md)** - This file, quick start and basic usage
- **[DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md)** - Complete guide for extending the SDK with custom services
- **[docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)** - Development setup and workflow
- **[scp/services/echo/echo.go](./scp/services/echo/echo.go)** - Complete example of interface-based service

## Why Interface-Based Design?

1. **Clear Contracts** - Interface explicitly shows all available operations
2. **Easy Testing** - Simple to create mocks for unit tests
3. **IDE Support** - Autocomplete shows all methods
4. **Type Safety** - Compile-time checking
5. **Encapsulation** - Implementation details hidden

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

When adding new services, please follow the interface-based design pattern shown in `services/echo/echo.go`.
