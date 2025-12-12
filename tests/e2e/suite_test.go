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

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"github.com/xuri/excelize/v2"

	"github.com/1Money-Co/1money-go-sdk/internal/utils"
	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/auto_conversion_rules"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
)

const (
	// CountryUSA is the country code for United States.
	CountryUSA = "USA"
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

// E2ETestSuite defines the integration test suite for the OneMoney client.
// This suite is used for end-to-end testing during development.
type E2ETestSuite struct {
	suite.Suite
	Client *onemoney.Client
	Ctx    context.Context
}

// SetupSuite runs once before all tests in the suite.
func (s *E2ETestSuite) SetupSuite() {
	// Load environment variables from .env file in project root
	projectRoot, err := utils.FindProjectRoot()
	if err == nil {
		envPath := filepath.Join(projectRoot, ".env")
		_ = godotenv.Load(envPath)
	}

	// Create client configuration
	cfg := &onemoney.Config{}

	// Create client
	client, err := onemoney.NewClient(cfg)
	if err != nil {
		s.T().Fatalf("failed to create client: %v", err)
	}

	s.Client = client
	s.Ctx = context.Background()
}

// SetupTest runs before each test.
func (*E2ETestSuite) SetupTest() {}

// TearDownTest runs after each test.
func (*E2ETestSuite) TearDownTest() {}

// TearDownSuite runs once after all tests.
func (*E2ETestSuite) TearDownSuite() {}

// CustomerDependentTestSuite is a base test suite for tests that require a customer.
// Each test suite that embeds this will create its own customer during SetupSuite.
type CustomerDependentTestSuite struct {
	E2ETestSuite
	CustomerID          string
	AssociatedPersonIDs []string
}

// SetupSuite creates or reuses a customer for the test suite.
// If an existing customer is found, it will be reused to speed up tests.
func (s *CustomerDependentTestSuite) SetupSuite() {
	s.E2ETestSuite.SetupSuite()

	// Try to reuse existing customer
	customerID, associatedPersonIDs, err := s.GetOrCreateTestCustomer()
	if err != nil {
		s.T().Fatalf("failed to get or create test customer: %v", err)
	}

	s.CustomerID = customerID
	s.AssociatedPersonIDs = associatedPersonIDs
}

// GetOrCreateTestCustomer returns an existing customer if available, otherwise creates a new one.
func (s *CustomerDependentTestSuite) GetOrCreateTestCustomer() (
	customerID string,
	associatedPersonIDs []string,
	err error,
) {
	// Try to find an existing customer
	listResp, err := s.Client.Customer.ListCustomers(s.Ctx, &customer.ListCustomersRequest{
		PageSize: 1,
	})
	if err == nil && listResp != nil && len(listResp.Customers) > 0 {
		existingCustomer := listResp.Customers[0]
		s.T().Logf("Reusing existing customer: %s (%s)", existingCustomer.CustomerID, existingCustomer.BusinessLegalName)

		// Get associated persons for the existing customer
		associatedPersonsResp, err := s.Client.Customer.ListAssociatedPersons(s.Ctx, existingCustomer.CustomerID)
		if err != nil {
			return "", nil, fmt.Errorf("ListAssociatedPersons failed: %w", err)
		}

		for i := range *associatedPersonsResp {
			associatedPersonIDs = append(associatedPersonIDs, (*associatedPersonsResp)[i].AssociatedPersonID)
		}

		return existingCustomer.CustomerID, associatedPersonIDs, nil
	}

	// No existing customer found, create a new one
	s.T().Log("No existing customer found, creating new test customer...")
	return s.CreateTestCustomer()
}

// TearDownSuite cleans up resources created during testing.
// This method cleans up auto conversion rules and external accounts.
func (s *CustomerDependentTestSuite) TearDownSuite() {
	if s.CustomerID == "" {
		return
	}

	s.T().Logf("Cleaning up resources for customer: %s", s.CustomerID)

	// Clean up auto conversion rules (soft delete)
	rules, err := s.Client.AutoConversionRules.ListRules(s.Ctx, s.CustomerID, nil)
	if err == nil && rules != nil {
		for i := range rules.Items {
			if rules.Items[i].Status == "ACTIVE" {
				if delErr := s.Client.AutoConversionRules.DeleteRule(s.Ctx, s.CustomerID, rules.Items[i].AutoConversionRuleID); delErr != nil {
					s.T().Logf("Failed to delete auto conversion rule %s: %v", rules.Items[i].AutoConversionRuleID, delErr)
				} else {
					s.T().Logf("Deleted auto conversion rule: %s", rules.Items[i].AutoConversionRuleID)
				}
			}
		}
	}

	// Clean up external accounts
	accounts, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
	if err == nil && accounts != nil {
		for i := range accounts {
			if err := s.Client.ExternalAccounts.RemoveExternalAccount(s.Ctx, s.CustomerID, accounts[i].ExternalAccountID); err != nil {
				s.T().Logf("Failed to delete external account %s: %v", accounts[i].ExternalAccountID, err)
			} else {
				s.T().Logf("Deleted external account: %s", accounts[i].ExternalAccountID)
			}
		}
	}

	s.T().Logf("Cleanup completed for customer: %s (Note: Customer cannot be deleted due to compliance requirements)", s.CustomerID)
}

// CreateTestCustomer creates a new customer with all required data for testing.
// Returns the customer ID, list of associated person IDs, and any error.
func (s *CustomerDependentTestSuite) CreateTestCustomer() (
	customerID string,
	associatedPersonIDs []string,
	err error,
) {
	faker := gofakeit.New(0)

	// Step 1: Create TOS link
	tosResp, err := s.Client.Customer.CreateTOSLink(s.Ctx, nil)
	if err != nil {
		return "", nil, fmt.Errorf("CreateTOSLink failed: %w", err)
	}

	// Step 2: Sign the agreement
	signResp, err := s.Client.Customer.SignTOSAgreement(s.Ctx, tosResp.SessionToken)
	if err != nil {
		return "", nil, fmt.Errorf("SignTOSAgreement failed: %w", err)
	}

	// Step 3: Create customer with associated persons
	associatedPersons := []customer.AssociatedPerson{
		FakeAssociatedPerson(faker),
		FakeAssociatedPerson(faker),
	}

	req := &customer.CreateCustomerRequest{
		BusinessLegalName:          faker.Company(),
		BusinessDescription:        faker.JobDescriptor() + " " + faker.BS(),
		BusinessRegistrationNumber: fmt.Sprintf("%s-%d", faker.LetterN(3), faker.Number(100000, 999999)),
		Email:                      faker.Email(),
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           "332999",
		RegisteredAddress: &customer.Address{
			StreetLine1: faker.Street(),
			StreetLine2: fmt.Sprintf("Suite %d", faker.Number(100, 999)),
			City:        faker.City(),
			State:       RandomUSState(faker),
			Country:     CountryUSA,
			PostalCode:  faker.Zip(),
			Subdivision: RandomUSState(faker),
		},
		DateOfIncorporation:            faker.Date().Format("2006-01-02"),
		SignedAgreementID:              signResp.SignedAgreementID,
		AssociatedPersons:              associatedPersons,
		SourceOfFunds:                  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth:                 []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},
		Documents:                      FakeCustomerDocuments(),
		AccountPurpose:                 customer.AccountPurposeTreasuryManagement,
		IsDAO:                          false,
		PubliclyTraded:                 false,
		EstimatedAnnualRevenueUSD:      customer.MoneyRange099999,
		ExpectedMonthlyFiatDeposits:    customer.MoneyRange099999,
		ExpectedMonthlyFiatWithdrawals: customer.MoneyRange099999,
		TaxID:                          fmt.Sprintf("%d-%d", faker.Number(10, 99), faker.Number(1000000, 9999999)),
		TaxType:                        customer.TaxIDTypeEIN,
		TaxCountry:                     CountryUSA,
	}

	resp, err := s.Client.Customer.CreateCustomer(s.Ctx, req)
	if err != nil {
		return "", nil, fmt.Errorf("CreateCustomer failed: %w", err)
	}

	// Get associated person IDs from the created customer
	associatedPersonsResp, err := s.Client.Customer.ListAssociatedPersons(s.Ctx, resp.CustomerID)
	if err != nil {
		return "", nil, fmt.Errorf("ListAssociatedPersons failed: %w", err)
	}

	for i := range *associatedPersonsResp {
		associatedPersonIDs = append(associatedPersonIDs, (*associatedPersonsResp)[i].AssociatedPersonID)
	}

	customerID = resp.CustomerID
	return customerID, associatedPersonIDs, nil
}

// EnsureExternalAccount ensures an approved external account exists for the customer.
// If no external account exists, it creates one and polls until it reaches APPROVED status.
// Returns the external account ID or an error if the account fails approval or times out.
func (s *CustomerDependentTestSuite) EnsureExternalAccount() (string, error) {
	const (
		pollInterval = 2 * time.Second
		maxWaitTime  = 10 * time.Second
	)

	// Try to get existing external accounts
	accounts, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
	if err != nil {
		return "", fmt.Errorf("ListExternalAccounts failed: %w", err)
	}

	var accountID string

	// If we have an approved account, return it
	for _, acc := range accounts {
		if acc.Status == string(external_accounts.BankAccountStatusAPPROVED) {
			return acc.ExternalAccountID, nil
		}
		// Remember a pending account to poll
		if acc.Status == string(external_accounts.BankAccountStatusPENDING) && accountID == "" {
			accountID = acc.ExternalAccountID
		}
	}

	// Create a new external account if none exists
	if accountID == "" {
		createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, s.CustomerID, FakeExternalAccountRequest())
		if err != nil {
			return "", fmt.Errorf("CreateExternalAccount failed: %w", err)
		}
		accountID = createResp.ExternalAccountID
	}

	// Poll until approved or failed
	deadline := time.Now().Add(maxWaitTime)
	for time.Now().Before(deadline) {
		acc, err := s.Client.ExternalAccounts.GetExternalAccount(s.Ctx, s.CustomerID, accountID)
		if err != nil {
			return "", fmt.Errorf("GetExternalAccount failed: %w", err)
		}

		switch acc.Status {
		case string(external_accounts.BankAccountStatusAPPROVED):
			return accountID, nil
		case string(external_accounts.BankAccountStatusFAILED):
			return "", fmt.Errorf("external account %s approval failed", accountID)
		}

		time.Sleep(pollInterval)
	}

	return "", fmt.Errorf("external account %s approval timed out after %v", accountID, maxWaitTime)
}

// EnsureTransaction ensures at least one transaction exists for the customer.
// Preferred order:
//  1. Reuse existing transaction if available
//  2. Try to create a simulated USD deposit and reuse its transaction if available
//
// If neither produces a transaction, an error is returned and callers may choose
// to skip tests that require persisted transaction history.
func (s *CustomerDependentTestSuite) EnsureTransaction() (string, error) {
	// Try to get existing transactions
	txResp, err := s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, nil)
	if err != nil {
		return "", fmt.Errorf("ListTransactions failed: %w", err)
	}

	// If we have transactions, return the first one
	if len(txResp.List) > 0 {
		return txResp.List[0].TransactionID, nil
	}

	// Try to create a simulated USD deposit; in some environments this may not
	// produce a persisted transaction, so we treat it as best-effort.
	simReq := &simulations.SimulateDepositRequest{
		Asset:  assets.AssetNameUSD,
		Amount: "100.00",
	}

	_, simErr := s.Client.Simulations.SimulateDeposit(s.Ctx, s.CustomerID, simReq)
	if simErr == nil {
		txResp, err = s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, nil)
		if err == nil && len(txResp.List) > 0 {
			return txResp.List[0].TransactionID, nil
		}
	}

	return "", fmt.Errorf("no transactions available for customer %s (simulated deposit did not create a transaction)", s.CustomerID)
}

// EnsureAutoConversionRule ensures an auto conversion rule exists for the customer.
// If no rule exists, it creates one and returns the ID.
func (s *CustomerDependentTestSuite) EnsureAutoConversionRule() (string, error) {
	// Try to get existing rules
	rules, err := s.Client.AutoConversionRules.ListRules(s.Ctx, s.CustomerID, nil)
	if err != nil {
		return "", fmt.Errorf("ListRules failed: %w", err)
	}

	// If we have an active rule, return it
	for i := range rules.Items {
		if rules.Items[i].Status == "ACTIVE" {
			return rules.Items[i].AutoConversionRuleID, nil
		}
	}

	// Create a new auto conversion rule using fake data
	createResp, err := s.Client.AutoConversionRules.CreateRule(s.Ctx, s.CustomerID, FakeAutoConversionRuleRequest())
	if err != nil {
		return "", fmt.Errorf("CreateRule failed: %w", err)
	}

	return createResp.AutoConversionRuleID, nil
}

// EnsureSignedAgreement creates a TOS link and signs it, returning the SignedAgreementID.
func (s *CustomerDependentTestSuite) EnsureSignedAgreement() (string, error) {
	// Create TOS link
	tosResp, err := s.Client.Customer.CreateTOSLink(s.Ctx, &customer.CreateTOSLinkRequest{
		RedirectUri: "https://example.com/redirect",
	})
	if err != nil {
		return "", fmt.Errorf("CreateTOSLink failed: %w", err)
	}

	// Sign the agreement
	signResp, err := s.Client.Customer.SignTOSAgreement(s.Ctx, tosResp.SessionToken)
	if err != nil {
		return "", fmt.Errorf("SignTOSAgreement failed: %w", err)
	}

	return signResp.SignedAgreementID, nil
}

// FakeExternalAccountRequest generates a fake external account request for testing.
// Uses uuid.New() for IdempotencyKey to ensure uniqueness across test runs.
func FakeExternalAccountRequest() *external_accounts.CreateReq {
	return &external_accounts.CreateReq{
		IdempotencyKey: uuid.New().String(),
		Network:        external_accounts.BankNetworkNameUSACH,
		Currency:       external_accounts.CurrencyUSD,
		CountryCode:    external_accounts.CountryCodeUSA,
		// https://qodex.ai/all-tools/routing-number-generator
		AccountNumber:   "5097935393",
		InstitutionID:   "327984566",
		InstitutionName: gofakeit.Company() + " Bank",
	}
}

// FakeEthereumAddress generates a fake Ethereum wallet address for testing.
// Returns a valid 42-character address (0x + 40 hex chars).
func FakeEthereumAddress() string {
	addrBytes := make([]byte, 20)
	for i := range addrBytes {
		addrBytes[i] = byte(gofakeit.Number(0, 255))
	}
	return fmt.Sprintf("0x%x", addrBytes)
}

// safeUint8 converts an int to uint8 with bounds checking to avoid overflow.
func safeUint8(n int) uint8 {
	if n < 0 {
		return 0
	}
	if n > 255 {
		return 255
	}
	return uint8(n)
}

// FakeImagePNG generates a valid PNG image as bytes.
// Uses Go's image package to create a real PNG image.
func FakeImagePNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a random color
	c := color.RGBA{
		R: safeUint8(gofakeit.Number(0, 255)),
		G: safeUint8(gofakeit.Number(0, 255)),
		B: safeUint8(gofakeit.Number(0, 255)),
		A: 255,
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
	return []customer.Document{
		{
			DocType:     customer.DocumentTypeFlowOfFunds,
			File:        customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
			Description: "Proof of Funds",
		},
		{
			DocType:     customer.DocumentTypeRegistrationDocument,
			File:        customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
			Description: "Certificate of Incorporation",
		},
		{
			DocType:     customer.DocumentTypeProofOfTaxIdentification,
			File:        customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
			Description: "W9 Form",
		},
		{
			DocType:     customer.DocumentTypeShareholderRegister,
			File:        customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
			Description: "Ownership Structure",
		},
		{
			DocType:     customer.DocumentTypeESignatureCertificate,
			File:        customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
			Description: "Authorized Representative List",
		},
		{
			DocType:     customer.DocumentTypeEvidenceOfGoodStanding,
			File:        customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
			Description: "Evidence of Good Standing",
		},
		{
			DocType:     customer.DocumentTypeProofOfAddress,
			File:        customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
			Description: "Proof of Address",
		},
	}
}

// TestClient_Initialization tests client initialization.
func (s *E2ETestSuite) TestClient_Initialization() {
	s.Require().NotNil(s.Client, "Client should not be nil")
	s.Require().NotNil(s.Client.Assets, "Assets service should be initialized")
	s.Require().NotNil(s.Client.AutoConversionRules, "AutoConversionRules service should be initialized")
	s.Require().NotNil(s.Client.Conversions, "Conversions service should be initialized")
	s.Require().NotNil(s.Client.Customer, "Customer service should be initialized")
	s.Require().NotNil(s.Client.Echo, "Echo service should be initialized")
	s.Require().NotNil(s.Client.ExternalAccounts, "ExternalAccounts service should be initialized")
	s.Require().NotNil(s.Client.Instructions, "Instructions service should be initialized")
	s.Require().NotNil(s.Client.Simulations, "Simulations service should be initialized")
	s.Require().NotNil(s.Client.Transactions, "Transactions service should be initialized")
	s.Require().NotNil(s.Client.Withdrawals, "Withdrawals service should be initialized")
	s.NotEmpty(s.Client.Version(), "Version should not be empty")
}

// TestE2ETestSuite runs the base E2E test suite.
func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

// PrettyJSON formats any value as indented JSON string.
func PrettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}

// FakeXLSXData generates a simple XLSX file as bytes for testing.
func FakeXLSXData() []byte {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	// Add some test data
	_ = f.SetCellValue("Sheet1", "A1", "Name")
	_ = f.SetCellValue("Sheet1", "B1", "Value")
	_ = f.SetCellValue("Sheet1", "A2", "Test")
	_ = f.SetCellValue("Sheet1", "B2", "Data")

	var buf bytes.Buffer
	_ = f.Write(&buf)
	return buf.Bytes()
}

// FakeAutoConversionRuleRequest generates a fake auto conversion rule request for testing.
// Creates a USD -> USDC (Polygon) conversion rule by default.
// Uses uuid.New() for IdempotencyKey to ensure uniqueness across test runs.
func FakeAutoConversionRuleRequest() *auto_conversion_rules.CreateRuleRequest {
	network := "POLYGON"
	return &auto_conversion_rules.CreateRuleRequest{
		IdempotencyKey: uuid.New().String(),
		Source: auto_conversion_rules.SourceAssetInfo{
			Asset:   "USD",
			Network: "ACH",
		},
		Destination: auto_conversion_rules.DestinationAssetInfo{
			Asset:   "USDC",
			Network: &network,
		},
	}
}

// FakeAssociatedPerson generates a fake associated person for testing.
func FakeAssociatedPerson(faker *gofakeit.Faker) customer.AssociatedPerson {
	// Randomly select gender
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
				ImageFront:             customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
				ImageBack:              customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
				NationalIdentityNumber: faker.LetterN(8) + faker.DigitN(4),
			},
		},
		CountryOfTax: CountryUSA,
		TaxType:      customer.TaxIDTypeSSN,
		TaxID:        faker.SSN(),
		POA:          customer.EncodeBase64ToDataURI(FakeImagePNG(100, 100), customer.ImageFormatPng),
		POAType:      "utility_bill",
	}
}
