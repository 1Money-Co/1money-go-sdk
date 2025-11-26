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

// Package service provides base functionality and common utilities for all service modules.
//
// This package implements the foundational layer for all 1Money API service clients,
// providing HTTP method wrappers, JSON serialization helpers, and shared service infrastructure.
//
// # Overview
//
// The service package offers:
//   - BaseService struct that all service implementations embed
//   - HTTP method wrappers (GET, POST, PUT, PATCH, DELETE)
//   - Generic JSON helpers with type-safe request/response handling
//   - Automatic error handling and response parsing
//   - Transport layer integration for authentication and signing
//
// # BaseService
//
// All service implementations should embed BaseService to inherit HTTP capabilities:
//
//	type MyService struct {
//	    service.BaseService
//	}
//
//	func NewMyService(t *transport.Transport) *MyService {
//	    return &MyService{
//	        BaseService: service.NewBaseService(t),
//	    }
//	}
//
// # Generic JSON Helpers
//
// The package provides type-safe generic functions for common HTTP operations:
//
//	// GET request with automatic unmarshaling
//	resp, err := service.GetJSON[MyResponse](ctx, &baseService, "/api/resource")
//
//	// POST request with automatic marshaling and unmarshaling
//	resp, err := service.PostJSON[MyRequest, MyResponse](
//	    ctx, &baseService, "/api/resource", req,
//	)
//
// These helpers automatically:
//   - Marshal request bodies to JSON
//   - Include proper Content-Type headers
//   - Unmarshal response bodies to typed structures
//   - Wrap responses in GenericResponse[T] for consistent error handling
//
// # Error Handling
//
// All service methods return errors that include:
//   - HTTP status codes for non-2xx responses
//   - Detailed error messages from the API
//   - Network and serialization errors
//
// Example error handling:
//
//	resp, err := svc.CreateResource(ctx, req)
//	if err != nil {
//	    var apiErr *transport.APIError
//	    if errors.As(err, &apiErr) {
//	        // Handle API-specific error
//	        log.Printf("API error: %s (code: %d)", apiErr.Message, apiErr.Code)
//	    }
//	    return err
//	}
//
// # Extending Services
//
// To create a new service module:
//
//  1. Create a new package under pkg/service/ (e.g., pkg/service/myservice)
//  2. Define a Service interface with your business methods
//  3. Create a private serviceImpl that embeds BaseService
//  4. Implement your business methods using the HTTP helpers
//  5. Export a NewService constructor
//
// Example:
//
//	package myservice
//
//	import (
//	    "context"
//	    "github.com/1Money-Co/1money-go-sdk/internal/transport"
//	    svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
//	)
//
//	type Service interface {
//	    GetResource(ctx context.Context, id string) (*Resource, error)
//	}
//
//	type serviceImpl struct {
//	    svc.BaseService
//	}
//
//	func NewService(t *transport.Transport) Service {
//	    return &serviceImpl{
//	        BaseService: svc.NewBaseService(t),
//	    }
//	}
//
//	func (s *serviceImpl) GetResource(ctx context.Context, id string) (*Resource, error) {
//	    resp, err := svc.GetJSON[Resource](ctx, &s.BaseService, "/api/resources/"+id)
//	    if err != nil {
//	        return nil, err
//	    }
//	    return &resp.Data, nil
//	}
package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/1Money-Co/1money-go-sdk/internal/transport"
)

// BaseService provides common functionality for all service implementations.
// Business modules should embed this struct to inherit transport capabilities.
type BaseService struct {
	transport *transport.Transport
}

// NewBaseService creates a new base service with the given transport.
func NewBaseService(t *transport.Transport) *BaseService {
	return &BaseService{transport: t}
}

// Get performs a GET request.
func (s *BaseService) Get(ctx context.Context, path string) (*transport.Response, error) {
	req := &transport.Request{
		Method: "GET",
		Path:   path,
	}
	return s.transport.Do(ctx, req)
}

// Post performs a POST request with the given body.
func (s *BaseService) Post(ctx context.Context, path string, body []byte) (*transport.Response, error) {
	req := &transport.Request{
		Method: "POST",
		Path:   path,
		Body:   body,
	}
	return s.transport.Do(ctx, req)
}

// Put performs a PUT request with the given body.
func (s *BaseService) Put(ctx context.Context, path string, body []byte) (*transport.Response, error) {
	req := &transport.Request{
		Method: "PUT",
		Path:   path,
		Body:   body,
	}
	return s.transport.Do(ctx, req)
}

// Delete performs a DELETE request.
func (s *BaseService) Delete(ctx context.Context, path string) (*transport.Response, error) {
	req := &transport.Request{
		Method: "DELETE",
		Path:   path,
	}
	return s.transport.Do(ctx, req)
}

// Patch performs a PATCH request with the given body.
func (s *BaseService) Patch(ctx context.Context, path string, body []byte) (*transport.Response, error) {
	req := &transport.Request{
		Method: "PATCH",
		Path:   path,
		Body:   body,
	}
	return s.transport.Do(ctx, req)
}

// Do performs a custom request with full control.
func (s *BaseService) Do(ctx context.Context, req *transport.Request) (*transport.Response, error) {
	return s.transport.Do(ctx, req)
}

// GetJSON performs a GET request and unmarshals the response directly into T.
func GetJSON[T any](ctx context.Context, s *BaseService, path string) (*T, error) {
	resp, err := s.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetJSONWithParams performs a GET request with query parameters and unmarshals the response directly into T.
func GetJSONWithParams[T any](ctx context.Context,
	s *BaseService,
	path string,
	params map[string]string,
) (*T, error) {
	req := &transport.Request{
		Method:      "GET",
		Path:        path,
		QueryParams: params,
	}
	resp, err := s.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// sendJSONRequest is a helper function that handles JSON marshaling/unmarshaling for HTTP requests.
// It marshals the request body, sends it using the provided method, and unmarshals the response directly.
func sendJSONRequest[Req, Resp any](ctx context.Context,
	path string,
	req Req,
	method func(context.Context, string, []byte) (*transport.Response, error),
) (*Resp, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := method(ctx, path, body)
	if err != nil {
		return nil, err
	}

	var result Resp
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// PostJSON performs a POST request with automatic JSON marshaling/unmarshaling.
// It marshals the request body and unmarshals the response directly into Resp.
func PostJSON[Req, Resp any](ctx context.Context, s *BaseService, path string, req Req) (*Resp, error) {
	return sendJSONRequest[Req, Resp](ctx, path, req, s.Post)
}

// PutJSON performs a PUT request with automatic JSON marshaling/unmarshaling.
// It marshals the request body and unmarshals the response directly into Resp.
func PutJSON[Req, Resp any](ctx context.Context, s *BaseService, path string, req Req) (*Resp, error) {
	return sendJSONRequest[Req, Resp](ctx, path, req, s.Put)
}

// PatchJSON performs a PATCH request with automatic JSON marshaling/unmarshaling.
// It marshals the request body and unmarshals the response directly into Resp.
func PatchJSON[Req, Resp any](ctx context.Context, s *BaseService, path string, req Req) (*Resp, error) {
	return sendJSONRequest[Req, Resp](ctx, path, req, s.Patch)
}

// DeleteJSON performs a DELETE request and unmarshals the response directly into T.
func DeleteJSON[T any](ctx context.Context, s *BaseService, path string) (*T, error) {
	resp, err := s.Do(ctx, &transport.Request{
		Method:  "DELETE",
		Path:    path,
		Headers: map[string]string{"Accept": "application/json"},
	})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 204 || len(resp.Body) == 0 {
		return nil, nil
	}

	fmt.Println(string(resp.Body))

	var result T
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
