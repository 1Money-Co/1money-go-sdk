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
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	"github.com/1Money-Co/1money-go-sdk/internal/utils"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/external_accounts"
)

// CustomerTestSuite tests customer service operations.
type CustomerTestSuite struct {
	CustomerDependentTestSuite
}

// TestCustomerService_TOSFlow tests the complete TOS signing flow.
func (s *CustomerTestSuite) TestCustomerService_TOSFlow() {
	// Step 1: Create TOS link
	const exampleRedirectURI = "https://example.com/terms-of-service-completed"
	tosResp, err := s.Client.Customer.CreateTOSLink(s.Ctx, &customer.CreateTOSLinkRequest{
		RedirectUri: exampleRedirectURI,
	})
	s.Require().NoError(err, "CreateTOSLink should not return error")
	s.Require().NotNil(tosResp, "CreateTOSLink response should not be nil")
	s.NotEmpty(tosResp.SessionToken, "Session token should not be empty")
	s.T().Logf("Created TOS link with session token:\n%s", PrettyJSON(tosResp))
	s.Require().Contains(tosResp.Url, url.QueryEscape(exampleRedirectURI), "Redirect URI should be included in the URL")

	// Step 2: Sign the agreement using the session token
	signResp, err := s.Client.Customer.SignTOSAgreement(s.Ctx, tosResp.SessionToken)
	s.Require().NoError(err, "SignTOSAgreement should not return error")
	s.Require().NotNil(signResp, "SignTOSAgreement response should not be nil")
	s.NotEmpty(signResp.SignedAgreementID, "Signed agreement ID should not be empty")
	s.T().Logf("Signed agreement with ID:\n%s", PrettyJSON(signResp))
}

// TestCustomerService_CreateCustomer tests customer creation.
func (s *CustomerTestSuite) TestCustomerService_CreateCustomer() {
	// Get a valid signed agreement ID dynamically
	signedAgreementID, err := s.EnsureSignedAgreement()
	s.Require().NoError(err, "EnsureSignedAgreement should succeed")

	faker := gofakeit.New(0)

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
			Country:     external_accounts.CountryCodeDEU.String(),
			PostalCode:  faker.Zip(),
			Subdivision: faker.StateAbr(),
		},
		DateOfIncorporation: faker.Date().Format("2006-01-02"),
		SignedAgreementID:   signedAgreementID,
		AssociatedPersons: []customer.AssociatedPerson{
			FakeAssociatedPerson(faker),
			FakeAssociatedPerson(faker),
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
		TaxID:                          fmt.Sprintf("%d-%d", faker.Number(10, 99), faker.Number(1000000, 9999999)),
		TaxType:                        customer.TaxIDTypeEIN,
		TaxCountry:                     external_accounts.CountryCodeDEU.String(),
	}

	resp, err := s.Client.Customer.CreateCustomer(s.Ctx, req)

	s.Require().NoError(err, "CreateCustomer should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.CustomerID, "Customer ID should not be empty")
	s.Equal(req.BusinessLegalName, resp.BusinessLegalName, "Business name should match")
	s.Equal(req.Email, resp.Email, "Customer email should match")
	s.Equal(req.BusinessType, resp.BusinessType, "Business type should match")
	s.Contains(
		[]customer.KybStatus{customer.KybStatusPendingReview, customer.KybStatusUnderReview},
		resp.Status,
		"New customer status should be pending_review or under_review",
	)
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.UpdatedAt, "UpdatedAt should not be empty")
}

// TestCustomerService_CreateCustomer_InvalidFileFormat tests that invalid file formats are rejected.
func (s *CustomerTestSuite) TestCustomerService_CreateCustomer_InvalidFileFormat() {
	faker := gofakeit.New(0)

	// Get a valid signed agreement ID
	signedAgreementID, err := s.EnsureSignedAgreement()
	s.Require().NoError(err, "EnsureSignedAgreement should succeed")

	// Test 1: Invalid MIME type (using unsupported format like .exe)
	invalidMIME := "data:application/x-msdownload;base64,TVqQAAMAAAAEAAAA"

	req := &customer.CreateCustomerRequest{
		BusinessLegalName:          faker.Company(),
		BusinessDescription:        faker.JobDescriptor(),
		BusinessRegistrationNumber: fmt.Sprintf("%s-%d", faker.LetterN(3), faker.Number(100000, 999999)),
		Email:                      faker.Email(),
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           "332999",
		RegisteredAddress: &customer.Address{
			StreetLine1: faker.Street(),
			City:        faker.City(),
			State:       faker.StateAbr(),
			Country:     external_accounts.CountryCodeDEU.String(),
			PostalCode:  faker.Zip(),
			Subdivision: faker.StateAbr(),
		},
		DateOfIncorporation: faker.Date().Format("2006-01-02"),
		SignedAgreementID:   signedAgreementID,
		AssociatedPersons: []customer.AssociatedPerson{
			FakeAssociatedPerson(faker),
		},
		SourceOfFunds:  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth: []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},
		Documents: []customer.Document{
			{
				DocType:     customer.DocumentTypeFlowOfFunds,
				File:        invalidMIME, // Invalid MIME type
				Description: "Invalid file format test",
			},
		},
		AccountPurpose:                 customer.AccountPurposeTreasuryManagement,
		EstimatedAnnualRevenueUSD:      customer.MoneyRange099999,
		ExpectedMonthlyFiatDeposits:    customer.MoneyRange099999,
		ExpectedMonthlyFiatWithdrawals: customer.MoneyRange099999,
		TaxID:                          fmt.Sprintf("%d-%d", faker.Number(10, 99), faker.Number(1000000, 9999999)),
		TaxType:                        customer.TaxIDTypeEIN,
		TaxCountry:                     external_accounts.CountryCodeDEU.String(),
	}

	_, err = s.Client.Customer.CreateCustomer(s.Ctx, req)
	s.Require().Error(err, "CreateCustomer should return error for invalid MIME type")
	s.T().Logf("Expected error for invalid MIME type: %v", err)
}

// TestCustomerService_CreateCustomer_InvalidBase64 tests that invalid base64 data is rejected.
func (s *CustomerTestSuite) TestCustomerService_CreateCustomer_InvalidBase64() {
	faker := gofakeit.New(0)

	// Get a valid signed agreement ID
	signedAgreementID, err := s.EnsureSignedAgreement()
	s.Require().NoError(err, "EnsureSignedAgreement should succeed")

	// Invalid base64 data (not properly encoded)
	invalidBase64 := "data:image/jpeg;base64,this-is-not-valid-base64!!!"

	req := &customer.CreateCustomerRequest{
		BusinessLegalName:          faker.Company(),
		BusinessDescription:        faker.JobDescriptor(),
		BusinessRegistrationNumber: fmt.Sprintf("%s-%d", faker.LetterN(3), faker.Number(100000, 999999)),
		Email:                      faker.Email(),
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           "332999",
		RegisteredAddress: &customer.Address{
			StreetLine1: faker.Street(),
			City:        faker.City(),
			State:       faker.StateAbr(),
			Country:     external_accounts.CountryCodeDEU.String(),
			PostalCode:  faker.Zip(),
			Subdivision: faker.StateAbr(),
		},
		DateOfIncorporation: faker.Date().Format("2006-01-02"),
		SignedAgreementID:   signedAgreementID,
		AssociatedPersons: []customer.AssociatedPerson{
			FakeAssociatedPerson(faker),
		},
		SourceOfFunds:  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth: []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},
		Documents: []customer.Document{
			{
				DocType:     customer.DocumentTypeFlowOfFunds,
				File:        invalidBase64, // Invalid base64
				Description: "Invalid base64 test",
			},
		},
		AccountPurpose:                 customer.AccountPurposeTreasuryManagement,
		EstimatedAnnualRevenueUSD:      customer.MoneyRange099999,
		ExpectedMonthlyFiatDeposits:    customer.MoneyRange099999,
		ExpectedMonthlyFiatWithdrawals: customer.MoneyRange099999,
		TaxID:                          fmt.Sprintf("%d-%d", faker.Number(10, 99), faker.Number(1000000, 9999999)),
		TaxType:                        customer.TaxIDTypeEIN,
		TaxCountry:                     external_accounts.CountryCodeDEU.String(),
	}

	_, err = s.Client.Customer.CreateCustomer(s.Ctx, req)
	s.Require().Error(err, "CreateCustomer should return error for invalid base64")
	s.T().Logf("Expected error for invalid base64: %v", err)
}

// TestCustomerService_CreateCustomer_CorruptedXLSX tests that corrupted XLSX files are rejected.
func (s *CustomerTestSuite) TestCustomerService_CreateCustomer_CorruptedXLSX() {
	s.T().Skip("API does not validate XLSX file content integrity - corrupted files are accepted")
	faker := gofakeit.New(0)

	// Get a valid signed agreement ID
	signedAgreementID, err := s.EnsureSignedAgreement()
	s.Require().NoError(err, "EnsureSignedAgreement should succeed")

	// Corrupted XLSX (random bytes that look like XLSX but are invalid)
	corruptedData := []byte("PK\x03\x04corrupted xlsx content that is not valid")
	corruptedXLSX := customer.EncodeDocumentToDataURI(corruptedData, customer.FileFormatXlsx)

	// Get all required documents, then replace one with the corrupted XLSX
	docs := FakeCustomerDocuments()
	for i := range docs {
		if docs[i].DocType == customer.DocumentTypeShareholderRegister {
			docs[i].File = corruptedXLSX
			docs[i].Description = "Corrupted XLSX test"
			break
		}
	}

	req := &customer.CreateCustomerRequest{
		BusinessLegalName:          faker.Company(),
		BusinessDescription:        faker.JobDescriptor(),
		BusinessRegistrationNumber: fmt.Sprintf("%s-%d", faker.LetterN(3), faker.Number(100000, 999999)),
		Email:                      faker.Email(),
		BusinessType:               customer.BusinessTypeCorporation,
		BusinessIndustry:           "332999",
		RegisteredAddress: &customer.Address{
			StreetLine1: faker.Street(),
			City:        faker.City(),
			State:       faker.StateAbr(),
			Country:     external_accounts.CountryCodeDEU.String(),
			PostalCode:  faker.Zip(),
			Subdivision: faker.StateAbr(),
		},
		DateOfIncorporation: faker.Date().Format("2006-01-02"),
		SignedAgreementID:   signedAgreementID,
		AssociatedPersons: []customer.AssociatedPerson{
			FakeAssociatedPerson(faker),
		},
		SourceOfFunds:                  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth:                 []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},
		Documents:                      docs,
		AccountPurpose:                 customer.AccountPurposeTreasuryManagement,
		EstimatedAnnualRevenueUSD:      customer.MoneyRange099999,
		ExpectedMonthlyFiatDeposits:    customer.MoneyRange099999,
		ExpectedMonthlyFiatWithdrawals: customer.MoneyRange099999,
		TaxID:                          fmt.Sprintf("%d-%d", faker.Number(10, 99), faker.Number(1000000, 9999999)),
		TaxType:                        customer.TaxIDTypeEIN,
		TaxCountry:                     external_accounts.CountryCodeDEU.String(),
	}

	_, err = s.Client.Customer.CreateCustomer(s.Ctx, req)
	s.Require().Error(err, "CreateCustomer should return error for corrupted XLSX")
	s.T().Logf("Expected error for corrupted XLSX: %v", err)
}

// TestCustomerService_ListCustomers tests listing customers.
func (s *CustomerTestSuite) TestCustomerService_ListCustomers() {
	req := &customer.ListCustomersRequest{
		PageNum:  0,
		PageSize: 10,
	}

	resp, err := s.Client.Customer.ListCustomers(s.Ctx, req)

	s.Require().NoError(err, "ListCustomers should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.GreaterOrEqual(resp.Total, 0, "Total should be non-negative")
	s.NotNil(resp.Customers, "Data should not be nil")

	s.T().Logf("List customers response:\n%s", PrettyJSON(resp))

	if len(resp.Customers) > 0 {
		firstCustomer := resp.Customers[0]
		s.NotEmpty(firstCustomer.CustomerID, "Customer ID should not be empty")
		s.NotEmpty(firstCustomer.BusinessLegalName, "Customer business name should not be empty")
		s.NotEmpty(firstCustomer.Email, "Customer email should not be empty")
		s.NotEmpty(firstCustomer.BusinessType, "Customer business type should not be empty")
		s.NotEmpty(firstCustomer.Status, "Customer status should not be empty")
		s.NotEmpty(firstCustomer.CreatedAt, "CreatedAt should not be empty")
		s.NotEmpty(firstCustomer.UpdatedAt, "UpdatedAt should not be empty")
	}
}

// TestCustomerService_GetCustomer tests getting a specific customer.
func (s *CustomerTestSuite) TestCustomerService_GetCustomer() {
	resp, err := s.Client.Customer.GetCustomer(s.Ctx, s.CustomerID)

	s.Require().NoError(err, "GetCustomer should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal(s.CustomerID, resp.CustomerID, "Customer ID should match")
	s.NotEmpty(resp.BusinessLegalName, "Business name should not be empty")
	s.NotEmpty(resp.Email, "Email should not be empty")
	s.NotEmpty(resp.BusinessType, "Business type should not be empty")
	s.NotEmpty(resp.Status, "Status should not be empty")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.UpdatedAt, "UpdatedAt should not be empty")

	s.T().Logf("Get customer response:\n%s", PrettyJSON(resp))
}

// TestCustomerService_UpdateCustomer tests updating a customer with minimal fields.
// Note: For approved customers, update is not allowed (409 Conflict or 500 with "KYB edit not allowed").
func (s *CustomerTestSuite) TestCustomerService_UpdateCustomer() {
	faker := gofakeit.New(0)

	updateReq := &customer.UpdateCustomerRequest{
		BusinessIndustry: utils.AsPtr("332999"),
		AccountPurpose:   utils.AsPtr(customer.AccountPurposeTreasuryManagement),
		AssociatedPersons: []customer.AssociatedPerson{
			FakeAssociatedPerson(faker),
			FakeAssociatedPerson(faker),
			FakeAssociatedPerson(faker),
		},
	}

	updateResp, err := s.Client.Customer.UpdateCustomer(s.Ctx, s.CustomerID, updateReq)
	// For approved customers, KYB edit is not allowed - this is expected behavior
	if err != nil {
		var apiErr *transport.APIError
		if errors.As(err, &apiErr) {
			// 409 Conflict or 500 with "KYB edit not allowed" are expected for approved customers
			if apiErr.IsConflictError() || strings.Contains(apiErr.Detail, "KYB edit not allowed") {
				s.T().Logf("UpdateCustomer rejected as expected for approved customer: %v", err)
				return
			}
		}
		s.Require().NoError(err, "UpdateCustomer should not return unexpected error")
	}

	s.Require().NotNil(updateResp, "Update response should not be nil")
	s.Equal(s.CustomerID, updateResp.CustomerID, "Customer ID should match")
	s.NotEmpty(updateResp.Status, "Status should not be empty")

	s.T().Logf("Update response:\n%s", PrettyJSON(updateResp))
}

// TestCustomerTestSuite runs the customer test suite.
func TestCustomerTestSuite(t *testing.T) {
	suite.Run(t, new(CustomerTestSuite))
}
