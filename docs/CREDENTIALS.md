# Credentials Configuration Guide

This SDK supports multiple methods for providing credentials, similar to the AWS SDK.

## Credential Provider Chain

The SDK uses a credential provider chain to search for credentials in the following order:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Credentials file** `~/.onemoney/credentials` (lowest priority)

The first valid credentials found are used.

## Method 1: Command-Line Flags

Provide credentials directly via command-line flags:

```bash
onemoney-cli --access-key YOUR_KEY --secret-key YOUR_SECRET echo get
```

Short form:

```bash
onemoney-cli -k YOUR_KEY -s YOUR_SECRET echo get
```

## Method 2: Environment Variables

### Option A: .env File (Recommended for Development)

Create a `.env` file in your project directory:

```bash
# .env
ONEMONEY_ACCESS_KEY=your-access-key
ONEMONEY_SECRET_KEY=your-secret-key
ONEMONEY_BASE_URL=http://localhost:9000
```

The CLI will automatically load the `.env` file when it starts. Then run commands:

```bash
onemoney-cli echo get
```

**Important**:
- Add `.env` to your `.gitignore` to avoid committing credentials
- The `.env` file must be in the current working directory where you run the CLI

### Option B: Export Environment Variables

Alternatively, you can export environment variables directly in your shell:

```bash
export ONEMONEY_ACCESS_KEY=your-access-key
export ONEMONEY_SECRET_KEY=your-secret-key
export ONEMONEY_BASE_URL=http://localhost:9000  # Optional
```

Then run commands without flags:

```bash
onemoney-cli echo get
```

## Method 3: Credentials File (Recommended)

Create a credentials file at `~/.onemoney/credentials`:

```ini
[default]
access_key = your-access-key
secret_key = your-secret-key
base_url = http://localhost:9000

[production]
access_key = prod-access-key
secret_key = prod-secret-key
base_url = https://api.production.com

[staging]
access_key = staging-access-key
secret_key = staging-secret-key
base_url = https://api.staging.com
```

### Using the Default Profile

```bash
onemoney-cli echo get
```

### Using a Specific Profile

```bash
onemoney-cli --profile production echo get
```

## SDK Usage Examples

### Example 1: Explicit Credentials

```go
import "github.com/1Money-Co/1money-go-sdk/scp"

client, err := scp.NewClient(&scp.Config{
    AccessKey: "your-access-key",
    SecretKey: "your-secret-key",
    BaseURL:   "http://localhost:9000",
})
if err != nil {
    log.Fatal(err)
}

resp, err := client.Echo.Get(context.Background())
```

### Example 2: Environment Variables

**Note**: The SDK itself does NOT automatically load `.env` files. The `.env` auto-loading feature is only available in the CLI tool.

For SDK usage, you have two options:

**Option A: Manually load .env in your code**

```go
import (
    "github.com/joho/godotenv"
    "github.com/1Money-Co/1money-go-sdk/scp"
)

func main() {
    // Manually load .env file
    _ = godotenv.Load()

    // Now environment variables are loaded
    client, err := scp.NewClient(&scp.Config{})
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.Echo.Get(context.Background())
}
```

**Option B: Export environment variables directly**

```go
// Set environment variables first:
// export ONEMONEY_ACCESS_KEY=xxx
// export ONEMONEY_SECRET_KEY=yyy

import "github.com/1Money-Co/1money-go-sdk/scp"

client, err := scp.NewClient(&scp.Config{})
if err != nil {
    log.Fatal(err)
}

resp, err := client.Echo.Get(context.Background())
```

### Example 3: Credentials File with Profile

```go
import "github.com/1Money-Co/1money-go-sdk/scp"

// Uses the "production" profile from ~/.onemoney/credentials
client, err := scp.NewClient(&scp.Config{
    Profile: "production",
})
if err != nil {
    log.Fatal(err)
}

resp, err := client.Echo.Get(context.Background())
```

## Priority Examples

### Scenario 1: All sources provided

```bash
# Credentials file has default profile
# Environment variables are set
# Command-line flags provided

onemoney-cli -k CLI_KEY -s CLI_SECRET echo get
```

**Result**: Uses `CLI_KEY` and `CLI_SECRET` (command-line has highest priority)

### Scenario 2: Environment and file

```bash
export ONEMONEY_ACCESS_KEY=ENV_KEY
export ONEMONEY_SECRET_KEY=ENV_SECRET

# ~/.onemoney/credentials exists with [default] profile

onemoney-cli echo get
```

**Result**: Uses `ENV_KEY` and `ENV_SECRET` (environment has priority over file)

### Scenario 3: File only with profile

```bash
onemoney-cli --profile staging echo get
```

**Result**: Uses credentials from `[staging]` profile in `~/.onemoney/credentials`

## Troubleshooting

### Understanding Error Messages

The SDK provides detailed error messages to help you diagnose credential issues. Each error message includes:

1. **Provider name**: Which provider failed
2. **Error type**: What went wrong
3. **Specific details**: Exactly what was missing or invalid

### Example Error Messages

#### Missing Environment Variables

```
Error: EnvProvider: missing required environment variables: ONEMONEY_ACCESS_KEY, ONEMONEY_SECRET_KEY: no credentials found
```

**Solution**: Set the missing environment variables:
```bash
export ONEMONEY_ACCESS_KEY=your-key
export ONEMONEY_SECRET_KEY=your-secret
```

#### Credentials File Not Found

```
Error: FileProvider: credentials file not found: /Users/you/.onemoney/credentials: no credentials found
```

**Solution**: Create the credentials file:
```bash
mkdir -p ~/.onemoney
cat > ~/.onemoney/credentials <<EOF
[default]
access_key = your-access-key
secret_key = your-secret-key
EOF
chmod 600 ~/.onemoney/credentials
```

#### Profile Not Found

```
Error: FileProvider: profile 'production' not found in /Users/you/.onemoney/credentials: no credentials found
```

**Solution**: Add the profile to your credentials file:
```ini
[production]
access_key = prod-access-key
secret_key = prod-secret-key
```

#### Missing Keys in Profile

```
Error: FileProvider: missing required keys in profile 'default': secret_key: no credentials found
```

**Solution**: Add the missing key to the profile in your credentials file.

#### Chain Provider Error (No Credentials from Any Source)

```
Error: ChainProvider: no credentials found: attempted to load credentials from 3 provider(s):
  - StaticProvider: missing required credentials: access_key, secret_key: no credentials found
  - EnvProvider: missing required environment variables: ONEMONEY_ACCESS_KEY, ONEMONEY_SECRET_KEY: no credentials found
  - FileProvider: credentials file not found: /Users/you/.onemoney/credentials: no credentials found
```

This comprehensive error shows:
- **All providers that were tried** (Static, Env, File)
- **Why each provider failed** (missing credentials, missing env vars, file not found)
- **What you need to fix** (set credentials via any of the three methods)

**Solution**: Choose one method and provide valid credentials:

1. **Command-line**: `onemoney-cli --access-key KEY --secret-key SECRET echo get`
2. **Environment**: `export ONEMONEY_ACCESS_KEY=KEY ONEMONEY_SECRET_KEY=SECRET`
3. **File**: Create `~/.onemoney/credentials` with `[default]` profile

### Common Issues and Solutions

#### 1. No credentials found

The SDK couldn't find valid credentials from any source. The error message will show which providers were checked and why each failed. Choose one method:

- Set command-line flags: `--access-key` and `--secret-key`
- Set environment variables: `ONEMONEY_ACCESS_KEY` and `ONEMONEY_SECRET_KEY`
- Create credentials file at `~/.onemoney/credentials`

#### 2. Profile name is incorrect

If using `--profile production`, ensure the profile exists in your credentials file:
```ini
[production]
access_key = your-key
secret_key = your-secret
```

#### 3. Invalid credentials format

Ensure your credentials file is valid INI format with the correct key names:
- `access_key` (not `accessKey` or `access-key`)
- `secret_key` (not `secretKey` or `secret-key`)
- `base_url` (optional)

### Check what credentials are being used

The SDK validates credentials at client creation time. If the client is created successfully, your credentials are valid.

### Enable Debug Logging

To see which provider successfully loaded credentials, you can examine the error messages when they fail. The SDK will use the first provider that succeeds without showing which one succeeded (for security).

## Best Practices

1. **Local Development (CLI)**: Use `.env` file in your project directory
2. **Local Development (SDK)**: Use credentials file with `[default]` profile
3. **CI/CD**: Use environment variables
4. **Production**: Use credentials file with specific profile (e.g., `[production]`)
5. **Testing**: Use command-line flags for quick tests

### .env File Best Practices

- Always add `.env` to your `.gitignore`
- Use `.env.example` file to document required variables (without actual values):
  ```bash
  # .env.example
  ONEMONEY_ACCESS_KEY=your-access-key-here
  ONEMONEY_SECRET_KEY=your-secret-key-here
  ONEMONEY_BASE_URL=http://localhost:9000
  ```
- Keep `.env` file in the project root directory
- Document the `.env` file usage in your project's README

## Security

- **Never commit credentials** to version control
- Add both `.env` and `~/.onemoney/` to your `.gitignore`:
  ```gitignore
  # Credentials
  .env
  .env.local
  .onemoney/
  ```
- Use appropriate file permissions:
  ```bash
  chmod 600 ~/.onemoney/credentials
  chmod 600 .env
  ```
- Rotate credentials regularly
- Use different credentials for different environments (dev, staging, prod)
