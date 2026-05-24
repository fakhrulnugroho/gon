package main

import (
	"context"
	"fmt"
	"gon/hexagonal/adapter"
	"gon/hexagonal/adapter/cli/common"
	"gon/hexagonal/adapter/cli/http"
	"gon/hexagonal/core/service"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	httpService := adapter.NewHttpService(&http.Client{})
	httpOutput := service.NewHttpOutput()
	versionService := service.NewVersionService("latest")

	commands := []*cli.Command{
		httpcli.GetCommand(httpService, httpOutput),
		httpcli.HttpCommand("post", httpService, httpOutput),
		httpcli.HttpCommand("put", httpService, httpOutput),
		httpcli.HttpCommand("delete", httpService, httpOutput),
		httpcli.HttpCommand("patch", httpService, httpOutput),
		common.VersionCommand(versionService),
	}

	cmd := &cli.Command{
		Name:  "gon",
		Usage: "An interactive HTTP client for terminal lovers",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("gon — An interactive HTTP client for terminal lovers")
			fmt.Println("Type 'help' for available commands")
			return nil
		},
		Commands: commands,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
