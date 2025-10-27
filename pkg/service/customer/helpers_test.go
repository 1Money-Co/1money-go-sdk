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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncodeBase64ToDataURI(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		format   ImageFormat
		expected string
	}{
		{
			name:     "JPEG format",
			data:     []byte("test data"),
			format:   ImageFormatJpeg,
			expected: "data:image/jpeg;base64,dGVzdCBkYXRh",
		},
		{
			name:     "PNG format",
			data:     []byte("test data"),
			format:   ImageFormatPng,
			expected: "data:image/png;base64,dGVzdCBkYXRh",
		},
		{
			name:     "HEIC format",
			data:     []byte("test data"),
			format:   ImageFormatHeic,
			expected: "data:image/heic;base64,dGVzdCBkYXRh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeBase64ToDataURI(tt.data, tt.format)
			if result != tt.expected {
				t.Errorf("EncodeBase64ToDataURI() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeStringToDataURI(t *testing.T) {
	tests := []struct {
		name       string
		base64Str  string
		format     ImageFormat
		wantPrefix string
	}{
		{
			name:       "JPEG with base64 string",
			base64Str:  "dGVzdCBkYXRh",
			format:     ImageFormatJpeg,
			wantPrefix: "data:image/jpeg;base64,dGVzdCBkYXRh",
		},
		{
			name:       "PNG with base64 string",
			base64Str:  "aGVsbG8=",
			format:     ImageFormatPng,
			wantPrefix: "data:image/png;base64,aGVsbG8=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeStringToDataURI(tt.base64Str, tt.format)
			if result != tt.wantPrefix {
				t.Errorf("EncodeStringToDataURI() = %v, want %v", result, tt.wantPrefix)
			}
		})
	}
}

func TestEncodeFileToDataURI(t *testing.T) {
	// Create temporary test files
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		fileName   string
		fileData   []byte
		format     ImageFormat
		wantPrefix string
		wantErr    bool
	}{
		{
			name:       "JPEG file with explicit format",
			fileName:   "test.jpg",
			fileData:   []byte("jpeg data"),
			format:     ImageFormatJpeg,
			wantPrefix: "data:image/jpeg;base64,",
			wantErr:    false,
		},
		{
			name:       "PNG file with auto-detect",
			fileName:   "test.png",
			fileData:   []byte("png data"),
			format:     "",
			wantPrefix: "data:image/png;base64,",
			wantErr:    false,
		},
		{
			name:       "TIFF file with auto-detect",
			fileName:   "test.tiff",
			fileData:   []byte("tiff data"),
			format:     "",
			wantPrefix: "data:image/tif;base64,",
			wantErr:    false,
		},
		{
			name:       "unsupported extension",
			fileName:   "test.pdf",
			fileData:   []byte("pdf data"),
			format:     "",
			wantPrefix: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.fileName)
			if err := os.WriteFile(filePath, tt.fileData, 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			result, err := EncodeFileToDataURI(filePath, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeFileToDataURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.HasPrefix(result, tt.wantPrefix) {
				t.Errorf("EncodeFileToDataURI() = %v, want prefix %v", result, tt.wantPrefix)
			}
		})
	}
}

func TestIsDataURI(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "valid data URI",
			s:    "data:image/jpeg;base64,dGVzdA==",
			want: true,
		},
		{
			name: "valid data URI PNG",
			s:    "data:image/png;base64,iVBORw0KGgo=",
			want: true,
		},
		{
			name: "plain base64",
			s:    "dGVzdCBkYXRh",
			want: false,
		},
		{
			name: "empty string",
			s:    "",
			want: false,
		},
		{
			name: "invalid prefix",
			s:    "data:text/plain;base64,dGVzdA==",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDataURI(tt.s); got != tt.want {
				t.Errorf("IsDataURI() = %v, want %v", got, tt.want)
			}
		})
	}
}
