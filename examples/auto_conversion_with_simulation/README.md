# Auto Conversion with Simulation

This example demonstrates how to set up automatic currency conversion rules using the 1Money SDK.

## What It Does

1. **Create Auto Conversion Rule** - Set up a rule to automatically convert USD to USDC
2. **Wait for Rule Activation** - Poll until the rule becomes ACTIVE
3. **Get Deposit Info** - Retrieve bank details including the **reference code**
4. **Simulate USD Deposit** - Trigger the rule with a deposit using the reference code
5. **Poll for Orders** - Watch for auto conversion orders created by the rule

## Business Scenario

Auto conversion rules allow you to automate currency conversion without manual intervention:

```
USD Deposit (with reference code) → Rule Matches → Auto Convert to USDC → Withdraw to Wallet
```

This is useful for:
- Automatically converting incoming fiat payments to stablecoins
- Setting up recurring conversion workflows
- Reducing manual treasury operations

## How Reference Code Works

**Important:** To trigger an auto conversion rule, deposits must include the rule's **reference code**.

1. When you create an auto conversion rule, 1Money provisions unique deposit info
2. For fiat rules (USD/ACH), this includes a **reference code** (e.g., `"ABC123"`)
3. When depositing funds, include this reference code in the wire transfer memo
4. 1Money uses the reference code to match the deposit to the correct rule

In sandbox, you must pass the `ReferenceCode` to `SimulateDeposit()`:

```go
// Get the reference code from the rule's deposit info
referenceCode := rule.SourceDepositInfo.Bank.ReferenceCode

// Simulate deposit WITH the reference code to trigger the rule
client.Simulations.SimulateDeposit(ctx, customerID, &simulations.SimulateDepositRequest{
    Asset:         assets.AssetNameUSD,
    Network:       "US_ACH",
    Amount:        "50.00",
    ReferenceCode: referenceCode,  // Required to trigger auto conversion!
})
```

Without the reference code, the deposit will be credited to the customer's balance but **will NOT trigger** the auto conversion rule.

## SDK Features Demonstrated

- `client.AutoConversionRules.CreateRule()` - Create an auto conversion rule
- `auto_conversion_rules.WaitForActive()` - Wait for rule activation
- `auto_conversion_rules.WaitForDepositInfoReady()` - Wait for deposit info (including reference code)
- `client.Simulations.SimulateDeposit()` - Simulate deposits with reference code
- `client.AutoConversionRules.ListOrders()` - Query orders created by a rule

## Prerequisites

```bash
# Required
ONEMONEY_ACCESS_KEY=your-access-key
ONEMONEY_SECRET_KEY=your-secret-key
ONEMONEY_CUSTOMER_ID=your-customer-id
ONEMONEY_TEST_WALLET_ADDRESS=0x...  # Destination wallet for converted USDC
```

## Run

```bash
go run ./examples/auto_conversion_with_simulation
```

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Auto Conversion Rule                           │
│                   (USD US_ACH → USDC Polygon)                       │
│                                                                     │
│  Deposit Info:                                                      │
│    - Reference Code: "ABC123"  ← Use this in deposits!              │
│    - Minimum Amount: "1.00"                                         │
└─────────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│  USD        │     │  Rule Matches   │     │   USDC      │
│  Deposit    │ ──▶ │  Reference Code │ ──▶ │   Sent to   │
│  (with ref) │     │  → Converts     │     │   Wallet    │
└─────────────┘     └─────────────────┘     └─────────────┘
```

## Rule Configuration

The example creates a rule with:
- **Source**: USD via US_ACH (bank transfer)
- **Destination**: USDC on Polygon network, sent to your wallet address

When a USD deposit arrives with the matching reference code:
1. Rule detects the deposit via reference code match
2. Creates an auto conversion order
3. Converts USD to USDC at market rate
4. Withdraws USDC to the configured wallet address

## Notes

- Auto conversion rules require deposit info to be provisioned before they can receive funds
- **The reference code is essential** - deposits without it won't trigger auto conversion
- For crypto→crypto rules, the rule provides a unique wallet address instead of a reference code
- In production, instruct your customers/partners to include the reference code in wire memos
- Rules can be paused, resumed, or deleted as needed
