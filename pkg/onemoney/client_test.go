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

package onemoney

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
)

// ClientTestSuite defines the integration test suite for the OneMoney client.
type ClientTestSuite struct {
	suite.Suite
	client *Client
	ctx    context.Context
}

// prettyJSON formats any value as indented JSON string.
func prettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}

// SetupSuite runs once before all tests in the suite.
func (s *ClientTestSuite) SetupSuite() {
	// Load environment variables from .env file if present
	_ = godotenv.Load()

	// Create client configuration
	cfg := &Config{
		BaseURL:   os.Getenv("ONEMONEY_BASE_URL"),
		AccessKey: os.Getenv("ONEMONEY_ACCESS_KEY"),
		SecretKey: os.Getenv("ONEMONEY_SECRET_KEY"),
		Timeout:   30 * time.Second,
	}

	// Skip tests if required environment variables are missing
	if cfg.BaseURL == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		s.T().Skipf("missing required environment variables (ONEMONEY_BASE_URL, ONEMONEY_ACCESS_KEY, ONEMONEY_SECRET_KEY)")
	}

	// Create client
	client, err := NewClient(cfg)
	if err != nil {
		s.T().Fatalf("failed to create client: %v", err)
	}

	s.client = client
	s.ctx = context.Background()
}

// SetupTest runs before each test.
func (*ClientTestSuite) SetupTest() {
	// Reset state if needed
}

// TearDownTest runs after each test.
func (*ClientTestSuite) TearDownTest() {
	// Cleanup if needed
}

// TearDownSuite runs once after all tests.
func (*ClientTestSuite) TearDownSuite() {
	// Final cleanup
}

// TestClient_Initialization tests client initialization.
func (s *ClientTestSuite) TestClient_Initialization() {
	// Assert
	s.Require().NotNil(s.client, "Client should not be nil")
	s.Require().NotNil(s.client.Echo, "Echo service should be initialized")
	s.Require().NotNil(s.client.Customer, "Customer service should be initialized")
	s.NotEmpty(s.client.Version(), "Version should not be empty")
}

// TestCustomerService_CreateCustomer tests customer creation.
func (s *ClientTestSuite) TestCustomerService_CreateCustomer() {
	// Arrange - Generate fake data using gofakeit
	faker := gofakeit.New(0)

	// Create at least one associated person
	associatedPerson := customer.AssociatedPerson{
		FirstName: faker.FirstName(),
		LastName:  faker.LastName(),
		Email:     faker.Email(),
		ResidentialAddress: &customer.Address{
			StreetLine1: faker.Street(),
			City:        faker.City(),
			State:       faker.StateAbr(),
			Country:     faker.Country(),
			PostalCode:  faker.Zip(),
		},
		BirthDate:           faker.Date().Format("2006-01-02"),
		CountryOfBirth:      faker.Country(),
		Gender:              customer.GenderM,
		PrimaryNationality:  faker.Country(),
		HasOwnership:        true,
		OwnershipPercentage: 100,
		HasControl:          true,
		IsSigner:            true,
		IsDirector:          true,
		IdentifyingInformation: []customer.IdentifyingInformation{
			{
				Type:           customer.IDTypeDriversLicense,
				IssuingCountry: "USA",
				ImageFront:     customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				ImageBack:      customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
			},
		},
		CountryOfTax: faker.Country(),
		TaxType:      customer.TaxIDTypeEIN,
		TaxIDNumber:  fmt.Sprintf("%d-%d", faker.Number(10, 99), faker.Number(1000000, 9999999)),
		POA:          customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg), // POA is required for directors and beneficial owners
	}

	req := &customer.CreateCustomerRequest{
		BusinessLegalName:          faker.Company(),
		BusinessDescription:        faker.JobDescriptor() + " " + faker.BS(),
		BusinessRegistrationNumber: fmt.Sprintf("%s-%d", faker.LetterN(3), faker.Number(100000, 999999)),
		Email:                      faker.Email(),
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           customer.BusinessIndustryTechnologyECommercePlatforms,
		RegisteredAddress: &customer.Address{
			StreetLine1: faker.Street(),
			StreetLine2: fmt.Sprintf("Suite %d", faker.Number(100, 999)),
			City:        faker.City(),
			State:       faker.StateAbr(),
			Country:     faker.Country(),
			PostalCode:  faker.Zip(),
			Subdivision: faker.State(),
		},
		DateOfIncorporation: faker.Date().Format("2006-01-02"),
		SignedAgreementID:   faker.UUID(),
		AssociatedPersons:   []customer.AssociatedPerson{associatedPerson},
		SourceOfFunds:       []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth:      []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},
		Documents: []customer.Document{
			{
				DocType:     customer.DocumentTypeCertificateOfIncorporation,
				File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				Description: "Certificate of Incorporation",
			},
			{
				DocType:     customer.DocumentTypeCertificateOfGoodStanding,
				File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				Description: "Certificate of Good Standing",
			},
			{
				DocType:     customer.DocumentTypeProofOfSourceOfFunds,
				File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				Description: "Proof of Source of Funds",
			},
			{
				DocType:     customer.DocumentTypeAuthorizedRepresentativeList,
				File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				Description: "Authorized Representative List",
			},
			{
				DocType:     customer.DocumentTypeOwnershipStructureCorp,
				File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				Description: "Ownership Structure - Corporation",
			},
			{
				DocType:     customer.DocumentTypeProofOfBusinessEntityAddress,
				File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				Description: "Proof of Business Entity Address",
			},
			{
				DocType:     customer.DocumentTypeCertificateOfIncumbencyOrRegisterOfDirectors,
				File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				Description: "Certificate of Incumbency",
			},
			{
				DocType:     customer.DocumentTypeMemorandumOfAssociationOrArticleOfAssociationOrEquivalentDocument,
				File:        customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				Description: "Memorandum of Association",
			},
		},
		AccountPurpose:                 customer.AccountPurposeTreasuryManagement,
		IsDAO:                          false,
		PubliclyTraded:                 false,
		EstimatedAnnualRevenueUSD:      customer.MoneyRange099999,
		ExpectedMonthlyFiatDeposits:    customer.MoneyRange099999,
		ExpectedMonthlyFiatWithdrawals: customer.MoneyRange099999,
		ConductsMoneyServices:          false,
		TaxID:                          fmt.Sprintf("%d-%d", faker.Number(10, 99), faker.Number(1000000, 9999999)),
		TaxType:                        customer.TaxIDTypeEIN,
	}

	// Act
	resp, err := s.client.Customer.CreateCustomer(s.ctx, req)

	// Assert
	s.Require().NoError(err, "CreateCustomer should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.ID, "Customer ID should not be empty")
	s.Equal(req.BusinessLegalName, resp.BusinessLegalName, "Business name should match")
	s.Equal(req.Email, resp.Email, "Customer email should match")
	s.Equal(req.BusinessType, resp.BusinessType, "Business type should match")
	s.NotEmpty(resp.Status, "Status should not be empty")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.UpdatedAt, "UpdatedAt should not be empty")
}

// TestCustomerService_ListCustomers tests listing customers.
func (s *ClientTestSuite) TestCustomerService_ListCustomers() {
	// Arrange
	req := &customer.ListCustomersRequest{
		Page:     0,
		PageSize: 10,
	}

	// Act
	resp, err := s.client.Customer.ListCustomers(s.ctx, req)

	// Assert
	s.Require().NoError(err, "ListCustomers should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.GreaterOrEqual(resp.Total, 0, "Total should be non-negative")
	s.NotNil(resp.Data, "Data should not be nil")

	s.T().Logf("List customers response:\n%s", prettyJSON(resp))

	// If there are customers, verify structure
	if len(resp.Data) > 0 {
		firstCustomer := resp.Data[0]
		s.NotEmpty(firstCustomer.ID, "Customer ID should not be empty")
		s.NotEmpty(firstCustomer.BusinessLegalName, "Customer business name should not be empty")
		s.NotEmpty(firstCustomer.Email, "Customer email should not be empty")
		s.NotEmpty(firstCustomer.BusinessType, "Customer business type should not be empty")
		s.NotEmpty(firstCustomer.Status, "Customer status should not be empty")
		s.NotEmpty(firstCustomer.CreatedAt, "CreatedAt should not be empty")
		s.NotEmpty(firstCustomer.UpdatedAt, "UpdatedAt should not be empty")
	}
}

// TestClientTestSuite runs the test suite.
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
