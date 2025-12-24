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
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/pkg/common"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/recipient"
)

// RecipientTestSuite tests recipient service operations.
type RecipientTestSuite struct {
	CustomerDependentTestSuite
}

// FakeRecipientRequest generates a fake recipient request for testing (individual type).
func FakeRecipientRequest() *recipient.CreateRecipientRequest {
	faker := gofakeit.New(0)
	firstName := faker.FirstName()
	lastName := faker.LastName()
	email := faker.Email()
	nickname := faker.Username()
	relationship := recipient.RecipientRelationshipVendor
	addressLine2 := faker.StreetNumber()
	region := RandomGermanState(faker)

	return &recipient.CreateRecipientRequest{
		IdempotencyKey: uuid.New().String(),
		RecipientType:  recipient.RecipientTypeIndividual,
		FirstName:      &firstName,
		LastName:       &lastName,
		Email:          &email,
		Nickname:       &nickname,
		Relationship:   &relationship,
		Address: &recipient.Address{
			CountryCode:  common.CountryCodeDEU,
			AddressLine1: faker.Street(),
			AddressLine2: &addressLine2,
			City:         faker.City(),
			Region:       &region,
			PostalCode:   faker.Zip(),
		},
	}
}

// FakeCompanyRecipientRequest generates a fake recipient request for testing (company type).
func FakeCompanyRecipientRequest() *recipient.CreateRecipientRequest {
	faker := gofakeit.New(0)
	companyName := faker.Company()
	email := faker.Email()
	nickname := faker.Username()
	relationship := recipient.RecipientRelationshipVendor
	addressLine2 := faker.StreetNumber()
	region := RandomGermanState(faker)

	return &recipient.CreateRecipientRequest{
		IdempotencyKey: uuid.New().String(),
		RecipientType:  recipient.RecipientTypeCompany,
		CompanyName:    &companyName,
		Email:          &email,
		Nickname:       &nickname,
		Relationship:   &relationship,
		Address: &recipient.Address{
			CountryCode:  common.CountryCodeDEU,
			AddressLine1: faker.Street(),
			AddressLine2: &addressLine2,
			City:         faker.City(),
			Region:       &region,
			PostalCode:   faker.Zip(),
		},
	}
}

// FakeRecipientBankAccountRequest generates a fake bank account request for testing.
func FakeRecipientBankAccountRequest() *recipient.BankAccountRequest {
	return &recipient.BankAccountRequest{
		IdempotencyKey:  uuid.New().String(),
		Network:         common.BankNetworkNameUSACH,
		Currency:        "USD",
		CountryCode:     common.CountryCodeUSA,
		AccountNumber:   "5097935393",
		InstitutionID:   "327984566",
		InstitutionName: gofakeit.Company() + " Bank",
	}
}

// FakeWalletAddressRequest generates a fake wallet address request for testing.
func FakeWalletAddressRequest() *recipient.WalletAddressRequest {
	nickname := "Test Wallet"
	return &recipient.WalletAddressRequest{
		Blockchain: string(assets.NetworkNamePOLYGON),
		Token:      "USDC",
		Address:    FakeEthereumAddress(),
		Nickname:   &nickname,
	}
}

// TestRecipient_List tests listing recipients with various scenarios.
func (s *RecipientTestSuite) TestRecipient_List() {
	s.Run("Empty", func() {
		// For a fresh customer, listing should succeed even with no recipients
		resp, err := s.Client.Recipient.ListRecipients(s.Ctx, s.CustomerID, nil)
		s.Require().NoError(err, "ListRecipients should succeed even with no recipients")
		s.Require().NotNil(resp, "Response should not be nil")
		s.T().Logf("Recipients list: %d recipients", len(resp.List))
	})

	s.Run("WithPagination", func() {
		resp, err := s.Client.Recipient.ListRecipients(s.Ctx, s.CustomerID, &recipient.ListRecipientsRequest{
			Page: 1,
			Size: 10,
		})
		s.Require().NoError(err, "ListRecipients with pagination should succeed")
		s.Require().NotNil(resp, "Response should not be nil")
		s.T().Logf("Recipients list with pagination: %d recipients", len(resp.List))
	})
}

// TestRecipient_CreateAndGet tests creating and retrieving a recipient.
func (s *RecipientTestSuite) TestRecipient_CreateAndGet() {
	createReq := FakeRecipientRequest()

	// Create recipient
	createResp, err := s.Client.Recipient.CreateRecipient(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRecipient should succeed")

	// Validate create response structure
	s.Require().NotNil(createResp, "Create response should not be nil")
	s.NotEmpty(createResp.RecipientID, "Recipient ID should not be empty")
	s.Equal(s.CustomerID, createResp.CustomerID, "CustomerID should match")
	s.Equal(recipient.RecipientTypeIndividual, createResp.RecipientType, "RecipientType should match")
	s.NotEmpty(createResp.FullName, "FullName should not be empty")
	s.NotEmpty(createResp.Status, "Status should not be empty")
	s.NotEmpty(createResp.CreatedAt, "CreatedAt should not be empty")
	s.NotEmpty(createResp.ModifiedAt, "ModifiedAt should not be empty")

	s.T().Logf("Created recipient: %s (status: %s)", createResp.RecipientID, createResp.Status)

	// Get recipient by ID
	getResp, err := s.Client.Recipient.GetRecipient(s.Ctx, s.CustomerID, createResp.RecipientID)
	s.Require().NoError(err, "GetRecipient should succeed")

	// Validate retrieved recipient matches created one
	s.Require().NotNil(getResp, "Get response should not be nil")
	s.Equal(createResp.RecipientID, getResp.RecipientID, "Recipient IDs should match")
	s.Equal(createResp.RecipientType, getResp.RecipientType, "RecipientType should match")
	s.Equal(createResp.FullName, getResp.FullName, "FullName should match")

	s.T().Logf("Retrieved recipient:\n%s", PrettyJSON(getResp))

	// Get recipient by idempotency key
	getByKeyResp, err := s.Client.Recipient.GetRecipientByIdempotencyKey(s.Ctx, s.CustomerID, createReq.IdempotencyKey)
	s.Require().NoError(err, "GetRecipientByIdempotencyKey should succeed")

	// Validate retrieved recipient matches created one
	s.Require().NotNil(getByKeyResp, "Get by key response should not be nil")
	s.Equal(createResp.RecipientID, getByKeyResp.RecipientID, "Recipient IDs should match")

	s.T().Logf("Retrieved recipient by idempotency key:\n%s", PrettyJSON(getByKeyResp))

	// List recipients and verify the created one is in the list
	listResp, err := s.Client.Recipient.ListRecipients(s.Ctx, s.CustomerID, nil)
	s.Require().NoError(err, "ListRecipients should succeed")
	s.Require().NotNil(listResp, "List response should not be nil")
	s.Require().NotEmpty(listResp.List, "Should have at least one recipient")

	found := false
	for i := range listResp.List {
		if listResp.List[i].RecipientID == createResp.RecipientID {
			found = true
			break
		}
	}
	s.True(found, "Created recipient should be in the list")
	s.T().Logf("Recipients list count: %d", len(listResp.List))
}

// TestRecipient_CreateCompany tests creating a company recipient.
func (s *RecipientTestSuite) TestRecipient_CreateCompany() {
	createReq := FakeCompanyRecipientRequest()

	// Create company recipient
	createResp, err := s.Client.Recipient.CreateRecipient(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRecipient for company should succeed")

	// Validate create response structure
	s.Require().NotNil(createResp, "Create response should not be nil")
	s.NotEmpty(createResp.RecipientID, "Recipient ID should not be empty")
	s.Equal(recipient.RecipientTypeCompany, createResp.RecipientType, "RecipientType should be company")
	s.NotEmpty(createResp.FullName, "FullName should not be empty")

	s.T().Logf("Created company recipient: %s (%s)", createResp.RecipientID, createResp.FullName)
}

// TestRecipient_CreateWithAccountsAndAddresses tests creating a recipient with bank accounts and wallet addresses in one call.
func (s *RecipientTestSuite) TestRecipient_CreateWithAccountsAndAddresses() {
	faker := gofakeit.New(0)
	firstName := faker.FirstName()
	lastName := faker.LastName()
	email := faker.Email()
	nickname := faker.Username()
	relationship := recipient.RecipientRelationshipFamily
	addressLine2 := faker.StreetNumber()
	region := RandomGermanState(faker)

	// Create multiple bank accounts
	bankAccounts := []recipient.BankAccountRequest{
		{
			Network:         common.BankNetworkNameUSACH,
			Currency:        "USD",
			CountryCode:     common.CountryCodeUSA,
			AccountNumber:   "1234567890",
			InstitutionID:   "021000021",
			InstitutionName: "Bank of America",
		},
		{
			Network:         common.BankNetworkNameUSACH,
			Currency:        "USD",
			CountryCode:     common.CountryCodeUSA,
			AccountNumber:   "9876543210",
			InstitutionID:   "026009593",
			InstitutionName: "Chase Bank",
		},
	}

	// Create multiple wallet addresses
	walletAddresses := []recipient.WalletAddressRequest{
		{
			Blockchain: string(assets.NetworkNamePOLYGON),
			Token:      "USDC",
			Address:    FakeEthereumAddress(),
		},
		{
			Blockchain: string(assets.NetworkNameETHEREUM),
			Token:      "USDT",
			Address:    FakeEthereumAddress(),
		},
	}

	createReq := &recipient.CreateRecipientRequest{
		IdempotencyKey: uuid.New().String(),
		RecipientType:  recipient.RecipientTypeIndividual,
		FirstName:      &firstName,
		LastName:       &lastName,
		Email:          &email,
		Nickname:       &nickname,
		Relationship:   &relationship,
		Address: &recipient.Address{
			CountryCode:  common.CountryCodeDEU,
			AddressLine1: faker.Street(),
			AddressLine2: &addressLine2,
			City:         faker.City(),
			Region:       &region,
			PostalCode:   faker.Zip(),
		},
		BankAccounts:    bankAccounts,
		WalletAddresses: walletAddresses,
	}

	// Create recipient with accounts and addresses
	createResp, err := s.Client.Recipient.CreateRecipient(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRecipient with accounts and addresses should succeed")
	s.Require().NotNil(createResp, "Create response should not be nil")
	s.NotEmpty(createResp.RecipientID, "Recipient ID should not be empty")

	s.T().Logf("Created recipient with accounts/addresses: %s", createResp.RecipientID)

	// Verify recipient can be retrieved via GetRecipient
	getResp, err := s.Client.Recipient.GetRecipient(s.Ctx, s.CustomerID, createResp.RecipientID)
	s.Require().NoError(err, "GetRecipient should succeed")
	s.Require().NotNil(getResp, "Get response should not be nil")
	s.Equal(createResp.RecipientID, getResp.RecipientID, "RecipientID should match")
	s.Equal(recipient.RecipientTypeIndividual, getResp.RecipientType, "RecipientType should match")
	s.Equal(recipient.RecipientStatusActive, getResp.Status, "Status should be active")
	s.NotEmpty(getResp.FullName, "FullName should be populated")
	s.NotNil(getResp.Relationship, "Relationship should not be nil")
	s.Equal(relationship, *getResp.Relationship, "Relationship should match")
	s.NotNil(getResp.Email, "Email should not be nil")
	s.Equal(email, *getResp.Email, "Email should match")
	s.NotNil(getResp.Nickname, "Nickname should not be nil")
	s.Equal(nickname, *getResp.Nickname, "Nickname should match")
	s.NotNil(getResp.Address, "Address should not be nil")
	s.Equal(common.CountryCodeDEU, getResp.Address.CountryCode, "Address country code should match")

	s.T().Logf("GetRecipient verified: FullName=%s, Status=%s, Email=%s",
		getResp.FullName, getResp.Status, *getResp.Email)

	// Verify bank accounts were created
	bankListResp, err := s.Client.Recipient.ListBankAccounts(s.Ctx, s.CustomerID, createResp.RecipientID, nil)
	s.Require().NoError(err, "ListBankAccounts should succeed")
	s.Require().NotNil(bankListResp, "Bank accounts list response should not be nil")
	s.Len(bankListResp.List, len(bankAccounts), "Should have %d bank accounts", len(bankAccounts))

	s.T().Logf("Created %d bank accounts:", len(bankListResp.List))
	for i := range bankListResp.List {
		s.NotEmpty(bankListResp.List[i].ExternalAccountID, "Bank account ID should not be empty")
		s.Equal(createResp.RecipientID, bankListResp.List[i].RecipientID, "RecipientID should match")
		s.T().Logf("  - %s: %s (%s)", bankListResp.List[i].ExternalAccountID,
			bankListResp.List[i].InstitutionName, bankListResp.List[i].Network)
	}

	// Verify wallet addresses were created
	walletListResp, err := s.Client.Recipient.ListWalletAddresses(s.Ctx, s.CustomerID, createResp.RecipientID, nil)
	s.Require().NoError(err, "ListWalletAddresses should succeed")
	s.Require().NotNil(walletListResp, "Wallet addresses list response should not be nil")
	s.Len(walletListResp.List, len(walletAddresses), "Should have %d wallet addresses", len(walletAddresses))

	s.T().Logf("Created %d wallet addresses:", len(walletListResp.List))
	for i := range walletListResp.List {
		s.NotEmpty(walletListResp.List[i].WalletAddressID, "Wallet address ID should not be empty")
		s.Equal(createResp.RecipientID, walletListResp.List[i].RecipientID, "RecipientID should match")
		s.T().Logf("  - %s: %s/%s", walletListResp.List[i].WalletAddressID,
			walletListResp.List[i].Blockchain, walletListResp.List[i].Token)
	}
}

// TestRecipient_Update tests updating a recipient.
func (s *RecipientTestSuite) TestRecipient_Update() {
	// First create a recipient
	createReq := FakeRecipientRequest()
	createResp, err := s.Client.Recipient.CreateRecipient(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRecipient should succeed")
	s.Require().NotNil(createResp, "Create response should not be nil")

	s.T().Logf("Created recipient for update: %s", createResp.RecipientID)

	// Update the recipient
	faker := gofakeit.New(1) // Different seed for different values
	updatedFirstName := faker.FirstName()
	updatedLastName := faker.LastName()
	updatedEmail := faker.Email()
	updatedNickname := "Updated Recipient"
	updatedRelationship := recipient.RecipientRelationshipContractor
	region := RandomGermanState(faker)

	updateReq := &recipient.UpdateRecipientRequest{
		RecipientType: recipient.RecipientTypeIndividual,
		FirstName:     &updatedFirstName,
		LastName:      &updatedLastName,
		Email:         &updatedEmail,
		Nickname:      &updatedNickname,
		Relationship:  &updatedRelationship,
		Address: &recipient.Address{
			CountryCode:  common.CountryCodeDEU,
			AddressLine1: faker.Street(),
			City:         faker.City(),
			Region:       &region,
			PostalCode:   faker.Zip(),
		},
	}

	updateResp, err := s.Client.Recipient.UpdateRecipient(s.Ctx, s.CustomerID, createResp.RecipientID, updateReq)
	s.Require().NoError(err, "UpdateRecipient should succeed")
	s.Require().NotNil(updateResp, "Update response should not be nil")
	s.Equal(createResp.RecipientID, updateResp.RecipientID, "Recipient ID should not change")
	s.Equal(&updatedNickname, updateResp.Nickname, "Nickname should be updated")
	s.Equal(&updatedRelationship, updateResp.Relationship, "Relationship should be updated")

	s.T().Logf("Updated recipient:\n%s", PrettyJSON(updateResp))
}

// TestRecipient_Delete tests deleting a recipient.
func (s *RecipientTestSuite) TestRecipient_Delete() {
	// First create a recipient to delete
	createReq := FakeRecipientRequest()
	createResp, err := s.Client.Recipient.CreateRecipient(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRecipient should succeed")
	s.Require().NotNil(createResp, "Create response should not be nil")

	s.T().Logf("Created recipient for deletion: %s", createResp.RecipientID)

	// Delete the recipient
	err = s.Client.Recipient.DeleteRecipient(s.Ctx, s.CustomerID, createResp.RecipientID)
	s.Require().NoError(err, "DeleteRecipient should succeed")

	s.T().Logf("Successfully deleted recipient: %s", createResp.RecipientID)

	// Verify deletion by trying to get the deleted recipient
	// Should return error or the recipient should no longer be active
	getResp, err := s.Client.Recipient.GetRecipient(s.Ctx, s.CustomerID, createResp.RecipientID)
	if err != nil {
		// Expected: recipient not found after deletion
		s.T().Logf("Get deleted recipient returned error (expected): %v", err)
	} else {
		// If API returns the recipient, it should NOT be active
		s.NotEqual(recipient.RecipientStatusActive, getResp.Status,
			"Deleted recipient should not have active status")
		s.T().Logf("Deleted recipient status: %s", getResp.Status)
	}
}

// TestRecipient_BankAccounts tests bank account operations for a recipient.
func (s *RecipientTestSuite) TestRecipient_BankAccounts() {
	// First create a recipient
	createReq := FakeRecipientRequest()
	recipientResp, err := s.Client.Recipient.CreateRecipient(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRecipient should succeed")
	s.Require().NotNil(recipientResp, "Create response should not be nil")

	recipientID := recipientResp.RecipientID
	s.T().Logf("Created recipient for bank account tests: %s", recipientID)

	s.Run("AddBankAccount", func() {
		bankReq := FakeRecipientBankAccountRequest()

		// Add bank account
		bankResp, err := s.Client.Recipient.AddBankAccount(s.Ctx, s.CustomerID, recipientID, bankReq)
		s.Require().NoError(err, "AddBankAccount should succeed")
		s.Require().NotNil(bankResp, "Bank account response should not be nil")
		s.NotEmpty(bankResp.ExternalAccountID, "ExternalAccountID should not be empty")
		s.Equal(recipientID, bankResp.RecipientID, "RecipientID should match")
		s.Equal(s.CustomerID, bankResp.CustomerID, "CustomerID should match")
		s.Equal(common.BankNetworkNameUSACH, bankResp.Network, "Network should match")
		s.Equal("USD", bankResp.Currency, "Currency should match")

		s.T().Logf("Added bank account: %s (status: %s)", bankResp.ExternalAccountID, bankResp.Status)

		// Get bank account by idempotency key
		getByKeyResp, err := s.Client.Recipient.GetBankAccountByIdempotencyKey(
			s.Ctx, s.CustomerID, recipientID, bankReq.IdempotencyKey)
		s.Require().NoError(err, "GetBankAccountByIdempotencyKey should succeed")
		s.Require().NotNil(getByKeyResp, "Get by key response should not be nil")
		s.Equal(bankResp.ExternalAccountID, getByKeyResp.ExternalAccountID, "ExternalAccountID should match")

		s.T().Logf("Retrieved bank account by idempotency key:\n%s", PrettyJSON(getByKeyResp))
	})

	s.Run("ListBankAccounts", func() {
		// List bank accounts
		listResp, err := s.Client.Recipient.ListBankAccounts(s.Ctx, s.CustomerID, recipientID, nil)
		s.Require().NoError(err, "ListBankAccounts should succeed")
		s.Require().NotNil(listResp, "List response should not be nil")
		s.T().Logf("Bank accounts list: %d accounts", len(listResp.List))

		// Verify all returned accounts have required fields
		for i := range listResp.List {
			s.NotEmpty(listResp.List[i].ExternalAccountID, "ExternalAccountID should not be empty")
			s.NotEmpty(listResp.List[i].Network, "Network should not be empty")
			s.NotEmpty(listResp.List[i].Currency, "Currency should not be empty")
		}
	})

	s.Run("ListBankAccountsWithFilter", func() {
		network := common.BankNetworkNameUSACH
		listResp, err := s.Client.Recipient.ListBankAccounts(s.Ctx, s.CustomerID, recipientID,
			&recipient.ListBankAccountsRequest{Network: &network})
		s.Require().NoError(err, "ListBankAccounts with filter should succeed")
		s.Require().NotNil(listResp, "List response should not be nil")
		s.T().Logf("Bank accounts with network %s: %d accounts", network, len(listResp.List))

		// Verify all returned accounts match the filter
		for i := range listResp.List {
			s.Equal(common.BankNetworkNameUSACH, listResp.List[i].Network, "Network should match filter")
		}
	})

	s.Run("DeleteBankAccount", func() {
		// First add a bank account to delete
		bankReq := FakeRecipientBankAccountRequest()
		bankResp, err := s.Client.Recipient.AddBankAccount(s.Ctx, s.CustomerID, recipientID, bankReq)
		s.Require().NoError(err, "AddBankAccount should succeed")
		s.Require().NotNil(bankResp, "Bank account response should not be nil")

		s.T().Logf("Created bank account for deletion: %s", bankResp.ExternalAccountID)

		// Get count before deletion
		listBefore, err := s.Client.Recipient.ListBankAccounts(s.Ctx, s.CustomerID, recipientID, nil)
		s.Require().NoError(err, "ListBankAccounts before delete should succeed")
		countBefore := len(listBefore.List)

		// Delete the bank account
		err = s.Client.Recipient.DeleteBankAccount(s.Ctx, s.CustomerID, recipientID, bankResp.ExternalAccountID)
		s.Require().NoError(err, "DeleteBankAccount should succeed")

		s.T().Logf("Successfully deleted bank account: %s", bankResp.ExternalAccountID)

		// Verify deletion by listing and checking the deleted account is not present
		listAfter, err := s.Client.Recipient.ListBankAccounts(s.Ctx, s.CustomerID, recipientID, nil)
		s.Require().NoError(err, "ListBankAccounts after delete should succeed")
		s.Len(listAfter.List, countBefore-1, "Bank account count should decrease by 1")

		// Verify the deleted account is not in the list
		for i := range listAfter.List {
			s.NotEqual(bankResp.ExternalAccountID, listAfter.List[i].ExternalAccountID,
				"Deleted bank account should not be in the list")
		}
	})
}

// TestRecipient_WalletAddresses tests wallet address operations for a recipient.
func (s *RecipientTestSuite) TestRecipient_WalletAddresses() {
	// First create a recipient
	createReq := FakeRecipientRequest()
	recipientResp, err := s.Client.Recipient.CreateRecipient(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateRecipient should succeed")
	s.Require().NotNil(recipientResp, "Create response should not be nil")

	recipientID := recipientResp.RecipientID
	s.T().Logf("Created recipient for wallet address tests: %s", recipientID)

	s.Run("AddWalletAddress", func() {
		walletReq := FakeWalletAddressRequest()

		// Add wallet address
		walletResp, err := s.Client.Recipient.AddWalletAddress(s.Ctx, s.CustomerID, recipientID, walletReq)
		s.Require().NoError(err, "AddWalletAddress should succeed")
		s.Require().NotNil(walletResp, "Wallet address response should not be nil")
		s.NotEmpty(walletResp.WalletAddressID, "WalletAddressID should not be empty")
		s.Equal(recipientID, walletResp.RecipientID, "RecipientID should match")
		s.Equal(s.CustomerID, walletResp.CustomerID, "CustomerID should match")
		s.Equal(walletReq.Blockchain, walletResp.Blockchain, "Blockchain should match")
		s.Equal(walletReq.Token, walletResp.Token, "Token should match")
		// Use case-insensitive comparison because server returns EIP-55 checksum format
		s.True(strings.EqualFold(walletReq.Address, walletResp.Address), "Address should match (case-insensitive)")

		s.T().Logf("Added wallet address: %s (%s/%s)", walletResp.WalletAddressID, walletResp.Blockchain, walletResp.Token)
	})

	s.Run("ListWalletAddresses", func() {
		// List wallet addresses
		listResp, err := s.Client.Recipient.ListWalletAddresses(s.Ctx, s.CustomerID, recipientID, nil)
		s.Require().NoError(err, "ListWalletAddresses should succeed")
		s.Require().NotNil(listResp, "List response should not be nil")
		s.T().Logf("Wallet addresses list: %d addresses", len(listResp.List))

		// Verify all returned addresses have required fields
		for i := range listResp.List {
			s.NotEmpty(listResp.List[i].WalletAddressID, "WalletAddressID should not be empty")
			s.NotEmpty(listResp.List[i].Blockchain, "Blockchain should not be empty")
			s.NotEmpty(listResp.List[i].Token, "Token should not be empty")
			s.NotEmpty(listResp.List[i].Address, "Address should not be empty")
		}
	})

	s.Run("ListWalletAddressesWithFilter", func() {
		blockchain := string(assets.NetworkNamePOLYGON)
		listResp, err := s.Client.Recipient.ListWalletAddresses(s.Ctx, s.CustomerID, recipientID,
			&recipient.ListWalletAddressesRequest{Blockchain: &blockchain})
		s.Require().NoError(err, "ListWalletAddresses with filter should succeed")
		s.Require().NotNil(listResp, "List response should not be nil")
		s.T().Logf("Wallet addresses with blockchain %s: %d addresses", blockchain, len(listResp.List))

		// Verify all returned addresses match the filter
		for i := range listResp.List {
			s.Equal(blockchain, listResp.List[i].Blockchain, "Blockchain should match filter")
		}
	})

	s.Run("DeleteWalletAddress", func() {
		// First add a wallet address to delete
		walletReq := FakeWalletAddressRequest()
		walletResp, err := s.Client.Recipient.AddWalletAddress(s.Ctx, s.CustomerID, recipientID, walletReq)
		s.Require().NoError(err, "AddWalletAddress should succeed")
		s.Require().NotNil(walletResp, "Wallet address response should not be nil")

		s.T().Logf("Created wallet address for deletion: %s", walletResp.WalletAddressID)

		// Get count before deletion
		listBefore, err := s.Client.Recipient.ListWalletAddresses(s.Ctx, s.CustomerID, recipientID, nil)
		s.Require().NoError(err, "ListWalletAddresses before delete should succeed")
		countBefore := len(listBefore.List)

		// Delete the wallet address
		err = s.Client.Recipient.DeleteWalletAddress(s.Ctx, s.CustomerID, recipientID, walletResp.WalletAddressID)
		s.Require().NoError(err, "DeleteWalletAddress should succeed")

		s.T().Logf("Successfully deleted wallet address: %s", walletResp.WalletAddressID)

		// Verify deletion by listing and checking the deleted address is not present
		listAfter, err := s.Client.Recipient.ListWalletAddresses(s.Ctx, s.CustomerID, recipientID, nil)
		s.Require().NoError(err, "ListWalletAddresses after delete should succeed")
		s.Len(listAfter.List, countBefore-1, "Wallet address count should decrease by 1")

		// Verify the deleted address is not in the list
		for i := range listAfter.List {
			s.NotEqual(walletResp.WalletAddressID, listAfter.List[i].WalletAddressID,
				"Deleted wallet address should not be in the list")
		}
	})
}

// TestRecipientTestSuite runs the recipient test suite.
func TestRecipientTestSuite(t *testing.T) {
	suite.Run(t, new(RecipientTestSuite))
}
