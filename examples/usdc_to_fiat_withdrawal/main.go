/*
 * Copyright 2025 1Money Co.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package main demonstrates a complete crypto-to-fiat withdrawal workflow.
//
// This example shows a common business scenario:
//  1. Receive USDC stablecoin - simulated deposit in sandbox
//  2. Convert USDC to USD fiat currency
//  3. Create an external bank account
//  4. Withdraw USD to the external bank account
//
// Prerequisites:
//   - Set ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY environment variables
//   - Set ONEMONEY_CUSTOMER_ID environment variable (from create_customer example)
//
// Run: go run ./examples/usdc_to_fiat_withdrawal
package main

import (
	"context"
	"log"
	"os"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/1Money-Co/1money-go-sdk/pkg/common"
	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/conversions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/transactions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/withdraws"
)

//nolint:funlen,gocyclo // Example code intentionally shows complete workflow in a single function for readability
func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	customerID := os.Getenv("ONEMONEY_CUSTOMER_ID")
	if customerID == "" {
		log.Fatal("ONEMONEY_CUSTOMER_ID environment variable is required")
	}

	client, err := onemoney.NewClient(&onemoney.Config{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Step 1: Simulate USDC crypto deposit (sandbox only)
	log.Println("step 1: simulating USDC deposit on Polygon")
	depositResp, err := client.Simulations.SimulateDeposit(ctx, customerID, &simulations.SimulateDepositRequest{
		Asset:   assets.AssetNameUSDC,
		Network: simulations.WalletNetworkNamePOLYGON,
		Amount:  "200.00",
	})
	if err != nil {
		log.Fatalf("failed to simulate deposit: %v", err)
	}
	log.Printf("USDC deposit initiated: simulation_id=%s amount=200.00 USDC", depositResp.SimulationID)

	// Step 2: Check balance
	log.Println("step 2: checking balances")
	balances, err := client.Assets.ListAssets(ctx, customerID, nil)
	if err != nil {
		log.Fatalf("failed to list assets: %v", err)
	}
	for _, b := range balances {
		if b.AvailableAmount != "0" {
			log.Printf("balance: asset=%s available=%s", b.Asset, b.AvailableAmount)
		}
	}

	// Step 3: Convert USDC to USD
	log.Println("step 3: converting USDC (Polygon) to USD")

	// 3a. Create quote
	quote, err := client.Conversions.CreateQuote(ctx, customerID, &conversions.CreateQuoteRequest{
		FromAsset: conversions.AssetInfo{
			Asset:   assets.AssetNameUSDC,
			Network: conversions.WalletNetworkNamePOLYGON,
			Amount:  "100.00",
		},
		ToAsset: conversions.AssetInfo{
			Asset: assets.AssetNameUSD,
		},
	})
	if err != nil {
		log.Fatalf("failed to create quote: %v", err)
	}
	log.Printf("quote created: quote_id=%s rate=%s pay=%s %s receive=%s %s",
		quote.QuoteID, quote.Rate, quote.UserPayAmount, quote.UserPayAsset,
		quote.UserObtainAmount, quote.UserObtainAsset)

	// 3b. Execute conversion
	hedge, err := client.Conversions.CreateHedge(ctx, customerID, &conversions.CreateHedgeRequest{
		QuoteID: quote.QuoteID,
	})
	if err != nil {
		log.Fatalf("failed to execute conversion: %v", err)
	}
	log.Printf("conversion executed: order_id=%s status=%s", hedge.OrderID, hedge.OrderStatus)

	// Step 4: Create external bank account for fiat withdrawal
	log.Println("step 4: creating external bank account")
	externalAccount, err := client.ExternalAccounts.CreateExternalAccount(ctx, customerID, &external_accounts.CreateReq{
		IdempotencyKey:  uuid.New().String(),
		Network:         common.BankNetworkNameUSACH,
		Currency:        common.CurrencyUSD,
		CountryCode:     common.CountryCodeUSA,
		AccountNumber:   "5097935393",
		InstitutionID:   "327984566",
		InstitutionName: gofakeit.Company() + " Bank",
	})
	if err != nil {
		log.Fatalf("failed to create external account: %v", err)
	}
	log.Printf("external account created: external_account_id=%s status=%s",
		externalAccount.ExternalAccountID, externalAccount.Status)

	// Wait for external account approval (usually instant in sandbox)
	if externalAccount.Status != string(common.BankAccountStatusAPPROVED) {
		log.Println("waiting for external account approval...")
		externalAccount, err = external_accounts.WaitForApproved(
			ctx, client.ExternalAccounts, customerID, externalAccount.ExternalAccountID,
			&external_accounts.WaitOptions{PrintProgress: true},
		)
		if err != nil {
			log.Fatalf("external account approval failed: %v", err)
		}
		log.Printf("external account approved: status=%s", externalAccount.Status)
	}

	// Step 5: Withdraw USD to external bank account
	log.Println("step 5: withdrawing USD to external bank account")
	withdrawal, err := client.Withdrawals.CreateWithdrawal(ctx, customerID, &withdraws.CreateWithdrawalRequest{
		IdempotencyKey:    uuid.New().String(),
		Amount:            "50.00",
		Asset:             assets.AssetNameUSD,
		Network:           assets.NetworkNameUSACH,
		ExternalAccountID: externalAccount.ExternalAccountID,
	})
	if err != nil {
		log.Fatalf("failed to create withdrawal: %v", err)
	}
	log.Printf("withdrawal submitted: transaction_id=%s status=%s amount=%s USD",
		withdrawal.TransactionID, withdrawal.Status, withdrawal.Amount)

	// Wait for withdrawal to settle (PENDING means ACH transfer is in progress)
	// Note: In production, ACH transfers typically take 1-3 business days
	if withdrawal.Status == string(transactions.TransactionStatusPENDING) {
		log.Println("withdrawal is processing (ACH transfer in progress)...")
		var tx *transactions.TransactionResponse
		tx, err = transactions.WaitForSettled(ctx, client.Transactions, customerID, withdrawal.TransactionID,
			&transactions.WaitOptions{PrintProgress: true})
		if err != nil {
			log.Fatalf("withdrawal failed: %v", err)
		}
		log.Printf("withdrawal completed: status=%s", tx.Status)
	}

	// Step 6: Show updated balances
	log.Println("step 6: final balances")
	balances, _ = client.Assets.ListAssets(ctx, customerID, nil)
	for _, b := range balances {
		if b.AvailableAmount != "0" {
			log.Printf("balance: asset=%s available=%s", b.Asset, b.AvailableAmount)
		}
	}

	// Step 7: List recent transactions to show the workflow history
	log.Println("step 7: listing recent transactions")
	txResp, err := client.Transactions.ListTransactions(ctx, customerID, &transactions.ListTransactionsRequest{
		Size: 10,
	})
	if err != nil {
		log.Fatalf("failed to list transactions: %v", err)
	}
	log.Printf("transaction history: total=%d", txResp.Total)
	for i := range txResp.List {
		tx := &txResp.List[i]
		log.Printf("  - %s: %s %s %s (%s) [%s]",
			tx.TransactionAction, tx.Amount, tx.Asset, tx.Network, tx.TransactionID[:8], tx.Status)
	}

	log.Println("")
	log.Println("=== workflow complete ===")
	log.Println("  USDC deposited → converted to USD → withdrawn to bank account")
}
