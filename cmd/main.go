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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"

	"github.com/1Money-Co/1money-go-sdk/cmd/loadtest"
)

const (
	defaultBaseURL = "http://localhost:9000"
	defaultTimeout = 30 * time.Second
)

var (
	// Global flags shared across all commands
	accessKey string
	secretKey string
	baseURL   string
	profile   string
	timeout   time.Duration
	pretty    bool
)

func main() {
	// Load .env file if it exists (silently ignore if not found)
	// This allows users to set ONEMONEY_ACCESS_KEY and ONEMONEY_SECRET_KEY in .env
	_ = godotenv.Load()

	app := &cli.App{
		Name:    "onemoney-cli",
		Usage:   "OneMoney API command-line interface",
		Version: ShortVersion(),
		Authors: []*cli.Author{
			{
				Name: "OneMoney",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "access-key",
				Aliases:     []string{"k"},
				Usage:       "API access key (overrides env vars and config file)",
				EnvVars:     []string{"ONEMONEY_ACCESS_KEY"},
				Destination: &accessKey,
			},
			&cli.StringFlag{
				Name:        "secret-key",
				Aliases:     []string{"s"},
				Usage:       "API secret key (overrides env vars and config file)",
				EnvVars:     []string{"ONEMONEY_SECRET_KEY"},
				Destination: &secretKey,
			},
			&cli.StringFlag{
				Name:        "base-url",
				Aliases:     []string{"u"},
				Usage:       "API base URL",
				Value:       defaultBaseURL,
				EnvVars:     []string{"ONEMONEY_BASE_URL"},
				Destination: &baseURL,
			},
			&cli.StringFlag{
				Name:        "profile",
				Usage:       "Profile to use from ~/.onemoney/credentials (default: \"default\")",
				Destination: &profile,
			},
			&cli.DurationFlag{
				Name:        "timeout",
				Aliases:     []string{"t"},
				Usage:       "Request timeout",
				Value:       defaultTimeout,
				Destination: &timeout,
			},
			&cli.BoolFlag{
				Name:        "pretty",
				Aliases:     []string{"p"},
				Usage:       "Pretty print JSON output",
				Destination: &pretty,
			},
		},
		Commands: []*cli.Command{
			versionCommand(),
			echoCommand(),
			loadtest.Command(),
		},
		Before: func(*cli.Context) error {
			// Credentials validation is now handled by the credential provider chain
			// No need to validate here as credentials can come from:
			// 1. Command-line flags
			// 2. Environment variables
			// 3. Config file
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// printJSON prints the given value as JSON (shared utility function).
func printJSON(v any) error {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(v, "", "  ")
	} else {
		output, err = json.Marshal(v)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
