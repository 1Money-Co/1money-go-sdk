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

### Using a `.env` File

For playing around with the examples, you can use a `.env` file to manage environment variables:

```bash
# Copy the example file
cp .env.example .env

# Edit with your credentials
vim .env  # or your preferred editor
```

The SDK examples use [godotenv](https://github.com/joho/godotenv) to automatically load variables from `.env`. Your `.env` file should look like:

```bash
ONEMONEY_ACCESS_KEY=your-access-key
ONEMONEY_SECRET_KEY=your-secret-key
ONEMONEY_SANDBOX=1
```

> **Note:** Never commit your `.env` file to version control. It's already in `.gitignore`.

See [.env.example](.env.example) for all available configuration options.

## Examples

The SDK includes runnable examples in the [`examples/`](examples/) directory demonstrating common workflows:

| Example                                                                        | Description                                        |
| ------------------------------------------------------------------------------ | -------------------------------------------------- |
| [`create_customer`](examples/create_customer/)                                 | Create a new business customer                     |
| [`fiat_to_usdc_withdrawal`](examples/fiat_to_usdc_withdrawal/)                 | USD deposit → convert to USDC → withdraw to wallet |
| [`usdc_to_fiat_withdrawal`](examples/usdc_to_fiat_withdrawal/)                 | USDC deposit → convert to USD → withdraw to bank   |
| [`auto_conversion_with_simulation`](examples/auto_conversion_with_simulation/) | Auto conversion rules with simulated deposits      |

### Running Examples

```bash
# Set up your environment first
cp .env.example .env
# Edit .env with your credentials

# Run an example
go run ./examples/create_customer
go run ./examples/fiat_to_usdc_withdrawal
```

Most examples require `ONEMONEY_CUSTOMER_ID` to be set. Run `create_customer` first to obtain one.

## License

Apache License 2.0
