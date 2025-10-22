package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// versionCommand returns the version command that displays detailed build information.
func versionCommand() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Show version information",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "short",
				Aliases: []string{"s"},
				Usage:   "Print only the version number",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("short") {
				fmt.Println(version)
			} else {
				fmt.Println(VersionInfo())
			}
			return nil
		},
	}
}
