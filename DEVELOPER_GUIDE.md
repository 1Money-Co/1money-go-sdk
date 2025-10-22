# OneMoney Go SDK - Developer Guide

This guide shows you how to extend the SDK with custom business modules using interface-based design.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Quick Start](#quick-start)
- [Creating Custom Services](#creating-custom-services)
- [Best Practices](#best-practices)
- [Testing](#testing)

## Architecture Overview

The SDK uses a layered architecture with interface-based service design:

```
┌─────────────────────────────────────────┐
│    Public API (services, client)        │
│    (Interface-based modules)            │
├─────────────────────────────────────────┤
│          Client Layer                   │
│    (Service Registry)                   │
├─────────────────────────────────────────┤
│      Internal Implementation            │
│    (Transport + Auth - not exposed)     │
├─────────────────────────────────────────┤
│         Transport Layer                 │
│      (HTTP Communication)               │
├─────────────────────────────────────────┤
│          Auth Layer                     │
│    (HMAC-SHA256 Signature)             │
└─────────────────────────────────────────┘
```

### Internal vs Public Packages

**Internal Packages** (`internal/`):
- `internal/auth` - Authentication and signature generation (implementation detail)
- `internal/transport` - HTTP transport layer (implementation detail)
- These packages are NOT accessible outside the module
- Users cannot import them directly

**Public Packages**:
- `client` - Main SDK client and service registry
- `services/*` - Business service modules
- `onemoney` - Extended client with pre-registered services

**Why Internal?**
Go's `internal` package mechanism ensures that implementation details remain hidden from users. This allows us to:
- Change internal implementations without breaking user code
- Maintain a stable public API
- Prevent users from depending on internal structures

## Quick Start

### Using Pre-Built Services

```go
import "github.com/1Money-Co/1money-go-sdk/onemoney"

// Create client with pre-registered services
c := onemoney.NewClient(&onemoney.Config{
    AccessKey: "your-access-key",
    SecretKey: "your-secret-key",
    BaseURL:   "http://localhost:9000",
})

// Use services directly
resp, err := c.Echo.Get(ctx)
resp, err := c.Echo.Post(ctx, &echo.Request{Message: "hello"})
```

## Creating Custom Services

### Step 1: Define Service Interface

**Key Principle**: Define an interface that lists all supported operations. This makes the service's capabilities immediately visible.

```go
package payment

import (
    "context"
)

// Service defines the payment service interface.
// All supported operations are visible here.
type Service interface {
    // Create creates a new payment.
    Create(ctx context.Context, req *CreateRequest) (*Payment, error)

    // Get retrieves a payment by ID.
    Get(ctx context.Context, id string) (*Payment, error)

    // Cancel cancels a pending payment.
    Cancel(ctx context.Context, id string) error

    // List retrieves a list of payments with optional filters.
    List(ctx context.Context, opts *ListOptions) ([]*Payment, error)
}
```

### Step 2: Define Data Models

```go
// Payment represents a payment in the system.
type Payment struct {
    ID       string  `json:"id"`
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
    Status   string  `json:"status"`
}

// CreateRequest represents the request to create a payment.
type CreateRequest struct {
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
}

// ListOptions provides filtering options for listing payments.
type ListOptions struct {
    Status string
    Limit  int
    Offset int
}
```

### Step 3: Implement the Service

```go
import (
    "encoding/json"
    "fmt"
    "github.com/1Money-Co/1money-go-sdk/client"
)

// serviceImpl is the concrete implementation of the payment service.
// It is private to enforce usage through the interface.
type serviceImpl struct {
    client.BaseService
}

// NewService creates a new payment service instance.
// Returns the interface type, not the implementation.
func NewService() Service {
    return &serviceImpl{}
}

// Create creates a new payment.
func (s *serviceImpl) Create(ctx context.Context, req *CreateRequest) (*Payment, error) {
    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    // Use inherited Post method from BaseService
    resp, err := s.Post(ctx, "/openapi/payments", body)
    if err != nil {
        return nil, fmt.Errorf("failed to create payment: %w", err)
    }

    var payment Payment
    if err := json.Unmarshal(resp.Body, &payment); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return &payment, nil
}

// Get retrieves a payment by ID.
func (s *serviceImpl) Get(ctx context.Context, id string) (*Payment, error) {
    path := fmt.Sprintf("/openapi/payments/%s", id)

    resp, err := s.BaseService.Get(ctx, path)
    if err != nil {
        return nil, fmt.Errorf("failed to get payment: %w", err)
    }

    var payment Payment
    if err := json.Unmarshal(resp.Body, &payment); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return &payment, nil
}

// Cancel cancels a pending payment.
func (s *serviceImpl) Cancel(ctx context.Context, id string) error {
    path := fmt.Sprintf("/openapi/payments/%s/cancel", id)

    _, err := s.Post(ctx, path, nil)
    if err != nil {
        return fmt.Errorf("failed to cancel payment: %w", err)
    }

    return nil
}

// List retrieves a list of payments with optional filters.
func (s *serviceImpl) List(ctx context.Context, opts *ListOptions) ([]*Payment, error) {
    path := "/openapi/payments"

    if opts != nil {
        path += buildQueryString(opts)
    }

    resp, err := s.BaseService.Get(ctx, path)
    if err != nil {
        return nil, fmt.Errorf("failed to list payments: %w", err)
    }

    var payments []*Payment
    if err := json.Unmarshal(resp.Body, &payments); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return payments, nil
}

func buildQueryString(opts *ListOptions) string {
    // Build query string from options
    // Implementation details...
    return ""
}
```

### Available BaseService Methods

Your service implementation automatically inherits these methods:

```go
s.Get(ctx, path)              // GET request
s.Post(ctx, path, body)       // POST request
s.Put(ctx, path, body)        // PUT request
s.Delete(ctx, path)           // DELETE request
s.Patch(ctx, path, body)      // PATCH request
s.Do(ctx, customRequest)      // Custom request with full control
```

## Service Registration

### Method 1: Using Base Client

```go
import (
    "github.com/1Money-Co/1money-go-sdk/client"
    "your-project/services/payment"
)

// Create base client
c := client.NewClient(accessKey, secretKey)

// Create and register service
paymentSvc := payment.NewService()

// Cast to client.Service for registration
if svc, ok := paymentSvc.(client.Service); ok {
    c.RegisterService("payment", svc)
}

// Use the service
payment, err := paymentSvc.Create(ctx, &payment.CreateRequest{
    Amount:   100.00,
    Currency: "USD",
})
```

### Method 2: Extending OneMoney Client

Create a custom client that includes your services:

```go
package myclient

import (
    "github.com/1Money-Co/1money-go-sdk/onemoney"
    "github.com/1Money-Co/1money-go-sdk/client"
    "your-project/services/payment"
)

// Client extends the OneMoney client with custom services.
type Client struct {
    *onemoney.Client

    // Custom services (using interface types)
    Payment payment.Service
}

// NewClient creates a client with all services pre-registered.
func NewClient(cfg *onemoney.Config) *Client {
    // Create base OneMoney client
    baseClient := onemoney.NewClient(cfg)

    // Create custom services
    paymentSvc := payment.NewService()

    // Register custom services
    if svc, ok := paymentSvc.(client.Service); ok {
        baseClient.RegisterCustomService("payment", svc)
    }

    return &Client{
        Client:  baseClient,
        Payment: paymentSvc,
    }
}
```

Usage:

```go
c := myclient.NewClient(&onemoney.Config{
    AccessKey: "your-key",
    SecretKey: "your-secret",
})

// Use built-in services
resp, err := c.Echo.Get(ctx)

// Use custom services
payment, err := c.Payment.Create(ctx, req)
```

## Best Practices

### 1. Interface-First Design

Always define the interface first. This makes the service's capabilities immediately visible:

```go
// ✅ Good - Interface clearly shows all operations
type Service interface {
    Create(ctx context.Context, req *CreateRequest) (*Resource, error)
    Get(ctx context.Context, id string) (*Resource, error)
    Update(ctx context.Context, id string, req *UpdateRequest) (*Resource, error)
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, opts *ListOptions) ([]*Resource, error)
}

// ❌ Bad - No interface, unclear what operations are available
type Service struct {
    client.BaseService
}
```

### 2. Return Interface, Not Implementation

```go
// ✅ Good - Return interface type
func NewService() Service {
    return &serviceImpl{}
}

// ❌ Bad - Expose implementation
func NewService() *serviceImpl {
    return &serviceImpl{}
}
```

### 3. Private Implementation

Keep the implementation struct private to enforce interface usage:

```go
// ✅ Good - Implementation is private
type serviceImpl struct {
    client.BaseService
}

// ❌ Bad - Implementation is exported
type ServiceImpl struct {
    client.BaseService
}
```

### 4. Error Handling

Always wrap errors with context:

```go
func (s *serviceImpl) Get(ctx context.Context, id string) (*Item, error) {
    resp, err := s.BaseService.Get(ctx, path)
    if err != nil {
        return nil, fmt.Errorf("failed to get item %s: %w", id, err)
    }

    if err := json.Unmarshal(resp.Body, &item); err != nil {
        return nil, fmt.Errorf("failed to parse item response: %w", err)
    }

    return &item, nil
}
```

### 5. Context Awareness

Always accept `context.Context` as the first parameter:

```go
// ✅ Good
func (s *Service) Get(ctx context.Context, id string) (*Item, error)

// ❌ Bad
func (s *Service) Get(id string) (*Item, error)
```

### 6. Documentation

Document all exported types and methods:

```go
// Service provides payment processing functionality.
//
// All payment operations including creation, retrieval, and cancellation
// are handled through this interface.
type Service interface {
    // Create creates a new payment transaction.
    //
    // The payment will be created in pending status and must be
    // confirmed before funds are transferred.
    Create(ctx context.Context, req *CreateRequest) (*Payment, error)
}
```

## Testing

### Unit Testing with Mocks

Since services are defined as interfaces, they're easy to mock:

```go
package myapp_test

import (
    "context"
    "testing"
    "your-project/services/payment"
)

// Mock implementation
type mockPaymentService struct{}

func (m *mockPaymentService) Create(ctx context.Context, req *payment.CreateRequest) (*payment.Payment, error) {
    return &payment.Payment{
        ID:     "mock-id",
        Amount: req.Amount,
        Status: "completed",
    }, nil
}

func (m *mockPaymentService) Get(ctx context.Context, id string) (*payment.Payment, error) {
    return &payment.Payment{ID: id}, nil
}

func (m *mockPaymentService) Cancel(ctx context.Context, id string) error {
    return nil
}

func (m *mockPaymentService) List(ctx context.Context, opts *payment.ListOptions) ([]*payment.Payment, error) {
    return []*payment.Payment{}, nil
}

// Test using mock
func TestProcessPayment(t *testing.T) {
    mockSvc := &mockPaymentService{}

    // Test your application code that uses payment.Service
    result, err := processPayment(mockSvc, 100.00)

    if err != nil {
        t.Fatalf("processPayment() error = %v", err)
    }

    if result.ID != "mock-id" {
        t.Errorf("got ID %s, want %s", result.ID, "mock-id")
    }
}
```

### Integration Testing

```go
func TestIntegration(t *testing.T) {
    // Use real client with test credentials
    c := onemoney.NewClient(&onemoney.Config{
        AccessKey: testAccessKey,
        SecretKey: testSecretKey,
        BaseURL:   testServerURL,
    })

    // Run integration tests
    resp, err := c.Echo.Get(context.Background())
    if err != nil {
        t.Fatalf("Echo.Get() error = %v", err)
    }

    if resp.Message == "" {
        t.Error("expected non-empty message")
    }
}
```

## Complete Example: Echo Service

See `services/echo/echo.go` for a complete example of an interface-based service:

```go
// Interface definition - all operations visible
type Service interface {
    Get(ctx context.Context) (*Response, error)
    Post(ctx context.Context, req *Request) (*Response, error)
}

// Private implementation
type serviceImpl struct {
    client.BaseService
}

// Constructor returns interface
func NewService() Service {
    return &serviceImpl{}
}
```

## Summary

The interface-based service design provides:

1. **Clear Contracts**: Interface defines all available operations
2. **Easy Mocking**: Interface-based design enables simple testing
3. **Encapsulation**: Implementation details are hidden
4. **Type Safety**: Compile-time checking of service usage
5. **Discoverability**: IDE autocomplete shows all available methods

Follow this pattern when creating new services to maintain consistency and quality across your SDK.
