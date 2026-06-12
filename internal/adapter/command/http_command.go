package command

import (
	"context"
	"fmt"
	"gon/internal/core/enums"
	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

func parseHeaders(headers []string) (map[string][]string, error) {
	result := make(map[string][]string)
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header %q: expected \"Key: Value\" format", h)
		}
		key := textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(parts[0]))
		if key == "" {
			return nil, fmt.Errorf("invalid header %q: empty key", h)
		}
		result[key] = append(result[key], strings.TrimSpace(parts[1]))
	}
	return result, nil
}

func parseQuery(query []string) (map[string][]string, error) {
	result := make(map[string][]string)
	for _, q := range query {
		parts := strings.SplitN(q, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid query %q: expected \"Key=Value\" format", q)
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("invalid query %q: empty key", q)
		}
		result[key] = append(result[key], strings.TrimSpace(parts[1]))
	}
	return result, nil
}

func resolveMode(cmd *cli.Command) enums.DisplayMode {
	mode := enums.DisplayModeNormal
	if cmd.Bool("minimal") {
		mode = enums.DisplayModeMinimal
	} else if cmd.Bool("normal") {
		mode = enums.DisplayModeNormal
	} else if cmd.Bool("full") {
		mode = enums.DisplayModeFull
	}
	return mode
}

func HttpCommand(method string, httpService driving.HttpService, environmentService driving.EnvironmentService, httpOutput driven.HttpOutput) *cli.Command {
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
			headers, err := parseHeaders(cmd.StringSlice("header"))
			if err != nil {
				return err
			}
			query, err := parseQuery(cmd.StringSlice("query"))
			if err != nil {
				return err
			}

			input := &payload.HttpExecuteInput{
				Method:  strings.ToUpper(method),
				URL:     cmd.StringArg("url"),
				Query:   query,
				Headers: headers,
			}

			if d := cmd.Duration("timeout"); d > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, d)
				defer cancel()
			}

			if cmd.String("json") != "" {
				input.Body = []byte(cmd.String("json"))
				if _, ok := headers["Content-Type"]; !ok {
					headers["Content-Type"] = append(headers["Content-Type"], "application/json")
				}
			}

			cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				env, err := environmentService.Resolve(ctx, cwd, cmd.String("env"))
				if err != nil {
					return err
				}

				mode := resolveMode(cmd)

				result, err := httpService.Execute(ctx, input, env)
			if err != nil {
				return err
			}
			httpOutput.Format(input, result, (mode))
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "env",
				Usage: "Environment to use for this request (overrides the active one)",
			},
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
			},
			&cli.BoolFlag{
				Name:  "full",
				Usage: `Full output, print request and response details`,
			},
			&cli.DurationFlag{
				Name:  "timeout",
				Value: 30 * time.Second,
				Usage: `Request timeout duration`,
			},
		},
	}
}
