# Go Project Development Guide (English Version)

**IMPORTANT: When writing code following this guide:**
- **ALL code, comments, documentation, variable names, function names, and any code-related text MUST be written in FULL ENGLISH**
- **ALL responses to users MUST be in Chinese (中文)**
- **Never mix Chinese in code, comments, or documentation**
- **CRITICAL: After making any code changes, you MUST run `just check` to verify code quality, formatting, linting, and tests**

This document is based on [Google Go Style Guide](https://google.github.io/styleguide/go/) to guide AI in writing Go code that meets Google standards.

## Core Principles (Priority Order)

1. **Clarity** - Code's purpose and rationale must be understandable to readers
2. **Simplicity** - Accomplish goals using the most straightforward approach
3. **Concision** - Maintain high signal-to-noise ratio in code
4. **Maintainability** - Enable easy future modifications
5. **Consistency** - Align with broader codebase patterns

## Mandatory Formatting Guidelines

### Tool Compliance
- All source files must conform to `gofmt` output format
- Use `go fmt` to automatically format code

### Naming Conventions

**Basic Rule: Use MixedCaps or mixedCaps, NEVER underscores**

```go
// Exported (public) - capitalize first letter
type UserAccount struct{}
const MaxLength = 100
func ParseRequest() {}

// Unexported (private) - lowercase first letter
type internalCache struct{}
const maxRetries = 3
func validateInput() {}
```

**Exceptions (underscores allowed):**
1. Packages imported only by generated code
2. Test function names in `*_test.go` files
3. Low-level OS/cgo libraries

### Package Naming

```go
// ✅ Recommended - concise, lowercase, unbroken
package tabwriter
package httputil

// ❌ Avoid - using underscores
package tab_writer
package http_util

// ❌ Avoid - overly generic names
package util
package common
package helpers
```

### Constant Naming

```go
// ✅ Name based on purpose, use camel case
const MaxPacketSize = 512
const DefaultTimeout = 30

// ❌ Avoid all caps with underscores
const MAX_PACKET_SIZE = 512
```

### Initialism Handling

Keep consistent casing within initialisms:

```go
// ✅ Correct
type XMLAPI struct{}      // Exported
type xmlAPI struct{}      // Unexported

type UserID int           // Exported
type userID int           // Unexported

// iOS special handling
type IOSApp struct{}      // Exported
type iOSApp struct{}      // Unexported

// ❌ Wrong - inconsistent casing
type XmlApi struct{}
type UserId int
```

### Receiver Naming

```go
// ✅ Short abbreviation, 1-2 letters, keep consistent
func (c *Client) Connect() {}
func (c *Client) Disconnect() {}

func (u *UserAccount) Validate() {}
func (u *UserAccount) Save() {}

// ❌ Avoid using full type name
func (client *Client) Connect() {}

// ❌ Avoid inconsistency
func (c *Client) Connect() {}
func (cl *Client) Disconnect() {}
```

### Getter Naming

```go
// ✅ Omit Get prefix
func (c *Client) Counts() int {}
func (u *User) Name() string {}

// Use Compute or Fetch for expensive operations
func (s *Stats) ComputeTotal() int {}
func (d *Database) FetchUsers() []User {}

// ❌ Avoid Get prefix
func (c *Client) GetCounts() int {}
```

### Variable Naming

**Scope principle:**
- Name length should be proportional to scope size
- Name length should be inversely proportional to usage frequency

```go
// ✅ Short scope uses short names
for i := 0; i < 10; i++ {}
if err := doSomething(); err != nil {}

// ✅ Long scope uses descriptive names
var authenticatedUserSessions map[string]*Session

// ✅ Don't include type in name
users := []User{}           // not userSlice
counts := map[string]int{}  // not countMap

// ❌ Avoid package name/export name repetition
// In user package
type UserManager struct{}  // ❌ Usage: user.UserManager
type Manager struct{}      // ✅ Usage: user.Manager
```

## Import Management

### Import Grouping

Organize into four groups in this order, separated by blank lines:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "os"

    // 2. Other packages
    "github.com/pkg/errors"
    "go.uber.org/zap"

    // 3. Protocol Buffers
    pb "myproject/gen/proto/go/myproject/v1"

    // 4. Side-effect imports (import for init only)
    _ "embed"
)
```

### Import Rules

```go
// ❌ Never use dot imports (except special test scenarios)
import . "fmt"

// ✅ Blank imports only for main package or tests
import _ "net/http/pprof"

// ✅ Import renaming - resolve conflicts or follow conventions
import (
    neturl "net/url"
    pb "myproject/gen/proto/go/myproject/v1"
)
```

## Error Handling

### Return Pattern

```go
// ✅ Error should be the last return value
func Open(name string) (*File, error) {}
func Parse(data []byte) (Result, error) {}

// ✅ Multiple returns - avoid in-band errors
func Lookup(key string) (value string, ok bool) {}

// ❌ Avoid using -1, nil, or empty string to signal errors
func Find(key string) string {
    // Returns "" for not found - bad
}
```

### Error Strings

```go
// ✅ Lowercase start, no punctuation (except proper nouns)
fmt.Errorf("something bad happened")
fmt.Errorf("failed to connect to database")
errors.New("invalid input")

// ❌ Avoid capitalization or punctuation
fmt.Errorf("Something bad happened.")
```

### Error Handling Strategy

```go
// ✅ Handle immediately, return early
func process() error {
    data, err := fetch()
    if err != nil {
        return fmt.Errorf("failed to fetch: %w", err)
    }
    // Continue normal flow
    return save(data)
}

// ❌ Never ignore errors
data, _ := fetch()  // Bad!

// ❌ Avoid nesting normal code
func process() error {
    data, err := fetch()
    if err == nil {
        // Normal code nested here - bad
        return save(data)
    }
    return err
}
```

### Error Wrapping

```go
// ✅ Use %w to allow callers to inspect error chain
if err := doSomething(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Callers can use
if errors.Is(err, ErrNotFound) {}
if errors.As(err, &specificErr) {}

// ✅ Use %v at system boundaries (RPC, storage)
// Transform domain-specific errors to canonical error spaces
return status.Errorf(codes.NotFound, "user not found: %v", err)

// ✅ When wrapping errors, place %w at end of string
return fmt.Errorf("failed to process request: %w", err)
```

### Structured Errors

```go
// ✅ Create errors that enable programmatic checking
var ErrNotFound = errors.New("not found")

type ValidationError struct {
    Field string
    Value interface{}
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("invalid %s: %v", e.Field, e.Value)
}

// Usage
if errors.Is(err, ErrNotFound) {
    // Handle not found case
}

var valErr *ValidationError
if errors.As(err, &valErr) {
    // Handle validation error
}
```

### Logging Strategy

```go
// ✅ Avoid both logging and returning errors
func process() error {
    if err := validate(); err != nil {
        // ❌ Don't do this
        log.Error("validation failed", err)
        return err  // Caller might also log
    }
    return nil
}

// ✅ Let caller decide whether to log
func process() error {
    if err := validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}

// Caller decides
if err := process(); err != nil {
    log.Error("processing error", err)
}
```

## Function Design

### Function Signature Formatting

```go
// ✅ Keep single line
func (c *Client) Connect(ctx context.Context, addr string, timeout time.Duration) error {}

// ✅ When too many parameters, extract to option struct
type ConnectOptions struct {
    Address string
    Timeout time.Duration
    Retry   int
}

func (c *Client) Connect(ctx context.Context, opts ConnectOptions) error {}
```

### Receiver Selection (Pointer vs Value)

**Use pointer receiver when:**
- Method needs to modify receiver
- Receiver contains fields that can't be safely copied (like `sync.Mutex`)
- Receiver is a large struct
- Need to support concurrent access
- Contains pointers to mutable data

**Maintain consistency:**
```go
// ✅ All methods for a type use same receiver style
type Client struct {
    conn net.Conn
}

func (c *Client) Connect() error {}     // All pointers
func (c *Client) Disconnect() error {}  // All pointers
func (c *Client) IsConnected() bool {}  // All pointers

// ❌ Avoid mixing
func (c *Client) Connect() error {}     // Pointer
func (c Client) Disconnect() error {}   // Value - inconsistent!
```

### Pass Value vs Pointer

```go
// ✅ Small structs pass by value
type Point struct {
    X, Y int
}
func Distance(p1, p2 Point) float64 {}

// ✅ Large structs or Protocol Buffers pass by pointer
type Config struct {
    // Many fields...
}
func Apply(cfg *Config) error {}

// ✅ Protocol Buffers always use pointers
func Process(req *pb.Request) (*pb.Response, error) {}
```

### Named Return Values

```go
// ✅ Use to clarify caller responsibility
func Split(path string) (dir, file string) {
    // Implementation
    return
}

// ✅ Use for deferred closures
func process() (err error) {
    f, err := os.Open("file")
    if err != nil {
        return err
    }
    defer func() {
        if closeErr := f.Close(); closeErr != nil && err == nil {
            err = closeErr
        }
    }()
    // Process file
    return nil
}

// ❌ Avoid creating repetition
func (n *Node) Parent() (node *Node) {}  // Redundant
func (n *Node) Parent() *Node {}         // Better

// ✅ Naked returns only in small functions
func add(a, b int) (result int) {
    result = a + b
    return  // OK, function is small
}
```

## Control Flow

### Conditional Statements

```go
// ✅ Keep single line or extract conditions
if user.IsActive && user.HasPermission("write") && user.Age >= 18 {
    // Handle
}

// ✅ Extract complex conditions as local variables
canWrite := user.IsActive &&
            user.HasPermission("write") &&
            user.Age >= 18
if canWrite {
    // Handle
}

// ✅ Variable on left side
if result == "foo" {  // ✅
if "foo" == result {  // ❌
```

### Loops

```go
// ✅ Keep single line
for i := 0; i < len(items); i++ {}
for key, value := range m {}

// ✅ Or extract conditions to loop body
for {
    item, err := next()
    if err != nil {
        break
    }
    process(item)
}
```

### Switch Statements

```go
// ✅ Keep cases on single line
switch status {
case "active":
    return true
case "inactive":
    return false
default:
    return false
}

// ✅ No need for break (Go auto-terminates)
switch x {
case 1:
    fmt.Println("one")
    // No break needed
case 2:
    fmt.Println("two")
}
```

## Types and Interfaces

### Interface Design

```go
// ✅ Define interfaces in consuming package, not implementing package
// storage package
type Repository interface {
    Save(item Item) error
    Find(id string) (Item, error)
}

// ✅ Return concrete types, not interfaces
func NewClient() *Client {}        // ✅
func NewClient() Interface {}      // ❌

// ✅ Create interfaces only when actually needed
// At least two implementations, or clear mocking requirement
```

### Struct Literals

```go
// ✅ Use field names for types from other packages
user := pkg.User{
    Name: "Alice",
    Age:  30,
}

// ✅ Omit zero-value fields (when clarity isn't lost)
config := Config{
    Host: "localhost",
    // Port: 0,  // Can omit
}

// ✅ Match brace indentation across multiple lines
config := Config{
    Database: DatabaseConfig{
        Host: "localhost",
        Port: 5432,
    },
    Cache: CacheConfig{
        TTL: time.Hour,
    },
}
```

### Nil Slices

```go
// ✅ Prefer nil initialization for local variables
var items []Item  // nil slice

// ✅ Don't distinguish between nil and empty slice in APIs
func process(items []Item) {
    if len(items) == 0 {  // Works for both nil and empty slice
        return
    }
}

// ❌ Not necessary
if items != nil && len(items) > 0 {}  // Redundant
if len(items) > 0 {}                  // Sufficient
```

### Type Aliases

```go
// ✅ Use type definitions to create new types
type UserID int64

// ❌ Avoid type aliases (except for package migration)
type UserID = int64  // Only for package migration
```

## Comments and Documentation

### Doc Comments

```go
// ✅ All exported names require doc comments
// Start with the name being described

// User represents a user account in the system.
type User struct {
    Name string
    Age  int
}

// NewUser creates a new user with the given name.
func NewUser(name string) *User {
    return &User{Name: name}
}

// ✅ Complex unexported types should also be documented
// cache stores user sessions for fast lookup.
type cache struct {
    sessions map[string]*Session
}
```

### Comment Length

```go
// ✅ Aim for 80-character lines (for narrow-screen readability)
// But don't obsess - break at punctuation and semantic units

// Process handles incoming requests, validates input, applies business logic,
// and returns the appropriate response. An error is returned if the request
// is invalid or processing fails.
func Process(req *Request) (*Response, error) {}
```

### Package Comments

```go
// ✅ Package comment must appear immediately before package clause
// Exactly one per package

// Package user provides user account management functionality.
//
// This package handles user creation, authentication, and authorization.
// It interacts with the database to persist user data.
package user
```

### Comment Style

```go
// ✅ Complete sentences should be capitalized and punctuated
// Process validates and saves the user.
func Process(u *User) error {}

// ✅ Fragments don't require punctuation
// maximum retries
const maxRetries = 3

// ✅ Explain "why", not just "what"
// Use buffered channel to avoid blocking producers under high load.
events := make(chan Event, 100)

// ❌ Avoid restating code
// Set x to 1
x := 1  // Useless comment
```

## Testing

### Test Function Naming

```go
// ✅ Test functions can use underscores
func TestUser_Create(t *testing.T) {}
func TestUser_Update_InvalidInput(t *testing.T) {}
```

### Failure Messages

```go
// ✅ Include function name, inputs, actual result, expected result
func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5
    if got != want {
        t.Errorf("Add(2, 3) = %d, want %d", got, want)
    }
}
```

### Table-Driven Tests

```go
// ✅ Use field names for clarity
func TestParse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Result
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "hello",
            want:    Result{Value: "hello"},
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   "",
            want:    Result{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse(%q) error = %v, wantErr %v",
                    tt.input, err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Parse(%q) = %v, want %v",
                    tt.input, got, tt.want)
            }
        })
    }
}
```

### Comparison

```go
// ✅ Use cmp.Equal and cmp.Diff
import "github.com/google/go-cmp/cmp"

if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("result mismatch (-want +got):\n%s", diff)
}

// ❌ Avoid assertion libraries or manual field comparison
```

### Fatal vs Error

```go
// ✅ Use t.Fatal for setup failures
func TestProcess(t *testing.T) {
    db, err := setupDatabase()
    if err != nil {
        t.Fatalf("failed to setup database: %v", err)
    }

    // Use t.Error for test failures (report all issues)
    result := Process(db)
    if result.Count != 5 {
        t.Errorf("got count %d, want 5", result.Count)
    }
    if result.Status != "ok" {
        t.Errorf("got status %q, want %q", result.Status, "ok")
    }
}

// ✅ In table-driven tests, use t.Fatal within subtests
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Parse(tt.input)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        // Continue testing
    })
}
```

## Concurrency

### Goroutine Lifetimes

```go
// ✅ Make clear when goroutines exit
func process(ctx context.Context) {
    go func() {
        <-ctx.Done()
        cleanup()
    }()
}

// ✅ Use WaitGroup to wait for completion
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        work(id)
    }(i)
}
wg.Wait()

// ❌ Never spawn goroutines without knowing how they terminate
go func() {
    for {
        // Run forever? When does it stop?
        work()
    }
}()
```

### Synchronous Functions Preferred

```go
// ✅ Prefer synchronous functions - let callers add concurrency
func Fetch(url string) ([]byte, error) {
    // Synchronous implementation
    return http.Get(url)
}

// Caller can call concurrently
go Fetch("http://example.com")

// ❌ Avoid async APIs (unless well-justified)
func FetchAsync(url string) <-chan Result {
    // Forces async
}
```

### Context Usage

```go
// ✅ Always pass context.Context as first parameter
func Process(ctx context.Context, data []byte) error {}

// ✅ Only use context.Background() in main, init, or test entrypoints
func main() {
    ctx := context.Background()
    // Use ctx
}

// ✅ Pass context through
func handler(ctx context.Context) error {
    return process(ctx, data)
}

// ❌ Never create custom context types
type MyContext struct {
    context.Context
    CustomField string
}
```

### Channel Direction

```go
// ✅ Always specify channel direction to prevent misuse
func producer(ch chan<- int) {
    ch <- 42
    // Cannot receive from ch - compile-time error
}

func consumer(ch <-chan int) {
    val := <-ch
    // Cannot send to ch - compile-time error
}

func process() {
    ch := make(chan int)
    go producer(ch)
    consumer(ch)
}
```

## API Design Best Practices

### Option Patterns

**Option Struct:**
```go
// ✅ For collecting related parameters
type ServerOptions struct {
    Host    string
    Port    int
    Timeout time.Duration
    MaxConn int
}

func NewServer(opts ServerOptions) *Server {
    // Apply defaults
    if opts.Port == 0 {
        opts.Port = 8080
    }
    return &Server{opts: opts}
}

// Usage
srv := NewServer(ServerOptions{
    Host:    "localhost",
    Timeout: 30 * time.Second,
    // Omit Port - will use default
})
```

**Variadic Options:**
```go
// ✅ For flexible configuration
type ServerOption func(*Server)

func WithPort(port int) ServerOption {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(d time.Duration) ServerOption {
    return func(s *Server) {
        s.timeout = d
    }
}

func NewServer(host string, opts ...ServerOption) *Server {
    s := &Server{
        host:    host,
        port:    8080,  // default
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage - simple calls stay clean
srv := NewServer("localhost")
srv := NewServer("localhost", WithPort(9000), WithTimeout(time.Minute))
```

## Dependency Philosophy

**Priority order:**
1. Core language constructs (channels, slices, maps, loops, structs)
2. Standard library tools
3. Internal project libraries, then external dependencies

```go
// ✅ Prefer standard library
import (
    "encoding/json"
    "net/http"
    "time"
)

// ✅ Only add external dependencies when necessary
import "github.com/google/uuid"
```

## Variable Declarations

```go
// ✅ Use := for non-zero initialization
name := "Alice"
count := 42
users := []User{{Name: "Bob"}}

// ✅ Use var declaration for zero values (indicates "empty but ready")
var buf bytes.Buffer  // Ready to use
var users []User      // nil slice, ready to append

// ✅ Composite literals for known initial values
config := Config{
    Host: "localhost",
    Port: 8080,
}

// ❌ Don't preallocate collections (unless empirical analysis)
users := make([]User, 0, 100)  // Usually unnecessary
users := []User{}               // Let runtime manage growth
```

## Package Design

```go
// ✅ Packages should contain conceptually related functionality
package user      // User management
package auth      // Authentication
package storage   // Data persistence

// ❌ Avoid generic names
package util      // Too generic
package helper    // Meaningless
package common    // Doesn't describe functionality

// ✅ Files within package should be organized by logical coupling
user/
  user.go          // Core types
  validation.go    // Validation logic
  repository.go    // Data access
```

## Panic and Recovery

```go
// ✅ Panics should rarely escape package boundaries
// Exception: API misuse detection
func (s *Stack) Pop() int {
    if len(s.items) == 0 {
        panic("Pop from empty stack")
    }
    return s.items[len(s.items)-1]
}

// ✅ Internal parsers may use panic as implementation detail
func (p *parser) parse() (result AST, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("parse error: %v", r)
        }
    }()
    // May panic internally
    return p.parseInternal(), nil
}
```

## Performance Considerations

```go
// ✅ Size hints for slices and maps should be based on empirical analysis
// Most code benefits from runtime-managed growth

// ❌ Avoid speculative optimization
users := make([]User, 0, 1000)  // Why 1000?

// ✅ Only preallocate when profiling shows benefit
// Benchmarking shows this hot path benefits from preallocation
results := make([]Result, 0, len(inputs))
```

## Toolchain

### Code Quality Verification

**MANDATORY: After any code changes, run the comprehensive check:**

```bash
# Run all checks: formatting, linting, testing, and building
just check
```

This command will automatically execute:
- Code formatting verification (`go fmt`)
- Linting (`golangci-lint run`)
- All tests with race detection
- Build verification
- Any other project-specific quality checks

**You MUST run `just check` before considering your work complete.**

### Required Tools

```bash
# Formatting
go fmt ./...

# Linting
golangci-lint run

# Testing
go test ./...
go test -race ./...  # Race detection
go test -cover ./... # Coverage

# Building
go build ./...
```

### Recommended Tools

```bash
# Static analysis
go vet ./...

# Dependency management
go mod tidy
go mod verify

# Documentation
godoc -http=:6060
```

## Quick Checklist

Before committing code, check:

- [ ] **CRITICAL: Run `just check` to verify all code quality checks pass**
- [ ] Run `go fmt` to format code
- [ ] All exported names have doc comments
- [ ] Use camel case naming, no underscores (except exceptions)
- [ ] Errors are last return value
- [ ] Error strings lowercase, no punctuation
- [ ] Error handling is explicit (not ignored)
- [ ] Use `%w` to wrap errors that need inspection
- [ ] Interfaces defined in consuming package
- [ ] Return concrete types, not interfaces
- [ ] Tests use table-driven approach
- [ ] Context as first parameter
- [ ] Channel direction specified
- [ ] Goroutines have clear exit conditions
- [ ] Package names concise, lowercase, no underscores
- [ ] Avoid generic package names (util, common)
- [ ] All tests pass
- [ ] Run race detector

**Note: Running `just check` covers most of the formatting, linting, testing, and building checks above.**

## Language Requirements for Code

**MANDATORY RULES:**

1. **All code identifiers in English:**
   ```go
   // ✅ CORRECT - All English
   type UserAccount struct {
       Name    string
       Email   string
       Balance float64
   }

   func (u *UserAccount) Deposit(amount float64) error {
       if amount <= 0 {
           return errors.New("amount must be positive")
       }
       u.Balance += amount
       return nil
   }

   // ❌ WRONG - Contains Chinese
   type 用户账户 struct {
       姓名 string
       邮箱 string
   }
   ```

2. **All comments in English:**
   ```go
   // ✅ CORRECT - English comments
   // Process validates the user input and returns the processed result.
   // It returns an error if validation fails.
   func Process(input string) (string, error) {
       // Check if input is empty
       if input == "" {
           return "", errors.New("input cannot be empty")
       }
       return input, nil
   }

   // ❌ WRONG - Chinese comments
   // 处理用户输入
   func Process(input string) (string, error) {
       // 检查输入是否为空
       if input == "" {
           return "", errors.New("input cannot be empty")
       }
       return input, nil
   }
   ```

3. **All documentation in English:**
   ```go
   // ✅ CORRECT - English documentation
   // Package account provides user account management functionality.
   //
   // This package handles account creation, balance management, and
   // transaction processing. All monetary values are represented as
   // float64 for precision.
   package account

   // ❌ WRONG - Chinese documentation
   // Package account 提供用户账户管理功能
   package account
   ```

4. **All error messages in English:**
   ```go
   // ✅ CORRECT - English error messages
   if user == nil {
       return errors.New("user cannot be nil")
   }

   if balance < 0 {
       return fmt.Errorf("insufficient balance: have %f, need %f",
           current, required)
   }

   // ❌ WRONG - Chinese error messages
   if user == nil {
       return errors.New("用户不能为空")
   }
   ```

5. **All test names and test output in English:**
   ```go
   // ✅ CORRECT - English test names and messages
   func TestUserAccount_Deposit(t *testing.T) {
       tests := []struct {
           name    string
           initial float64
           deposit float64
           want    float64
           wantErr bool
       }{
           {
               name:    "valid deposit",
               initial: 100.0,
               deposit: 50.0,
               want:    150.0,
               wantErr: false,
           },
           {
               name:    "negative deposit",
               initial: 100.0,
               deposit: -50.0,
               want:    100.0,
               wantErr: true,
           },
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               account := &UserAccount{Balance: tt.initial}
               err := account.Deposit(tt.deposit)

               if (err != nil) != tt.wantErr {
                   t.Errorf("Deposit() error = %v, wantErr %v",
                       err, tt.wantErr)
                   return
               }

               if account.Balance != tt.want {
                   t.Errorf("Deposit() balance = %v, want %v",
                       account.Balance, tt.want)
               }
           })
       }
   }

   // ❌ WRONG - Chinese test names
   func TestUserAccount_存款(t *testing.T) {
       // ...
   }
   ```

6. **All log messages in English:**
   ```go
   // ✅ CORRECT - English log messages
   log.Info("user logged in successfully",
       zap.String("user_id", userID),
       zap.String("ip", ipAddress))

   log.Error("failed to connect to database",
       zap.Error(err),
       zap.String("host", dbHost))

   // ❌ WRONG - Chinese log messages
   log.Info("用户登录成功")
   ```

## Complete Example

```go
// ✅ CORRECT - Full English implementation

// Package account provides financial account management functionality.
//
// This package implements core banking operations including deposits,
// withdrawals, and balance inquiries. All monetary operations use
// float64 for precision and return errors for invalid operations.
package account

import (
    "errors"
    "fmt"
    "sync"
)

var (
    // ErrInsufficientBalance is returned when withdrawal exceeds balance.
    ErrInsufficientBalance = errors.New("insufficient balance")

    // ErrInvalidAmount is returned when amount is zero or negative.
    ErrInvalidAmount = errors.New("amount must be positive")
)

// Account represents a user's financial account.
type Account struct {
    // mu protects concurrent access to balance
    mu      sync.RWMutex
    id      string
    owner   string
    balance float64
}

// NewAccount creates a new account with the given owner.
func NewAccount(id, owner string) *Account {
    return &Account{
        id:      id,
        owner:   owner,
        balance: 0,
    }
}

// Deposit adds the specified amount to the account balance.
// It returns an error if the amount is not positive.
func (a *Account) Deposit(amount float64) error {
    if amount <= 0 {
        return ErrInvalidAmount
    }

    a.mu.Lock()
    defer a.mu.Unlock()

    a.balance += amount
    return nil
}

// Withdraw removes the specified amount from the account balance.
// It returns an error if the amount is invalid or exceeds the balance.
func (a *Account) Withdraw(amount float64) error {
    if amount <= 0 {
        return ErrInvalidAmount
    }

    a.mu.Lock()
    defer a.mu.Unlock()

    if a.balance < amount {
        return fmt.Errorf("%w: have %.2f, need %.2f",
            ErrInsufficientBalance, a.balance, amount)
    }

    a.balance -= amount
    return nil
}

// Balance returns the current account balance.
func (a *Account) Balance() float64 {
    a.mu.RLock()
    defer a.mu.RUnlock()
    return a.balance
}
```

## Reference Resources

- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)

---

**Remember: Code is written for humans to read, it just happens to be executable by machines. Always prioritize clarity and maintainability.**

**CRITICAL REMINDER: ALL code, comments, documentation, variable names, function names, error messages, log messages, and any text in code files MUST be in FULL ENGLISH. Respond to users in Chinese, but keep all code artifacts in English.**
