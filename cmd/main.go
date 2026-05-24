package main

import (
	"context"
	"fmt"
	"gon/internal/adapter"
	"gon/internal/adapter/cli/common"
	"gon/internal/adapter/cli/httpcli"
	"gon/internal/core/service"
	"gon/internal/version"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	httpService := adapter.NewHttpService(&http.Client{})
	httpOutput := service.NewHttpOutput()
	versionService := service.NewVersionService(version.Version, version.OS, version.Arch)

	commands := []*cli.Command{
		httpcli.GetCommand(httpService, httpOutput),
		httpcli.HttpCommand("post", httpService, httpOutput),
		httpcli.HttpCommand("put", httpService, httpOutput),
		httpcli.HttpCommand("delete", httpService, httpOutput),
		httpcli.HttpCommand("patch", httpService, httpOutput),
		common.VersionCommand(versionService),
	}

	command := &cli.Command{
		Name:  "gon",
		Usage: "An interactive HTTP client for terminal lovers",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("gon — An interactive HTTP client for terminal lovers")
			fmt.Println("Type 'help' for available commands")
			return nil
		},
		Commands: commands,
	}

	if err := command.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
