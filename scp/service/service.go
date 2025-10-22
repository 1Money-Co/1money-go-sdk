// Package service provides base functionality for all service modules.
package service

import (
	"context"

	"github.com/1Money-Co/1money-go-sdk/internal/transport"
)

// BaseService provides common functionality for all service implementations.
// Business modules should embed this struct to inherit transport capabilities.
type BaseService struct {
	transport *transport.Transport
}

// NewBaseService creates a new base service with the given transport.
func NewBaseService(t *transport.Transport) BaseService {
	return BaseService{transport: t}
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
