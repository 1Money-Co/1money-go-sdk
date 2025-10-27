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
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/1Money-Co/1money-go-sdk/pkg/onemoney"
	"github.com/1Money-Co/1money-go-sdk/pkg/service/echo"
)

// echoCommand returns the echo command with all its subcommands.
func echoCommand() *cli.Command {
	return &cli.Command{
		Name:    "echo",
		Aliases: []string{"e"},
		Usage:   "Test echo service",
		Subcommands: []*cli.Command{
			{
				Name:   "get",
				Usage:  "Send a GET echo request",
				Action: echoGet,
			},
			{
				Name:  "post",
				Usage: "Send a POST echo request",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "message",
						Aliases: []string{"m"},
						Usage:   "Message to send",
						Value:   "Hello from CLI",
					},
				},
				Action: echoPost,
			},
		},
		Action: echoGet,
	}
}

func echoGet(*cli.Context) error {
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx := context.Background()

	resp, err := client.Echo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to perform GET echo: %w", err)
	}

	return printJSON(resp)
}

func echoPost(c *cli.Context) error {
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx := context.Background()

	message := c.String("message")
	if c.NArg() > 0 {
		message = c.Args().First()
	}

	req := &echo.Request{
		Message: message,
	}

	resp, err := client.Echo.Post(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to perform POST echo: %w", err)
	}

	return printJSON(resp)
}

func createClient() (*onemoney.Client, error) {
	return onemoney.NewClient(&onemoney.Config{
		AccessKey: accessKey,
		SecretKey: secretKey,
		BaseURL:   baseURL,
		Profile:   profile,
		Timeout:   timeout,
	})
}
