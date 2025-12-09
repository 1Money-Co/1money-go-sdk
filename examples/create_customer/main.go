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
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"

	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/joho/godotenv"
)

// generateSampleImage generates a valid PNG image for testing purposes.
// In production, you should use real document images.
func generateSampleImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a light gray color
	c := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(fmt.Sprintf("failed to encode PNG: %v", err))
	}
	return buf.Bytes()
}

func main() {
	// Load .env file if it exists (silently ignore if not found)
	// This allows users to set ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY in .env
	_ = godotenv.Load()
	// Create a client (credentials can be provided via environment variables
	// ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY, or via config file)
	client, err := onemoney.NewClient(&onemoney.Config{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Step 1: Create TOS link to get session token
	fmt.Println("Creating TOS link...")
	tosResp, err := client.Customer.CreateTOSLink(ctx)
	if err != nil {
		log.Fatalf("failed to create TOS link: %v", err)
	}
	fmt.Printf("TOS link created. Session token: %s\n", tosResp.SessionToken)

	// Step 2: Sign the agreement using the session token
	fmt.Println("Signing TOS agreement...")
	signResp, err := client.Customer.SignTOSAgreement(ctx, tosResp.SessionToken)
	if err != nil {
		log.Fatalf("failed to sign TOS agreement: %v", err)
	}
	fmt.Printf("TOS agreement signed. Signed agreement ID: %s\n", signResp.SignedAgreementID)

	// Step 3: Build the create customer request
	req := &customer.CreateCustomerRequest{
		// Business information
		BusinessLegalName:          "Acme Corporation",
		BusinessDescription:        "Software development and consulting services",
		BusinessRegistrationNumber: "REG-123456",
		Email:                      "contact@acme-corp.example.com",
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           "332999", // All Other Miscellaneous Fabricated Metal Product Manufacturing

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

		// Signed agreement ID from TOS flow
		SignedAgreementID: signResp.SignedAgreementID,

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
						ImageFront:             customer.EncodeBase64ToDataURI(generateSampleImage(100, 100), customer.ImageFormatPng),
						ImageBack:              customer.EncodeBase64ToDataURI(generateSampleImage(100, 100), customer.ImageFormatPng),
						NationalIdentityNumber: "D1234567",
					},
				},
				CountryOfTax: "USA",
				TaxType:      customer.TaxIDTypeSSN,
				TaxID:        "123-45-6789",
				POA:          customer.EncodeBase64ToDataURI(generateSampleImage(100, 100), customer.ImageFormatPng),
				POAType:      "utility_bill",
			},
		},

		// Source of funds and wealth
		SourceOfFunds:  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth: []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},

		// Required documents for Corporation in US region
		// Note: In production, use real document files with customer.EncodeFileToDataURI() or customer.EncodeDocumentFileToDataURI()
		Documents: []customer.Document{
			{
				DocType:     customer.DocumentTypeFlowOfFunds,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(100, 100), customer.FileFormatPng),
				Description: "Proof of Funds",
			},
			{
				DocType:     customer.DocumentTypeRegistrationDocument,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(100, 100), customer.FileFormatPng),
				Description: "Certificate of Incorporation",
			},
			{
				DocType:     customer.DocumentTypeProofOfTaxIdentification,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(100, 100), customer.FileFormatPng),
				Description: "W9 Form",
			},
			{
				DocType:     customer.DocumentTypeShareholderRegister,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(100, 100), customer.FileFormatPng),
				Description: "Ownership Structure",
			},
			{
				DocType:     customer.DocumentTypeESignatureCertificate,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(100, 100), customer.FileFormatPng),
				Description: "Authorized Representative List",
			},
			{
				DocType:     customer.DocumentTypeEvidenceOfGoodStanding,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(100, 100), customer.FileFormatPng),
				Description: "Evidence of Good Standing",
			},
			{
				DocType:     customer.DocumentTypeProofOfAddress,
				File:        customer.EncodeDocumentToDataURI(generateSampleImage(100, 100), customer.FileFormatPng),
				Description: "Proof of Address",
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

	// Step 4: Create the customer
	fmt.Println("Creating customer...")
	resp, err := client.Customer.CreateCustomer(ctx, req)
	if err != nil {
		log.Fatalf("failed to create customer: %v", err)
	}

	fmt.Println("Customer created successfully!")
	fmt.Println("Customer ID:", resp.CustomerID)
	fmt.Println("Status:", resp.Status)

	// Step 5: Get the created customer
	fmt.Println("\nRetrieving customer details...")
	customerResp, err := client.Customer.GetCustomer(ctx, resp.CustomerID)
	if err != nil {
		log.Fatalf("failed to get customer: %v", err)
	}

	fmt.Println("Customer retrieved successfully!")
	fmt.Println("Customer ID:", customerResp.CustomerID)
	fmt.Println("Business Legal Name:", customerResp.BusinessLegalName)
	fmt.Println("Email:", customerResp.Email)
	fmt.Println("Business Type:", customerResp.BusinessType)
	fmt.Println("Status:", customerResp.Status)
	if customerResp.CreatedAt != "" {
		fmt.Println("Created At:", customerResp.CreatedAt)
	}
	if customerResp.UpdatedAt != "" {
		fmt.Println("Updated At:", customerResp.UpdatedAt)
	}

	// Step 6: List all customers
	fmt.Println("\nListing all customers...")
	listReq := &customer.ListCustomersRequest{
		PageSize: 10,
		PageNum:  0,
	}
	listResp, err := client.Customer.ListCustomers(ctx, listReq)
	if err != nil {
		log.Fatalf("failed to list customers: %v", err)
	}

	fmt.Printf("Found %d customer(s):\n", listResp.Total)
	for i, c := range listResp.Customers {
		fmt.Printf("\nCustomer %d:\n", i+1)
		fmt.Println("  Customer ID:", c.CustomerID)
		fmt.Println("  Business Legal Name:", c.BusinessLegalName)
		fmt.Println("  Email:", c.Email)
		fmt.Println("  Business Type:", c.BusinessType)
		fmt.Println("  Status:", c.Status)
		if c.CreatedAt != "" {
			fmt.Println("  Created At:", c.CreatedAt)
		}
		if c.UpdatedAt != "" {
			fmt.Println("  Updated At:", c.UpdatedAt)
		}
	}
}
