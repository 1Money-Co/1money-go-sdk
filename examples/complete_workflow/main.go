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
//  3. Get deposit instructions
//  4. Simulate deposits (sandbox only)
//  5. Create external bank accounts
//  6. Set up auto conversion rules
//  7. Perform manual asset conversions
//  8. Create withdrawals
//  9. View transaction history
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
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/auto_conversion_rules"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/conversions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/instructions"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/withdraws"
)

func main() {
	_ = godotenv.Load()

	client, err := onemoney.NewClient(&onemoney.Config{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Step 1: Get or use existing customer
	customerID := getOrCreateCustomer(ctx, client)
	fmt.Printf("\n=== Using Customer ID: %s ===\n\n", customerID)

	// Step 2: View initial asset balances
	fmt.Println("=== Step 1: View Initial Asset Balances ===")
	viewAssetBalances(ctx, client, customerID)

	// Step 3: Get deposit instructions
	fmt.Println("\n=== Step 2: Get Deposit Instructions ===")
	getDepositInstructions(ctx, client, customerID)

	// Step 4: Simulate deposits (sandbox only)
	fmt.Println("\n=== Step 3: Simulate Deposits (Sandbox Only) ===")
	simulateDeposits(ctx, client, customerID)

	// Step 5: View updated asset balances
	fmt.Println("\n=== Step 4: View Updated Asset Balances ===")
	viewAssetBalances(ctx, client, customerID)

	// Step 6: Create external bank account
	fmt.Println("\n=== Step 5: Create External Bank Account ===")
	externalAccountID := createExternalAccount(ctx, client, customerID)

	// Step 7: Create auto conversion rule
	fmt.Println("\n=== Step 6: Create Auto Conversion Rule ===")
	autoConversionRuleID := createAutoConversionRule(ctx, client, customerID)

	// Step 8: Perform manual asset conversion
	fmt.Println("\n=== Step 7: Perform Manual Asset Conversion ===")
	conversionOrderID := performConversion(ctx, client, customerID)

	// Step 9: View asset balances after conversion
	fmt.Println("\n=== Step 8: View Asset Balances After Conversion ===")
	viewAssetBalances(ctx, client, customerID)

	// Step 10: Create withdrawal
	fmt.Println("\n=== Step 9: Create Withdrawal ===")
	withdrawalID := createWithdrawal(ctx, client, customerID, externalAccountID)

	// Step 11: View transaction history
	fmt.Println("\n=== Step 10: View Transaction History ===")
	viewTransactionHistory(ctx, client, customerID)

	// Summary
	fmt.Println("\n=== Workflow Summary ===")
	fmt.Printf("Customer ID:              %s\n", customerID)
	fmt.Printf("External Account ID:     %s\n", externalAccountID)
	if autoConversionRuleID != "" {
		fmt.Printf("Auto Conversion Rule ID: %s\n", autoConversionRuleID)
	} else {
		fmt.Println("Auto Conversion Rule ID: (not created - fiat account may need verification)")
	}
	if conversionOrderID != "" {
		fmt.Printf("Conversion Order ID:     %s\n", conversionOrderID)
	} else {
		fmt.Println("Conversion Order ID:     (not created - conversion pair may not be supported)")
	}
	fmt.Printf("Withdrawal ID:           %s\n", withdrawalID)
	fmt.Println("\nComplete workflow executed successfully!")
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
			fmt.Printf("Using existing customer: %s\n", customerID)
			return customerID
		}
		fmt.Printf("Customer %s not found, will create new one\n", customerID)
	}

	// Create new customer (simplified - see create_customer example for full details)
	fmt.Println("Creating new customer...")
	tosResp, err := client.Customer.CreateTOSLink(ctx, &customer.CreateTOSLinkRequest{
		RedirectUri: "https://example.com/tos-completed",
	})
	if err != nil {
		log.Fatalf("failed to create TOS link: %v", err)
	}

	signResp, err := client.Customer.SignTOSAgreement(ctx, tosResp.SessionToken)
	if err != nil {
		log.Fatalf("failed to sign TOS agreement: %v", err)
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
		log.Fatalf("failed to create customer: %v", err)
	}

	fmt.Printf("Customer created: %s (Status: %s)\n", resp.CustomerID, resp.Status)
	return resp.CustomerID
}

func viewAssetBalances(ctx context.Context, client *onemoney.Client, customerID string) {
	assetList, err := client.Assets.ListAssets(ctx, customerID, nil)
	if err != nil {
		log.Fatalf("failed to list assets: %v", err)
	}

	if len(assetList) == 0 {
		fmt.Println("No assets found for this customer.")
		return
	}

	fmt.Printf("Found %d asset(s):\n", len(assetList))
	for i, a := range assetList {
		network := "<fiat>"
		if a.Network != nil {
			network = *a.Network
		}
		fmt.Printf("  [%d] %s (%s) - Available: %s, Unavailable: %s\n",
			i+1, a.Asset, network, a.AvailableAmount, a.UnavailableAmount)
	}
}

func getDepositInstructions(ctx context.Context, client *onemoney.Client, customerID string) {
	// Get USD deposit instruction via ACH
	fmt.Println("Getting USD deposit instruction via ACH...")
	usdAch, err := client.Instructions.GetDepositInstruction(ctx, customerID, assets.AssetNameUSD, assets.NetworkNameUSACH)
	if err != nil {
		log.Printf("failed to get USD ACH instruction: %v", err)
	} else {
		printInstruction("USD / US_ACH", usdAch)
	}

	// Get USDT deposit instruction on Ethereum
	fmt.Println("\nGetting USDT deposit instruction on ETHEREUM...")
	usdtEth, err := client.Instructions.GetDepositInstruction(ctx, customerID, assets.AssetNameUSDT, assets.NetworkNameETHEREUM)
	if err != nil {
		log.Printf("failed to get USDT ETHEREUM instruction: %v", err)
	} else {
		printInstruction("USDT / ETHEREUM", usdtEth)
	}
}

func printInstruction(label string, instr *instructions.InstructionResponse) {
	fmt.Printf("%s instruction:\n", label)
	fmt.Printf("  Asset: %s, Network: %s\n", instr.Asset, instr.Network)
	if instr.BankInstruction != nil {
		fmt.Printf("  Bank: %s, Account: %s\n",
			instr.BankInstruction.BankName, instr.BankInstruction.AccountNumber)
	}
	if instr.WalletInstruction != nil {
		fmt.Printf("  Wallet Address: %s\n", instr.WalletInstruction.WalletAddress)
	}
}

func simulateDeposits(ctx context.Context, client *onemoney.Client, customerID string) {
	// Simulate USD deposit
	fmt.Println("Simulating USD deposit ($100.00)...")
	usdResp, err := client.Simulations.SimulateDeposit(ctx, customerID, &simulations.SimulateDepositRequest{
		Asset:  assets.AssetNameUSD,
		Amount: "100.00",
	})
	if err != nil {
		log.Printf("failed to simulate USD deposit: %v", err)
	} else {
		fmt.Printf("  Simulation ID: %s, Status: %s\n", usdResp.SimulationID, usdResp.Status)
	}

	// Simulate USDT deposit on Ethereum
	fmt.Println("\nSimulating USDT deposit on ETHEREUM (50.00)...")
	usdtResp, err := client.Simulations.SimulateDeposit(ctx, customerID, &simulations.SimulateDepositRequest{
		Asset:   assets.AssetNameUSDT,
		Network: simulations.WalletNetworkNameETHEREUM,
		Amount:  "50.00",
	})
	if err != nil {
		log.Printf("failed to simulate USDT deposit: %v", err)
	} else {
		fmt.Printf("  Simulation ID: %s, Status: %s\n", usdtResp.SimulationID, usdtResp.Status)
	}
}

func createExternalAccount(ctx context.Context, client *onemoney.Client, customerID string) string {
	idempotencyKey := uuid.New().String()
	createReq := &external_accounts.CreateReq{
		IdempotencyKey:  idempotencyKey,
		Network:         external_accounts.BankNetworkNameUSACH,
		Currency:        external_accounts.CurrencyUSD,
		CountryCode:     external_accounts.CountryCodeUSA,
		AccountNumber:   "123456789",
		InstitutionID:   "021000021",
		InstitutionName: "Example Bank",
	}

	fmt.Println("Creating external bank account...")
	created, err := client.ExternalAccounts.CreateExternalAccount(ctx, customerID, createReq)
	if err != nil {
		log.Fatalf("failed to create external account: %v", err)
	}

	fmt.Printf("External account created: %s (Status: %s)\n",
		created.ExternalAccountID, created.Status)
	return created.ExternalAccountID
}

func createAutoConversionRule(ctx context.Context, client *onemoney.Client, customerID string) string {
	destNetwork := "POLYGON"
	idempotencyKey := uuid.New().String()
	createReq := &auto_conversion_rules.CreateRuleRequest{
		IdempotencyKey: idempotencyKey,
		Source: auto_conversion_rules.SourceAssetInfo{
			Asset:   "USD",
			Network: "ACH", // Use "ACH" instead of "US_ACH" for auto conversion rules
		},
		Destination: auto_conversion_rules.DestinationAssetInfo{
			Asset:   "USDC",
			Network: &destNetwork,
		},
	}

	fmt.Println("Creating USD -> USDC (POLYGON) auto conversion rule...")
	created, err := client.AutoConversionRules.CreateRule(ctx, customerID, createReq)
	if err != nil {
		// This error is expected if the customer's fiat account is not yet verified.
		// In production, customers need to complete KYC/verification before creating auto conversion rules.
		log.Printf("failed to create auto conversion rule (this may be normal if fiat account is not verified): %v", err)
		fmt.Println("Note: Auto conversion rules require a verified fiat account. Skipping this step.")
		return "" // Return empty string to indicate rule was not created
	}

	fmt.Printf("Auto conversion rule created: %s (Status: %s)\n",
		created.AutoConversionRuleID, created.Status)
	return created.AutoConversionRuleID
}

func performConversion(ctx context.Context, client *onemoney.Client, customerID string) string {
	// Create a quote for USDT -> USD conversion (crypto to fiat)
	fmt.Println("Creating USDT -> USD conversion quote...")
	quoteReq := &conversions.CreateQuoteRequest{
		FromAsset: conversions.AssetInfo{
			Asset:   assets.AssetNameUSDT,
			Amount:  "10.00",
			Network: conversions.WalletNetworkNameETHEREUM,
		},
		ToAsset: conversions.AssetInfo{
			Asset: assets.AssetNameUSD,
		},
	}

	quoteResp, err := client.Conversions.CreateQuote(ctx, customerID, quoteReq)
	if err != nil {
		log.Printf("failed to create conversion quote: %v", err)
		fmt.Println("Note: Conversion quotes may require specific account setup or supported pairs. Skipping this step.")
		return "" // Return empty string to indicate conversion was not performed
	}

	fmt.Printf("Quote created: %s\n", quoteResp.QuoteID)
	fmt.Printf("  Rate: %s, Valid until: %s\n",
		quoteResp.Rate, quoteResp.ValidUntilTimestamp)

	// Execute hedge
	fmt.Println("\nExecuting hedge...")
	hedgeResp, err := client.Conversions.CreateHedge(ctx, customerID, &conversions.CreateHedgeRequest{
		QuoteID: quoteResp.QuoteID,
	})
	if err != nil {
		log.Printf("failed to create hedge: %v", err)
		fmt.Println("Note: Hedge execution may require sufficient balances. Skipping this step.")
		return "" // Return empty string to indicate hedge was not executed
	}

	fmt.Printf("Hedge created: %s (Status: %s)\n",
		hedgeResp.OrderID, hedgeResp.OrderStatus)
	return hedgeResp.OrderID
}

func createWithdrawal(ctx context.Context, client *onemoney.Client, customerID, externalAccountID string) string {
	// Create fiat withdrawal
	fmt.Println("Creating USD withdrawal via ACH...")
	withdrawalKey := uuid.New().String()
	withdrawalReq := &withdraws.CreateWithdrawalRequest{
		IdempotencyKey:    withdrawalKey,
		Amount:            "10.00",
		Asset:             assets.AssetNameUSD,
		Network:           assets.NetworkNameUSACH,
		ExternalAccountID: externalAccountID,
	}

	withdrawalResp, err := client.Withdrawals.CreateWithdrawal(ctx, customerID, withdrawalReq)
	if err != nil {
		log.Fatalf("failed to create withdrawal: %v", err)
	}

	fmt.Printf("Withdrawal created: %s (Status: %s)\n",
		withdrawalResp.TransactionID, withdrawalResp.Status)
	return withdrawalResp.TransactionID
}

func viewTransactionHistory(ctx context.Context, client *onemoney.Client, customerID string) {
	listResp, err := client.Transactions.ListTransactions(ctx, customerID, nil)
	if err != nil {
		log.Fatalf("failed to list transactions: %v", err)
	}

	fmt.Printf("Found %d transaction(s) (total=%d):\n", len(listResp.List), listResp.Total)
	for i := range listResp.List {
		tx := &listResp.List[i]
		fmt.Printf("  [%d] %s - %s %s %s - Amount: %s, Status: %s\n",
			i+1,
			tx.TransactionID,
			tx.TransactionAction,
			tx.Asset,
			tx.Network,
			tx.Amount,
			tx.Status,
		)
	}
}
