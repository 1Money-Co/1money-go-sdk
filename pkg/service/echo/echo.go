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

// Package echo provides a simple echo service for demonstrating SDK usage.
package echo

import (
	"context"
	"encoding/json"
	"fmt"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
)

// Service defines the echo service interface.
// All supported operations are visible in this interface.
type Service interface {
	// Get performs a GET echo request.
	Get(ctx context.Context) (*Response, error)

	// Post performs a POST echo request with the given message.
	Post(ctx context.Context, req *Request) (*Response, error)
}

// Request represents an echo request.
type Request struct {
	Message string `json:"message"`
}

// Response represents an echo response.
type Response struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp,omitempty"`
}

// serviceImpl is the concrete implementation of the echo service (private).
type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new echo service instance with the given base service.
// Returns interface type, not implementation.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// Get performs a GET echo request.
func (s *serviceImpl) Get(ctx context.Context) (*Response, error) {
	resp, err := s.BaseService.Get(ctx, "/echo")
	if err != nil {
		return nil, fmt.Errorf("failed to perform GET echo: %w", err)
	}

	var result Response
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// Post performs a POST echo request with the given message.
func (s *serviceImpl) Post(ctx context.Context, req *Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.BaseService.Post(ctx, "/echo", body)
	if err != nil {
		return nil, fmt.Errorf("failed to perform POST echo: %w", err)
	}

	var result Response
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
