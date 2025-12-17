# Create Customer

This example demonstrates how to onboard a new business customer using the 1Money SDK.

## What It Does

1. **Create TOS Link** - Generate a Terms of Service agreement link
2. **Sign TOS Agreement** - Programmatically sign the agreement (for sandbox testing)
3. **Create Customer** - Submit business KYB (Know Your Business) information
4. **Wait for KYB Approval** - Poll until the customer is approved (auto-approved in sandbox)
5. **Wait for Fiat Account** - Wait for the fiat account to be provisioned

## Business Scenario

Before a business can use 1Money services, they must complete customer onboarding:

```
TOS Agreement → KYB Submission → KYB Review → Account Provisioning → Ready
```

## SDK Features Demonstrated

- `client.Customer.CreateTOSLink()` - Generate TOS agreement URL
- `client.Customer.SignTOSAgreement()` - Sign the agreement
- `client.Customer.CreateCustomer()` - Submit customer with KYB data
- `customer.WaitForKybApproved()` - Helper to poll for KYB approval
- `customer.WaitForFaitAccount()` - Helper to wait for account provisioning

## Prerequisites

```bash
# Required
ONEMONEY_ACCESS_KEY=your-access-key
ONEMONEY_SECRET_KEY=your-secret-key
```

## Run

```bash
go run ./examples/create_customer
```

## Output

After successful execution, you'll receive a `customer_id` that you can use with other examples:

```bash
export ONEMONEY_CUSTOMER_ID=<customer_id_from_output>
```

## Notes

- This example uses test data from `pkg/testdata` for ID documents and proof of address
- In production, you would collect real business information and documents
- The customer structure includes associated persons (beneficial owners, directors, signers)
