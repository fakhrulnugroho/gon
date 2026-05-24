package common

import (
	"context"
	"fmt"
	"gon/hexagonal/core/port/driven"

	"github.com/urfave/cli/v3"
)

func VersionCommand(versionService driven.VersionService) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print the version",
		Action: func(ctx context.Context, c *cli.Command) error {
			fmt.Println(versionService.GetVersion())
			return nil
		},
	}
}
