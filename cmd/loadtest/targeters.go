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
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
)

func defaultHeaders(apiKey string) http.Header {
	return http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"Bearer " + apiKey},
	}
}

// 1. Create TOS signing link
func createTOSLinkTargeter(ctx *loadtestContext) vegeta.Targeter {
	body, _ := json.Marshal(customer.CreateTOSLinkRequest{
		RedirectUrl: "https://example.com/redirect",
	})
	return vegeta.NewStaticTargeter(vegeta.Target{
		Method: http.MethodPost,
		URL:    ctx.client.Config.BaseURL + "/v1/customers/tos_links",
		Body:   body,
		Header: defaultHeaders(ctx.client.Config.AccessKey),
	})
}

// 2. Create Customer
// Each request needs a fresh signedAgreementID (one-time use)
func createCustomerTargeter(ctx *loadtestContext) vegeta.Targeter {
	faker := gofakeit.New(0)
	var idx int
	var mu sync.Mutex

	log.Debugw("targeter created", "poolSize", len(ctx.signedAgreementIDs))

	return func(tgt *vegeta.Target) error {
		mu.Lock()
		if idx >= len(ctx.signedAgreementIDs) {
			mu.Unlock()
			log.Debugw("pool exhausted", "idx", idx)
			return fmt.Errorf("signed agreement pool exhausted")
		}
		signedAgreementID := ctx.signedAgreementIDs[idx]
		idx++
		currentIdx := idx
		mu.Unlock()

		if currentIdx <= 3 {
			log.Debugw("preparing request", "idx", currentIdx)
		}

		body, _ := json.Marshal(FakeCreateCustomerRequest(faker, signedAgreementID))
		tgt.Method = http.MethodPost
		tgt.URL = ctx.client.Config.BaseURL + "/v1/customers"
		tgt.Body = body
		tgt.Header = defaultHeaders(ctx.client.Config.AccessKey)
		return nil
	}
}

// prepareSignedAgreements generates signed agreements concurrently (max 10 TPS)
func prepareSignedAgreements(ctx *loadtestContext, count int) error {
	const maxTPS = 10
	log.Infow("preparing signed agreements", "count", count, "maxTPS", maxTPS)
	results := make(chan string, count)
	errs := make(chan error, 1)
	ticker := time.NewTicker(time.Second / maxTPS)
	defer ticker.Stop()

	var wg sync.WaitGroup
	var errOnce sync.Once

	for i := range count {
		<-ticker.C // rate limit
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id, err := getSignedAgreementID(ctx)
			if err != nil {
				errOnce.Do(func() {
					errs <- fmt.Errorf("failed at %d: %w", idx, err)
				})
				return
			}
			results <- id
		}(i)
	}

	// Wait and close channels
	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	// Collect results
	for id := range results {
		ctx.signedAgreementIDs = append(ctx.signedAgreementIDs, id)
		if len(ctx.signedAgreementIDs)%10 == 0 {
			log.Infow("progress", "prepared", len(ctx.signedAgreementIDs), "total", count)
		}
	}

	// Check for errors
	if err := <-errs; err != nil {
		return err
	}

	log.Infow("preparation complete", "prepared", len(ctx.signedAgreementIDs))
	return nil
}

// getSignedAgreementID creates a TOS link and signs it to get a one-time signedAgreementID
func getSignedAgreementID(ctx *loadtestContext) (string, error) {
	bgCtx := context.Background()
	tosResp, err := ctx.client.Customer.CreateTOSLink(bgCtx, &customer.CreateTOSLinkRequest{
		RedirectUrl: "https://example.com/redirect",
	})
	if err != nil {
		return "", err
	}

	signResp, err := ctx.client.Customer.SignTOSAgreement(bgCtx, tosResp.SessionToken)
	if err != nil {
		return "", err
	}

	return signResp.SignedAgreementID, nil
}

// 3. List Customers
func listCustomersTargeter(ctx *loadtestContext) vegeta.Targeter {
	return vegeta.NewStaticTargeter(vegeta.Target{
		Method: http.MethodGet,
		URL:    ctx.client.Config.BaseURL + "/v1/customers?page_size=10",
		Header: defaultHeaders(ctx.client.Config.AccessKey),
	})
}

// 4. Create External Account
func createExternalAccountTargeter(ctx *loadtestContext) vegeta.Targeter {
	faker := gofakeit.New(0)
	return func(tgt *vegeta.Target) error {
		req := FakeExternalAccountRequest(faker)
		body, _ := json.Marshal(req)
		tgt.Method = http.MethodPost
		tgt.URL = ctx.client.Config.BaseURL + "/v1/customers/" + ctx.customerID + "/external-accounts"
		tgt.Body = body
		tgt.Header = defaultHeaders(ctx.client.Config.AccessKey)
		tgt.Header.Set("Idempotency-Key", req.IdempotencyKey)
		return nil
	}
}

// 5. Get External Account By ID
func getExternalAccountTargeter(ctx *loadtestContext) vegeta.Targeter {
	if ctx.externalAccountID == "" {
		// Fallback to list if no specific account
		return vegeta.NewStaticTargeter(vegeta.Target{
			Method: http.MethodGet,
			URL:    ctx.client.Config.BaseURL + "/v1/customers/" + ctx.customerID + "/external-accounts/list",
			Header: defaultHeaders(ctx.client.Config.AccessKey),
		})
	}
	return vegeta.NewStaticTargeter(vegeta.Target{
		Method: http.MethodGet,
		URL:    ctx.client.Config.BaseURL + "/v1/customers/" + ctx.customerID + "/external-accounts/" + ctx.externalAccountID,
		Header: defaultHeaders(ctx.client.Config.AccessKey),
	})
}

// 6. Create Auto-Conversion Rule
func createAutoConversionRuleTargeter(ctx *loadtestContext) vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		req := FakeAutoConversionRuleRequest()
		body, _ := json.Marshal(req)
		tgt.Method = http.MethodPost
		tgt.URL = ctx.client.Config.BaseURL + "/v1/customers/" + ctx.customerID + "/auto-conversion-rules"
		tgt.Body = body
		tgt.Header = defaultHeaders(ctx.client.Config.AccessKey)
		tgt.Header.Set("Idempotency-Key", req.IdempotencyKey)
		return nil
	}
}

// 7. List Auto-Conversion Rules
func listAutoConversionRulesTargeter(ctx *loadtestContext) vegeta.Targeter {
	return vegeta.NewStaticTargeter(vegeta.Target{
		Method: http.MethodGet,
		URL:    ctx.client.Config.BaseURL + "/v1/customers/" + ctx.customerID + "/auto-conversion-rules/list",
		Header: defaultHeaders(ctx.client.Config.AccessKey),
	})
}

// 8. List All Transactions
func listTransactionsTargeter(ctx *loadtestContext) vegeta.Targeter {
	return vegeta.NewStaticTargeter(vegeta.Target{
		Method: http.MethodGet,
		URL:    ctx.client.Config.BaseURL + "/v1/customers/" + ctx.customerID + "/transactions",
		Header: defaultHeaders(ctx.client.Config.AccessKey),
	})
}
