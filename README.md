# 1Money Go SDK

Official Go SDK for the 1Money API.

## Installation

```bash
go get github.com/1Money-Co/1money-go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
)

func main() {
    // Create client
    client, err := onemoney.NewClient(&onemoney.Config{
        AccessKey: "your-access-key",
        SecretKey: "your-secret-key",
        BaseURL:   "http://localhost:9000",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Use services
    resp, err := client.Customer.CreateCustomer(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Client Initialization

### With Explicit Credentials

```go
client, err := onemoney.NewClient(&onemoney.Config{
    AccessKey: "your-access-key",
    SecretKey: "your-secret-key",
})
```

### With Environment Variables

```bash
export ONEMONEY_ACCESS_KEY="your-access-key"
export ONEMONEY_SECRET_KEY="your-secret-key"
export ONEMONEY_BASE_URL="https://api.1money.co"
```

```go
client, err := onemoney.NewClient(&onemoney.Config{})
```

### With Credentials File

Create `~/.onemoney/credentials`:

```ini
[default]
access_key = your-access-key
secret_key = your-secret-key
base_url = http://localhost:9000

[production]
access_key = prod-access-key
secret_key = prod-secret-key
base_url = https://api.1money.co
```

```go
// Use default profile
client, err := onemoney.NewClient(&onemoney.Config{})

// Use specific profile
client, err := onemoney.NewClient(&onemoney.Config{
    Profile: "production",
})
```

See [docs/CREDENTIALS.md](./docs/CREDENTIALS.md) for more details.

### With Custom Options

```go
client, err := onemoney.NewClient(&onemoney.Config{
    AccessKey: "your-access-key",
    SecretKey: "your-secret-key",
},
    onemoney.WithBaseURL("https://api.1money.co"),
    onemoney.WithTimeout(60*time.Second),
    onemoney.WithHTTPClient(customHTTPClient),
)
```

## Using Services

### Customer Service

```go
resp, err := client.Customer.CreateCustomer(ctx, &customer.CreateCustomerRequest{
    BusinessLegalName: "Acme Corp",
    Email:             "contact@acme.com",
    // ... other fields
})
```

### Echo Service

```go
// GET request
resp, err := client.Echo.Get(ctx)

// POST request
resp, err := client.Echo.Post(ctx, &echo.Request{
    Message: "hello",
})
```

## Documentation

```bash
# View complete API documentation
just docs
# Open http://localhost:7070
```

## Development

```bash
# Install tools
just init

# Run tests
just test

# Format & lint
just check
```

See `just --list` for all commands.

## License

Apache License 2.0
