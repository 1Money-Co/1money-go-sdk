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

package transactions

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/1Money-Co/1money-go-sdk/internal/utils"
	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
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

// DefaultWaitOptions returns the default wait options.
func DefaultWaitOptions() WaitOptions {
	return WaitOptions{
		PollInterval: 5 * time.Second,
		MaxWaitTime:  10 * time.Minute,
	}
}

// TransactionCondition is a function that checks if a transaction meets a condition.
type TransactionCondition func(*TransactionResponse) bool

// WaitFor polls until the condition returns true.
// Returns the transaction response when condition is met, or an error on timeout/failure.
func WaitFor(
	ctx context.Context,
	service Service,
	customerID svc.CustomerID,
	transactionID string,
	condition TransactionCondition,
	opts *WaitOptions,
) (*TransactionResponse, error) {
	defaults := DefaultWaitOptions()
	if opts == nil {
		opts = &defaults
	}

	utilOpts := &utils.WaitOptions{
		PollInterval:  opts.PollInterval,
		MaxWaitTime:   opts.MaxWaitTime,
		Logger:        opts.Logger,
		LogMessage:    "polling transaction status",
		PrintProgress: opts.PrintProgress,
	}

	return utils.WaitFor(
		ctx,
		func(ctx context.Context) (*TransactionResponse, error) {
			return service.GetTransaction(ctx, customerID, transactionID)
		},
		utils.Condition[TransactionResponse](condition),
		func(tx *TransactionResponse) string { return tx.Status.String() },
		"transaction",
		transactionID,
		utilOpts,
	)
}

// WaitForSettled polls until the transaction status is no longer PENDING.
// Returns the transaction response when settled (COMPLETED, FAILED, or REVERSED).
func WaitForSettled(
	ctx context.Context,
	service Service,
	customerID svc.CustomerID,
	transactionID string,
	opts *WaitOptions,
) (*TransactionResponse, error) {
	return WaitFor(ctx, service, customerID, transactionID, func(tx *TransactionResponse) bool {
		return tx.Status != TransactionStatusPENDING
	}, opts)
}

// WaitForCompleted polls until the transaction status becomes COMPLETED.
// Returns an error if the status becomes FAILED or REVERSED.
func WaitForCompleted(
	ctx context.Context,
	service Service,
	customerID svc.CustomerID,
	transactionID string,
	opts *WaitOptions,
) (*TransactionResponse, error) {
	tx, err := WaitFor(ctx, service, customerID, transactionID, func(tx *TransactionResponse) bool {
		return tx.Status != TransactionStatusPENDING
	}, opts)
	if err != nil {
		return nil, err
	}

	if tx.Status == TransactionStatusFAILED {
		return tx, fmt.Errorf("transaction %s failed", transactionID)
	}
	if tx.Status == TransactionStatusREVERSED {
		return tx, fmt.Errorf("transaction %s was reversed", transactionID)
	}

	return tx, nil
}
