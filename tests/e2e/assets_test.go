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

// TestAssets_ListAssets tests listing assets with various filters.
func (s *AssetsTestSuite) TestAssets_ListAssets() {
	tests := []struct {
		name    string
		req     *assets.ListAssetsRequest
		checkFn func(resp []assets.AssetResponse)
	}{
		{
			name: "ListAll",
			req:  nil,
			checkFn: func(resp []assets.AssetResponse) {
				for _, asset := range resp {
					s.NotEmpty(asset.CustomerID, "Customer ID should not be empty")
					s.NotEmpty(asset.Asset, "Asset name should not be empty")
					s.NotEmpty(asset.AvailableAmount, "Available amount should not be empty")
					s.NotEmpty(asset.CreatedAt, "CreatedAt should not be empty")
					s.NotEmpty(asset.ModifiedAt, "ModifiedAt should not be empty")
				}
			},
		},
		{
			name: "FilterByAsset",
			req:  &assets.ListAssetsRequest{Asset: assets.AssetNameUSD},
			checkFn: func(resp []assets.AssetResponse) {
				for _, asset := range resp {
					s.Equal(string(assets.AssetNameUSD), asset.Asset, "Asset should be USD")
				}
			},
		},
		{
			name: "FilterByNetwork",
			req:  &assets.ListAssetsRequest{Network: assets.NetworkNameETHEREUM},
		},
		{
			name: "WithSortOrderDesc",
			req:  &assets.ListAssetsRequest{SortOrder: assets.SortOrderDESC},
		},
		{
			name: "WithSortOrderAsc",
			req:  &assets.ListAssetsRequest{SortOrder: assets.SortOrderASC},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := s.Client.Assets.ListAssets(s.Ctx, testCustomerID, tc.req)
			s.Require().NoError(err, "ListAssets should succeed")
			s.Require().NotNil(resp, "Response should not be nil")
			s.T().Logf("%s response:\n%s", tc.name, PrettyJSON(resp))

			if tc.checkFn != nil {
				tc.checkFn(resp)
			}
		})
	}
}

// TestAssetsTestSuite runs the assets test suite.
func TestAssetsTestSuite(t *testing.T) {
	suite.Run(t, new(AssetsTestSuite))
}
