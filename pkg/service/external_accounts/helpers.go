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

package external_accounts

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/1Money-Co/1money-go-sdk/internal/utils"
	"github.com/1Money-Co/1money-go-sdk/pkg/common"
	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
)

// WaitOptions configures the polling behavior for wait functions.
type WaitOptions struct {
	// PollInterval is the interval between polling attempts. Default: 2s.
	PollInterval time.Duration
	// MaxWaitTime is the maximum duration to wait. Default: 2m.
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
		PollInterval: 2 * time.Second,
		MaxWaitTime:  2 * time.Minute,
	}
}

// ExternalAccountCondition is a function that checks if an external account meets a condition.
type ExternalAccountCondition func(*Resp) bool

// WaitFor polls until the condition returns true.
// Returns the external account response when condition is met, or an error on timeout/failure.
func WaitFor(
	ctx context.Context,
	service Service,
	customerID svc.CustomerID,
	id svc.ExternalAccountID,
	condition ExternalAccountCondition,
	opts *WaitOptions,
) (*Resp, error) {
	defaults := DefaultWaitOptions()
	if opts == nil {
		opts = &defaults
	}

	utilOpts := &utils.WaitOptions{
		PollInterval:  opts.PollInterval,
		MaxWaitTime:   opts.MaxWaitTime,
		Logger:        opts.Logger,
		LogMessage:    "polling external account status",
		PrintProgress: opts.PrintProgress,
	}

	return utils.WaitFor(
		ctx,
		func(ctx context.Context) (*Resp, error) {
			return service.GetExternalAccount(ctx, customerID, id)
		},
		utils.Condition[Resp](condition),
		func(a *Resp) string { return a.Status },
		"external_account",
		id,
		utilOpts,
	)
}

// WaitForApproved polls until the external account's status becomes APPROVED.
// Returns an error if the status becomes FAILED or timeout occurs.
func WaitForApproved(
	ctx context.Context,
	service Service,
	customerID svc.CustomerID,
	id svc.ExternalAccountID,
	opts *WaitOptions,
) (*Resp, error) {
	account, err := WaitFor(ctx, service, customerID, id, func(a *Resp) bool {
		return a.Status == string(common.BankAccountStatusAPPROVED) || a.Status == string(common.BankAccountStatusFAILED)
	}, opts)
	if err != nil {
		return nil, err
	}

	if account.Status == string(common.BankAccountStatusFAILED) {
		return account, fmt.Errorf("external account %s approval failed", id)
	}

	return account, nil
}
