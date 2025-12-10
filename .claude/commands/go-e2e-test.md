# Go E2E Test Writing Guide

Guide for writing end-to-end (e2e) tests in Go using `github.com/stretchr/testify/suite`. This skill provides patterns and best practices for implementing business domain e2e tests.

## Core Architecture

### Test Suite Hierarchy

```
E2ETestSuite (base)
    └── CustomerDependentTestSuite (with customer setup)
            └── DomainTestSuite (e.g., ExternalAccountsTestSuite)
```

### Base Test Suite Structure

```go
// Base suite - provides client and context
type E2ETestSuite struct {
    suite.Suite
    Client *onemoney.Client
    Ctx    context.Context
}

func (s *E2ETestSuite) SetupSuite() {
    // One-time setup before all tests
    client, err := onemoney.NewClient(cfg)
    if err != nil {
        s.T().Fatalf("failed to create client: %v", err)
    }
    s.Client = client
    s.Ctx = context.Background()
}
```

### Domain Test Suite Structure

```go
// Domain-specific suite embedding the base suite
type ExternalAccountsTestSuite struct {
    CustomerDependentTestSuite  // Embeds base suite with customer setup
}

// Entry point - MUST have this function
func TestExternalAccountsTestSuite(t *testing.T) {
    suite.Run(t, new(ExternalAccountsTestSuite))
}
```

## Test Method Patterns

### Scenario-Based Testing with Subtests

Use `s.Run()` for multiple scenarios within a single test method:

```go
func (s *ExternalAccountsTestSuite) TestExternalAccounts_List() {
    s.Run("Empty", func() {
        resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
        s.Require().NoError(err, "should succeed even with no accounts")
        s.Require().NotNil(resp, "Response should not be nil")
    })

    s.Run("WithData", func() {
        _, err := s.EnsureExternalAccount()
        s.Require().NoError(err)

        resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
        s.Require().NoError(err)
        s.Require().NotEmpty(resp)
    })

    s.Run("FilterByStatus", func() {
        req := &external_accounts.ListExternalAccountsRequest{
            Status: external_accounts.BankAccountStatusAPPROVED,
        }
        resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, req)
        s.Require().NoError(err)

        for i := range resp {
            s.Equal("APPROVED", resp[i].Status)
        }
    })
}
```

### Table-Driven Tests with Subtests (Preferred)

For testing multiple variations of the same operation:

```go
func (s *CustomerTestSuite) TestCustomerService_CreateCustomer_ValidationErrors() {
    tests := []struct {
        name        string
        setup       func() *customer.CreateCustomerRequest
        wantErr     bool
        errContains string
    }{
        {
            name: "InvalidMIMEType",
            setup: func() *customer.CreateCustomerRequest {
                req := s.baseRequest()
                req.Documents[0].File = "data:application/x-msdownload;base64,TVqQAAMAAAAEAAAA"
                return req
            },
            wantErr:     true,
            errContains: "invalid file format",
        },
        {
            name: "InvalidBase64",
            setup: func() *customer.CreateCustomerRequest {
                req := s.baseRequest()
                req.Documents[0].File = "data:image/jpeg;base64,not-valid-base64!!!"
                return req
            },
            wantErr:     true,
            errContains: "invalid base64",
        },
        {
            name: "MissingRequiredField",
            setup: func() *customer.CreateCustomerRequest {
                req := s.baseRequest()
                req.Email = ""
                return req
            },
            wantErr:     true,
            errContains: "email required",
        },
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            req := tt.setup()
            _, err := s.Client.Customer.CreateCustomer(s.Ctx, req)

            if tt.wantErr {
                s.Require().Error(err)
                if tt.errContains != "" {
                    s.Contains(err.Error(), tt.errContains)
                }
            } else {
                s.Require().NoError(err)
            }
        })
    }
}
```

### Comprehensive CRUD Test Pattern

```go
func (s *ExternalAccountsTestSuite) TestExternalAccounts_CreateAndGet() {
    // Create
    createReq := FakeExternalAccountRequest()
    createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, s.CustomerID, createReq)
    s.Require().NoError(err, "CreateExternalAccount should succeed")
    s.Require().NotNil(createResp)
    s.NotEmpty(createResp.ExternalAccountID)
    s.T().Logf("Created external account:\n%s", PrettyJSON(createResp))

    // Get by ID
    getResp, err := s.Client.ExternalAccounts.GetExternalAccount(s.Ctx, s.CustomerID, createResp.ExternalAccountID)
    s.Require().NoError(err, "GetExternalAccount should succeed")
    s.Equal(createResp.ExternalAccountID, getResp.ExternalAccountID)

    // Get by idempotency key
    getByKeyResp, err := s.Client.ExternalAccounts.GetExternalAccountByIdempotencyKey(s.Ctx, s.CustomerID, createReq.IdempotencyKey)
    s.Require().NoError(err)
    s.Equal(createResp.ExternalAccountID, getByKeyResp.ExternalAccountID)
}
```

## Assertion Guidelines

### Use `s.Require()` vs `s.Assert()`

- `s.Require()` - **Stops test execution** on failure. Use for critical preconditions.
- `s.Assert()` or direct calls like `s.Equal()` - **Continues execution** on failure. Use for non-critical assertions.

```go
// CRITICAL: Use Require for prerequisites
s.Require().NoError(err, "API call must succeed")
s.Require().NotNil(resp, "Response is required for subsequent checks")
s.Require().NotEmpty(resp.Items, "Must have items to continue")

// NON-CRITICAL: Use Assert for additional validations
s.NotEmpty(item.Name, "Name should not be empty")
s.Equal(expected, actual, "Values should match")
```

### Common Assertion Patterns

```go
// Error handling
s.Require().NoError(err, "operation should succeed")
s.Require().Error(err, "operation should fail")

// Nil checks
s.Require().NotNil(resp, "Response should not be nil")
s.Nil(resp.OptionalField, "Optional field should be nil")

// Empty checks
s.NotEmpty(resp.ID, "ID should not be empty")
s.Empty(resp.Errors, "Should have no errors")

// Equality
s.Equal(expected, actual, "should match")
s.NotEqual(a, b, "should differ")

// Collection assertions
s.Len(items, 3, "should have exactly 3 items")
s.GreaterOrEqual(len(items), 1, "should have at least 1 item")
s.Contains(items, expectedItem, "should contain item")
```

## Helper Patterns

### Ensure Functions (Guarantee Prerequisites)

Create `EnsureXxx` methods to guarantee test prerequisites exist:

```go
func (s *CustomerDependentTestSuite) EnsureExternalAccount() (string, error) {
    // Try to get existing
    accounts, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
    if err != nil {
        return "", fmt.Errorf("ListExternalAccounts failed: %w", err)
    }

    // Return existing if available
    if len(accounts) > 0 {
        return accounts[0].ExternalAccountID, nil
    }

    // Create new if needed
    createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, s.CustomerID, FakeExternalAccountRequest())
    if err != nil {
        return "", fmt.Errorf("CreateExternalAccount failed: %w", err)
    }

    return createResp.ExternalAccountID, nil
}
```

### Fake Data Generators

Use `gofakeit` for generating realistic test data:

```go
func FakeExternalAccountRequest() *external_accounts.CreateExternalAccountRequest {
    faker := gofakeit.New(0)
    return &external_accounts.CreateExternalAccountRequest{
        IdempotencyKey:       faker.UUID(),
        BankNetworkName:      external_accounts.BankNetworkNameUSACH,
        Currency:             external_accounts.CurrencyUSD,
        BankName:             faker.Company() + " Bank",
        BankAccountOwnerName: faker.Name(),
        BankAccountNumber:    faker.DigitN(9),
        BankRoutingNumber:    faker.DigitN(9),
    }
}
```

### Logging Helper

```go
func PrettyJSON(v any) string {
    b, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return fmt.Sprintf("%+v", v)
    }
    return string(b)
}

// Usage in tests
s.T().Logf("Created account:\n%s", PrettyJSON(resp))
```

## File Organization

```
tests/e2e/
├── suite_test.go           # Base suites, helpers, fake data generators
├── customer_test.go        # Customer domain tests
├── external_accounts_test.go
├── transactions_test.go
├── withdrawals_test.go
└── ...
```

## Test Naming Conventions

### Test Method Names

```go
// Pattern: Test{Domain}_{Operation}
func (s *CustomerTestSuite) TestCustomerService_CreateCustomer() {}
func (s *CustomerTestSuite) TestCustomerService_ListCustomers() {}
func (s *CustomerTestSuite) TestCustomerService_GetCustomer() {}
func (s *CustomerTestSuite) TestCustomerService_UpdateCustomer() {}

// Pattern: Test{Domain}_{Operation}_{Scenario}
func (s *CustomerTestSuite) TestCustomerService_CreateCustomer_InvalidFileFormat() {}
func (s *CustomerTestSuite) TestCustomerService_CreateCustomer_InvalidBase64() {}
```

### Subtest Names (s.Run)

Use descriptive, PascalCase names:

```go
s.Run("Empty", func() { ... })
s.Run("WithData", func() { ... })
s.Run("WithPagination", func() { ... })
s.Run("FilterByStatus", func() { ... })
s.Run("FilterByAsset", func() { ... })
```

## Complete Example

```go
package e2e

import (
    "testing"

    "github.com/stretchr/testify/suite"

    "github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
)

type ExternalAccountsTestSuite struct {
    CustomerDependentTestSuite
}

func (s *ExternalAccountsTestSuite) TestExternalAccounts_List() {
    s.Run("Empty", func() {
        resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
        s.Require().NoError(err, "ListExternalAccounts should succeed")
        s.Require().NotNil(resp)
        s.T().Logf("External accounts: %d", len(resp))
    })

    s.Run("WithData", func() {
        _, err := s.EnsureExternalAccount()
        s.Require().NoError(err)

        resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
        s.Require().NoError(err)
        s.Require().NotEmpty(resp)

        for i := range resp {
            s.NotEmpty(resp[i].ExternalAccountID)
            s.NotEmpty(resp[i].CustomerID)
            s.NotEmpty(resp[i].Status)
        }
    })

    s.Run("FilterByStatus", func() {
        req := &external_accounts.ListExternalAccountsRequest{
            Status: external_accounts.BankAccountStatusAPPROVED,
        }

        resp, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, req)
        s.Require().NoError(err)

        for i := range resp {
            s.Equal("APPROVED", resp[i].Status)
        }
    })
}

func (s *ExternalAccountsTestSuite) TestExternalAccounts_CreateAndGet() {
    createReq := FakeExternalAccountRequest()

    // Create
    createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, s.CustomerID, createReq)
    s.Require().NoError(err)
    s.NotEmpty(createResp.ExternalAccountID)
    s.T().Logf("Created:\n%s", PrettyJSON(createResp))

    // Get
    getResp, err := s.Client.ExternalAccounts.GetExternalAccount(s.Ctx, s.CustomerID, createResp.ExternalAccountID)
    s.Require().NoError(err)
    s.Equal(createResp.ExternalAccountID, getResp.ExternalAccountID)
}

func (s *ExternalAccountsTestSuite) TestExternalAccounts_Delete() {
    // Create account to delete
    createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, s.CustomerID, FakeExternalAccountRequest())
    s.Require().NoError(err)

    // Delete
    err = s.Client.ExternalAccounts.DeleteExternalAccount(s.Ctx, s.CustomerID, createResp.ExternalAccountID)
    s.Require().NoError(err)
    s.T().Logf("Deleted: %s", createResp.ExternalAccountID)
}

func TestExternalAccountsTestSuite(t *testing.T) {
    suite.Run(t, new(ExternalAccountsTestSuite))
}
```

## Checklist for New E2E Tests

1. [ ] Create domain-specific test suite embedding appropriate base suite
2. [ ] Add `TestXxxTestSuite(t *testing.T)` entry point function
3. [ ] Group related scenarios using `s.Run()` subtests
4. [ ] Use table-driven tests for multiple input variations
5. [ ] Use `s.Require()` for critical assertions, `s.Assert()` for non-critical
6. [ ] Create `EnsureXxx` helpers for test prerequisites
7. [ ] Create `FakeXxx` helpers for generating test data
8. [ ] Log important responses using `PrettyJSON()` helper
9. [ ] Follow naming conventions: `Test{Domain}_{Operation}` or `Test{Domain}_{Operation}_{Scenario}`
