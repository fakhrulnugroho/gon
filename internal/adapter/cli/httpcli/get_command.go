package httpcli

import (
	"context"
	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"
	"strings"

	"github.com/urfave/cli/v3"
)

func GetCommand(httpService driven.HttpService, httpOutput driving.HttpOutput) *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Send an HTTP GET request",
		ArgsUsage: "<url>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "url",
				UsageText: "URL to send the GET request to",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			headers := make(map[string][]string)
			for _, h := range c.StringSlice("header") {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					headers[key] = append(headers[key], val)
				}
			}

			input := &payload.HttpExecuteInput{
				Method:  "GET",
				URL:     c.StringArg("url"),
				Headers: headers,
			}
			result, err := httpService.Execute(ctx, input)
			if err != nil {
				return err
			}
			httpOutput.Format(input, result)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "header",
				Usage: `HTTP header in "Key: Value" format, can be repeated`,
			},
		},
	}
}
