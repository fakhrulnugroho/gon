package command

import (
	"context"
	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"
	"strings"

	"github.com/urfave/cli/v3"
)

func HttpCommand(method string, httpService driving.HttpService, httpOutput driven.HttpOutput) *cli.Command {
	return &cli.Command{
		Name:      method,
		Usage:     "Send an HTTP " + strings.ToUpper(method) + " request",
		ArgsUsage: "<url>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "url",
				UsageText: "URL to send the " + strings.ToUpper(method) + " request to",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			headers := make(map[string][]string)
			query := make(map[string][]string)
			for _, h := range cmd.StringSlice("header") {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					headers[key] = append(headers[key], val)
				}
			}

			for _, h := range cmd.StringSlice("query") {
				parts := strings.SplitN(h, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					query[key] = append(query[key], val)
				}
			}

			input := &payload.HttpExecuteInput{
				Method:  strings.ToUpper(method),
				URL:     cmd.StringArg("url"),
				Query:   query,
				Headers: headers,
			}

			if cmd.String("json") != "" {
				input.Body = []byte(cmd.String("json"))
				headers["Content-Type"] = append(headers["Content-Type"], "application/json")
			}

			mode := 1
			if cmd.Bool("minimal") {
				mode = 0
			} else if cmd.Bool("normal") {
				mode = 1
			} else if cmd.Bool("full") {
				mode = 2
			}

			result, err := httpService.Execute(ctx, input)
			if err != nil {
				return err
			}
			httpOutput.Format(input, result, mode)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "json",
				Value: "",
				Usage: "JSON body for the request",
			},
			&cli.StringSliceFlag{
				Name:  "header",
				Usage: `HTTP header in "Key: Value" format, can be repeated`,
			},
			&cli.StringSliceFlag{
				Name:  "query",
				Usage: `HTTP query parameter in "Key=Value" format, can be repeated`,
			},
			&cli.BoolFlag{
				Name:  "minimal",
				Usage: `Minimal output, only print status code and headers`,
			},
			&cli.BoolFlag{
				Name:  "normal",
				Usage: `Normal output, print status code, headers, and body`,
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "full",
				Usage: `Full output, print request and response details`,
			},
		},
	}
}
