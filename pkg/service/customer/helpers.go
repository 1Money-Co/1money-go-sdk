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

package customer

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	svc "github.com/1Money-Co/1money-go-sdk/pkg/service"
)

// EncodeFileToDataURI reads a file and encodes it as a data-uri string.
// The format parameter specifies the image format (jpeg, jpg, png, heic, tif).
// If format is empty, it will be detected from the file extension.
//
// Example:
//
//	dataURI, err := customer.EncodeFileToDataURI("/path/to/image.jpg", customer.ImageFormatJPEG)
//	if err != nil {
//	    return err
//	}
//	req.AssociatedPersons[0].POA = dataURI
func EncodeFileToDataURI(filePath string, format ImageFormat) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Auto-detect format from extension if not provided
	if format == "" {
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filePath)), ".")
		switch ext {
		case "jpg", "jpeg":
			format = ImageFormatJpeg
		case "png":
			format = ImageFormatPng
		case "heic":
			format = ImageFormatHeic
		case "tif", "tiff":
			format = ImageFormatTif
		default:
			return "", fmt.Errorf("unsupported file extension: %s (supported: jpg, jpeg, png, heic, tif, tiff)", ext)
		}
	}

	return EncodeBase64ToDataURI(data, format), nil
}

// EncodeDocumentFileToDataURI reads a file and encodes it as a data-uri string.
// Supports all document formats: images (jpeg, jpg, png), PDF, CSV, XLS, XLSX.
// If format is empty, it will be detected from the file extension.
//
// Example:
//
//	dataURI, err := customer.EncodeDocumentFileToDataURI("/path/to/doc.pdf", customer.FileFormatPdf)
//	if err != nil {
//	    return err
//	}
//	doc.File = dataURI
func EncodeDocumentFileToDataURI(filePath string, format FileFormat) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Auto-detect format from extension if not provided
	if format == "" {
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filePath)), ".")
		switch ext {
		case "jpg", "jpeg":
			format = FileFormatJpeg
		case "png":
			format = FileFormatPng
		case "heic":
			format = FileFormatHeic
		case "tif", "tiff":
			format = FileFormatTif
		case "pdf":
			format = FileFormatPdf
		case "csv":
			format = FileFormatCsv
		case "xls":
			format = FileFormatXls
		case "xlsx":
			format = FileFormatXlsx
		default:
			return "", fmt.Errorf("unsupported file extension: %s (supported: jpg, jpeg, png, heic, tif, pdf, csv, xls, xlsx)", ext)
		}
	}

	return EncodeDocumentToDataURI(data, format), nil
}

// fileFormatToMIME returns the MIME type for a given file format.
func fileFormatToMIME(format FileFormat) string {
	switch format {
	case FileFormatJpeg, FileFormatJpg:
		return "image/jpeg"
	case FileFormatPng:
		return "image/png"
	case FileFormatHeic:
		return "image/heic"
	case FileFormatTif:
		return "image/tiff"
	case FileFormatPdf:
		return "application/pdf"
	case FileFormatCsv:
		return "text/csv"
	case FileFormatXls:
		return "application/xls"
	case FileFormatXlsx:
		return "application/xlsx"
	default:
		return "application/octet-stream"
	}
}

// EncodeDocumentToDataURI converts binary data to a data-uri string.
// Supports all document formats: images, PDF, CSV, XLS, XLSX.
//
// Example:
//
//	data := []byte{...}
//	dataURI := customer.EncodeDocumentToDataURI(data, customer.FileFormatPdf)
func EncodeDocumentToDataURI(data []byte, format FileFormat) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	mime := fileFormatToMIME(format)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded)
}

// EncodeBase64ToDataURI converts base64-encoded data to a data-uri string.
// The format parameter specifies the image format (jpeg, jpg, png, heic, tif).
//
// Example:
//
//	base64Data := []byte("iVBORw0KGgoAAAANSUhEUgAA...")
//	dataURI := customer.EncodeBase64ToDataURI(base64Data, customer.ImageFormatPNG)
func EncodeBase64ToDataURI(data []byte, format ImageFormat) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:image/%s;base64,%s", format, encoded)
}

// EncodeStringToDataURI converts a base64-encoded string to a data-uri string.
// The format parameter specifies the image format (jpeg, jpg, png, heic, tif).
//
// Example:
//
//	base64Str := "iVBORw0KGgoAAAANSUhEUgAA..."
//	dataURI := customer.EncodeStringToDataURI(base64Str, customer.ImageFormatPNG)
func EncodeStringToDataURI(base64Str string, format ImageFormat) string {
	return fmt.Sprintf("data:image/%s;base64,%s", format, base64Str)
}

// IsDataURI checks if a string is already in data-uri format with a supported MIME type.
// Supported MIME types: image/jpeg, image/png, image/heic, image/tiff,
// application/pdf, text/csv, application/xls, application/xlsx.
func IsDataURI(s string) bool {
	if !strings.HasPrefix(s, "data:") {
		return false
	}

	// Check for supported MIME types
	supportedPrefixes := []string{
		"data:image/jpeg;",
		"data:image/png;",
		"data:image/heic;",
		"data:image/tiff;",
		"data:application/pdf;",
		"data:text/csv;",
		"data:application/xls;",
		"data:application/xlsx;",
	}

	for _, prefix := range supportedPrefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}

	return false
}

// WaitOptions configures the polling behavior for wait functions.
type WaitOptions struct {
	// PollInterval is the interval between polling attempts. Default: 5s.
	PollInterval time.Duration
	// MaxWaitTime is the maximum duration to wait. Default: 10m.
	MaxWaitTime time.Duration
}

// DefaultWaitOptions returns the default wait options.
func DefaultWaitOptions() WaitOptions {
	return WaitOptions{
		PollInterval: 1 * time.Second,
		MaxWaitTime:  60 * time.Minute,
	}
}

// CustomerCondition is a function that checks if a customer meets a condition.
type CustomerCondition func(*CustomerResponse) bool

// WaitFor polls until the condition returns true.
// Returns the customer response when condition is met, or an error on timeout/failure.
func WaitFor(ctx context.Context,
	service Service,
	customerID svc.CustomerID,
	condition CustomerCondition,
	opts *WaitOptions,
) (*CustomerResponse, error) {
	if opts == nil {
		defaults := DefaultWaitOptions()
		opts = &defaults
	}

	deadline := time.Now().Add(opts.MaxWaitTime)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		cust, err := service.GetCustomer(ctx, customerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get customer: %w", err)
		}

		if condition(cust) {
			return cust, nil
		}

		time.Sleep(opts.PollInterval)
	}

	return nil, fmt.Errorf("timeout waiting for customer %s after %v", customerID, opts.MaxWaitTime)
}

// WaitForKybApproved polls until the customer's KYB status becomes APPROVED.
func WaitForKybApproved(ctx context.Context, service Service, customerID svc.CustomerID, opts *WaitOptions) (*CustomerResponse, error) {
	return WaitFor(ctx, service, customerID, func(c *CustomerResponse) bool {
		return c.Status == KybStatusApproved
	}, opts)
}

// WaitForKybDecision polls until the customer's KYB status becomes APPROVED or REJECTED.
// Returns the customer response and nil error if approved, or an error if rejected or timeout.
func WaitForKybDecision(ctx context.Context, service Service, customerID svc.CustomerID, opts *WaitOptions) (*CustomerResponse, error) {
	cust, err := WaitFor(ctx, service, customerID, func(c *CustomerResponse) bool {
		return c.Status == KybStatusApproved || c.Status == KybStatusRejected
	}, opts)
	if err != nil {
		return nil, err
	}

	if cust.Status == KybStatusRejected {
		return cust, fmt.Errorf("KYB rejected for customer %s", customerID)
	}

	return cust, nil
}

// fiatAccountWaitDuration is the delay for waiting on fiat account setup.
const fiatAccountWaitDuration = 60 * time.Second

func WaitForFaitAccount() {
	time.Sleep(fiatAccountWaitDuration)
}
