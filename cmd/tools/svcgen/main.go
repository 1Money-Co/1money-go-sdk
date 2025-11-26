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

// Package main provides a code generator for creating new service modules.
//
// This tool generates boilerplate code for new services following the project's
// architecture patterns and conventions.
//
// Usage:
//
//	go run internal/tools/servicegen/main.go <service-name>
//	go generate -run servicegen <service-name>
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// dirPerm defines the permission bits for created directories.
	dirPerm = 0o755
)

const serviceTemplate = `/*
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

// Package {{.PackageName}} provides {{.ServiceName}} service functionality.
//
// This package implements the {{.ServiceName}} service client for the 1Money platform.
//
// # Basic Usage
//
//	import (
//	    "context"
//	    onemoney "github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
//	    "github.com/1Money-Co/1money-go-sdk/pkg/service/{{.PackageName}}"
//	)
//
//	// Create client
//	client, err := onemoney.NewClient(&onemoney.Config{
//	    AccessKey: "your-access-key",
//	    SecretKey: "your-secret-key",
//	})
//
//	// Use the {{.PackageName}} service
//	// resp, err := client.{{.ServiceName}}.YourMethod(context.Background(), req)
package {{.PackageName}}

import (
	"context"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
)

// Service defines the {{.PackageName}} service interface for managing {{.ServiceName}} operations.
type Service interface {
	// TODO: Add your service methods here
	// Example:
	// Get{{.ServiceName}}(ctx context.Context, id string) (*{{.ServiceName}}Response, error)
	// List{{.ServiceName}}s(ctx context.Context, req *List{{.ServiceName}}sRequest) (*List{{.ServiceName}}sResponse, error)
}

type serviceImpl struct {
	*svc.BaseService
}

// NewService creates a new {{.PackageName}} service instance with the given base service.
func NewService(base *svc.BaseService) Service {
	return &serviceImpl{
		BaseService: base,
	}
}

// TODO: Implement your service methods here
// Example:
//
// func (s *serviceImpl) Get{{.ServiceName}}(ctx context.Context, id string) (*{{.ServiceName}}Response, error) {
//     resp, err := svc.GetJSON[{{.ServiceName}}Response](
//         ctx,
//         s.BaseService,
//         "/api/v1/{{.PackageName}}/"+id,
//     )
//     if err != nil {
//         return nil, err
//     }
//     return &resp.Data, nil
// }
`

const testTemplate = `/*
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

package {{.PackageName}}_test

import (
	"testing"

	"github.com/1Money-Co/1money-go-sdk/internal/auth"
	"github.com/1Money-Co/1money-go-sdk/internal/transport"
	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/{{.PackageName}}"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	service := {{.PackageName}}.NewService(&base)

	// Assert
	require.NotNil(t, service)
	assert.Implements(t, (*{{.PackageName}}.Service)(nil), service)
}

// TODO: Add more tests for your service methods
`

type templateData struct {
	PackageName string
	ServiceName string
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <service-name>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s payment\n", os.Args[0])
		os.Exit(1)
	}

	serviceName := flag.Arg(0)
	packageName := strings.ToLower(serviceName)

	// Validate service name
	if packageName == "" {
		fmt.Fprintf(os.Stderr, "Error: service name cannot be empty\n")
		os.Exit(1)
	}

	// Create service directory
	serviceDir := filepath.Join("pkg", "service", packageName)
	if err := os.MkdirAll(serviceDir, dirPerm); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", serviceDir, err)
		os.Exit(1)
	}

	// Prepare template data
	caser := cases.Title(language.English)
	data := templateData{
		PackageName: packageName,
		ServiceName: caser.String(serviceName),
	}

	// Generate service file
	servicePath := filepath.Join(serviceDir, "service.go")
	if err := generateFile(servicePath, serviceTemplate, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating service file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Generated: %s\n", servicePath)

	// Generate test file
	testPath := filepath.Join(serviceDir, "service_test.go")
	if err := generateFile(testPath, testTemplate, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating test file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Generated: %s\n", testPath)

	fmt.Printf("\n🎉 Service '%s' created successfully!\n", serviceName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Implement your service methods in %s\n", servicePath)
	fmt.Printf("  2. Add tests in %s\n", testPath)
	fmt.Printf("  3. Register the service in pkg/onemoney/client.go\n")
}

func generateFile(path, tmpl string, data templateData) error {
	t, err := template.New("service").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
