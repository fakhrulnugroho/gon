package command

import (
	"context"
	"fmt"
	"os"

	"gon/internal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func CollectionInitCommand(collectionService driving.CollectionService) *cli.Command {
	return &cli.Command{
		Name:      "collection",
		Usage:     "Manage collections",
		ArgsUsage: "init <name>",
		Commands: []*cli.Command{
			{
				Name:      "init",
				Usage:     "Create a new collection folder (nesting allowed, e.g. auth/admin)",
				ArgsUsage: "<name>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name", UsageText: "Collection path, e.g. auth or auth/admin"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					name := cmd.StringArg("name")
					if name == "" {
						return fmt.Errorf("collection name is required")
					}
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					if err := collectionService.Create(ctx, cwd, name); err != nil {
						return err
					}
					fmt.Printf("Collection '%s' created\n", name)
					return nil
				},
			},
		},
	}
}
