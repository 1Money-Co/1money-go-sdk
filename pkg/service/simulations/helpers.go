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

package simulations

import (
	"context"
	"time"

	"go.uber.org/zap"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/transactions"
)

// WaitOptions configures the polling behavior for wait functions.
type WaitOptions struct {
	// PollInterval is the interval between polling attempts. Default: 5s.
	PollInterval time.Duration
	// MaxWaitTime is the maximum duration to wait. Default: 10m.
	MaxWaitTime time.Duration
	// Logger is an optional zap logger for logging polling progress.
	Logger *zap.Logger
	// PrintProgress prints polling progress to stdout using standard log package.
	// This is useful for examples and debugging when zap logger is not available.
	PrintProgress bool
}

// WaitForSettled polls until the simulated deposit transaction status is no longer PENDING.
// The simulationID from SimulateDepositResponse can be used as a transaction ID.
// Returns the transaction response when settled (COMPLETED, FAILED, or REVERSED).
func WaitForSettled(
	ctx context.Context,
	txService transactions.Service,
	customerID svc.CustomerID,
	simulationID string,
	opts *WaitOptions,
) (*transactions.TransactionResponse, error) {
	txOpts := toTransactionWaitOptions(opts)
	return transactions.WaitForSettled(ctx, txService, customerID, simulationID, txOpts)
}

// WaitForCompleted polls until the simulated deposit transaction status becomes COMPLETED.
// Returns an error if the status becomes FAILED or REVERSED.
func WaitForCompleted(
	ctx context.Context,
	txService transactions.Service,
	customerID svc.CustomerID,
	simulationID string,
	opts *WaitOptions,
) (*transactions.TransactionResponse, error) {
	txOpts := toTransactionWaitOptions(opts)
	return transactions.WaitForCompleted(ctx, txService, customerID, simulationID, txOpts)
}

func toTransactionWaitOptions(opts *WaitOptions) *transactions.WaitOptions {
	if opts == nil {
		return nil
	}
	return &transactions.WaitOptions{
		PollInterval:  opts.PollInterval,
		MaxWaitTime:   opts.MaxWaitTime,
		Logger:        opts.Logger,
		PrintProgress: opts.PrintProgress,
	}
}
