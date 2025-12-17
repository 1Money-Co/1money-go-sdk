# USDC to Fiat Withdrawal

This example demonstrates a complete crypto-to-fiat on-ramp workflow using the 1Money SDK.

## What It Does

1. **Simulate USDC Deposit** - Add crypto funds to the account (sandbox only)
2. **Check Balances** - View available asset balances
3. **Convert USDC → USD** - Create a quote and execute the conversion
4. **Create External Bank Account** - Register a bank account for withdrawal
5. **Withdraw USD** - Send USD to the external bank account via ACH
6. **List Transactions** - View the transaction history

## Business Scenario

A common use case for businesses that need to convert crypto to fiat:

```
USDC (Crypto Deposit) → Convert to USD → Withdraw to Bank
```

This flow is useful for:
- Liquidating crypto revenue
- Paying bills and salaries in fiat
- Treasury management
- Off-ramping DeFi earnings

## SDK Features Demonstrated

- `client.Simulations.SimulateDeposit()` - Simulate crypto deposits in sandbox
- `client.Assets.ListAssets()` - Query account balances
- `client.Conversions.CreateQuote()` - Get conversion rates
- `client.Conversions.CreateHedge()` - Execute conversions
- `client.ExternalAccounts.CreateExternalAccount()` - Register bank accounts
- `external_accounts.WaitForApproved()` - Wait for bank account approval
- `client.Withdrawals.CreateWithdrawal()` - Withdraw to bank accounts
- `transactions.WaitForSettled()` - Wait for withdrawal settlement

## Prerequisites

```bash
# Required
ONEMONEY_ACCESS_KEY=your-access-key
ONEMONEY_SECRET_KEY=your-secret-key
ONEMONEY_CUSTOMER_ID=your-customer-id
```

## Run

```bash
go run ./examples/usdc_to_fiat_withdrawal
```

## Flow Diagram

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Deposit    │     │   Convert   │     │   Create    │     │  Withdraw   │
│    USDC     │ ──▶ │  USDC→USD   │ ──▶ │   Bank Acct │ ──▶ │     USD     │
│  (Crypto)   │     │             │     │             │     │   (Fiat)    │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
```

## Notes

- External bank accounts require approval before they can receive withdrawals
- In sandbox, bank accounts are auto-approved
- Fiat withdrawals use ACH (US) network and may take 1-3 business days to settle
- The example waits for settlement to demonstrate the full lifecycle
