package command

import (
	"context"
	"fmt"
	"gon/internal/core/port/driving"
	"os"

	"github.com/urfave/cli/v3"
)

func WorkspaceInitCommand(workspaceService driving.WorkspaceService) *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize a new workspace",
		Action: func(ctx context.Context, c *cli.Command) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			err = workspaceService.Create(ctx, cwd)
			if err != nil {
				return err
			}
			fmt.Println("Workspace initialized successfully")
			return nil
		},
	}
}
