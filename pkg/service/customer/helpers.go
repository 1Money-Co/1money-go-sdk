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
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// IsDataURI checks if a string is already in data-uri format.
func IsDataURI(s string) bool {
	return strings.HasPrefix(s, "data:image/")
}
