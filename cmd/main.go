package main

import (
	"context"
	"fmt"
	"gon/internal/adapter"
	"gon/internal/adapter/cli/common"
	"gon/internal/adapter/cli/httpcli"
	"gon/internal/color"
	"gon/internal/core/service"
	"gon/internal/version"
	"net/http"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/mattn/go-shellwords"
	"github.com/urfave/cli/v3"
)

func cliApp() *cli.Command {
	httpService := adapter.NewHttpService(&http.Client{})
	httpOutput := service.NewHttpOutput()
	versionService := service.NewVersionService(version.Version, version.OS, version.Arch)

	httpCommands := []*cli.Command{
		httpcli.GetCommand(httpService, httpOutput),
		httpcli.HttpCommand("post", httpService, httpOutput),
		httpcli.HttpCommand("put", httpService, httpOutput),
		httpcli.HttpCommand("delete", httpService, httpOutput),
		httpcli.HttpCommand("patch", httpService, httpOutput),
	}
	utilityCommands := []*cli.Command{
		common.VersionCommand(versionService),
	}

	groups := []common.CommandGroup{
		{Name: "HTTP Commands", Commands: httpCommands},
		{Name: "Common", Commands: utilityCommands},
	}

	helpCmd := common.HelpCommand(groups)
	utilityCommands = append(utilityCommands, helpCmd)
	groups[1].Commands = utilityCommands

	commands := append(httpCommands, utilityCommands...)
	return &cli.Command{
		Name:            "gon",
		Usage:           "An interactive HTTP client for terminal lovers",
		HideHelp:        true,
		HideHelpCommand: true,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if firstCmd := cmd.Args().First(); firstCmd != "" {
				fmt.Printf("Command '%s' not found\n", firstCmd)
				fmt.Println("Type 'help' for available commands")
				return nil
			}
			fmt.Println("gon — An interactive HTTP client for terminal lovers")
			fmt.Println("Type 'help' for available commands")
			return nil
		},
		Commands: commands,
	}
}

func repl() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          color.Info("gon> "),
		HistoryFile:     "/tmp/gon.history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		fmt.Println("errors ", err)
		os.Exit(1)
	}

	defer rl.Close()

	fmt.Println("gon — An interactive HTTP client for terminal lovers")
	fmt.Println("Type 'help' for available commands")

	for {
		line, err := rl.Readline()

		if err != nil {
			break
		}

		input := strings.TrimSpace(line)

		if input == "" {
			continue
		}

		args, _ := shellwords.Parse(line)
		if strings.HasPrefix(args[0], "-") {
			fmt.Println("Invalid command: options must be specified before the command")
			continue
		}
		if err := cliApp().Run(context.Background(), append([]string{"gon"}, args...)); err != nil {
			fmt.Println(err)
		}
		fmt.Println()
	}
}

func main() {
	args := os.Args

	if len(args) > 1 {
		if err := cliApp().Run(context.Background(), args); err != nil {
			fmt.Println(err)
		}
	} else {
		repl()
	}
}
