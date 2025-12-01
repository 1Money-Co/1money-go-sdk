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

package transactions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/1Money-Co/1money-go-sdk/internal/auth"
	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/transactions"
)

func TestNewService(t *testing.T) {
	// Arrange
	creds := auth.NewCredentials("test-key", "test-secret")
	signer := auth.NewSigner(creds)
	tr := transport.NewTransport(&transport.Config{
		BaseURL: "http://localhost:9000",
	}, signer)
	base := svc.NewBaseService(tr)

	// Act
	service := transactions.NewService(base)

	// Assert
	require.NotNil(t, service)
	assert.Implements(t, (*transactions.Service)(nil), service)
}

// TODO: Add more tests for your service methods
