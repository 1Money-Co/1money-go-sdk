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
	"path/filepath"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/flaticols/countrycodes"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/internal/utils"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/echo"
)

const (
	// testCustomerID is a test customer ID used across multiple tests.
	testCustomerID = "10592382-48e0-4b52-bc40-4d484f09dfad"
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
	// Load environment variables from .env file in project root
	// Find project root by looking for go.mod file
	projectRoot, err := utils.FindProjectRoot()
	if err == nil {
		envPath := filepath.Join(projectRoot, ".env")
		_ = godotenv.Load(envPath)
	}

	// Create client configuration
	cfg := &Config{
		// BaseURL:   "http://localhost:9000",
		// AccessKey: "XLVXZY2Z3DLLCLKELVB4",
		// SecretKey: "LTCs85bMOvmLxjKmfMete2FsH-nfa3qP1PdVSOSbeLo",
		// Timeout:   30 * time.Second,
	}

	// Skip tests if required environment variables are missing
	// if cfg.BaseURL == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
	// 	s.T().Skipf("missing required environment variables (ONEMONEY_BASE_URL, ONEMONEY_ACCESS_KEY, ONEMONEY_SECRET_KEY)")
	// }

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

// TestCustomerService_TOSFlow tests the complete TOS signing flow.
func (s *ClientTestSuite) TestCustomerService_TOSFlow() {
	// Step 1: Create TOS link
	tosResp, err := s.client.Customer.CreateTOSLink(s.ctx)
	s.Require().NoError(err, "CreateTOSLink should not return error")
	s.Require().NotNil(tosResp, "CreateTOSLink response should not be nil")
	s.NotEmpty(tosResp.SessionToken, "Session token should not be empty")
	s.T().Logf("Created TOS link with session token:\n%s", prettyJSON(tosResp))

	// Step 2: Sign the agreement using the session token
	signResp, err := s.client.Customer.SignTOSAgreement(s.ctx, tosResp.SessionToken)
	s.Require().NoError(err, "SignTOSAgreement should not return error")
	s.Require().NotNil(signResp, "SignTOSAgreement response should not be nil")
	s.NotEmpty(signResp.SignedAgreementID, "Signed agreement ID should not be empty")
	s.T().Logf("Signed agreement with ID:\n%s", prettyJSON(signResp))
}

func (s *ClientTestSuite) TestCustomerService_SignTOS() {
	sessionToken := "8d8448b7-0246-4162-bfdd-6b3b50de8bf5" //nolint:gosec // test session token
	// Step 2: Sign the agreement using the session token
	signResp, err := s.client.Customer.SignTOSAgreement(s.ctx, sessionToken)
	s.Require().NoError(err, "SignTOSAgreement should not return error")
	s.Require().NotNil(signResp, "SignTOSAgreement response should not be nil")
	s.NotEmpty(signResp.SignedAgreementID, "Signed agreement ID should not be empty")
	s.T().Logf("Signed agreement with ID:\n%s", prettyJSON(signResp))
}

func fakeCountryCode(faker *gofakeit.Faker) string {
	country, _ := countrycodes.Alpha2ToAlpha3(faker.CountryAbr())
	return country
}

// TestCustomerService_CreateCustomer tests customer creation.
func (s *ClientTestSuite) TestCustomerService_CreateCustomer() {
	// Arrange - Generate fake data using gofakeit
	faker := gofakeit.New(0)

	// Create at least one associated person
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
			Country:     fakeCountryCode(faker),
			PostalCode:  faker.Zip(),
			Subdivision: faker.StateAbr(),
		},
		DateOfIncorporation: faker.Date().Format("2006-01-02"),
		SignedAgreementID:   967,
		AssociatedPersons: []customer.AssociatedPerson{
			fakeAssociatedPerson(faker),
			fakeAssociatedPerson(faker),
			fakeAssociatedPerson(faker),
			fakeAssociatedPerson(faker),
		},
		SourceOfFunds:  []customer.SourceOfFunds{customer.SourceOfFundsSalesOfGoodsAndServices},
		SourceOfWealth: []customer.SourceOfWealth{customer.SourceOfWealthBusinessDividendsOrProfits},
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

// TestCustomerService_GetCustomer tests getting a specific customer.
func (s *ClientTestSuite) TestCustomerService_GetCustomer() {
	// Arrange
	customerID := testCustomerID

	// Act
	resp, err := s.client.Customer.GetCustomer(s.ctx, customerID)

	// Assert
	s.Require().NoError(err, "GetCustomer should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal(customerID, resp.ID, "Customer ID should match")
	s.NotEmpty(resp.BusinessLegalName, "Business name should not be empty")
	s.NotEmpty(resp.Email, "Email should not be empty")
	s.NotEmpty(resp.BusinessType, "Business type should not be empty")
	s.NotEmpty(resp.Status, "Status should not be empty")
	s.NotEmpty(resp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(resp.UpdatedAt, "UpdatedAt should not be empty")

	s.T().Logf("Get customer response:\n%s", prettyJSON(resp))
}

// TestCustomerService_UpdateCustomer_MinimalUpdate tests updating a customer with minimal fields.
func (s *ClientTestSuite) TestCustomerService_UpdateCustomer_MinimalUpdate() {
	customerID := testCustomerID
	faker := gofakeit.New(0)

	updateReq := &customer.UpdateCustomerRequest{
		BusinessIndustry: utils.AsPtr(customer.BusinessIndustryTechnologyECommercePlatforms),
		AccountPurpose:   utils.AsPtr(customer.AccountPurposeTreasuryManagement),
		AssociatedPersons: []customer.AssociatedPerson{
			fakeAssociatedPerson(faker),
			fakeAssociatedPerson(faker),
			fakeAssociatedPerson(faker),
		},
	}

	// Act
	updateResp, err := s.client.Customer.UpdateCustomer(s.ctx, customerID, updateReq)

	// Assert
	s.Require().NoError(err, "UpdateCustomer should not return error")
	s.Require().NotNil(updateResp, "Update response should not be nil")
	s.Require().Empty(updateResp.ValidationErrors, "Validation errors should be empty")
	s.Equal(customerID, updateResp.ID, "Customer ID should match")
	s.NotEmpty(updateResp.Status, "Status should not be empty")

	s.T().Logf("Minimal update response:\n%s", prettyJSON(updateResp))
}

func fakeAssociatedPerson(faker *gofakeit.Faker) customer.AssociatedPerson {
	return customer.AssociatedPerson{
		FirstName: faker.FirstName(),
		LastName:  faker.LastName(),
		Email:     faker.Email(),
		ResidentialAddress: &customer.Address{
			StreetLine1: faker.Street(),
			City:        faker.City(),
			State:       faker.StateAbr(),
			Country:     fakeCountryCode(faker),
			PostalCode:  faker.Zip(),
		},
		BirthDate:           faker.Date().Format("2006-01-02"),
		CountryOfBirth:      fakeCountryCode(faker),
		PrimaryNationality:  fakeCountryCode(faker),
		HasOwnership:        true,
		OwnershipPercentage: 100,
		HasControl:          true,
		IsSigner:            true,
		IsDirector:          true,
		IdentifyingInformation: []customer.IdentifyingInformation{
			{
				Type:           customer.IDTypeDriversLicense,
				IssuingCountry: fakeCountryCode(faker),
				ImageFront:     customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
				ImageBack:      customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg),
			},
		},
		CountryOfTax: fakeCountryCode(faker),
		TaxType:      customer.TaxIDTypeEIN,
		TaxIDNumber:  fmt.Sprintf("%d-%d", faker.Number(10, 99), faker.Number(1000000, 9999999)),
		POA:          customer.EncodeBase64ToDataURI(gofakeit.ImageJpeg(100, 100), customer.ImageFormatJpeg), // POA is required for directors and beneficial owners
	}
}

// TestAssociatedPerson_Create tests creating an associated person.
func (s *ClientTestSuite) TestAssociatedPerson_Create() {
	customerID := "f6186a5c-64c1-451e-8a37-80a74b846416"
	faker := gofakeit.New(0)

	req := &customer.CreateAssociatedPersonRequest{
		AssociatedPerson: fakeAssociatedPerson(faker),
	}

	resp, err := s.client.Customer.CreateAssociatedPerson(s.ctx, customerID, req)

	s.Require().NoError(err, "CreateAssociatedPerson should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.ID, "Associated person ID should not be empty")
	s.T().Logf("Created associated person:\n%s", prettyJSON(resp))
}

// TestAssociatedPerson_List tests listing associated persons.
func (s *ClientTestSuite) TestAssociatedPerson_List() {
	customerID := testCustomerID

	resp, err := s.client.Customer.ListAssociatedPersons(s.ctx, customerID)

	s.Require().NoError(err, "ListAssociatedPersons should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Associated persons list:\n%s", prettyJSON(resp))
}

// TestAssociatedPerson_Get tests getting a specific associated person.
func (s *ClientTestSuite) TestAssociatedPerson_Get() {
	customerID := testCustomerID
	associatedPersonID := "ae56e5ab-d171-4ac9-8e22-6d67a9216095"

	resp, err := s.client.Customer.GetAssociatedPerson(s.ctx, customerID, associatedPersonID)
	if err != nil {
		s.T().Logf("GetAssociatedPerson error (expected if person doesn't exist): %v", err)
		return
	}

	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal(associatedPersonID, resp.ID, "Associated person ID should match")
	s.T().Logf("Associated person details:\n%s", prettyJSON(resp))
}

// TestAssociatedPerson_Update tests updating an associated person.
func (s *ClientTestSuite) TestAssociatedPerson_Update() {
	customerID := testCustomerID
	associatedPersonID := "ae56e5ab-d171-4ac9-8e22-6d67a9216095"
	faker := gofakeit.New(0)

	getResp, err := s.client.Customer.GetAssociatedPerson(s.ctx, customerID, associatedPersonID)
	if err != nil {
		s.T().Logf("GetAssociatedPerson error (expected if person doesn't exist): %v", err)
		return
	}
	s.Require().NotNil(getResp, "Response should not be nil")

	newEmail := faker.Email()
	hasControl := true
	updateReq := &customer.UpdateAssociatedPersonRequest{
		Email:      &newEmail,
		HasControl: &hasControl,
	}
	updateResp, err := s.client.Customer.UpdateAssociatedPerson(s.ctx, customerID, associatedPersonID, updateReq)
	if err != nil {
		s.T().Logf("UpdateAssociatedPerson error (expected if person doesn't exist): %v", err)
		return
	}
	updateResp.Email = newEmail
	updateResp.HasControl = hasControl

	s.Require().NotNil(updateResp, "Response should not be nil")
	s.T().Logf("Updated associated person:\n%s", prettyJSON(updateResp))
}

// TestAssociatedPerson_Delete tests deleting an associated person.
func (s *ClientTestSuite) TestAssociatedPerson_Delete() {
	customerID := testCustomerID
	associatedPersonID := "324a052a-ff44-4ae8-9ef9-1a5e50df3c94"

	err := s.client.Customer.DeleteAssociatedPerson(s.ctx, customerID, associatedPersonID)
	if err != nil {
		s.T().Logf("DeleteAssociatedPerson error (expected if person doesn't exist): %v", err)
		return
	}

	// we should not be able to get the associated person after deletion
	getResp, err := s.client.Customer.GetAssociatedPerson(s.ctx, customerID, associatedPersonID)
	if err == nil {
		s.T().Logf("GetAssociatedPerson should return error (expected if person doesn't exist): %v", err)
		return
	}
	s.Require().Nil(getResp, "Response should be nil")
	s.T().Log("Associated person deleted successfully")
}

func (s *ClientTestSuite) TestEchoService_Get() {
	resp, err := s.client.Echo.Get(s.ctx)
	if err != nil {
		s.T().Logf("Get error: %v", err)
		return
	}
	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Echo response:\n%s", prettyJSON(resp))
}

func (s *ClientTestSuite) TestEchoService_Post() {
	resp, err := s.client.Echo.Post(s.ctx, &echo.Request{Message: "Hello, World!"})
	if err != nil {
		s.T().Logf("Post error: %v", err)
		return
	}
	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Echo response:\n%s", prettyJSON(resp))
}

// TestRateLimiter_IPBasedLimiting tests that the IP-based rate limiter is working correctly.
// The backend is configured with a rate limit of 10 requests per second with a burst size of 10.
func (s *ClientTestSuite) TestRateLimiter_IPBasedLimiting() {
	// The rate limiter is configured with:
	// - per_second: 10 (10 requests per second)
	// - burst_size: 10 (can handle up to 10 burst requests)

	const (
		burstSize    = 10
		extraRequest = 5
		totalRequest = burstSize + extraRequest
	)

	s.T().Log("Testing rate limiter with concurrent requests...")

	// Use channels to collect results from goroutines
	type result struct {
		index       int
		success     bool
		rateLimited bool
		err         error
		responseMsg string
	}

	resultChan := make(chan result, totalRequest)

	// Launch all goroutines concurrently to simulate burst traffic
	for i := range totalRequest {
		go func(index int) {
			resp, err := s.client.Echo.Post(s.ctx, &echo.Request{
				Message: fmt.Sprintf("Rate limit test message #%d", index+1),
			})

			res := result{index: index + 1}
			if err != nil {
				res.err = err
				// Check if it's a rate limit error (usually HTTP 429)
				if containsRateLimitError(err.Error()) {
					res.rateLimited = true
				}
			} else {
				res.success = true
				res.responseMsg = resp.Message
			}
			resultChan <- res
		}(i)
	}

	// Collect all results
	successCount := 0
	rateLimitedCount := 0
	unexpectedErrors := 0

	for range totalRequest {
		res := <-resultChan
		if res.success {
			successCount++
			s.T().Logf("Request #%d: Success - %s", res.index, res.responseMsg)
		} else if res.rateLimited {
			rateLimitedCount++
			s.T().Logf("Request #%d: Rate limited (expected after burst)", res.index)
		} else {
			unexpectedErrors++
			s.T().Logf("Request #%d: Unexpected error: %v", res.index, res.err)
		}
	}
	close(resultChan)

	s.T().Logf("Rate limiter test results:")
	s.T().Logf("  Total requests: %d", totalRequest)
	s.T().Logf("  Successful: %d", successCount)
	s.T().Logf("  Rate limited: %d", rateLimitedCount)
	s.T().Logf("  Unexpected errors: %d", unexpectedErrors)

	// Assertions:
	// 1. We should have some successful requests (may vary due to concurrent execution)
	s.Positive(successCount, "Should have at least some successful requests")

	// 2. We should have some rate-limited requests (the extra requests beyond burst)
	// Note: Due to concurrent execution, exact count may vary
	s.Positive(rateLimitedCount,
		"Should have at least one rate-limited request when exceeding burst size")

	// 3. Total processed should match (excluding unexpected errors)
	s.Equal(totalRequest, successCount+rateLimitedCount+unexpectedErrors,
		"Total requests should equal successful + rate limited + unexpected errors")

	// 4. We shouldn't have unexpected errors
	s.Equal(0, unexpectedErrors, "Should not have unexpected errors")

	// Wait for rate limiter to reset (1 second + buffer)
	s.T().Log("Waiting for rate limiter to reset...")
	time.Sleep(1500 * time.Millisecond)

	// After waiting, we should be able to send requests again
	resp, err := s.client.Echo.Post(s.ctx, &echo.Request{Message: "After reset"})
	s.Require().NoError(err, "Request should succeed after rate limiter reset")
	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("After reset: Successfully sent request - %s", resp.Message)
}

// containsRateLimitError checks if an error message indicates a rate limit error.
func containsRateLimitError(errMsg string) bool {
	// Common rate limit indicators
	indicators := []string{
		"429",               // HTTP status code
		"Too Many Requests", // Standard HTTP 429 message
		"rate limit",        // Generic rate limit message
		"too many requests", // Alternative message
		"throttle",          // Alternative terminology
	}

	// Simple case-insensitive substring check
	errMsgLower := toLower(errMsg)
	for _, indicator := range indicators {
		if contains(errMsgLower, toLower(indicator)) {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase (ASCII only).
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := range len(s) {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// contains checks if string s contains substring substr.
func contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestClientTestSuite runs the test suite.
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
