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

// Package testdata provides embedded test files for e2e tests and examples.
package testdata

import (
	"embed"

	"github.com/1Money-Co/1money-go-sdk/pkg/service/customer"
)

//go:embed *.png *.jpg
var fs embed.FS

// File paths for embedded test data.
const (
	fileIDFront  = "1128061-Germany-ID_front.png"
	fileIDBack   = "012e6a1-Germany-ID_back.png"
	filePassport = "29346237-germany-passport.jpg"
	filePOA      = "62349849-poa-krause-green.jpg"
)

// IDFront returns the ID front image as a data URI.
func IDFront() string {
	data, _ := fs.ReadFile(fileIDFront)
	return customer.EncodeBase64ToDataURI(data, customer.ImageFormatPng)
}

// IDBack returns the ID back image as a data URI.
func IDBack() string {
	data, _ := fs.ReadFile(fileIDBack)
	return customer.EncodeBase64ToDataURI(data, customer.ImageFormatPng)
}

// Passport returns the passport image as a data URI.
func Passport() string {
	data, _ := fs.ReadFile(filePassport)
	return customer.EncodeBase64ToDataURI(data, customer.ImageFormatJpeg)
}

// POA returns the proof of address image as a data URI.
func POA() string {
	data, _ := fs.ReadFile(filePOA)
	return customer.EncodeBase64ToDataURI(data, customer.ImageFormatJpeg)
}

// POAAsDocument returns the proof of address image as a document data URI.
// This uses the FileFormat type instead of ImageFormat for document uploads.
func POAAsDocument() string {
	data, _ := fs.ReadFile(filePOA)
	return customer.EncodeDocumentToDataURI(data, customer.FileFormatJpeg)
}
