package command

import (
	"context"
	"fmt"
	"os"

	"gon/internal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func EnvCommand(environmentService driving.EnvironmentService) *cli.Command {
	return &cli.Command{
		Name:      "env",
		Usage:     "Manage environments",
		ArgsUsage: "new|list|use <name>",
		Commands: []*cli.Command{
			{
				Name:      "new",
				Usage:     "Create a new environment (e.g. env new dev)",
				ArgsUsage: "<name>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name", UsageText: "Environment name, e.g. dev"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					name := cmd.StringArg("name")
					if name == "" {
						return fmt.Errorf("environment name is required")
					}
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					if err := environmentService.Create(ctx, cwd, name); err != nil {
						return err
					}
					fmt.Printf("Environment '%s' created\n", name)
					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List environments (active marked with *)",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					names, active, err := environmentService.List(ctx, cwd)
					if err != nil {
						return err
					}
					if len(names) == 0 {
						fmt.Println("No environments. Create one with 'env new <name>'")
						return nil
					}
					for _, name := range names {
						marker := " "
						if name == active {
							marker = "*"
						}
						fmt.Printf("%s %s\n", marker, name)
					}
					return nil
				},
			},
			{
				Name:      "use",
				Usage:     "Set the active environment (e.g. env use prod)",
				ArgsUsage: "<name>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name", UsageText: "Environment name to activate"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					name := cmd.StringArg("name")
					if name == "" {
						return fmt.Errorf("environment name is required")
					}
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					if err := environmentService.Use(ctx, cwd, name); err != nil {
						return err
					}
					fmt.Printf("Active environment set to '%s'\n", name)
					return nil
				},
			},
		},
	}
}
