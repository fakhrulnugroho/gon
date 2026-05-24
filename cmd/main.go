package main

import (
	"fmt"
	"gon/internal/color"
	"gon/internal/command"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

func interactive() {
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

		handleInput(input)
	}
}

func main() {
	args := os.Args

	if len(args) > 1 {
		handleInput(strings.Join(args[1:], " "))
	} else {
		interactive()
	}

}

func handleInput(input string) {
	parts := strings.Fields(input)

	name := parts[0]

	cmd, ok := command.Find(name)

	if !ok {
		fmt.Println("unknown command")
		return
	}

	cmd.Execute(parts[1:])
}
