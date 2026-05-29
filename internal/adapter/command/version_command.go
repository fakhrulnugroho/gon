package command

import (
	"context"
	"fmt"
	"gon/internal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func VersionCommand(versionService driving.VersionService) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print the version",
		Action: func(ctx context.Context, c *cli.Command) error {
			fmt.Println(versionService.GetVersion())
			return nil
		},
	}
}
