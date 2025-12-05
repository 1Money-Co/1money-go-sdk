# Run E2E Tests and Discover Edge Cases

Guide for running Go e2e tests, analyzing results, proactively discovering boundary issues, and adding test cases.

## Running Tests

### Run All E2E Tests

```bash
go test -v ./tests/e2e/... -count=1
```

### Run Specific Test Suite

```bash
# Run a specific test suite
go test -v ./tests/e2e/... -run TestExternalAccountsTestSuite -count=1

# Run a specific test method
go test -v ./tests/e2e/... -run TestExternalAccountsTestSuite/TestExternalAccounts_List -count=1

# Run a specific subtest
go test -v ./tests/e2e/... -run "TestExternalAccountsTestSuite/TestExternalAccounts_List/FilterByStatus" -count=1
```

### Run with Timeout

```bash
go test -v ./tests/e2e/... -timeout 10m -count=1
```

### Run with Coverage

```bash
go test -v ./tests/e2e/... -coverprofile=coverage.out -count=1
go tool cover -html=coverage.out
```

## Workflow: Analyze Service and Discover Edge Cases

When tasked with improving test coverage for a service, follow this workflow:

### Step 1: Understand the Service Interface

Read the service definition to understand all available operations:

```bash
# Find and read the service file
cat pkg/service/{domain}/service.go
```

Key things to identify:
- All methods in the `Service` interface
- Request/response types and their fields
- Optional vs required fields
- Field constraints (enums, ranges, formats)

### Step 2: Analyze Existing Test Coverage

```bash
# Read existing tests
cat tests/e2e/{domain}_test.go
```

Create a coverage matrix:

| Method | Happy Path | Empty/Nil | Invalid Input | Edge Cases |
|--------|-----------|-----------|---------------|------------|
| Create | ? | ? | ? | ? |
| Get    | ? | ? | ? | ? |
| List   | ? | ? | ? | ? |
| Update | ? | ? | ? | ? |
| Delete | ? | ? | ? | ? |

### Step 3: Identify Missing Test Scenarios

#### Category 1: Input Validation Boundaries

For each request field, consider:

| Field Type | Test Cases |
|------------|------------|
| **String** | Empty string, whitespace only, max length, special characters, unicode, SQL injection patterns |
| **Integer** | Zero, negative, max int, boundary values (e.g., page size 0, 1, 100, 101) |
| **Enum** | Invalid enum value, empty string, case sensitivity |
| **UUID** | Invalid format, empty, non-existent ID |
| **Email** | Invalid format, empty, special domains |
| **Date** | Invalid format, future date, past date, boundary dates |
| **Array** | Empty array, single item, max items, duplicate items |
| **Nested Object** | Nil object, partial fields, all fields |
| **Optional Field** | Nil/omitted, empty value, valid value |

#### Category 2: State Transitions

For resources with status/state:
- Test operations on resources in each possible state
- Test invalid state transitions
- Test concurrent operations

#### Category 3: Pagination Boundaries

```go
// Pagination edge cases
tests := []struct {
    name string
    page int
    size int
}{
    {"FirstPage", 1, 10},
    {"ZeroPage", 0, 10},           // Should default or error?
    {"NegativePage", -1, 10},      // Should error
    {"ZeroSize", 1, 0},            // Should default or error?
    {"MaxSize", 1, 100},           // At limit
    {"OverMaxSize", 1, 101},       // Over limit - should cap or error?
    {"LargePageNumber", 9999, 10}, // Beyond data - should return empty
}
```

#### Category 4: Idempotency

```go
// Test idempotency key behavior
s.Run("DuplicateIdempotencyKey", func() {
    req := FakeCreateRequest()

    // First create
    resp1, err := s.Client.Service.Create(s.Ctx, s.CustomerID, req)
    s.Require().NoError(err)

    // Second create with same idempotency key
    resp2, err := s.Client.Service.Create(s.Ctx, s.CustomerID, req)
    s.Require().NoError(err)

    // Should return same resource, not create duplicate
    s.Equal(resp1.ID, resp2.ID)
})

s.Run("GetByIdempotencyKey", func() {
    req := FakeCreateRequest()
    createResp, err := s.Client.Service.Create(s.Ctx, s.CustomerID, req)
    s.Require().NoError(err)

    getResp, err := s.Client.Service.GetByIdempotencyKey(s.Ctx, s.CustomerID, req.IdempotencyKey)
    s.Require().NoError(err)
    s.Equal(createResp.ID, getResp.ID)
})
```

#### Category 5: Not Found Scenarios

```go
s.Run("GetNonExistent", func() {
    _, err := s.Client.Service.Get(s.Ctx, s.CustomerID, "non-existent-uuid")
    s.Require().Error(err)
    // Verify it's a 404-type error
})

s.Run("DeleteNonExistent", func() {
    err := s.Client.Service.Delete(s.Ctx, s.CustomerID, "non-existent-uuid")
    // Should this error or succeed silently?
})
```

#### Category 6: Authorization Boundaries

```go
s.Run("AccessOtherCustomerResource", func() {
    // Create resource for customer A
    resp, err := s.Client.Service.Create(s.Ctx, s.CustomerID, FakeCreateRequest())
    s.Require().NoError(err)

    // Try to access with different customer ID
    _, err = s.Client.Service.Get(s.Ctx, "other-customer-id", resp.ID)
    s.Require().Error(err) // Should be forbidden
})
```

### Step 4: Add Test Cases

Use table-driven tests for systematic coverage:

```go
func (s *ServiceTestSuite) TestService_Create_ValidationErrors() {
    tests := []struct {
        name    string
        setup   func() *CreateRequest
        wantErr bool
        errType string // "validation", "not_found", "conflict", etc.
    }{
        {
            name: "EmptyRequiredField",
            setup: func() *CreateRequest {
                req := FakeCreateRequest()
                req.RequiredField = ""
                return req
            },
            wantErr: true,
            errType: "validation",
        },
        {
            name: "InvalidEnumValue",
            setup: func() *CreateRequest {
                req := FakeCreateRequest()
                req.Status = "INVALID_STATUS"
                return req
            },
            wantErr: true,
            errType: "validation",
        },
        {
            name: "NegativeAmount",
            setup: func() *CreateRequest {
                req := FakeCreateRequest()
                req.Amount = "-100.00"
                return req
            },
            wantErr: true,
            errType: "validation",
        },
        {
            name: "InvalidUUIDFormat",
            setup: func() *CreateRequest {
                req := FakeCreateRequest()
                req.ReferenceID = "not-a-uuid"
                return req
            },
            wantErr: true,
            errType: "validation",
        },
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            req := tt.setup()
            _, err := s.Client.Service.Create(s.Ctx, s.CustomerID, req)

            if tt.wantErr {
                s.Require().Error(err, "expected error for %s", tt.name)
                s.T().Logf("Expected error: %v", err)
            } else {
                s.Require().NoError(err)
            }
        })
    }
}
```

## Edge Case Discovery Checklist

### For Create Operations

- [ ] All required fields provided
- [ ] Missing each required field individually
- [ ] Empty string for required string fields
- [ ] Invalid format for typed fields (UUID, email, date)
- [ ] Invalid enum values
- [ ] Boundary values for numeric fields (0, negative, max)
- [ ] String length boundaries (empty, 1 char, max length, over max)
- [ ] Special characters in string fields
- [ ] Duplicate idempotency key
- [ ] Unicode/emoji in text fields

### For Get Operations

- [ ] Valid existing ID
- [ ] Non-existent ID (404)
- [ ] Invalid ID format
- [ ] ID belonging to different customer (403)
- [ ] Get by idempotency key (if supported)

### For List Operations

- [ ] Empty list (no data)
- [ ] List with data
- [ ] Pagination: page 0, 1, negative, large number
- [ ] Page size: 0, 1, max, over max
- [ ] Filter by each supported filter
- [ ] Filter with invalid values
- [ ] Sort by each supported field
- [ ] Sort direction (asc/desc)
- [ ] Combined filters

### For Update Operations

- [ ] Update single field
- [ ] Update multiple fields
- [ ] Update with no changes
- [ ] Update non-existent resource
- [ ] Update with invalid values
- [ ] Concurrent updates (optimistic locking)
- [ ] Update immutable fields (should fail)

### For Delete Operations

- [ ] Delete existing resource
- [ ] Delete already deleted resource
- [ ] Delete non-existent resource
- [ ] Delete resource with dependencies
- [ ] Verify resource is actually deleted (GET after DELETE)

## Example: Complete Edge Case Test Suite

```go
func (s *AutoConversionRulesTestSuite) TestAutoConversionRules_Create_EdgeCases() {
    tests := []struct {
        name    string
        setup   func() *auto_conversion_rules.CreateRuleRequest
        wantErr bool
    }{
        {
            name: "ValidRequest",
            setup: func() *auto_conversion_rules.CreateRuleRequest {
                return FakeAutoConversionRuleRequest()
            },
            wantErr: false,
        },
        {
            name: "EmptyIdempotencyKey",
            setup: func() *auto_conversion_rules.CreateRuleRequest {
                req := FakeAutoConversionRuleRequest()
                req.IdempotencyKey = ""
                return req
            },
            wantErr: true,
        },
        {
            name: "InvalidSourceAsset",
            setup: func() *auto_conversion_rules.CreateRuleRequest {
                req := FakeAutoConversionRuleRequest()
                req.Source.Asset = "INVALID"
                return req
            },
            wantErr: true,
        },
        {
            name: "InvalidSourceNetwork",
            setup: func() *auto_conversion_rules.CreateRuleRequest {
                req := FakeAutoConversionRuleRequest()
                req.Source.Network = "INVALID_NETWORK"
                return req
            },
            wantErr: true,
        },
        {
            name: "MismatchedAssetNetwork",
            setup: func() *auto_conversion_rules.CreateRuleRequest {
                req := FakeAutoConversionRuleRequest()
                // USD with crypto network
                req.Source.Asset = "USD"
                req.Source.Network = "POLYGON"
                return req
            },
            wantErr: true,
        },
        {
            name: "CryptoDestinationWithoutNetwork",
            setup: func() *auto_conversion_rules.CreateRuleRequest {
                req := FakeAutoConversionRuleRequest()
                req.Destination.Asset = "USDC"
                req.Destination.Network = nil
                return req
            },
            wantErr: true,
        },
        {
            name: "SameSourceAndDestination",
            setup: func() *auto_conversion_rules.CreateRuleRequest {
                req := FakeAutoConversionRuleRequest()
                network := "POLYGON"
                req.Source.Asset = "USDC"
                req.Source.Network = "POLYGON"
                req.Destination.Asset = "USDC"
                req.Destination.Network = &network
                return req
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            req := tt.setup()
            resp, err := s.Client.AutoConversionRules.CreateRule(s.Ctx, s.CustomerID, req)

            if tt.wantErr {
                s.Require().Error(err, "expected error for case: %s", tt.name)
                s.T().Logf("[%s] Expected error: %v", tt.name, err)
            } else {
                s.Require().NoError(err, "unexpected error for case: %s", tt.name)
                s.NotEmpty(resp.AutoConversionRuleID)
                s.T().Logf("[%s] Created rule: %s", tt.name, resp.AutoConversionRuleID)
            }
        })
    }
}

func (s *AutoConversionRulesTestSuite) TestAutoConversionRules_List_Pagination() {
    tests := []struct {
        name        string
        page        int
        size        int
        expectEmpty bool
        expectError bool
    }{
        {"DefaultPagination", 0, 0, false, false},
        {"FirstPage", 1, 10, false, false},
        {"SmallPageSize", 1, 1, false, false},
        {"MaxPageSize", 1, 100, false, false},
        {"OverMaxPageSize", 1, 101, false, true}, // or should cap?
        {"LargePageNumber", 9999, 10, true, false},
        {"ZeroPageSize", 1, 0, false, false}, // should use default
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            req := &auto_conversion_rules.ListRulesRequest{
                Page: tt.page,
                Size: tt.size,
            }

            resp, err := s.Client.AutoConversionRules.ListRules(s.Ctx, s.CustomerID, req)

            if tt.expectError {
                s.Require().Error(err)
                return
            }

            s.Require().NoError(err)
            s.Require().NotNil(resp)

            if tt.expectEmpty {
                s.Empty(resp.Items, "expected empty result for large page number")
            }

            s.T().Logf("[%s] Total: %d, Items: %d", tt.name, resp.Total, len(resp.Items))
        })
    }
}
```

## Running and Analyzing Results

### Identify Failures

```bash
# Run tests and capture output
go test -v ./tests/e2e/... -count=1 2>&1 | tee test_output.log

# Find failures
grep -E "(FAIL|Error|panic)" test_output.log
```

### Analyze Error Patterns

When tests fail, categorize:

1. **Expected Failures** - Tests correctly catching API validation
2. **Unexpected Failures** - Bugs in test or API
3. **Flaky Tests** - Timing/ordering issues

### Document Findings

After running edge case tests, document:

```go
// TestAutoConversionRules_Create_EdgeCases_Findings documents discovered behaviors:
//
// 1. Empty idempotency key: Returns 400 "idempotency_key is required"
// 2. Invalid source asset: Returns 400 "invalid asset"
// 3. Mismatched asset/network: Returns 400 "network not supported for asset"
// 4. Same source and destination: Returns 400 "source and destination must differ"
// 5. Over max page size: API caps at 100, does not error
```
