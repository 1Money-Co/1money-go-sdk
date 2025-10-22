# SDK Version Information

The OneMoney Go SDK includes comprehensive version tracking at multiple levels.

## 1. SDK Package Version

The SDK version is defined in `version.go` at the root level:

```go
package onemoney

const Version = "0.1.0"
```

## 2. Client Version Method

You can retrieve the SDK version from the client instance:

```go
import "github.com/1Money-Co/1money-go-sdk/scp"

client, err := scp.NewClient(&scp.Config{
    AccessKey: "your-key",
    SecretKey: "your-secret",
})
if err != nil {
    log.Fatal(err)
}

// Get SDK version
version := client.Version()
fmt.Printf("Using SDK version: %s\n", version)
// Output: Using SDK version: 0.1.0
```

## 3. Automatic User-Agent Header

Every HTTP request automatically includes a `User-Agent` header with SDK version information:

```
User-Agent: OneMoney-Go-SDK/0.1.0 (Go/go1.25.2; darwin/arm64)
```

This helps the server:
- Track which SDK versions are being used
- Identify compatibility issues
- Provide version-specific support
- Gather usage analytics

**Format:**
```
OneMoney-Go-SDK/<version> (Go/<go-version>; <os>/<arch>)
```

**Example values:**
- `OneMoney-Go-SDK/0.1.0 (Go/go1.25.2; darwin/arm64)` - macOS on Apple Silicon
- `OneMoney-Go-SDK/0.1.0 (Go/go1.25.2; linux/amd64)` - Linux on x86_64
- `OneMoney-Go-SDK/0.1.0 (Go/go1.25.2; windows/amd64)` - Windows on x86_64

## 4. CLI Version Information

The CLI tool includes build-time version injection:

```bash
# Short version
onemoney-cli --version
# Output: 0.1.0 (commit: a1b2c3d)

# Detailed version information
onemoney-cli version
# Output:
# Version:    0.1.0
# Git Commit: a1b2c3d
# Build Time: 2025-10-22_06:35:28
# Go Version: go1.25.2

# Just version number
onemoney-cli version --short
# Output: 0.1.0
```

## Version Management

### Updating the Version

To release a new version:

1. **Update `version.go`:**
   ```go
   const Version = "0.2.0"  // Update this
   ```

2. **Build with version injection:**
   ```bash
   just build-cli
   # or
   just build-release  # for all platforms
   ```

3. **Verify:**
   ```bash
   just version
   # Output:
   # Version:    0.2.0
   # Git Commit: <current-commit>
   # Build Time: <current-time>
   ```

### Build-Time Version Injection

The CLI build process automatically injects version information:

```bash
# Manual build with version injection
go build -ldflags="-X main.version=0.1.0 \
  -X main.gitCommit=$(git rev-parse --short HEAD) \
  -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
  -o onemoney-cli ./cmd
```

Or use the Justfile:
```bash
just build-cli      # Single platform
just build-release  # All platforms
```

## Benefits

### For Developers
- **Easy debugging:** Know exactly which version users are running
- **Feature flags:** Enable/disable features based on version
- **Deprecation warnings:** Notify users of old SDK versions

### For Users
- **Transparency:** Always know which SDK version you're using
- **Support:** Easier to get help with version-specific issues
- **Updates:** Know when to upgrade

### For Server-Side
- **Analytics:** Track SDK version adoption
- **Compatibility:** Handle different SDK versions appropriately
- **Debugging:** Correlate issues with specific SDK versions

## Example: Version-Based Feature Toggle

```go
import (
    "fmt"
    "strings"

    "github.com/1Money-Co/1money-go-sdk/scp"
)

func main() {
    client, _ := scp.NewClient(&scp.Config{
        AccessKey: "key",
        SecretKey: "secret",
    })

    version := client.Version()

    // Example: Enable feature for v0.2.0+
    if strings.HasPrefix(version, "0.2.") ||
       strings.HasPrefix(version, "0.3.") {
        fmt.Println("New feature available!")
    }
}
```

## Example: Logging with Version

```go
import (
    "log"

    "github.com/1Money-Co/1money-go-sdk/scp"
)

func main() {
    client, err := scp.NewClient(&scp.Config{
        AccessKey: "key",
        SecretKey: "secret",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Log SDK version for debugging
    log.Printf("Initialized OneMoney SDK version %s", client.Version())

    // Your application code...
}
```

## Server-Side Version Detection

On the server side, you can read the `User-Agent` header:

```
User-Agent: OneMoney-Go-SDK/0.1.0 (Go/go1.25.2; darwin/arm64)
```

Parse this to:
- Detect SDK version: `0.1.0`
- Detect Go version: `go1.25.2`
- Detect platform: `darwin/arm64`

Use this information for:
- Version-specific API responses
- Deprecation warnings
- Usage analytics
- Compatibility checks
