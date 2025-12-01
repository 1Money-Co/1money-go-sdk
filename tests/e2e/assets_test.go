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

	"github.com/stretchr/testify/suite"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/assets"
)

// AssetsTestSuite tests assets service operations.
type AssetsTestSuite struct {
	E2ETestSuite
}

// TestAssets_ListAll tests listing all assets for a customer.
func (s *AssetsTestSuite) TestAssets_ListAll() {
	resp, err := s.Client.Assets.ListAssets(s.Ctx, testCustomerID, nil)
	s.Require().NoError(err, "ListAssets should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Assets list:\n%s", PrettyJSON(resp))

	for _, asset := range resp {
		s.NotEmpty(asset.CustomerID, "Customer ID should not be empty")
		s.NotEmpty(asset.Asset, "Asset name should not be empty")
		s.NotEmpty(asset.AvailableAmount, "Available amount should not be empty")
		s.NotEmpty(asset.CreatedAt, "CreatedAt should not be empty")
		s.NotEmpty(asset.ModifiedAt, "ModifiedAt should not be empty")
	}
}

// TestAssets_ListByAsset tests listing assets filtered by asset name.
func (s *AssetsTestSuite) TestAssets_ListByAsset() {
	req := &assets.ListAssetsRequest{
		Asset: assets.AssetNameUSD,
	}

	resp, err := s.Client.Assets.ListAssets(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "ListAssets should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("USD assets:\n%s", PrettyJSON(resp))

	for _, asset := range resp {
		s.Equal(string(assets.AssetNameUSD), asset.Asset, "Asset should be USD")
	}
}

// TestAssets_ListByNetwork tests listing assets filtered by network.
func (s *AssetsTestSuite) TestAssets_ListByNetwork() {
	req := &assets.ListAssetsRequest{
		Network: assets.NetworkNameETHEREUM,
	}

	resp, err := s.Client.Assets.ListAssets(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "ListAssets should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Ethereum network assets:\n%s", PrettyJSON(resp))
}

// TestAssets_ListWithSortOrder tests listing assets with sort order.
func (s *AssetsTestSuite) TestAssets_ListWithSortOrder() {
	req := &assets.ListAssetsRequest{
		SortOrder: assets.SortOrderDESC,
	}

	resp, err := s.Client.Assets.ListAssets(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "ListAssets should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.T().Logf("Assets (DESC order):\n%s", PrettyJSON(resp))
}

// TestAssetsTestSuite runs the assets test suite.
func TestAssetsTestSuite(t *testing.T) {
	suite.Run(t, new(AssetsTestSuite))
}
