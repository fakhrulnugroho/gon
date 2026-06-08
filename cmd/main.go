package main

import (
	"context"
	"fmt"
	"gon/internal/adapter/command"
	"gon/internal/adapter/formatter"
	"gon/internal/adapter/output"
	"gon/internal/adapter/repository"
	"gon/internal/core/domain"
	"gon/internal/core/service"
	"gon/internal/utility"
	"gon/internal/version"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/mattn/go-shellwords"
	"github.com/urfave/cli/v3"
)

func cli_app(workspace *domain.Workspace) *cli.Command {
	httpService := service.NewHttpService(workspace, &http.Client{Timeout: 30 * time.Second})
	jsonFormatter := formatter.NewJsonFormatter()
	keyPairFormatter := formatter.NewKeyPairFormatter()
	httpOutput := output.NewHttpOutput(jsonFormatter, keyPairFormatter)
	versionService := service.NewVersionService(version.Version, version.OS, version.Arch)
	workspaceRepository := repository.NewWorkspaceRepository()
	workspaceService := service.NewWorkspaceService(workspaceRepository)

	httpCommands := []*cli.Command{
		command.HttpCommand(strings.ToLower(http.MethodGet), httpService, httpOutput),
		command.HttpCommand(strings.ToLower(http.MethodPost), httpService, httpOutput),
		command.HttpCommand(strings.ToLower(http.MethodPut), httpService, httpOutput),
		command.HttpCommand(strings.ToLower(http.MethodDelete), httpService, httpOutput),
		command.HttpCommand(strings.ToLower(http.MethodPatch), httpService, httpOutput),
	}
	utilityCommands := []*cli.Command{
		command.VersionCommand(versionService),
	}

	workspaceCommands := []*cli.Command{
		command.WorkspaceInitCommand(workspaceService),
	}

	groups := []command.CommandGroup{
		{Name: "HTTP Commands", Commands: httpCommands},
		{Name: "Workspace", Commands: workspaceCommands},
		{Name: "Common", Commands: utilityCommands},
	}

	helpCmd := command.HelpCommand(groups)
	utilityCommands = append(utilityCommands, helpCmd)
	groups[2].Commands = utilityCommands

	commands := append(httpCommands, workspaceCommands...)
	commands = append(commands, utilityCommands...)
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

type painter struct{}

func (p *painter) Paint(line []rune, pos int) []rune {
	input := string(line[:pos])
	coloredInput := utility.ColorSuccess(input)
	return []rune(coloredInput)
}

func NewPainter() readline.Painter {
	return &painter{}
}

func repl() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("errors ", err)
		os.Exit(1)
	}

	workspaceRepository := repository.NewWorkspaceRepository()
	workspace, err := workspaceRepository.Load(context.Background(), cwd)

	prompt := utility.ColorInfo("gon> ")
	historyFile := "/tmp/gon.history"

	if workspace != nil {
		prompt = utility.ColorInfo("gon(" + workspace.Name + ")> ")
		cacheDirectory := filepath.Join(cwd, ".gon", "cache")
		os.Mkdir(cacheDirectory, 0755)
		historyFile = filepath.Join(cacheDirectory, workspace.Name+".history")
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		Painter:         NewPainter(),
	})

	if err != nil {
		fmt.Println("errors ", err)
		os.Exit(1)
	}

	defer rl.Close()

	fmt.Println("gon — An interactive HTTP client for terminal lovers")
	fmt.Println("Type 'help' for available commands")

	gon_app := cli_app(workspace)

	for {
		line, err := rl.Readline()

		if err != nil {
			break
		}

		input := strings.TrimSpace(line)

		if input == "" {
			continue
		}

		args, err := shellwords.Parse(line)
		if err != nil || len(args) == 0 {
			fmt.Println("Invalid input:", line)
			continue
		}
		if strings.HasPrefix(args[0], "-") {
			fmt.Println("Invalid command: options must be specified before the command")
			continue
		}
		if err := gon_app.Run(context.Background(), append([]string{"gon"}, args...)); err != nil {
			fmt.Println(err)
		}
		fmt.Println()
	}
}

func main() {
	args := os.Args

	if len(args) > 1 {
		if err := cli_app(nil).Run(context.Background(), args); err != nil {
			fmt.Println(err)
		}
	} else {
		repl()
	}
}
