package httpcli

import (
	"context"
	"gon/hexagonal/core/payload"
	"gon/hexagonal/core/port/driven"
	"gon/hexagonal/core/port/driving"
	"strings"

	"github.com/urfave/cli/v3"
)

func HttpCommand(method string, httpService driven.HttpService, httpOutput driving.HttpOutput) *cli.Command {
	return &cli.Command{
		Name:  method,
		Usage: "Send an HTTP " + strings.ToUpper(method) + " request",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := &payload.HttpExecuteInput{
				Method: strings.ToUpper(method),
			}
			if cmd.NArg() > 0 {
				input.URL = cmd.Args().First()
			}

			if cmd.String("json") != "" {
				input.Body = []byte(cmd.String("json"))
				input.Headers = map[string][]string{
					"Content-Type": {"application/json"},
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
				Usage: "JSON payload for the POST request",
			},
		},
	}
}
