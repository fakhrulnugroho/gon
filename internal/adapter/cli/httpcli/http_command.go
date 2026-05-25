package httpcli

import (
	"context"
	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"
	"strings"

	"github.com/urfave/cli/v3"
)

func HttpCommand(method string, httpService driven.HttpService, httpOutput driving.HttpOutput) *cli.Command {
	return &cli.Command{
		Name:  method,
		Usage: "Send an HTTP " + strings.ToUpper(method) + " request",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "url",
				UsageText: "URL to send the " + strings.ToUpper(method) + " request to",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			headers := make(map[string][]string)
			for _, h := range cmd.StringSlice("header") {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					headers[key] = append(headers[key], val)
				}
			}

			input := &payload.HttpExecuteInput{
				Method:  strings.ToUpper(method),
				URL:     cmd.StringArg("url"),
				Headers: headers,
			}

			if cmd.String("json") != "" {
				input.Body = []byte(cmd.String("json"))
				if _, exists := input.Headers["Content-Type"]; !exists {
					input.Headers["Content-Type"] = []string{"application/json"}
				}
			}

			result, err := httpService.Execute(ctx, input)
			if err != nil {
				return err
			}
			httpOutput.Format(input, result)
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
		},
	}
}
