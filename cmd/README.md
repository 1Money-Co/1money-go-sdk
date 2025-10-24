# 1Money CLI Tool

Command-line interface for testing and interacting with the 1Money API.

## Installation

### Using Just

```bash
just build-cli
# Binary created at: bin/onemoney-cli
```

### Using Go

```bash
# Build locally
cd cmd && go build -o onemoney-cli

# Or install globally
go install github.com/1Money-Co/1money-go-sdk/cmd@latest
```

## Quick Start

```bash
# Set credentials via environment variables
export ONEMONEY_ACCESS_KEY="your-access-key"
export ONEMONEY_SECRET_KEY="your-secret-key"

# Run echo test
./onemoney-cli echo

# Or pass credentials as flags
./onemoney-cli -k KEY -s SECRET echo
```

## Commands

### Echo Test

```bash
# GET request
./onemoney-cli echo

# POST request with message
./onemoney-cli echo post -m "Hello World"
```

### Custom Requests

```bash
# GET request
./onemoney-cli request --path /openapi/users

# POST request with JSON data
./onemoney-cli request \
  --method POST \
  --path /openapi/users \
  --data '{"name":"John"}' \
  --pretty
```

## Global Flags

| Flag | Short | Description | Default | Env Var |
|------|-------|-------------|---------|---------|
| `--access-key` | `-k` | API access key | *required* | `ONEMONEY_ACCESS_KEY` |
| `--secret-key` | `-s` | API secret key | *required* | `ONEMONEY_SECRET_KEY` |
| `--base-url` | `-u` | API base URL | `http://localhost:9000` | `ONEMONEY_BASE_URL` |
| `--timeout` | `-t` | Request timeout | `30s` | - |
| `--pretty` | `-p` | Pretty print JSON | `false` | - |
| `--help` | `-h` | Show help | - | - |
| `--version` | `-v` | Show version | - | - |

## Examples

### Pretty Print Response

```bash
./onemoney-cli -k KEY -s SECRET --pretty echo
```

### Custom Timeout

```bash
./onemoney-cli --timeout 60s request --path /openapi/slow-endpoint
```

### Using Environment Variables

```bash
export ONEMONEY_ACCESS_KEY="your-key"
export ONEMONEY_SECRET_KEY="your-secret"
export ONEMONEY_BASE_URL="https://api.1money.co"

./onemoney-cli echo
```

## Help

```bash
# Global help
./onemoney-cli --help

# Command help
./onemoney-cli echo --help
./onemoney-cli request --help
```

## Version

```bash
./onemoney-cli --version
```

See the [main README](../README.md) for more information about the SDK.
