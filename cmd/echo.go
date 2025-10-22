package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/1Money-Co/1money-go-sdk/scp"
	"github.com/1Money-Co/1money-go-sdk/scp/service/echo"
)

// echoCommand returns the echo command with all its subcommands.
func echoCommand() *cli.Command {
	return &cli.Command{
		Name:    "echo",
		Aliases: []string{"e"},
		Usage:   "Test echo service",
		Subcommands: []*cli.Command{
			{
				Name:  "get",
				Usage: "Send a GET echo request",
				Action: func(c *cli.Context) error {
					return echoGet(c)
				},
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
				Action: func(c *cli.Context) error {
					return echoPost(c)
				},
			},
		},
		Action: func(c *cli.Context) error {
			return echoGet(c)
		},
	}
}

func echoGet(c *cli.Context) error {
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

func createClient() (*scp.Client, error) {
	return scp.NewClient(&scp.Config{
		AccessKey: accessKey,
		SecretKey: secretKey,
		BaseURL:   baseURL,
		Profile:   profile,
		Timeout:   timeout,
	})
}
