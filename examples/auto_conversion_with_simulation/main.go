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

// Package main demonstrates the lifecycle of an auto conversion rule together
// with a sandbox deposit simulation that triggers the rule.
//
// This example shows how to:
//   - Create a fiat→crypto auto conversion rule (USD → USDC on Polygon)
//   - Wait for the rule to become ACTIVE and deposit info to be ready
//   - Get the reference code from deposit info (required to trigger the rule!)
//   - Simulate a USD deposit with the reference code
//   - Poll for auto conversion orders created by the rule
//
// Key concept: The reference code is essential for triggering auto conversion.
// Without it, deposits go to the customer's balance but won't trigger the rule.
//
// Prerequisites:
//   - Set ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY environment variables
//   - Set ONEMONEY_CUSTOMER_ID to an existing customer ID
//   - Set ONEMONEY_TEST_WALLET_ADDRESS for the destination wallet
//
// Run:
//
//	go run ./examples/auto_conversion_with_simulation
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/auto_conversion_rules"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
)

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

	log.Printf("starting auto conversion rule demo: customer_id=%s", customerID)

	// Step 1: Create a fiat→crypto auto conversion rule (USD ACH → USDC Polygon).
	log.Println("step 1: creating fiat→crypto auto conversion rule")
	ruleID := createFiatToCryptoRule(ctx, client, customerID)
	if ruleID == "" {
		log.Fatal("failed to create auto conversion rule; see logs for details")
	}

	// Step 2: Wait for deposit info to be ready.
	// The deposit info contains the reference code needed to trigger the rule.
	log.Println("step 2: waiting for deposit info (includes reference code)")
	rule, err := auto_conversion_rules.WaitForDepositInfoReady(ctx, client.AutoConversionRules, customerID, ruleID,
		&auto_conversion_rules.WaitOptions{PrintProgress: true, PollInterval: 1 * time.Second, MaxWaitTime: 2 * time.Minute})
	if err != nil {
		log.Fatalf("failed waiting for deposit info: %v", err)
	}

	// Log the deposit info - the reference code is the key!
	referenceCode := logDepositInfo(rule)
	if referenceCode == "" {
		log.Fatal("no reference code available - cannot trigger auto conversion")
	}

	// Step 3: Simulate a USD deposit WITH the reference code.
	// IMPORTANT: The reference code is required to trigger the auto conversion rule.
	// Without it, the deposit would go to the customer's balance but NOT trigger conversion.
	log.Println("step 3: simulating USD deposit with reference code")
	log.Printf("  using reference_code=%s (this links the deposit to the rule)", referenceCode)
	simResp, err := client.Simulations.SimulateDeposit(ctx, customerID, &simulations.SimulateDepositRequest{
		Asset:         assets.AssetNameUSD,
		Network:       "US_ACH",
		Amount:        "50.00",
		ReferenceCode: referenceCode, // Required to trigger auto conversion!
	})
	if err != nil {
		log.Fatalf("failed to simulate deposit: %v", err)
	}
	log.Printf("USD deposit simulated: simulation_id=%s status=%s amount=50.00 USD",
		simResp.SimulationID, simResp.Status)

	// Step 4: Poll for auto conversion orders created by this rule.
	// The rule should automatically create an order when the deposit arrives.
	log.Println("step 4: polling for auto conversion orders (rule should trigger automatically)")
	if err := waitForAutoConversionOrder(ctx, client, customerID, ruleID); err != nil {
		log.Printf("WARNING: no auto conversion orders detected within timeout: %v", err)
	}

	log.Println("")
	log.Println("=== workflow complete ===")
	log.Println("  Rule created → Deposit simulated with reference code → Auto conversion triggered")
}

// createFiatToCryptoRule creates a USD (ACH) → USDC (Polygon) auto conversion rule
// and waits for it to become ACTIVE.
func createFiatToCryptoRule(
	ctx context.Context,
	client *onemoney.Client,
	customerID string,
) string {
	destNetwork := "POLYGON"

	withdrawalWalletAddress := os.Getenv("ONEMONEY_TEST_WALLET_ADDRESS")
	if withdrawalWalletAddress == "" {
		log.Fatalf("missing wallet address: %s", withdrawalWalletAddress)
	}

	req := &auto_conversion_rules.CreateRuleRequest{
		IdempotencyKey: uuid.New().String(),
		Source: auto_conversion_rules.SourceAssetInfo{
			Asset:   "USD",
			Network: "US_ACH",
		},
		Destination: auto_conversion_rules.DestinationAssetInfo{
			Asset:         "USDC",
			Network:       &destNetwork,
			WalletAddress: &withdrawalWalletAddress,
		},
	}

	rule, err := client.AutoConversionRules.CreateRule(ctx, customerID, req)
	if err != nil {
		log.Printf("WARNING: failed to create fiat→crypto auto conversion rule: %v", err)
		return ""
	}

	log.Printf("auto conversion rule created: rule_id=%s status=%s nickname=%s",
		rule.AutoConversionRuleID, rule.Status, rule.Nickname)

	if rule.Status != "ACTIVE" {
		log.Println("waiting for auto conversion rule to become ACTIVE")
		rule, err = auto_conversion_rules.WaitForActive(ctx, client.AutoConversionRules, customerID, rule.AutoConversionRuleID, nil)
		if err != nil {
			log.Printf("WARNING: auto conversion rule activation failed: %v", err)
			return rule.AutoConversionRuleID
		}
		log.Println("auto conversion rule is now ACTIVE")
	}

	return rule.AutoConversionRuleID
}

// logDepositInfo prints the deposit information and returns the reference code (for bank) or wallet address (for crypto).
func logDepositInfo(rule *auto_conversion_rules.RuleResponse) string {
	if rule.SourceDepositInfo == nil {
		log.Printf("rule has no deposit info (yet): rule_id=%s", rule.AutoConversionRuleID)
		return ""
	}

	if rule.SourceDepositInfo.Bank != nil {
		info := rule.SourceDepositInfo.Bank
		log.Printf("deposit info ready (bank):")
		log.Printf("  rule_id=%s", rule.AutoConversionRuleID)
		log.Printf("  reference_code=%s  ← use this in deposits!", info.ReferenceCode)
		log.Printf("  minimum_amount=%s", info.MinimumDepositAmount)
		return info.ReferenceCode
	}

	if rule.SourceDepositInfo.Crypto != nil {
		info := rule.SourceDepositInfo.Crypto
		log.Printf("deposit info ready (crypto):")
		log.Printf("  rule_id=%s", rule.AutoConversionRuleID)
		log.Printf("  wallet_address=%s  ← send crypto here!", info.WalletAddress)
		log.Printf("  minimum_amount=%s", info.MinimumDepositAmount)
		log.Printf("  contract_address=%s", info.ContractAddress)
		return info.WalletAddress
	}

	return ""
}

// waitForAutoConversionOrder polls for orders created by the given rule.
func waitForAutoConversionOrder(
	ctx context.Context,
	client *onemoney.Client,
	customerID, ruleID string,
) error {
	const (
		pollInterval = 1 * time.Second
		maxWaitTime  = 60 * time.Second
	)

	start := time.Now()
	deadline := start.Add(maxWaitTime)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		orders, err := client.AutoConversionRules.ListOrders(ctx, customerID, ruleID, &auto_conversion_rules.ListOrdersRequest{
			Page: 1,
			Size: 10,
		})
		if err != nil {
			return fmt.Errorf("failed to list orders: %w", err)
		}

		log.Printf("polling auto conversion rule_id=%s orders: elapsed=%.1fs count=%d",
			ruleID, time.Since(start).Seconds(), len(orders.Items))

		if len(orders.Items) > 0 {
			log.Printf("auto conversion orders found: rule_id=%s total=%d returned=%d",
				ruleID, orders.Total, len(orders.Items))
			for i := range orders.Items {
				order := &orders.Items[i]
				log.Printf("auto conversion order: order_id=%s status=%s initial_amount=%s initial_asset=%s",
					order.AutoConversionOrderID, order.Status, order.Receipt.Initial.Amount, order.Receipt.Initial.Asset)
			}
			return nil
		}

		time.Sleep(pollInterval)
	}

	return fmt.Errorf("timeout waiting for auto conversion orders after %v", maxWaitTime)
}
