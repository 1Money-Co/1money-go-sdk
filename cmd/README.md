# OneMoney CLI Tool

A command-line interface tool for testing and interacting with the OneMoney API. Built with [urfave/cli](https://github.com/urfave/cli).

## Installation

### Build from source

```bash
cd cmd
go build -o onemoney-cli
```

Or use Just:

```bash
just build-cli
```

### Install globally

```bash
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

## Global Flags

These flags are available for all commands:

| Flag | Short | Description | Default | Env Var |
|------|-------|-------------|---------|---------|
| `--access-key` | `-k` | API access key | *required* | `ONEMONEY_ACCESS_KEY` |
| `--secret-key` | `-s` | API secret key | *required* | `ONEMONEY_SECRET_KEY` |
| `--base-url` | `-u` | API base URL | `http://localhost:9000` | `ONEMONEY_BASE_URL` |
| `--timeout` | `-t` | Request timeout | `30s` | - |
| `--pretty` | `-p` | Pretty print JSON | `false` | - |
| `--help` | `-h` | Show help | - | - |
| `--version` | `-v` | Show version | - | - |

## Commands

### `echo` - Test Echo Service

Test the echo service with GET or POST requests.

#### Subcommands

**`echo get`** - Send a GET echo request

```bash
./onemoney-cli echo get
```

**`echo post`** - Send a POST echo request

```bash
# Default message
./onemoney-cli echo post

# Custom message via flag
./onemoney-cli echo post --message "Hello World"

# Custom message via argument
./onemoney-cli echo post "Hello World"
```

Flags:
- `--message`, `-m` - Message to send (default: "Hello from CLI")

#### Default Action

Running `echo` without a subcommand defaults to `get`:

```bash
./onemoney-cli echo
# Same as: ./onemoney-cli echo get
```

### `request` - Make Custom HTTP Requests

Make custom HTTP requests to any API endpoint.

```bash
# GET request
./onemoney-cli request --path /openapi/users

# POST request
./onemoney-cli request \
  --method POST \
  --path /openapi/users \
  --data '{"name":"John","email":"john@example.com"}'

# PUT request
./onemoney-cli request \
  --method PUT \
  --path /openapi/users/123 \
  --data '{"name":"John Smith"}'

# DELETE request
./onemoney-cli request --method DELETE --path /openapi/users/123
```

Aliases: `req`, `r`

Flags:
- `--method`, `-X` - HTTP method: GET, POST, PUT, DELETE (default: GET)
- `--path` - API endpoint path (default: /openapi/echo)
- `--data`, `-d` - Request body (JSON string)

## Usage Examples

### Basic Echo Test

```bash
./onemoney-cli -k "ZTDAZGUWIVBDU1UNX0NZ" -s "nNkADJdyGzRuGO8QDSmyqRnpz70wsTBdekVpVxKQSvI" echo
```

### Echo with Custom Message

```bash
./onemoney-cli \
  -k "ZTDAZGUWIVBDU1UNX0NZ" \
  -s "nNkADJdyGzRuGO8QDSmyqRnpz70wsTBdekVpVxKQSvI" \
  echo post -m "Hello from CLI"
```

### Pretty Print JSON

```bash
./onemoney-cli -k KEY -s SECRET --pretty echo
```

Output:
```json
{
  "message": "hello",
  "timestamp": "2025-01-22T10:30:00Z"
}
```

### Custom API Call

```bash
./onemoney-cli \
  -k KEY -s SECRET \
  request --path /openapi/balance
```

### POST with JSON Data

```bash
./onemoney-cli \
  -k KEY -s SECRET \
  request \
  --method POST \
  --path /openapi/users \
  --data '{"name":"Alice","email":"alice@example.com","age":30}' \
  --pretty
```

### Using Environment Variables

```bash
# Set credentials
export ONEMONEY_ACCESS_KEY="your-access-key"
export ONEMONEY_SECRET_KEY="your-secret-key"
export ONEMONEY_BASE_URL="https://api.onemoney.com"

# Run commands without flags
./onemoney-cli echo
./onemoney-cli request --path /openapi/users
```

### Using Aliases

```bash
# Short flags
./onemoney-cli -k KEY -s SECRET -p echo

# Command aliases
./onemoney-cli req --path /openapi/users  # Same as 'request'
./onemoney-cli r --path /openapi/users    # Even shorter
./onemoney-cli e                           # Same as 'echo'
```

### Different Base URL

```bash
./onemoney-cli \
  -k KEY -s SECRET \
  -u "https://api.onemoney.com" \
  echo
```

### Custom Timeout

```bash
./onemoney-cli \
  -k KEY -s SECRET \
  --timeout 60s \
  request --path /openapi/slow-endpoint
```

## Output Format

The CLI outputs responses in JSON format:

```json
{"message":"hello","timestamp":"2025-01-22T10:30:00Z"}
```

Use `--pretty` flag for human-readable formatting:

```json
{
  "message": "hello",
  "timestamp": "2025-01-22T10:30:00Z"
}
```

## Error Handling

Errors are printed to stderr with descriptive messages:

```bash
$ ./onemoney-cli echo
Error: access-key is required
```

Exit codes:
- `0` - Success
- `1` - Error occurred

## Help

Show help for any command:

```bash
# Global help
./onemoney-cli --help
./onemoney-cli -h

# Command help
./onemoney-cli echo --help
./onemoney-cli request --help

# Subcommand help
./onemoney-cli echo post --help
```

## Version

Show CLI version:

```bash
./onemoney-cli --version
```

## Advanced Usage

### Shell Completion (Bash)

Generate bash completion:

```bash
./onemoney-cli --generate-bash-completion
```

Add to your `.bashrc`:

```bash
eval "$(./onemoney-cli --generate-bash-completion)"
```

### Configuration File (Future)

Future enhancement: Support for config file:

```yaml
# ~/.onemoney-cli.yml
access_key: your-access-key
secret_key: your-secret-key
base_url: https://api.onemoney.com
timeout: 30s
```

### Multiple Profiles (Future)

Future enhancement: Support for multiple profiles:

```bash
./onemoney-cli --profile prod echo
./onemoney-cli --profile dev echo
```

## Development

### Building

```bash
# Development build
go build -o onemoney-cli

# Production build with optimizations
go build -ldflags="-s -w" -o onemoney-cli

# Cross-platform builds
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o onemoney-cli-linux-amd64
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o onemoney-cli-darwin-amd64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o onemoney-cli-windows-amd64.exe
```

Or use Just:

```bash
just build-cli          # Development build
just build-release      # All platforms
```

### Testing

```bash
# Run tests
go test -v

# Test with real API (requires credentials)
export ONEMONEY_ACCESS_KEY="test-key"
export ONEMONEY_SECRET_KEY="test-secret"
go run main.go echo
```

### Adding New Commands

1. Define the command in `main.go`:

```go
{
    Name:    "newcommand",
    Aliases: []string{"nc"},
    Usage:   "Description of new command",
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:  "option",
            Usage: "Option description",
        },
    },
    Action: func(c *cli.Context) error {
        return handleNewCommand(c)
    },
}
```

2. Implement the handler:

```go
func handleNewCommand(c *cli.Context) error {
    client := createClient()
    // Implementation
    return nil
}
```

## Troubleshooting

### Issue: "access-key is required"

**Solution**: Provide credentials via flags or environment variables:

```bash
# Via flags
./onemoney-cli -k KEY -s SECRET echo

# Via environment
export ONEMONEY_ACCESS_KEY="KEY"
export ONEMONEY_SECRET_KEY="SECRET"
./onemoney-cli echo
```

### Issue: Connection refused

**Solution**: Check if the API server is running and the base URL is correct:

```bash
./onemoney-cli -u "http://localhost:9000" echo
```

### Issue: Request timeout

**Solution**: Increase timeout:

```bash
./onemoney-cli --timeout 60s echo
```

### Issue: Invalid JSON in response

**Solution**: Use `--pretty` to see formatted output and check for errors:

```bash
./onemoney-cli --pretty echo
```

## Resources

- [urfave/cli Documentation](https://cli.urfave.org/)
- [OneMoney API Documentation](https://api.onemoney.com/docs)
- [Source Code](https://github.com/1Money-Co/1money-go-sdk)

## Contributing

To add new features or fix bugs:

1. Fork the repository
2. Create your feature branch
3. Add tests if applicable
4. Submit a pull request

See [DEVELOPER_GUIDE.md](../DEVELOPER_GUIDE.md) for more details.
