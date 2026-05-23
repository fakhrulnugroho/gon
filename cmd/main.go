package main

import (
	"fmt"
	"gon/internal/color"
	"gon/internal/command"
	"strings"

	"github.com/chzyer/readline"
)

func main() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          color.Info("gon> "),
		HistoryFile:     "/tmp/gon.history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		panic(err)
	}

	defer rl.Close()

	fmt.Println("Welcome to gon")
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

		command.Init()
		handleInput(input)
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

	if cmd.Group() == "http" {
		cmd.Execute(parts[1:])
		return
	}
	cmd.Execute(parts[1:])
}
