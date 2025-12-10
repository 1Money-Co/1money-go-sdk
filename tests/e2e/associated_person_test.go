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
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
)

// AssociatedPersonTestSuite tests associated person operations.
type AssociatedPersonTestSuite struct {
	CustomerDependentTestSuite
}

// TestAssociatedPerson_Create tests creating an associated person.
func (s *AssociatedPersonTestSuite) TestAssociatedPerson_Create() {
	faker := gofakeit.New(0)

	req := &customer.CreateAssociatedPersonRequest{
		AssociatedPerson: FakeAssociatedPerson(faker),
	}

	resp, err := s.Client.Customer.CreateAssociatedPerson(s.Ctx, s.CustomerID, req)

	s.Require().NoError(err, "CreateAssociatedPerson should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.AssociatedPersonID, "Associated person ID should not be empty")
	s.T().Logf("Created associated person:\n%s", PrettyJSON(resp))
}

// TestAssociatedPerson_List tests listing associated persons.
func (s *AssociatedPersonTestSuite) TestAssociatedPerson_List() {
	resp, err := s.Client.Customer.ListAssociatedPersons(s.Ctx, s.CustomerID)

	s.Require().NoError(err, "ListAssociatedPersons should not return error")
	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Associated persons list:\n%s", PrettyJSON(resp))
}

// TestAssociatedPerson_Get tests getting a specific associated person.
func (s *AssociatedPersonTestSuite) TestAssociatedPerson_Get() {
	resp, err := s.Client.Customer.GetAssociatedPerson(s.Ctx, s.CustomerID, s.AssociatedPersonIDs[0])
	s.Require().NoError(err, "GetAssociatedPerson should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.Equal(s.AssociatedPersonIDs[0], resp.AssociatedPersonID, "Associated person ID should match")
	s.T().Logf("Associated person details:\n%s", PrettyJSON(resp))
}

// TestAssociatedPerson_Update tests updating an associated person.
func (s *AssociatedPersonTestSuite) TestAssociatedPerson_Update() {
	faker := gofakeit.New(0)

	getResp, err := s.Client.Customer.GetAssociatedPerson(s.Ctx, s.CustomerID, s.AssociatedPersonIDs[0])
	s.Require().NoError(err, "GetAssociatedPerson should succeed")
	s.Require().NotNil(getResp, "Response should not be nil")

	newEmail := faker.Email()
	hasControl := true
	updateReq := &customer.UpdateAssociatedPersonRequest{
		Email:      &newEmail,
		HasControl: &hasControl,
	}
	updateResp, err := s.Client.Customer.UpdateAssociatedPerson(s.Ctx, s.CustomerID, s.AssociatedPersonIDs[0], updateReq)
	s.Require().NoError(err, "UpdateAssociatedPerson should succeed")
	s.Require().NotNil(updateResp, "Response should not be nil")
	s.Equal(newEmail, updateResp.Email, "Email should be updated")
	s.Equal(hasControl, updateResp.HasControl, "HasControl should be updated")
	s.T().Logf("Updated associated person:\n%s", PrettyJSON(updateResp))
}

// TestAssociatedPerson_Delete tests deleting an associated person.
// Creates its own associated person to delete, avoiding test order dependency.
func (s *AssociatedPersonTestSuite) TestAssociatedPerson_Delete() {
	// Create a new associated person specifically for deletion test
	faker := gofakeit.New(0)
	createReq := &customer.CreateAssociatedPersonRequest{
		AssociatedPerson: FakeAssociatedPerson(faker),
	}

	createResp, err := s.Client.Customer.CreateAssociatedPerson(s.Ctx, s.CustomerID, createReq)
	s.Require().NoError(err, "CreateAssociatedPerson should succeed")
	s.Require().NotNil(createResp, "Create response should not be nil")

	personIDToDelete := createResp.AssociatedPersonID
	s.T().Logf("Created associated person for deletion: %s", personIDToDelete)

	// Delete the newly created associated person
	err = s.Client.Customer.DeleteAssociatedPerson(s.Ctx, s.CustomerID, personIDToDelete)
	s.Require().NoError(err, "DeleteAssociatedPerson should succeed")

	// Verify deletion - should return error
	getResp, err := s.Client.Customer.GetAssociatedPerson(s.Ctx, s.CustomerID, personIDToDelete)
	s.Require().Error(err, "GetAssociatedPerson should return error after deletion")
	s.Require().Nil(getResp, "Response should be nil")
	s.T().Log("Associated person deleted successfully")
}

// TestAssociatedPerson_FileSizeLimit tests that files larger than 3MB are rejected.
func (s *AssociatedPersonTestSuite) TestAssociatedPerson_FileSizeLimit() {
	faker := gofakeit.New(0)

	// Generate data larger than 3MB (3 * 1024 * 1024 = 3145728 bytes)
	// We need slightly more to ensure we exceed the limit after base64 encoding
	oversizedData := make([]byte, 3*1024*1024+1)
	for i := range oversizedData {
		oversizedData[i] = byte(i % 256)
	}
	oversizedDataURI := customer.EncodeBase64ToDataURI(oversizedData, customer.ImageFormatJpeg)

	s.Run("OversizedPOA", func() {
		person := FakeAssociatedPerson(faker)
		person.POA = oversizedDataURI

		req := &customer.CreateAssociatedPersonRequest{
			AssociatedPerson: person,
		}

		resp, err := s.Client.Customer.CreateAssociatedPerson(s.Ctx, s.CustomerID, req)
		s.Require().Error(err, "CreateAssociatedPerson with oversized POA should fail")
		s.Require().Nil(resp, "Response should be nil")
	})

	s.Run("OversizedImageFront", func() {
		person := FakeAssociatedPerson(faker)
		person.IdentifyingInformation[0].ImageFront = oversizedDataURI

		req := &customer.CreateAssociatedPersonRequest{
			AssociatedPerson: person,
		}

		resp, err := s.Client.Customer.CreateAssociatedPerson(s.Ctx, s.CustomerID, req)
		s.Require().Error(err, "CreateAssociatedPerson with oversized ImageFront should fail")
		s.Require().Nil(resp, "Response should be nil")
	})

	s.Run("OversizedImageBack", func() {
		person := FakeAssociatedPerson(faker)
		person.IdentifyingInformation[0].ImageBack = oversizedDataURI

		req := &customer.CreateAssociatedPersonRequest{
			AssociatedPerson: person,
		}

		resp, err := s.Client.Customer.CreateAssociatedPerson(s.Ctx, s.CustomerID, req)
		s.Require().Error(err, "CreateAssociatedPerson with oversized ImageBack should fail")
		s.Require().Nil(resp, "Response should be nil")
	})
}

// TestAssociatedPersonTestSuite runs the associated person test suite.
func TestAssociatedPersonTestSuite(t *testing.T) {
	suite.Run(t, new(AssociatedPersonTestSuite))
}
