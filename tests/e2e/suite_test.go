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
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
)

const (
	// testCustomerID is a test customer ID used across multiple tests.
	testCustomerID = "ee71219c-1baa-4b8c-b860-08822ea53b8e"
	// testAssociatedPersonID is a test associated person ID used across multiple tests.
	testAssociatedPersonID = "96d2727d-4373-4b21-a60c-dae81d763902"
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

// TestClient_Initialization tests client initialization.
func (s *E2ETestSuite) TestClient_Initialization() {
	s.Require().NotNil(s.Client, "Client should not be nil")
	s.Require().NotNil(s.Client.Assets, "Assets service should be initialized")
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

// FakeAssociatedPerson generates a fake associated person for testing.
func FakeAssociatedPerson(faker *gofakeit.Faker) customer.AssociatedPerson {
	return customer.AssociatedPerson{
		FirstName: faker.FirstName(),
		LastName:  faker.LastName(),
		Email:     faker.Email(),
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
				Type:           customer.IDTypeDriversLicense,
				IssuingCountry: CountryUSA,
				ImageFront:     customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				ImageBack:      customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
			},
		},
		CountryOfTax: CountryUSA,
		TaxType:      customer.TaxIDTypeSSN,
		TaxID:        faker.SSN(),
		POA:          customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
	}
}
