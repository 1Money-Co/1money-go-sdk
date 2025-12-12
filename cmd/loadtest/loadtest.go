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

package loadtest

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
)

var log *zap.SugaredLogger

func init() {
	var logger *zap.Logger
	if os.Getenv("ONEMONEY_DEBUG") == "1" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	log = logger.Sugar()
}

// Command returns the loadtest CLI command.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "loadtest",
		Usage: "Run load tests against API endpoints in sequence",
		Description: `Run load tests in sequence to avoid rate limiting.

Test sequence:
  1. Create TOS signing link
  2. Create Customer
  3. List Customers
  4. Create External Account
  5. Get External Account By ID
  6. Create Auto-Conversion Rule (500 expected if no verified fiat account)
  7. List Auto-Conversion Rules
  8. List All Transactions

Output is Vegeta format, can be piped to vegeta report:

  onemoney loadtest | vegeta report
  onemoney loadtest | vegeta report -type=hist`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "rate",
				Aliases: []string{"r"},
				Usage:   "Requests per second per endpoint",
				Value:   5,
			},
			&cli.DurationFlag{
				Name:    "duration",
				Aliases: []string{"d"},
				Usage:   "Duration per endpoint",
				Value:   10 * time.Second,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file for report (default: stdout)",
			},
		},
		Action: runLoadtest,
	}
}

// testCase represents a single load test case
type testCase struct {
	name     string
	targeter func(ctx *loadtestContext) vegeta.Targeter
	setup    func(ctx *loadtestContext) error // optional setup before running
}

// loadtestContext holds state across test cases
type loadtestContext struct {
	client             *onemoney.Client
	customerID         string
	externalAccountID  string
	signedAgreementIDs []string // pre-generated for create-customer test
}

func runLoadtest(c *cli.Context) error {
	rate := c.Int("rate")
	duration := c.Duration("duration")
	outputFile := c.String("output")

	// Create SDK client - loads config from env automatically
	client, err := onemoney.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Debug: print config
	fmt.Fprintf(os.Stderr, "Config: BaseURL=%s, AccessKey=%s..., Sandbox=%v\n",
		client.Config.BaseURL,
		client.Config.AccessKey[:min(8, len(client.Config.AccessKey))],
		client.Config.Sandbox)

	if client.Config.AccessKey == "" {
		panic("AccessKey is empty - check ONEMONEY_ACCESS_KEY env var")
	}

	ctx := &loadtestContext{
		client: client,
	}

	// Pool size for create-customer (need pre-generated signed agreements)
	poolSize := rate * int(duration.Seconds())

	// Define test cases in order
	testCases := []testCase{
		{name: "create-tos-link", targeter: createTOSLinkTargeter},
		{name: "create-customer", targeter: createCustomerTargeter, setup: func(ctx *loadtestContext) error {
			return prepareSignedAgreements(ctx, poolSize)
		}},
		{name: "list-customers", targeter: listCustomersTargeter},
		{name: "create-external-account", targeter: createExternalAccountTargeter, setup: setupCustomerID},
		{name: "get-external-account", targeter: getExternalAccountTargeter, setup: setupExternalAccount},
		{name: "create-auto-conversion-rule", targeter: createAutoConversionRuleTargeter}, // 500 expected if no verified fiat account
		{name: "list-auto-conversion-rules", targeter: listAutoConversionRulesTargeter},
		{name: "list-transactions", targeter: listTransactionsTargeter},
	}

	// Collect results per test case and all results for final report
	var allResults []*vegeta.Result
	caseResults := make(map[string][]*vegeta.Result)

	for _, tc := range testCases {
		// Create fresh attacker for each test case
		attacker := vegeta.NewAttacker()
		// Run setup if needed
		if tc.setup != nil {
			if err := tc.setup(ctx); err != nil {
				panic(fmt.Sprintf("setup failed for %s: %v", tc.name, err))
			}
		}

		// Panic if missing required context
		if needsCustomer(tc.name) && ctx.customerID == "" {
			panic(fmt.Sprintf("missing customer ID for %s", tc.name))
		}

		fmt.Fprintf(os.Stderr, "→ %s (%d req/s, %s)\n", tc.name, rate, duration)

		var caseRes []*vegeta.Result
		targeter := tc.targeter(ctx)
		resultCount := 0
		for res := range attacker.Attack(targeter, vegeta.Rate{Freq: rate, Per: time.Second}, duration, tc.name) {
			resultCount++
			caseRes = append(caseRes, res)
			allResults = append(allResults, res)
			if resultCount <= 3 {
				log.Debugw("attack result",
					"seq", resultCount,
					"code", res.Code,
					"latency", res.Latency,
				)
			}
			// Log errors with response body for debugging
			if res.Error != "" || (res.Code >= 400 && res.Code < 600) {
				log.Warnw("request failed",
					"seq", resultCount,
					"code", res.Code,
					"error", res.Error,
					"body", string(res.Body),
				)
			}
		}
		log.Debugw("attack loop finished", "totalResults", resultCount)
		caseResults[tc.name] = caseRes
	}

	// Generate report
	return writeReport(caseResults, allResults, outputFile)
}

func needsCustomer(name string) bool {
	switch name {
	case "create-external-account", "get-external-account",
		"create-auto-conversion-rule", "list-auto-conversion-rules", "list-transactions":
		return true
	}
	return false
}

// Setup functions - create resources needed for subsequent tests

func setupCustomerID(ctx *loadtestContext) error {
	// Skip if already have customerID
	if ctx.customerID != "" {
		return nil
	}

	// Try to get existing customer first
	listResp, err := ctx.client.Customer.ListCustomers(context.Background(), nil)
	if err == nil && listResp != nil && len(listResp.Customers) > 0 {
		ctx.customerID = listResp.Customers[0].CustomerID
		return nil
	}

	return fmt.Errorf("no existing customer found - run create-customer first")
}

func setupExternalAccount(ctx *loadtestContext) error {
	if ctx.customerID == "" {
		return fmt.Errorf("no customer ID")
	}

	accounts, err := ctx.client.ExternalAccounts.ListExternalAccounts(context.Background(), ctx.customerID, nil)
	if err != nil {
		return err
	}

	if len(accounts) > 0 {
		ctx.externalAccountID = accounts[0].ExternalAccountID
	}
	return nil
}

// testCaseOrder defines the order for reporting test cases.
var testCaseOrder = []string{
	"create-tos-link",
	"create-customer",
	"list-customers",
	"create-external-account",
	"get-external-account",
	"create-auto-conversion-rule",
	"list-auto-conversion-rules",
	"list-transactions",
}

// writeReport generates a text report from results and writes to file or stdout.
func writeReport(caseResults map[string][]*vegeta.Result, allResults []*vegeta.Result, outputFile string) error {
	// Determine output destination
	var out io.Writer = os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		out = f
	}

	// Write per-case reports
	for _, name := range testCaseOrder {
		results, ok := caseResults[name]
		if !ok || len(results) == 0 {
			continue
		}

		var metrics vegeta.Metrics
		for _, res := range results {
			metrics.Add(res)
		}
		metrics.Close()

		fmt.Fprintf(out, "\n=== %s ===\n", name)
		reporter := vegeta.NewTextReporter(&metrics)
		if err := reporter.Report(out); err != nil {
			return fmt.Errorf("failed to write report for %s: %w", name, err)
		}
	}

	// Write total report
	var totalMetrics vegeta.Metrics
	for _, res := range allResults {
		totalMetrics.Add(res)
	}
	totalMetrics.Close()

	fmt.Fprintf(out, "\n=== TOTAL ===\n")
	reporter := vegeta.NewTextReporter(&totalMetrics)
	if err := reporter.Report(out); err != nil {
		return fmt.Errorf("failed to write total report: %w", err)
	}

	if outputFile != "" {
		fmt.Fprintf(os.Stderr, "\nReport written to: %s\n", outputFile)
	}

	return nil
}
