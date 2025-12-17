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

// Package main demonstrates a complete fiat-to-crypto withdrawal workflow.
//
// This example shows a common business scenario:
//  1. Receive fiat currency (USD) - simulated deposit in sandbox
//  2. Convert USD to USDC stablecoin
//  3. Withdraw USDC to an external wallet
//
// Prerequisites:
//   - Set ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY environment variables
//   - Set CUSTOMER_ID environment variable (from create_customer example)
//   - Optionally set WALLET_ADDRESS for the destination wallet
//
// Run: go run ./examples/fiat_to_usdc_withdrawal
package main

import (
	"context"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/conversions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/transactions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/withdraws"
)

//nolint:funlen // Example code intentionally shows complete workflow in a single function for readability
func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	customerID := os.Getenv("ONEMONEY_CUSTOMER_ID")
	if customerID == "" {
		log.Fatal("ONEMONEY_CUSTOMER_ID environment variable is required")
	}

	withdrawalWalletAddress := os.Getenv("ONEMONEY_TEST_WALLET_ADDRESS")
	if withdrawalWalletAddress == "" {
		log.Fatalf("missing wallet address: %s", withdrawalWalletAddress)
	}

	client, err := onemoney.NewClient(&onemoney.Config{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Step 1: Simulate USD fiat deposit (sandbox only)
	log.Println("step 1: simulating USD deposit")
	depositResp, err := client.Simulations.SimulateDeposit(ctx, customerID, &simulations.SimulateDepositRequest{
		Asset:   assets.AssetNameUSD,
		Amount:  "100.00",
		Network: simulations.WalletNetworkNameUSACH,
	})
	if err != nil {
		log.Fatalf("failed to simulate deposit: %v", err)
	}
	log.Printf("USD deposit initiated: simulation_id=%s amount=100.00 USD", depositResp.SimulationID)

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

	// Step 3: Convert USD to USDC
	log.Println("step 3: converting USD to USDC (Polygon)")

	// 3a. Create quote
	quote, err := client.Conversions.CreateQuote(ctx, customerID, &conversions.CreateQuoteRequest{
		FromAsset: conversions.AssetInfo{
			Asset:  assets.AssetNameUSD,
			Amount: "50.00",
		},
		ToAsset: conversions.AssetInfo{
			Asset:   assets.AssetNameUSDC,
			Network: conversions.WalletNetworkNamePOLYGON,
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

	// Step 4: Withdraw USDC to external wallet
	log.Println("step 4: withdrawing USDC to external wallet")
	withdrawal, err := client.Withdrawals.CreateWithdrawal(ctx, customerID, &withdraws.CreateWithdrawalRequest{
		IdempotencyKey: uuid.New().String(),
		Amount:         "49.00",
		Asset:          assets.AssetNameUSDC,
		Network:        assets.NetworkNamePOLYGON,
		WalletAddress:  withdrawalWalletAddress,
	})
	if err != nil {
		log.Fatalf("failed to create withdrawal: %v", err)
	}
	log.Printf("withdrawal submitted: transaction_id=%s status=%s amount=%s USDC",
		withdrawal.TransactionID, withdrawal.Status, withdrawal.Amount)
	log.Println("note: crypto withdrawals require on-chain confirmation and may remain in PENDING status for several minutes;")
	log.Println("      this example does not wait for final on-chain settlement. You can poll the Transactions API to track status.")

	// Final: Show updated balances
	log.Println("step 5: final balances")
	balances, _ = client.Assets.ListAssets(ctx, customerID, nil)
	for _, b := range balances {
		if b.AvailableAmount != "0" {
			log.Printf("balance: asset=%s available=%s", b.Asset, b.AvailableAmount)
		}
	}

	// Step 6: List recent transactions to show the workflow history
	log.Println("step 6: listing recent transactions")
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
	log.Println("  USD deposited → converted to USDC → withdrawn to external wallet")
}
