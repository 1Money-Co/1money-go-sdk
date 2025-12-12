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

// Package main demonstrates a complete end-to-end workflow using the 1Money SDK.
//
// This example shows how to:
//  1. Check or create a customer
//  2. View asset balances
//  3. Get deposit instructions (fiat and crypto)
//  4. Simulate deposits (sandbox only)
//  5. Create external bank accounts
//  6. Manage auto conversion rules (create, list, get, delete)
//  7. Perform manual asset conversions (crypto↔fiat)
//  8. Create withdrawals (fiat and crypto)
//  9. View transaction history with filtering
//
// This is a comprehensive example that demonstrates the full lifecycle
// of a customer's interaction with the 1Money platform.
package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/auto_conversion_rules"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/conversions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/instructions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/transactions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/withdraws"
)

var logger *zap.Logger

func main() {
	// Initialize logger
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer func() { _ = logger.Sync() }()

	_ = godotenv.Load()

	// Setup context with signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	startWebhookServer()

	client, err := onemoney.NewClient(&onemoney.Config{})
	if err != nil {
		logger.Fatal("failed to create client", zap.Error(err))
	}

	// Phase 1: Customer Setup
	checkContext(ctx)
	logSection("Phase 1: Customer Setup")
	customerID := getOrCreateCustomer(ctx, client)

	// Phase 2: View Initial State
	checkContext(ctx)
	logSection("Phase 2: Initial Asset Balances")
	viewAssetBalances(ctx, client, customerID)

	// Phase 3: Deposit Instructions
	checkContext(ctx)
	logSection("Phase 3: Deposit Instructions")
	getDepositInstructions(ctx, client, customerID)

	// Phase 4: Simulate Deposits (Sandbox)
	checkContext(ctx)
	logSection("Phase 4: Simulate Deposits (Sandbox)")
	simulateDeposits(ctx, client, customerID)

	// Phase 5: View Updated Balances
	checkContext(ctx)
	logSection("Phase 5: Updated Asset Balances")
	viewAssetBalances(ctx, client, customerID)

	// Phase 6: External Accounts
	checkContext(ctx)
	logSection("Phase 6: External Bank Accounts")
	externalAccountID := createExternalAccount(ctx, client, customerID)

	// Phase 7: Auto Conversion Rules
	checkContext(ctx)
	logSection("Phase 7: Auto Conversion Rules")
	manageAutoConversionRules(ctx, client, customerID, externalAccountID)

	// Phase 8: Manual Conversions
	checkContext(ctx)
	logSection("Phase 8: Manual Conversions")
	performConversions(ctx, client, customerID)

	// Phase 9: View Balances After Conversion
	checkContext(ctx)
	logSection("Phase 9: Balances After Conversion")
	viewAssetBalances(ctx, client, customerID)

	// Phase 10: Withdrawals
	checkContext(ctx)
	logSection("Phase 10: Withdrawals")
	performWithdrawals(ctx, client, customerID, externalAccountID)

	// Phase 11: Transaction History
	checkContext(ctx)
	logSection("Phase 11: Transaction History")
	viewTransactionHistory(ctx, client, customerID)

	// Summary
	logSection("Workflow Complete")
	logger.Info("all phases completed successfully",
		zap.String("customer_id", customerID),
	)
}

// logSection prints a clear section header for better readability.
func logSection(name string) {
	logger.Info("════════════════════════════════════════")
	logger.Info(name)
	logger.Info("════════════════════════════════════════")
}

// checkContext checks if the context has been cancelled and exits if so.
func checkContext(ctx context.Context) {
	select {
	case <-ctx.Done():
		logger.Info("workflow interrupted, exiting gracefully")
		os.Exit(0)
	default:
	}
}

const imageSize = 100

// generateSampleImage generates a valid PNG image for testing purposes.
// In production, you should use real document images.
func generateSampleImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
	// Fill with a light gray color
	c := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	for y := range imageSize {
		for x := range imageSize {
			img.Set(x, y, c)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(fmt.Sprintf("failed to encode PNG: %v", err))
	}
	return buf.Bytes()
}

func getOrCreateCustomer(ctx context.Context, client *onemoney.Client) string {
	// Try to use existing customer ID from environment
	customerID := os.Getenv("EXAMPLE_CUSTOMER_ID")
	if customerID != "" {
		// Verify customer exists
		_, err := client.Customer.GetCustomer(ctx, customerID)
		if err == nil {
			logger.Info("using existing customer", zap.String("customer_id", customerID))
			return customerID
		}
		logger.Warn("customer not found, will create new one", zap.String("customer_id", customerID))
	}

	// Create new customer (simplified - see create_customer example for full details)
	logger.Info("creating new customer")
	tosResp, err := client.Customer.CreateTOSLink(ctx, &customer.CreateTOSLinkRequest{
		RedirectUri: "https://example.com/tos-completed",
	})
	if err != nil {
		logger.Fatal("failed to create TOS link", zap.Error(err))
	}

	signResp, err := client.Customer.SignTOSAgreement(ctx, tosResp.SessionToken)
	if err != nil {
		logger.Fatal("failed to sign TOS agreement", zap.Error(err))
	}

	// Note: In production, you would provide real business information and documents
	req := &customer.CreateCustomerRequest{
		BusinessLegalName:          "Example Corp",
		BusinessDescription:        "Example business for SDK demonstration",
		BusinessRegistrationNumber: fmt.Sprintf("REG-%d", time.Now().Unix()),
		Email:                      fmt.Sprintf("example-%d@example.com", time.Now().Unix()),
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           "332999",
		RegisteredAddress: &customer.Address{
			StreetLine1: "123 Example St",
			City:        "San Francisco",
			State:       "CA",
			Country:     "USA",
			PostalCode:  "94102",
			Subdivision: "CA",
		},
		DateOfIncorporation: "2020-01-15",
		SignedAgreementID:   signResp.SignedAgreementID,
		AssociatedPersons: []customer.AssociatedPerson{
			{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Gender:    customer.GenderMale,
				ResidentialAddress: &customer.Address{
					StreetLine1: "456 Residential St",
					City:        "San Francisco",
					State:       "CA",
					Country:     "USA",
					PostalCode:  "94103",
					Subdivision: "CA",
				},
				BirthDate:           "1985-06-15",
				CountryOfBirth:      "USA",
				PrimaryNationality:  "USA",
				HasOwnership:        true,
				OwnershipPercentage: 100,
				HasControl:          true,
				IsSigner:            true,
				IsDirector:          true,
				IdentifyingInformation: []customer.IdentifyingInformation{
					{
						Type:                   customer.IDTypeDriversLicense,
						IssuingCountry:         "USA",
						ImageFront:             customer.EncodeBase64ToDataURI(generateSampleImage(), customer.ImageFormatPng),
						ImageBack:              customer.EncodeBase64ToDataURI(generateSampleImage(), customer.ImageFormatPng),
						NationalIdentityNumber: "D1234567",
					},
				},
				CountryOfTax: "USA",
				TaxType:      customer.TaxIDTypeSSN,
				TaxID:        "123-45-6789",
				POA:          customer.EncodeBase64ToDataURI(generateSampleImage(), customer.ImageFormatPng),
				POAType:      "utility_bill",
			},
		},
		SourceOfFunds:  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth: []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},
		// Required documents for Corporation in US region
		// Note: In production, use real document files
		Documents: []customer.Document{
			{
				DocType:     customer.DocumentTypeFlowOfFunds,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(), customer.FileFormatPng),
				Description: "Proof of Funds",
			},
			{
				DocType:     customer.DocumentTypeRegistrationDocument,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(), customer.FileFormatPng),
				Description: "Certificate of Incorporation",
			},
			{
				DocType:     customer.DocumentTypeProofOfTaxIdentification,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(), customer.FileFormatPng),
				Description: "W9 Form",
			},
			{
				DocType:     customer.DocumentTypeShareholderRegister,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(), customer.FileFormatPng),
				Description: "Ownership Structure",
			},
			{
				DocType:     customer.DocumentTypeESignatureCertificate,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(), customer.FileFormatPng),
				Description: "Authorized Representative List",
			},
			{
				DocType:     customer.DocumentTypeEvidenceOfGoodStanding,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(), customer.FileFormatPng),
				Description: "Evidence of Good Standing",
			},
			{
				DocType:     customer.DocumentTypeProofOfAddress,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(), customer.FileFormatPng),
				Description: "Proof of Address",
			},
		},
		AccountPurpose:                 customer.AccountPurposeTreasuryManagement,
		EstimatedAnnualRevenueUSD:      customer.MoneyRange099999,
		ExpectedMonthlyFiatDeposits:    customer.MoneyRange099999,
		ExpectedMonthlyFiatWithdrawals: customer.MoneyRange099999,
		TaxID:                          "12-3456789",
		TaxType:                        customer.TaxIDTypeEIN,
		TaxCountry:                     "USA",
	}

	resp, err := client.Customer.CreateCustomer(ctx, req)
	if err != nil {
		logger.Fatal("failed to create customer", zap.Error(err))
	}

	logger.Info("customer created",
		zap.String("customer_id", resp.CustomerID),
		zap.String("status", string(resp.Status)),
	)
	return resp.CustomerID
}

func viewAssetBalances(ctx context.Context, client *onemoney.Client, customerID string) {
	assetList, err := client.Assets.ListAssets(ctx, customerID, nil)
	if err != nil {
		logger.Warn("failed to list assets", zap.Error(err))
		return
	}

	if len(assetList) == 0 {
		logger.Info("no assets found for this customer")
		return
	}

	logger.Info("found assets", zap.Int("count", len(assetList)))
	for _, a := range assetList {
		network := "<fiat>"
		if a.Network != nil {
			network = *a.Network
		}
		logger.Info("asset",
			zap.String("asset", a.Asset),
			zap.String("network", network),
			zap.String("available", a.AvailableAmount),
			zap.String("unavailable", a.UnavailableAmount),
		)
	}
}

// getDepositInstructions demonstrates getting deposit instructions for various asset/network combinations.
func getDepositInstructions(ctx context.Context, client *onemoney.Client, customerID string) {
	// Define instructions to fetch - covers fiat and crypto networks
	instructionTypes := []struct {
		asset   assets.AssetName
		network assets.NetworkName
		label   string
	}{
		// Fiat deposit instructions
		{assets.AssetNameUSD, assets.NetworkNameUSACH, "USD via ACH"},
		{assets.AssetNameUSD, assets.NetworkNameUSFEDWIRE, "USD via Fedwire"},
		// Crypto deposit instructions
		{assets.AssetNameUSDT, assets.NetworkNameETHEREUM, "USDT on Ethereum"},
		{assets.AssetNameUSDC, assets.NetworkNamePOLYGON, "USDC on Polygon"},
		{assets.AssetNameUSDC, assets.NetworkNameETHEREUM, "USDC on Ethereum"},
	}

	for _, instr := range instructionTypes {
		logger.Info("getting deposit instruction", zap.String("type", instr.label))
		resp, err := client.Instructions.GetDepositInstruction(ctx, customerID, instr.asset, instr.network)
		if err != nil {
			logger.Warn("failed to get instruction (may require verified account)",
				zap.String("type", instr.label),
				zap.Error(err),
			)
			continue
		}
		logInstruction(instr.label, resp)
	}
}

func logInstruction(label string, instr *instructions.InstructionResponse) {
	fields := []zap.Field{
		zap.String("label", label),
		zap.String("asset", instr.Asset),
		zap.String("network", instr.Network),
	}
	if instr.BankInstruction != nil {
		fields = append(fields,
			zap.String("bank_name", instr.BankInstruction.BankName),
			zap.String("account_number", instr.BankInstruction.AccountNumber),
		)
	}
	if instr.WalletInstruction != nil {
		fields = append(fields, zap.String("wallet_address", instr.WalletInstruction.WalletAddress))
	}
	logger.Info("deposit instruction", fields...)
}

// simulateDeposits demonstrates simulating deposits for various assets (sandbox only).
func simulateDeposits(ctx context.Context, client *onemoney.Client, customerID string) {
	// Define deposits to simulate - covers fiat and multiple crypto networks
	deposits := []struct {
		asset   assets.AssetName
		network simulations.WalletNetworkName
		amount  string
		label   string
	}{
		// Fiat deposit
		{assets.AssetNameUSD, "", "500.00", "USD Fiat"},
		// Crypto deposits on various networks
		{assets.AssetNameUSDT, simulations.WalletNetworkNameETHEREUM, "100.00", "USDT on Ethereum"},
		{assets.AssetNameUSDC, simulations.WalletNetworkNamePOLYGON, "200.00", "USDC on Polygon"},
		{assets.AssetNameUSDC, simulations.WalletNetworkNameETHEREUM, "100.00", "USDC on Ethereum"},
	}

	for _, dep := range deposits {
		logger.Info("simulating deposit",
			zap.String("type", dep.label),
			zap.String("amount", dep.amount),
		)

		req := &simulations.SimulateDepositRequest{
			Asset:  dep.asset,
			Amount: dep.amount,
		}
		if dep.network != "" {
			req.Network = dep.network
		}

		resp, err := client.Simulations.SimulateDeposit(ctx, customerID, req)
		if err != nil {
			logger.Warn("failed to simulate deposit",
				zap.String("type", dep.label),
				zap.Error(err),
			)
			continue
		}

		logger.Info("deposit simulated",
			zap.String("type", dep.label),
			zap.String("simulation_id", resp.SimulationID),
			zap.String("status", resp.Status),
		)
	}
}

func createExternalAccount(ctx context.Context, client *onemoney.Client, customerID string) string {
	const (
		pollInterval = 2 * time.Second
		maxWaitTime  = 10 * time.Second
	)

	idempotencyKey := uuid.New().String()
	createReq := &external_accounts.CreateReq{
		IdempotencyKey:  idempotencyKey,
		Network:         external_accounts.BankNetworkNameUSACH,
		Currency:        external_accounts.CurrencyUSD,
		CountryCode:     external_accounts.CountryCodeUSA,
		AccountNumber:   "79308791878",
		InstitutionID:   "072704185",
		InstitutionName: "Example Bank",
	}

	logger.Info("creating external bank account")
	created, err := client.ExternalAccounts.CreateExternalAccount(ctx, customerID, createReq)
	if err != nil {
		logger.Fatal("failed to create external account", zap.Error(err))
	}

	logger.Info("external account created",
		zap.String("external_account_id", created.ExternalAccountID),
		zap.String("status", created.Status),
	)

	// Poll until approved or failed (required before withdrawals can be made)
	if created.Status != "APPROVED" {
		logger.Info("waiting for external account approval")
		deadline := time.Now().Add(maxWaitTime)

		for time.Now().Before(deadline) {
			acc, err := client.ExternalAccounts.GetExternalAccount(ctx, customerID, created.ExternalAccountID)
			if err != nil {
				logger.Fatal("failed to get external account status", zap.Error(err))
			}

			logger.Debug("polling external account",
				zap.String("external_account_id", created.ExternalAccountID),
				zap.String("status", acc.Status),
			)

			switch acc.Status {
			case "APPROVED":
				logger.Info("external account approved")
				return created.ExternalAccountID
			case "FAILED":
				logger.Fatal("external account approval failed")
			}

			time.Sleep(pollInterval)
		}

		logger.Fatal("external account approval timed out", zap.Duration("timeout", maxWaitTime))
	}

	return created.ExternalAccountID
}

// manageAutoConversionRules demonstrates the full lifecycle of auto conversion rules.
func manageAutoConversionRules(ctx context.Context, client *onemoney.Client, customerID, externalAccountID string) {
	// 1. Create fiat→crypto rule (USD ACH → USDC Polygon)
	logger.Info("creating fiat→crypto auto conversion rule")
	rule1ID := createFiatToCryptoRule(ctx, client, customerID)

	// 2. Create crypto→fiat rule (USDC Polygon → USD with external account withdrawal)
	logger.Info("creating crypto→fiat auto conversion rule")
	rule2ID := createCryptoToFiatRule(ctx, client, customerID, externalAccountID)

	// 3. List all rules
	listAutoConversionRules(ctx, client, customerID)

	// 4. Get rule details (shows deposit info)
	if rule1ID != "" {
		getAutoConversionRuleDetails(ctx, client, customerID, rule1ID)
	}

	// 5. List orders for a rule (execution history)
	if rule1ID != "" {
		listAutoConversionOrders(ctx, client, customerID, rule1ID)
	}

	// 6. Delete a rule (soft delete → INACTIVE)
	if rule2ID != "" {
		deleteAutoConversionRule(ctx, client, customerID, rule2ID)
	}
}

func createFiatToCryptoRule(ctx context.Context, client *onemoney.Client, customerID string) string {
	destNetwork := "POLYGON"
	req := &auto_conversion_rules.CreateRuleRequest{
		IdempotencyKey: uuid.New().String(),
		Source: auto_conversion_rules.SourceAssetInfo{
			Asset:   "USD",
			Network: "ACH", // Use "ACH" instead of "US_ACH" for auto conversion rules
		},
		Destination: auto_conversion_rules.DestinationAssetInfo{
			Asset:   "USDC",
			Network: &destNetwork,
		},
	}

	created, err := client.AutoConversionRules.CreateRule(ctx, customerID, req)
	if err != nil {
		logger.Warn("failed to create fiat→crypto rule (may require verified fiat account)",
			zap.Error(err),
		)
		return ""
	}

	logger.Info("fiat→crypto rule created",
		zap.String("rule_id", created.AutoConversionRuleID),
		zap.String("nickname", created.Nickname),
		zap.String("status", created.Status),
	)
	return created.AutoConversionRuleID
}

func createCryptoToFiatRule(ctx context.Context, client *onemoney.Client, customerID, externalAccountID string) string {
	if externalAccountID == "" {
		logger.Warn("skipping crypto→fiat rule creation (no external account)")
		return ""
	}

	req := &auto_conversion_rules.CreateRuleRequest{
		IdempotencyKey: uuid.New().String(),
		Source: auto_conversion_rules.SourceAssetInfo{
			Asset:   "USDC",
			Network: "POLYGON",
		},
		Destination: auto_conversion_rules.DestinationAssetInfo{
			Asset:             "USD",
			ExternalAccountID: &externalAccountID, // Auto-withdraw to external account
		},
	}

	created, err := client.AutoConversionRules.CreateRule(ctx, customerID, req)
	if err != nil {
		logger.Warn("failed to create crypto→fiat rule",
			zap.Error(err),
		)
		return ""
	}

	logger.Info("crypto→fiat rule created",
		zap.String("rule_id", created.AutoConversionRuleID),
		zap.String("nickname", created.Nickname),
		zap.String("status", created.Status),
		zap.String("external_account_id", externalAccountID),
	)
	return created.AutoConversionRuleID
}

func listAutoConversionRules(ctx context.Context, client *onemoney.Client, customerID string) {
	resp, err := client.AutoConversionRules.ListRules(ctx, customerID, &auto_conversion_rules.ListRulesRequest{
		Page: 1,
		Size: 10,
	})
	if err != nil {
		logger.Warn("failed to list auto conversion rules", zap.Error(err))
		return
	}

	logger.Info("auto conversion rules",
		zap.Int64("total", resp.Total),
		zap.Int("returned", len(resp.Items)),
	)

	for _, rule := range resp.Items {
		logger.Info("rule",
			zap.String("rule_id", rule.AutoConversionRuleID),
			zap.String("nickname", rule.Nickname),
			zap.String("status", rule.Status),
			zap.String("source", fmt.Sprintf("%s/%s", rule.Source.Asset, rule.Source.Network)),
		)
	}
}

func getAutoConversionRuleDetails(ctx context.Context, client *onemoney.Client, customerID, ruleID string) {
	rule, err := client.AutoConversionRules.GetRule(ctx, customerID, ruleID)
	if err != nil {
		logger.Warn("failed to get rule details", zap.Error(err))
		return
	}

	logger.Info("rule details",
		zap.String("rule_id", rule.AutoConversionRuleID),
		zap.String("nickname", rule.Nickname),
		zap.String("status", rule.Status),
	)

	// Log deposit info if available
	if rule.SourceDepositInfo != nil {
		if rule.SourceDepositInfo.Bank != nil {
			logger.Info("rule deposit info (bank)",
				zap.String("reference_code", rule.SourceDepositInfo.Bank.ReferenceCode),
				zap.String("minimum_amount", rule.SourceDepositInfo.Bank.MinimumDepositAmount),
			)
		}
		if rule.SourceDepositInfo.Crypto != nil {
			logger.Info("rule deposit info (crypto)",
				zap.String("wallet_address", rule.SourceDepositInfo.Crypto.WalletAddress),
				zap.String("minimum_amount", rule.SourceDepositInfo.Crypto.MinimumDepositAmount),
			)
		}
	}
}

func listAutoConversionOrders(ctx context.Context, client *onemoney.Client, customerID, ruleID string) {
	resp, err := client.AutoConversionRules.ListOrders(ctx, customerID, ruleID, &auto_conversion_rules.ListOrdersRequest{
		Page: 1,
		Size: 10,
	})
	if err != nil {
		logger.Warn("failed to list auto conversion orders", zap.Error(err))
		return
	}

	logger.Info("auto conversion orders",
		zap.String("rule_id", ruleID),
		zap.Int64("total", resp.Total),
		zap.Int("returned", len(resp.Items)),
	)

	for _, order := range resp.Items {
		logger.Info("order",
			zap.String("order_id", order.AutoConversionOrderID),
			zap.String("status", order.Status),
			zap.String("initial_amount", order.Receipt.Initial.Amount),
			zap.String("initial_asset", order.Receipt.Initial.Asset),
		)
	}
}

func deleteAutoConversionRule(ctx context.Context, client *onemoney.Client, customerID, ruleID string) {
	logger.Info("deleting auto conversion rule (soft delete)", zap.String("rule_id", ruleID))

	err := client.AutoConversionRules.DeleteRule(ctx, customerID, ruleID)
	if err != nil {
		logger.Warn("failed to delete rule", zap.Error(err))
		return
	}

	// Verify deletion (status should be INACTIVE)
	rule, err := client.AutoConversionRules.GetRule(ctx, customerID, ruleID)
	if err != nil {
		logger.Warn("failed to verify rule deletion", zap.Error(err))
		return
	}

	logger.Info("rule deleted (soft delete)",
		zap.String("rule_id", ruleID),
		zap.String("new_status", rule.Status),
	)
}

// performConversions demonstrates manual asset conversions with network validation.
func performConversions(ctx context.Context, client *onemoney.Client, customerID string) {
	// Define conversions to perform - covers crypto↔fiat scenarios
	conversionTypes := []struct {
		fromAsset   assets.AssetName
		fromNetwork conversions.WalletNetworkName
		fromAmount  string
		toAsset     assets.AssetName
		toNetwork   conversions.WalletNetworkName
		label       string
	}{
		// Crypto to Fiat
		{
			fromAsset:   assets.AssetNameUSDC,
			fromNetwork: conversions.WalletNetworkNamePOLYGON,
			fromAmount:  "50.00",
			toAsset:     assets.AssetNameUSD,
			toNetwork:   "",
			label:       "USDC Polygon → USD",
		},
		// Fiat to Crypto
		{
			fromAsset:   assets.AssetNameUSD,
			fromNetwork: "",
			fromAmount:  "50.00",
			toAsset:     assets.AssetNameUSDC,
			toNetwork:   conversions.WalletNetworkNameETHEREUM,
			label:       "USD → USDC Ethereum",
		},
	}

	for _, conv := range conversionTypes {
		logger.Info("performing conversion", zap.String("type", conv.label))

		// Create quote
		quoteReq := &conversions.CreateQuoteRequest{
			FromAsset: conversions.AssetInfo{
				Asset:   conv.fromAsset,
				Amount:  conv.fromAmount,
				Network: conv.fromNetwork,
			},
			ToAsset: conversions.AssetInfo{
				Asset:   conv.toAsset,
				Network: conv.toNetwork,
			},
		}

		quote, err := client.Conversions.CreateQuote(ctx, customerID, quoteReq)
		if err != nil {
			logger.Warn("failed to create quote (may require sufficient balance)",
				zap.String("type", conv.label),
				zap.Error(err),
			)
			continue
		}

		// Log quote with network fields for verification
		logger.Info("quote created",
			zap.String("quote_id", quote.QuoteID),
			zap.String("rate", quote.Rate),
			zap.String("user_pay_asset", quote.UserPayAsset),
			zap.String("user_pay_network", quote.UserPayNetwork),
			zap.String("user_obtain_asset", quote.UserObtainAsset),
			zap.String("user_obtain_network", quote.UserObtainNetwork),
			zap.String("valid_until", quote.ValidUntilTimestamp),
		)

		// Execute hedge
		hedge, err := client.Conversions.CreateHedge(ctx, customerID, &conversions.CreateHedgeRequest{
			QuoteID: quote.QuoteID,
		})
		if err != nil {
			logger.Warn("failed to execute hedge",
				zap.String("type", conv.label),
				zap.Error(err),
			)
			continue
		}

		logger.Info("hedge executed",
			zap.String("order_id", hedge.OrderID),
			zap.String("status", hedge.OrderStatus),
			zap.String("user_pay_network", hedge.UserPayNetwork),
			zap.String("user_obtain_network", hedge.UserObtainNetwork),
		)

		// Get order details
		order, err := client.Conversions.GetOrder(ctx, customerID, hedge.OrderID)
		if err != nil {
			logger.Warn("failed to get order details", zap.Error(err))
			continue
		}

		logger.Info("order details",
			zap.String("order_id", order.OrderID),
			zap.String("status", order.OrderStatus),
			zap.String("fee", order.Fee),
			zap.String("fee_currency", order.FeeCurrency),
		)
	}
}

// performWithdrawals demonstrates both fiat and crypto withdrawals.
func performWithdrawals(ctx context.Context, client *onemoney.Client, customerID, externalAccountID string) {
	// 1. Fiat withdrawal (USD via ACH to external bank account)
	createFiatWithdrawal(ctx, client, customerID, externalAccountID)

	// 2. Crypto withdrawal (USDT to external wallet address)
	createCryptoWithdrawal(ctx, client, customerID)
}

func createFiatWithdrawal(ctx context.Context, client *onemoney.Client, customerID, externalAccountID string) {
	if externalAccountID == "" {
		logger.Warn("skipping fiat withdrawal (no external account)")
		return
	}

	logger.Info("creating fiat withdrawal (USD via ACH)")
	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey:    uuid.New().String(),
		Amount:            "10.00",
		Asset:             assets.AssetNameUSD,
		Network:           assets.NetworkNameUSACH,
		ExternalAccountID: externalAccountID,
	}

	resp, err := client.Withdrawals.CreateWithdrawal(ctx, customerID, req)
	if err != nil {
		logger.Warn("failed to create fiat withdrawal", zap.Error(err))
		return
	}

	logger.Info("fiat withdrawal created",
		zap.String("transaction_id", resp.TransactionID),
		zap.String("status", resp.Status),
		zap.String("amount", req.Amount),
		zap.String("asset", string(req.Asset)),
	)
}

func createCryptoWithdrawal(ctx context.Context, client *onemoney.Client, customerID string) {
	// Example external wallet address (replace with real address in production)
	externalWallet := "0x742d35Cc6634C0532925a3b844Bc9e7595f2bD38"

	logger.Info("creating crypto withdrawal (USDT to external wallet)")
	req := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey: uuid.New().String(),
		Amount:         "10.00",
		Asset:          assets.AssetNameUSDT,
		Network:        assets.NetworkNameETHEREUM,
		WalletAddress:  externalWallet,
	}

	resp, err := client.Withdrawals.CreateWithdrawal(ctx, customerID, req)
	if err != nil {
		logger.Warn("failed to create crypto withdrawal (may require sufficient balance)",
			zap.Error(err),
		)
		return
	}

	logger.Info("crypto withdrawal created",
		zap.String("transaction_id", resp.TransactionID),
		zap.String("status", resp.Status),
		zap.String("amount", req.Amount),
		zap.String("asset", string(req.Asset)),
		zap.String("network", string(req.Network)),
		zap.String("wallet_address", externalWallet),
	)
}

// viewTransactionHistory demonstrates querying transactions with filtering and pagination.
func viewTransactionHistory(ctx context.Context, client *onemoney.Client, customerID string) {
	// 1. List all transactions
	logger.Info("listing all transactions")
	listResp, err := client.Transactions.ListTransactions(ctx, customerID, nil)
	if err != nil {
		logger.Warn("failed to list transactions", zap.Error(err))
		return
	}

	logger.Info("transaction summary",
		zap.Int("returned", len(listResp.List)),
		zap.Int("total", listResp.Total),
	)

	for _, tx := range listResp.List {
		logger.Info("transaction",
			zap.String("id", tx.TransactionID),
			zap.String("action", tx.TransactionAction),
			zap.String("asset", tx.Asset),
			zap.String("network", tx.Network),
			zap.String("amount", tx.Amount),
			zap.String("status", tx.Status),
		)
	}

	// 2. Filter by asset (USD only)
	logger.Info("filtering transactions by USD")
	usdResp, err := client.Transactions.ListTransactions(ctx, customerID, &transactions.ListTransactionsRequest{
		Asset: assets.AssetNameUSD,
	})
	if err != nil {
		logger.Warn("failed to filter transactions by USD", zap.Error(err))
	} else {
		logger.Info("USD transactions",
			zap.Int("returned", len(usdResp.List)),
			zap.Int("total", usdResp.Total),
		)
	}

	// 3. Pagination example
	logger.Info("paginated query (page 1, size 5)")
	pagedResp, err := client.Transactions.ListTransactions(ctx, customerID, &transactions.ListTransactionsRequest{
		Page: 1,
		Size: 5,
	})
	if err != nil {
		logger.Warn("failed to get paginated transactions", zap.Error(err))
	} else {
		logger.Info("paginated result",
			zap.Int("returned", len(pagedResp.List)),
			zap.Int("total", pagedResp.Total),
		)
	}

	// 4. Get single transaction details
	if len(listResp.List) > 0 {
		txID := listResp.List[0].TransactionID
		logger.Info("getting transaction details", zap.String("transaction_id", txID))

		tx, err := client.Transactions.GetTransaction(ctx, customerID, txID)
		if err != nil {
			logger.Warn("failed to get transaction details", zap.Error(err))
			return
		}

		logger.Info("transaction details",
			zap.String("id", tx.TransactionID),
			zap.String("idempotency_key", tx.IdempotencyKey),
			zap.String("action", tx.TransactionAction),
			zap.String("amount", tx.Amount),
			zap.String("asset", tx.Asset),
			zap.String("network", tx.Network),
			zap.String("fee", tx.TransactionFee),
			zap.String("status", tx.Status),
			zap.String("source", tx.Source.AddressID),
			zap.String("destination", tx.Destination.AddressID),
			zap.String("created_at", tx.CreatedAt),
		)
	}
}
