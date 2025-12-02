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
		return "application/vnd.ms-excel"
	case FileFormatXlsx:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
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

// IsDataURI checks if a string is already in data-uri format.
func IsDataURI(s string) bool {
	return strings.HasPrefix(s, "data:")
}
