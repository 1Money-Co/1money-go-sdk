# 1Money Go SDK

Official Go SDK for the 1Money API.

## Installation

```bash
go get github.com/1Money-Co/1money-go-sdk
```

## Usage

```go
import (
    "context"

    "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
)

client, err := onemoney.NewClient(&onemoney.Config{
    AccessKey: "your-access-key",
    SecretKey: "your-secret-key",
})

// Use services
resp, err := client.Customer.CreateCustomer(ctx, req)
```

## Configuration

Credentials are loaded in order of priority:

1. **Config fields** - `AccessKey` and `SecretKey` in `Config`
2. **Environment variables** - `ONEMONEY_ACCESS_KEY`, `ONEMONEY_SECRET_KEY`, `ONEMONEY_BASE_URL`
3. **Credentials file** - `~/.onemoney/credentials` with profile support

## License

Apache License 2.0
