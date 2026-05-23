package command

import (
	"flag"
	"fmt"
	"gon/internal/color"
	"os"
	"strings"
)

type HelpCommand struct{}
type ExitCommand struct{}
type ClearCommand struct{}

func (c HelpCommand) Name() string {
	return "help"
}

func (c HelpCommand) Group() string {
	return "common"
}

func (c HelpCommand) Description() string {
	return "Print this help message"
}

func (c HelpCommand) Execute(args []string) {
	fmt.Println("Available commands:")
	for _, cmd := range commands {
		fmt.Println("  " + color.Info(strings.ToLower(cmd.Name())))
		if desc := cmd.Description(); desc != "" {
			fmt.Println("     " + desc)
		}
	}
}

func (c ExitCommand) Name() string {
	return "exit"
}

func (c ExitCommand) Group() string {
	return "common"
}

func (c ExitCommand) Description() string {
	return "Exit the application"
}

func (c ExitCommand) Execute(args []string) {
	fmt.Println("bye!")
	os.Exit(0)
}

func (c ClearCommand) Name() string {
	return "clear"
}

func (c ClearCommand) Group() string {
	return "common"
}

func (c ClearCommand) Description() string {
	return "Clear the terminal screen"
}

func (c ClearCommand) Execute(args []string) {
	fmt.Print("\033[H\033[2J")
}

func parseOutputMode(cmdName string, flagArgs []string) string {
	fs := flag.NewFlagSet(cmdName, flag.ContinueOnError)
	minimal := fs.Bool("minimal", false, "show minimal output")
	fs.Parse(flagArgs)
	if *minimal {
		return "minimal"
	}
	return "normal"
}
