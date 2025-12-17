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

// This example demonstrates how to create a new business customer.
//
// Prerequisites:
//   - Set ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY environment variables
//
// Run: go run ./examples/create_customer
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/joho/godotenv"

	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
	"github.com/1Money-Co/1money-go-sdk/pkg/testdata"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	// Create client (credentials from env vars)
	client, err := onemoney.NewClient(&onemoney.Config{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Step 1: Create TOS link and sign agreement
	log.Println("creating TOS link")
	tosResp, err := client.Customer.CreateTOSLink(ctx, &customer.CreateTOSLinkRequest{
		RedirectUri: "https://example.com/redirect",
	})
	if err != nil {
		log.Fatalf("failed to create TOS link: %v", err)
	}
	log.Printf("TOS link created: url=%s", tosResp.Url)

	log.Println("signing TOS agreement")
	signResp, err := client.Customer.SignTOSAgreement(ctx, tosResp.SessionToken)
	if err != nil {
		log.Fatalf("failed to sign TOS agreement: %v", err)
	}
	log.Printf("TOS agreement signed: signed_agreement_id=%s", signResp.SignedAgreementID)

	// Step 2: Create customer with KYB information
	log.Println("creating customer")
	req := buildCustomerRequest(signResp.SignedAgreementID)
	resp, err := client.Customer.CreateCustomer(ctx, req)
	if err != nil {
		log.Fatalf("failed to create customer: %v", err)
	}
	log.Printf("customer created: customer_id=%s status=%s", resp.CustomerID, resp.Status)

	// Step 3: Wait for KYB approval (sandbox auto-approves)
	// In production, you might want to implement a webhook to get notified of status changes instead of polling.
	log.Println("waiting for KYB approval")
	if _, err = customer.WaitForKybApproved(ctx, client.Customer, resp.CustomerID, &customer.WaitOptions{
		PrintProgress: true,
	}); err != nil {
		log.Fatalf("KYB approval failed: %v", err)
	}
	log.Println("KYB approved, now we are waiting for fiat account provisioning")

	// Wait for fiat account provisioning (fixed delay, no polling needed)
	log.Println("waiting 60s for fiat account provisioning...")
	customer.WaitForFaitAccount()
	log.Println("fiat account has been created, customer is ready to use")

	// Verify customer details
	getResp, err := client.Customer.GetCustomer(ctx, resp.CustomerID)
	if err != nil {
		log.Fatalf("failed to get customer: %v", err)
	}
	log.Printf("customer details: customer_id=%s business_name=%s status=%s",
		getResp.CustomerID, getResp.BusinessLegalName, getResp.Status)

	log.Printf("export this for other examples: export ONEMONEY_CUSTOMER_ID=%s", resp.CustomerID)
}

func buildCustomerRequest(signedAgreementID string) *customer.CreateCustomerRequest {
	// nationalIDLength is the length of the generated national identity number.
	const nationalIDLength = 12
	// Note: In production, you would provide real business information and documents
	return &customer.CreateCustomerRequest{
		BusinessLegalName:          "Example Corp",
		BusinessDescription:        "Example business for SDK demonstration",
		BusinessRegistrationNumber: fmt.Sprintf("REG-%d", time.Now().Unix()),
		Email:                      fmt.Sprintf("example-%d@example.com", time.Now().Unix()),
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           "332999",
		RegisteredAddress: &customer.Address{
			StreetLine1: "123 Example St",
			City:        "Munich",
			State:       "BY",
			Country:     "DEU",
			PostalCode:  "80331",
			Subdivision: "BY",
		},
		DateOfIncorporation: "2020-01-15",
		SignedAgreementID:   signedAgreementID,
		AssociatedPersons: []customer.AssociatedPerson{
			{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Gender:    customer.GenderMale,
				ResidentialAddress: &customer.Address{
					StreetLine1: "456 Residential St",
					City:        "Munich",
					State:       "BY",
					Country:     "DEU",
					PostalCode:  "80333",
					Subdivision: "BY",
				},
				BirthDate:           "1985-06-15",
				CountryOfBirth:      string(external_accounts.CountryCodeDEU),
				PrimaryNationality:  string(external_accounts.CountryCodeDEU),
				HasOwnership:        true,
				OwnershipPercentage: 100,
				HasControl:          true,
				IsSigner:            true,
				IsDirector:          true,
				IdentifyingInformation: []customer.IdentifyingInformation{
					{
						Type:                   customer.IDTypeNationalId,
						IssuingCountry:         string(external_accounts.CountryCodeDEU),
						ImageFront:             testdata.IDFront(),
						ImageBack:              testdata.IDBack(),
						NationalIdentityNumber: gofakeit.LetterN(nationalIDLength),
					},
				},
				CountryOfTax: string(external_accounts.CountryCodeDEU),
				TaxType:      customer.TaxIDTypeSSN,
				TaxID:        "123-45-6789",
				POA:          testdata.POA(),
				POAType:      "utility_bill",
			},
		},
		SourceOfFunds:  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth: []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},
		// Required documents for Corporation in US region
		// Uses embedded test images from pkg/testdata
		Documents: []customer.Document{
			{
				DocType:     customer.DocumentTypeFlowOfFunds,
				File:        testdata.POAAsDocument(),
				Description: "Proof of Funds",
			},
			{
				DocType:     customer.DocumentTypeRegistrationDocument,
				File:        testdata.POAAsDocument(),
				Description: "Certificate of Incorporation",
			},
			{
				DocType:     customer.DocumentTypeProofOfTaxIdentification,
				File:        testdata.POAAsDocument(),
				Description: "W9 Form",
			},
			{
				DocType:     customer.DocumentTypeShareholderRegister,
				File:        testdata.POAAsDocument(),
				Description: "Ownership Structure",
			},
			{
				DocType:     customer.DocumentTypeESignatureCertificate,
				File:        testdata.POAAsDocument(),
				Description: "Authorized Representative List",
			},
			{
				DocType:     customer.DocumentTypeEvidenceOfGoodStanding,
				File:        testdata.POAAsDocument(),
				Description: "Evidence of Good Standing",
			},
			{
				DocType:     customer.DocumentTypeProofOfAddress,
				File:        testdata.POAAsDocument(),
				Description: "Proof of Address",
			},
		},
		AccountPurpose:                 customer.AccountPurposeTreasuryManagement,
		EstimatedAnnualRevenueUSD:      customer.MoneyRange099999,
		ExpectedMonthlyFiatDeposits:    customer.MoneyRange099999,
		ExpectedMonthlyFiatWithdrawals: customer.MoneyRange099999,
		TaxID:                          "12-3456789",
		TaxType:                        customer.TaxIDTypeEIN,
		TaxCountry:                     "DEU",
	}
}
