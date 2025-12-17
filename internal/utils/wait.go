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

// Package utils provides common utilities for the 1Money SDK.
package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
)

// WaitOptions configures the polling behavior for wait functions.
type WaitOptions struct {
	// PollInterval is the interval between polling attempts. Default: 5s.
	PollInterval time.Duration
	// MaxWaitTime is the maximum duration to wait. Default: 10m.
	MaxWaitTime time.Duration
	// Logger is an optional zap logger for logging polling progress.
	Logger *zap.Logger
	// LogMessage is the message to log on each polling iteration.
	LogMessage string
	// LogFields are additional fields to include in log messages.
	LogFields []zap.Field
	// PrintProgress prints polling progress to stdout using standard log package.
	// This is useful for examples and debugging when zap logger is not available.
	PrintProgress bool
}

// DefaultWaitOptions returns the default wait options.
func DefaultWaitOptions() WaitOptions {
	return WaitOptions{
		PollInterval: 5 * time.Second,
		MaxWaitTime:  10 * time.Minute,
		LogMessage:   "polling status",
	}
}

// MergeWaitOptions merges the provided options with defaults for zero values.
func MergeWaitOptions(opts *WaitOptions, defaults WaitOptions) WaitOptions {
	if opts == nil {
		return defaults
	}

	result := *opts
	if result.PollInterval == 0 {
		result.PollInterval = defaults.PollInterval
	}
	if result.MaxWaitTime == 0 {
		result.MaxWaitTime = defaults.MaxWaitTime
	}
	if result.LogMessage == "" {
		result.LogMessage = defaults.LogMessage
	}
	return result
}

// Condition is a function that checks if a resource meets a condition.
type Condition[T any] func(*T) bool

// Getter is a function that fetches the current state of a resource.
type Getter[T any] func(ctx context.Context) (*T, error)

// StatusExtractor is a function that extracts a status string from a resource for logging.
type StatusExtractor[T any] func(*T) string

// WaitFor polls until the condition returns true.
// Returns the resource when condition is met, or an error on timeout/failure.
//
// Example:
//
//	tx, err := utils.WaitFor(ctx,
//	    func(ctx context.Context) (*Transaction, error) {
//	        return service.GetTransaction(ctx, customerID, txID)
//	    },
//	    func(tx *Transaction) bool {
//	        return tx.Status != "PENDING"
//	    },
//	    func(tx *Transaction) string {
//	        return tx.Status
//	    },
//	    "transaction_id", txID,
//	    &utils.WaitOptions{Logger: logger},
//	)
func WaitFor[T any](
	ctx context.Context,
	getter Getter[T],
	condition Condition[T],
	statusExtractor StatusExtractor[T],
	resourceName string,
	resourceID string,
	opts *WaitOptions,
) (*T, error) {
	defaults := DefaultWaitOptions()
	merged := MergeWaitOptions(opts, defaults)

	start := time.Now()
	deadline := start.Add(merged.MaxWaitTime)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resource, err := getter(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", resourceName, err)
		}

		status := ""
		if statusExtractor != nil {
			status = statusExtractor(resource)
		}

		if merged.Logger != nil {
			fields := []zap.Field{
				zap.Float64("elapsed_seconds", time.Since(start).Seconds()),
				zap.String(resourceName+"_id", resourceID),
			}
			if status != "" {
				fields = append(fields, zap.String("status", status))
			}
			fields = append(fields, merged.LogFields...)
			merged.Logger.Info(merged.LogMessage, fields...)
		} else if merged.PrintProgress {
			log.Printf("%s: %s=%s elapsed=%.1fs status=%s",
				merged.LogMessage, resourceName, resourceID, time.Since(start).Seconds(), status)
		}

		if condition(resource) {
			return resource, nil
		}

		time.Sleep(merged.PollInterval)
	}

	return nil, fmt.Errorf("timeout waiting for %s %s after %v", resourceName, resourceID, merged.MaxWaitTime)
}
