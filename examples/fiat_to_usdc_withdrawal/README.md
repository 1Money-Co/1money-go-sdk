# Fiat to USDC Withdrawal

This example demonstrates a complete fiat-to-crypto off-ramp workflow using the 1Money SDK.

## What It Does

1. **Simulate USD Deposit** - Add fiat funds to the account (sandbox only)
2. **Check Balances** - View available asset balances
3. **Convert USD → USDC** - Create a quote and execute the conversion
4. **Withdraw USDC** - Send USDC to an external wallet on Polygon
5. **List Transactions** - View the transaction history

## Business Scenario

A common use case for businesses that need to convert fiat revenue to stablecoins:

```
USD (Bank Deposit) → Convert to USDC → Withdraw to Wallet
```

This flow is useful for:
- Treasury management (holding stablecoins)
- Cross-border payments
- DeFi integrations
- Payroll in crypto

## SDK Features Demonstrated

- `client.Simulations.SimulateDeposit()` - Simulate deposits in sandbox
- `client.Assets.ListAssets()` - Query account balances
- `client.Conversions.CreateQuote()` - Get conversion rates
- `client.Conversions.CreateHedge()` - Execute conversions
- `client.Withdrawals.CreateWithdrawal()` - Withdraw to external wallets
- `client.Transactions.ListTransactions()` - Query transaction history

## Prerequisites

```bash
# Required
ONEMONEY_ACCESS_KEY=your-access-key
ONEMONEY_SECRET_KEY=your-secret-key
ONEMONEY_CUSTOMER_ID=your-customer-id
ONEMONEY_TEST_WALLET_ADDRESS=0x...  # Polygon wallet address
```

## Run

```bash
go run ./examples/fiat_to_usdc_withdrawal
```

## Flow Diagram

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Deposit    │     │   Convert   │     │  Withdraw   │
│    USD      │ ──▶ │  USD→USDC   │ ──▶ │    USDC     │
│  (Fiat)     │     │             │     │  (Crypto)   │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Notes

- Conversions use a quote-then-execute pattern for price transparency
- The quote includes the exchange rate and fee breakdown
- Withdrawals require an idempotency key to prevent double-spending
