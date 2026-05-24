package httpcli

import (
	"context"
	"gon/hexagonal/core/payload"
	"gon/hexagonal/core/port/driven"
	"gon/hexagonal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func GetCommand(httpService driven.HttpService, httpOutput driving.HttpOutput) *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Send an HTTP GET request",
		Action: func(ctx context.Context, c *cli.Command) error {
			input := &payload.HttpExecuteInput{
				Method: "GET",
				URL:    c.Args().First(),
			}
			result, err := httpService.Execute(ctx, input)
			if err != nil {
				return err
			}
			httpOutput.Format(input, result)
			return nil
		},
	}
}
