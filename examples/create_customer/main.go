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

// Package main demonstrates how to create a customer using the 1Money SDK.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
)

func main() {
	// Create a client (credentials can be provided via environment variables
	// ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY, or via config file)
	client, err := onemoney.NewClient(&onemoney.Config{
		AccessKey: "your-access-key",
		SecretKey: "your-secret-key",
	})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Build the create customer request
	req := &customer.CreateCustomerRequest{
		// Business information
		BusinessLegalName:          "Acme Corporation",
		BusinessDescription:        "Software development and consulting services",
		BusinessRegistrationNumber: "REG-123456",
		Email:                      "contact@acme-corp.example.com",
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           "541511", // Custom Computer Programming Services

		// Registered address
		RegisteredAddress: &customer.Address{
			StreetLine1: "123 Business Ave",
			StreetLine2: "Suite 100",
			City:        "San Francisco",
			State:       "CA",
			Country:     "USA",
			PostalCode:  "94102",
			Subdivision: "CA",
		},

		// Incorporation details
		DateOfIncorporation: "2020-01-15",

		// Signed agreement ID from TOS flow (obtain via CreateTOSLink + SignTOSAgreement)
		SignedAgreementID: "your-signed-agreement-id",

		// Associated persons (owners, directors, signers)
		AssociatedPersons: []customer.AssociatedPerson{
			{
				FirstName: "John",
				LastName:  "Smith",
				Email:     "john.smith@acme-corp.example.com",
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
						ImageFront:             "data:image/jpeg;base64,...", // Base64 encoded front image
						ImageBack:              "data:image/jpeg;base64,...", // Base64 encoded back image
						NationalIdentityNumber: "D1234567",
					},
				},
				CountryOfTax: "USA",
				TaxType:      customer.TaxIDTypeSSN,
				TaxID:        "123-45-6789",
				POA:          "data:image/jpeg;base64,...", // Proof of address image
				POAType:      "utility_bill",
			},
		},

		// Source of funds and wealth
		SourceOfFunds:  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth: []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},

		// Required documents
		Documents: []customer.Document{
			{
				DocType:     customer.DocumentTypeRegistrationDocument,
				File:        "data:image/jpeg;base64,...", // Certificate of incorporation
				Description: "Certificate of Incorporation",
			},
			{
				DocType:     customer.DocumentTypeProofOfTaxIdentification,
				File:        "data:application/pdf;base64,...", // W9 form as PDF
				Description: "W9 Form",
			},
		},

		// Account details
		AccountPurpose:                 customer.AccountPurposeTreasuryManagement,
		IsDAO:                          false,
		PubliclyTraded:                 false,
		EstimatedAnnualRevenueUSD:      customer.MoneyRange099999,
		ExpectedMonthlyFiatDeposits:    customer.MoneyRange099999,
		ExpectedMonthlyFiatWithdrawals: customer.MoneyRange099999,

		// Tax information
		TaxID:      "12-3456789", // EIN
		TaxType:    customer.TaxIDTypeEIN,
		TaxCountry: "USA",
	}

	// Create the customer
	ctx := context.Background()
	resp, err := client.Customer.CreateCustomer(ctx, req)
	if err != nil {
		log.Fatalf("failed to create customer: %v", err)
	}

	fmt.Println("Customer created successfully!")
	fmt.Println("Customer ID:", resp.CustomerID)
	fmt.Println("Status:", resp.Status)
}
