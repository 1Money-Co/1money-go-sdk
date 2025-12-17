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

package auto_conversion_rules

import (
	"context"
	"fmt"
	"log"
	"time"
)

// WaitOptions configures the polling behavior for wait functions.
type WaitOptions struct {
	// PollInterval is the interval between polling attempts. Default: 2s.
	PollInterval time.Duration
	// MaxWaitTime is the maximum duration to wait. Default: 60s.
	MaxWaitTime time.Duration
	// PrintProgress prints polling progress to stdout using standard log package.
	// This is useful for examples and debugging.
	PrintProgress bool
}

// DefaultWaitOptions returns the default wait options.
func DefaultWaitOptions() WaitOptions {
	return WaitOptions{
		PollInterval: 2 * time.Second,
		MaxWaitTime:  60 * time.Second,
	}
}

// RuleCondition is a function that checks if a rule meets a condition.
type RuleCondition func(*RuleResponse) bool

// WaitFor polls until the condition returns true.
// Returns the rule response when condition is met, or an error on timeout/failure.
func WaitFor(
	ctx context.Context, svc Service, customerID, ruleID string,
	condition RuleCondition, opts *WaitOptions,
) (*RuleResponse, error) {
	if opts == nil {
		defaults := DefaultWaitOptions()
		opts = &defaults
	}

	start := time.Now()
	deadline := start.Add(opts.MaxWaitTime)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		rule, err := svc.GetRule(ctx, customerID, ruleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get rule: %w", err)
		}

		if opts.PrintProgress {
			log.Printf("polling rule status: rule=%s elapsed=%.1fs status=%s deposit_info_status=%s",
				ruleID, time.Since(start).Seconds(), rule.Status, rule.DepositInfoStatus)
		}

		if condition(rule) {
			return rule, nil
		}

		time.Sleep(opts.PollInterval)
	}

	return nil, fmt.Errorf("timeout waiting for rule %s after %v", ruleID, opts.MaxWaitTime)
}

// WaitForActive polls until the rule's Status becomes ACTIVE.
func WaitForActive(ctx context.Context, svc Service, customerID, ruleID string, opts *WaitOptions) (*RuleResponse, error) {
	return WaitFor(ctx, svc, customerID, ruleID, func(r *RuleResponse) bool {
		return r.Status == RuleStatusACTIVE
	}, opts)
}

// WaitForDepositInfoReady polls until the rule's DepositInfoStatus is no longer PENDING.
func WaitForDepositInfoReady(ctx context.Context, svc Service, customerID, ruleID string, opts *WaitOptions) (*RuleResponse, error) {
	return WaitFor(ctx, svc, customerID, ruleID, func(r *RuleResponse) bool {
		return r.DepositInfoStatus != DepositInfoStatusPENDING
	}, opts)
}
