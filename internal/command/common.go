package command

import (
	"fmt"
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

func (c HelpCommand) Execute(args []string) {
	fmt.Println("Available commands:")
	for _, cmd := range commands {
		fmt.Println("- " + strings.ToLower(cmd.Name()))
	}
}

func (c ExitCommand) Name() string {
	return "exit"
}

func (c ExitCommand) Group() string {
	return "common"
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

func (c ClearCommand) Execute(args []string) {
	fmt.Print("\033[H\033[2J")
}
