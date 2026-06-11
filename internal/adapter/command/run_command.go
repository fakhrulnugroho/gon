package command

import (
	"context"
	"os"
	"time"

	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func RunCommand(requestService driving.RequestService, httpOutput driven.HttpOutput) *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Run a saved request by path (e.g. run auth/login)",
		ArgsUsage: "<path>",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "path", UsageText: "Path to the saved request, e.g. auth/login"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			headers, err := parseHeaders(cmd.StringSlice("header"))
			if err != nil {
				return err
			}
			query, err := parseQuery(cmd.StringSlice("query"))
			if err != nil {
				return err
			}

			overrides := &payload.HttpExecuteInput{Headers: headers, Query: query}
			if cmd.String("json") != "" {
				overrides.Body = []byte(cmd.String("json"))
				if _, ok := headers["Content-Type"]; !ok {
					headers["Content-Type"] = append(headers["Content-Type"], "application/json")
				}
			}

			if d := cmd.Duration("timeout"); d > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, d)
				defer cancel()
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			result, err := requestService.Run(ctx, cwd, cmd.StringArg("path"), overrides)
			if err != nil {
				return err
			}
			httpOutput.Format(overrides, result, resolveMode(cmd))
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "json", Value: "", Usage: "JSON body override for the request"},
			&cli.StringSliceFlag{Name: "header", Usage: `HTTP header override in "Key: Value" format, can be repeated`},
			&cli.StringSliceFlag{Name: "query", Usage: `HTTP query override in "Key=Value" format, can be repeated`},
			&cli.BoolFlag{Name: "minimal", Usage: "Minimal output, only print status code and headers"},
			&cli.BoolFlag{Name: "normal", Usage: "Normal output, print status code, headers, and body"},
			&cli.BoolFlag{Name: "full", Usage: "Full output, print request and response details"},
			&cli.DurationFlag{Name: "timeout", Value: 30 * time.Second, Usage: "Request timeout duration"},
		},
	}
}
