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
	"path/filepath"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
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

// SetupSuite creates a new customer for the test suite.
func (s *CustomerDependentTestSuite) SetupSuite() {
	s.E2ETestSuite.SetupSuite()

	customerID, associatedPersonIDs, err := s.CreateTestCustomer()
	if err != nil {
		s.T().Fatalf("failed to create test customer: %v", err)
	}

	s.CustomerID = customerID
	s.AssociatedPersonIDs = associatedPersonIDs
	s.T().Logf("Created test customer: %s with %d associated persons", customerID, len(associatedPersonIDs))
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
	tosResp, err := s.Client.Customer.CreateTOSLink(s.Ctx)
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
			State:       faker.StateAbr(),
			Country:     CountryUSA,
			PostalCode:  faker.Zip(),
			Subdivision: faker.StateAbr(),
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

// EnsureExternalAccount ensures an external account exists for the customer.
// If no external account exists, it creates one and returns the ID.
func (s *CustomerDependentTestSuite) EnsureExternalAccount() (string, error) {
	// Try to get existing external accounts
	accounts, err := s.Client.ExternalAccounts.ListExternalAccounts(s.Ctx, s.CustomerID, nil)
	if err != nil {
		return "", fmt.Errorf("ListExternalAccounts failed: %w", err)
	}

	// If we have an account, return the first one
	if len(accounts) > 0 {
		return accounts[0].ExternalAccountID, nil
	}

	// Create a new external account using fake data
	createResp, err := s.Client.ExternalAccounts.CreateExternalAccount(s.Ctx, s.CustomerID, FakeExternalAccountRequest())
	if err != nil {
		return "", fmt.Errorf("CreateExternalAccount failed: %w", err)
	}

	return createResp.ExternalAccountID, nil
}

// EnsureTransaction ensures at least one transaction exists for the customer.
// If no transactions exist, it simulates a USD deposit and returns the transaction ID.
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

	// Simulate a deposit to create a transaction
	simReq := &simulations.SimulateDepositRequest{
		Asset:  assets.AssetNameUSD,
		Amount: "100.00",
	}

	_, err = s.Client.Simulations.SimulateDeposit(s.Ctx, s.CustomerID, simReq)
	if err != nil {
		return "", fmt.Errorf("SimulateDeposit failed: %w", err)
	}

	// Get the newly created transaction
	txResp, err = s.Client.Transactions.ListTransactions(s.Ctx, s.CustomerID, nil)
	if err != nil {
		return "", fmt.Errorf("ListTransactions after deposit failed: %w", err)
	}

	if len(txResp.List) == 0 {
		return "", fmt.Errorf("no transactions found after simulating deposit")
	}

	return txResp.List[0].TransactionID, nil
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
	tosResp, err := s.Client.Customer.CreateTOSLink(s.Ctx)
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
func FakeExternalAccountRequest() *external_accounts.CreateExternalAccountRequest {
	faker := gofakeit.New(0)
	return &external_accounts.CreateExternalAccountRequest{
		IdempotencyKey:       faker.UUID(),
		BankNetworkName:      external_accounts.BankNetworkNameUSACH,
		Currency:             external_accounts.CurrencyUSD,
		BankName:             faker.Company() + " Bank",
		BankAccountOwnerName: faker.Name(),
		BankAccountNumber:    faker.DigitN(9),
		BankRoutingNumber:    faker.DigitN(9),
	}
}

// FakeEthereumAddress generates a fake Ethereum wallet address for testing.
func FakeEthereumAddress() string {
	faker := gofakeit.New(0)
	return "0x" + faker.HexUint(160)
}

// FakeCustomerDocuments generates fake documents required for customer creation.
func FakeCustomerDocuments() []customer.Document {
	return []customer.Document{
		{
			DocType:     customer.DocumentTypeFlowOfFunds,
			File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
			Description: "Proof of Funds",
		},
		{
			DocType:     customer.DocumentTypeRegistrationDocument,
			File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
			Description: "Certificate of Incorporation",
		},
		{
			DocType:     customer.DocumentTypeProofOfTaxIdentification,
			File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
			Description: "W9 Form",
		},
		{
			DocType:     customer.DocumentTypeShareholderRegister,
			File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
			Description: "Ownership Structure",
		},
		{
			DocType:     customer.DocumentTypeESignatureCertificate,
			File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
			Description: "Authorized Representative List",
		},
		{
			DocType:     customer.DocumentTypeEvidenceOfGoodStanding,
			File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
			Description: "Evidence of Good Standing",
		},
		{
			DocType:     customer.DocumentTypeProofOfAddress,
			File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
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
func FakeAutoConversionRuleRequest() *auto_conversion_rules.CreateRuleRequest {
	faker := gofakeit.New(0)
	network := "POLYGON"
	return &auto_conversion_rules.CreateRuleRequest{
		IdempotencyKey: faker.UUID(),
		Source: auto_conversion_rules.SourceAssetInfo{
			Asset:   "USD",
			Network: "US_ACH",
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
			State:       faker.StateAbr(),
			Country:     CountryUSA,
			PostalCode:  faker.Zip(),
			Subdivision: faker.StateAbr(),
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
				ImageFront:             customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				ImageBack:              customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				NationalIdentityNumber: faker.LetterN(8) + faker.DigitN(4),
			},
		},
		CountryOfTax: CountryUSA,
		TaxType:      customer.TaxIDTypeSSN,
		TaxID:        faker.SSN(),
		POA:          customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
		POAType:      "utility_bill",
	}
}
