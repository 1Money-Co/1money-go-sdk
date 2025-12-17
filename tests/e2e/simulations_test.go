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
	"github.com/1Money-Co/1money-go-sdk/pkg/service/simulations"
)

// SimulationsTestSuite tests simulations service operations.
// NOTE: These tests only work in sandbox/non-production environments.
type SimulationsTestSuite struct {
	CustomerDependentTestSuite
}

type simulateDepositTestCase struct {
	name    string
	asset   assets.AssetName
	network simulations.WalletNetworkName
	amount  string
}

func (s *SimulationsTestSuite) TestSimulations_SimulateDeposit() {
	testCases := []simulateDepositTestCase{
		{
			name:    "USD",
			asset:   assets.AssetNameUSD,
			amount:  "100.00",
			network: simulations.WalletNetworkNameUSACH,
		},
		{
			name:    "USDT_Ethereum",
			asset:   assets.AssetNameUSDT,
			network: simulations.WalletNetworkNameETHEREUM,
			amount:  "50.00",
		},
		{
			name:    "USDC_Polygon",
			asset:   assets.AssetNameUSDC,
			network: simulations.WalletNetworkNamePOLYGON,
			amount:  "25.00",
		},
		{
			name:    "USDT_Solana",
			asset:   assets.AssetNameUSDT,
			network: simulations.WalletNetworkNameSOLANA,
			amount:  "75.00",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := &simulations.SimulateDepositRequest{
				Asset:   tc.asset,
				Network: tc.network,
				Amount:  tc.amount,
			}

			resp, err := s.Client.Simulations.SimulateDeposit(s.Ctx, s.CustomerID, req)
			s.Require().NoError(err, "SimulateDeposit should succeed")
			s.Require().NotNil(resp, "Response should not be nil")

			// Validate response structure
			s.NotEmpty(resp.SimulationID)
			s.NotEmpty(resp.Status)
			s.NotEmpty(resp.CreatedAt)
			s.NotEmpty(resp.ModifiedAt)
			s.Contains([]string{"PENDING", "COMPLETED", "FAILED", "REVERSED"}, resp.Status.String())

			s.T().Logf("Simulated %s deposit:\n%s", tc.name, PrettyJSON(resp))
		})
	}
}

// TestSimulationsTestSuite runs the simulations test suite.
func TestSimulationsTestSuite(t *testing.T) {
	suite.Run(t, new(SimulationsTestSuite))
}
