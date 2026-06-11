package command

import (
	"context"
	"fmt"
	"os"

	"gon/internal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func RequestNewCommand(requestService driving.RequestService) *cli.Command {
	return &cli.Command{
		Name:      "request",
		Usage:     "Manage saved requests",
		ArgsUsage: "new <path>",
		Commands: []*cli.Command{
			{
				Name:      "new",
				Usage:     "Scaffold a new request file (e.g. request new auth/login --method POST)",
				ArgsUsage: "<path>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "path", UsageText: "Request path, e.g. auth/login"},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "method", Value: "GET", Usage: "HTTP method for the new request"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					requestPath := cmd.StringArg("path")
					if requestPath == "" {
						return fmt.Errorf("request path is required")
					}
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					if err := requestService.Create(ctx, cwd, requestPath, cmd.String("method")); err != nil {
						return err
					}
					fmt.Printf("Request '%s' created\n", requestPath)
					return nil
				},
			},
		},
	}
}
