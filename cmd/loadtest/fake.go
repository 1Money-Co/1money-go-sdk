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

package loadtest

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/auto_conversion_rules"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
)

// CountryUSA is the country code for United States.
const CountryUSA = "USA"

// Fake data generation constants.
const (
	maxUint8Value       = 255
	fakeImageSize       = 100
	idLetterCount       = 8
	idDigitCount        = 4
	accountNumberDigits = 9
	regNumMin           = 100000
	regNumMax           = 999999
	suiteNumMin         = 100
	suiteNumMax         = 999
	taxIDPrefixMin      = 10
	taxIDPrefixMax      = 99
	taxIDSuffixMin      = 1000000
	taxIDSuffixMax      = 9999999
	regLetterCount      = 3
)

// ValidUSStates contains valid US state codes for API validation.
var ValidUSStates = []string{
	"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA",
	"HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD",
	"MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ",
	"NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC",
	"SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY",
	"DC",
}

// RandomUSState returns a random valid US state code.
func RandomUSState(faker *gofakeit.Faker) string {
	return ValidUSStates[faker.Number(0, len(ValidUSStates)-1)]
}

// safeUint8 converts an int to uint8 with bounds checking to avoid overflow.
func safeUint8(n int) uint8 {
	if n < 0 {
		return 0
	}
	if n > maxUint8Value {
		return maxUint8Value
	}
	return uint8(n)
}

// FakeImagePNG generates a valid PNG image as bytes.
func FakeImagePNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	c := color.RGBA{
		R: safeUint8(gofakeit.Number(0, maxUint8Value)),
		G: safeUint8(gofakeit.Number(0, maxUint8Value)),
		B: safeUint8(gofakeit.Number(0, maxUint8Value)),
		A: maxUint8Value,
	}
	for y := range height {
		for x := range width {
			img.Set(x, y, c)
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// FakeCustomerDocuments generates fake documents required for customer creation.
func FakeCustomerDocuments() []customer.Document {
	fakeImage := func() string {
		return customer.EncodeBase64ToDataURI(FakeImagePNG(fakeImageSize, fakeImageSize), customer.ImageFormatPng)
	}
	return []customer.Document{
		{
			DocType:     customer.DocumentTypeFlowOfFunds,
			File:        fakeImage(),
			Description: "Proof of Funds",
		},
		{
			DocType:     customer.DocumentTypeRegistrationDocument,
			File:        fakeImage(),
			Description: "Certificate of Incorporation",
		},
		{
			DocType:     customer.DocumentTypeProofOfTaxIdentification,
			File:        fakeImage(),
			Description: "W9 Form",
		},
		{
			DocType:     customer.DocumentTypeShareholderRegister,
			File:        fakeImage(),
			Description: "Ownership Structure",
		},
		{
			DocType:     customer.DocumentTypeESignatureCertificate,
			File:        fakeImage(),
			Description: "Authorized Representative List",
		},
		{
			DocType:     customer.DocumentTypeEvidenceOfGoodStanding,
			File:        fakeImage(),
			Description: "Evidence of Good Standing",
		},
		{
			DocType:     customer.DocumentTypeProofOfAddress,
			File:        fakeImage(),
			Description: "Proof of Address",
		},
	}
}

// FakeAssociatedPerson generates a fake associated person for testing.
func FakeAssociatedPerson(faker *gofakeit.Faker) customer.AssociatedPerson {
	gender := customer.GenderMale
	if faker.Bool() {
		gender = customer.GenderFemale
	}

	return customer.AssociatedPerson{
		FirstName: faker.FirstName(),
		LastName:  faker.LastName(),
		Email:     faker.Email(),
		Gender:    gender,
		ResidentialAddress: &customer.Address{
			StreetLine1: faker.Street(),
			City:        faker.City(),
			State:       RandomUSState(faker),
			Country:     CountryUSA,
			PostalCode:  faker.Zip(),
			Subdivision: RandomUSState(faker),
		},
		BirthDate:           faker.Date().Format("2006-01-02"),
		CountryOfBirth:      CountryUSA,
		PrimaryNationality:  CountryUSA,
		HasOwnership:        true,
		OwnershipPercentage: 100,
		HasControl:          true,
		IsSigner:            true,
		IsDirector:          true,
		IdentifyingInformation: []customer.IdentifyingInformation{
			{
				Type:                   customer.IDTypeDriversLicense,
				IssuingCountry:         CountryUSA,
				ImageFront:             customer.EncodeBase64ToDataURI(FakeImagePNG(fakeImageSize, fakeImageSize), customer.ImageFormatPng),
				ImageBack:              customer.EncodeBase64ToDataURI(FakeImagePNG(fakeImageSize, fakeImageSize), customer.ImageFormatPng),
				NationalIdentityNumber: faker.LetterN(idLetterCount) + faker.DigitN(idDigitCount),
			},
		},
		CountryOfTax: CountryUSA,
		TaxType:      customer.TaxIDTypeSSN,
		TaxID:        faker.SSN(),
		POA:          customer.EncodeBase64ToDataURI(FakeImagePNG(fakeImageSize, fakeImageSize), customer.ImageFormatPng),
		POAType:      "utility_bill",
	}
}

// FakeCreateCustomerRequest generates a fake customer creation request.
func FakeCreateCustomerRequest(faker *gofakeit.Faker, signedAgreementID string) *customer.CreateCustomerRequest {
	return &customer.CreateCustomerRequest{
		BusinessLegalName:          faker.Company(),
		BusinessDescription:        faker.JobDescriptor() + " " + faker.BS(),
		BusinessRegistrationNumber: fmt.Sprintf("%s-%d", faker.LetterN(regLetterCount), faker.Number(regNumMin, regNumMax)),
		Email:                      faker.Email(),
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           "332999",
		RegisteredAddress: &customer.Address{
			StreetLine1: faker.Street(),
			StreetLine2: fmt.Sprintf("Suite %d", faker.Number(suiteNumMin, suiteNumMax)),
			City:        faker.City(),
			State:       RandomUSState(faker),
			Country:     CountryUSA,
			PostalCode:  faker.Zip(),
			Subdivision: RandomUSState(faker),
		},
		DateOfIncorporation: faker.Date().Format("2006-01-02"),
		SignedAgreementID:   signedAgreementID,
		AssociatedPersons: []customer.AssociatedPerson{
			FakeAssociatedPerson(faker),
			FakeAssociatedPerson(faker),
		},
		SourceOfFunds:                  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth:                 []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},
		Documents:                      FakeCustomerDocuments(),
		AccountPurpose:                 customer.AccountPurposeTreasuryManagement,
		IsDAO:                          false,
		PubliclyTraded:                 false,
		EstimatedAnnualRevenueUSD:      customer.MoneyRange099999,
		ExpectedMonthlyFiatDeposits:    customer.MoneyRange099999,
		ExpectedMonthlyFiatWithdrawals: customer.MoneyRange099999,
		TaxID: fmt.Sprintf("%d-%d",
			faker.Number(taxIDPrefixMin, taxIDPrefixMax),
			faker.Number(taxIDSuffixMin, taxIDSuffixMax)),
		TaxType:    customer.TaxIDTypeEIN,
		TaxCountry: CountryUSA,
	}
}

// FakeExternalAccountRequest generates a fake external account request.
func FakeExternalAccountRequest(faker *gofakeit.Faker) *external_accounts.CreateReq {
	return &external_accounts.CreateReq{
		IdempotencyKey:  uuid.New().String(),
		Network:         external_accounts.BankNetworkNameUSACH,
		Currency:        external_accounts.CurrencyUSD,
		CountryCode:     external_accounts.CountryCodeUSA,
		AccountNumber:   faker.DigitN(accountNumberDigits),
		InstitutionID:   faker.DigitN(accountNumberDigits),
		InstitutionName: faker.Company() + " Bank",
	}
}

// FakeAutoConversionRuleRequest generates a fake auto conversion rule request.
func FakeAutoConversionRuleRequest() *auto_conversion_rules.CreateRuleRequest {
	network := "POLYGON"
	return &auto_conversion_rules.CreateRuleRequest{
		IdempotencyKey: uuid.New().String(),
		Source: auto_conversion_rules.SourceAssetInfo{
			Asset:   "USD",
			Network: "ACH", // Use "ACH" instead of "US_ACH" for auto conversion rules
		},
		Destination: auto_conversion_rules.DestinationAssetInfo{
			Asset:   "USDC",
			Network: &network,
		},
	}
}
