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
	E2ETestSuite
}

// TestSimulations_SimulateDeposit_USD tests simulating a USD deposit.
func (s *SimulationsTestSuite) TestSimulations_SimulateDeposit_USD() {
	req := &simulations.SimulateDepositRequest{
		Asset:  assets.AssetNameUSD,
		Amount: "100.00",
	}

	resp, err := s.Client.Simulations.SimulateDeposit(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "SimulateDeposit should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.SimulationID, "SimulationID should not be empty")
	s.NotEmpty(resp.Status, "Status should not be empty")
	s.T().Logf("Simulated USD deposit:\n%s", PrettyJSON(resp))
}

// TestSimulations_SimulateDeposit_USDT_Ethereum tests simulating a USDT deposit on Ethereum.
func (s *SimulationsTestSuite) TestSimulations_SimulateDeposit_USDT_Ethereum() {
	req := &simulations.SimulateDepositRequest{
		Asset:   assets.AssetNameUSDT,
		Network: simulations.WalletNetworkNameETHEREUM,
		Amount:  "50.00",
	}

	resp, err := s.Client.Simulations.SimulateDeposit(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "SimulateDeposit should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.SimulationID, "SimulationID should not be empty")
	s.T().Logf("Simulated USDT Ethereum deposit:\n%s", PrettyJSON(resp))
}

// TestSimulations_SimulateDeposit_USDC_Polygon tests simulating a USDC deposit on Polygon.
func (s *SimulationsTestSuite) TestSimulations_SimulateDeposit_USDC_Polygon() {
	req := &simulations.SimulateDepositRequest{
		Asset:   assets.AssetNameUSDC,
		Network: simulations.WalletNetworkNamePOLYGON,
		Amount:  "25.00",
	}

	resp, err := s.Client.Simulations.SimulateDeposit(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "SimulateDeposit should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.SimulationID, "SimulationID should not be empty")
	s.T().Logf("Simulated USDC Polygon deposit:\n%s", PrettyJSON(resp))
}

// TestSimulations_SimulateDeposit_USDT_Solana tests simulating a USDT deposit on Solana.
func (s *SimulationsTestSuite) TestSimulations_SimulateDeposit_USDT_Solana() {
	req := &simulations.SimulateDepositRequest{
		Asset:   assets.AssetNameUSDT,
		Network: simulations.WalletNetworkNameSOLANA,
		Amount:  "75.00",
	}

	resp, err := s.Client.Simulations.SimulateDeposit(s.Ctx, testCustomerID, req)
	s.Require().NoError(err, "SimulateDeposit should succeed")

	s.Require().NotNil(resp, "Response should not be nil")
	s.NotEmpty(resp.SimulationID, "SimulationID should not be empty")
	s.T().Logf("Simulated USDT Solana deposit:\n%s", PrettyJSON(resp))
}

// TestSimulationsTestSuite runs the simulations test suite.
func TestSimulationsTestSuite(t *testing.T) {
	suite.Run(t, new(SimulationsTestSuite))
}
