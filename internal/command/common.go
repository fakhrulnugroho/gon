package command

import (
	"flag"
	"fmt"
	"gon/internal/color"
	"gon/internal/version"
	"os"
	"strings"
)

type HelpCommand struct{}
type ExitCommand struct{}
type ClearCommand struct{}
type VersionCommand struct{}

func (c VersionCommand) Name() string {
	return "version"
}

func (c VersionCommand) Group() string {
	return "common"
}

func (c VersionCommand) Description() string {
	return "Print the version"
}

func (c VersionCommand) Execute(args []string) {
	fmt.Printf("gon %s (%s/%s)\n", version.Version, version.OS, version.Arch)
}

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
	grouped := make(map[string][]Command)
	groupOrder := []string{}

	for _, cmd := range commands {
		g := cmd.Group()
		if _, seen := grouped[g]; !seen {
			groupOrder = append(groupOrder, g)
		}
		grouped[g] = append(grouped[g], cmd)
	}

	// fixed order: common first, then the rest alphabetically
	ordered := []string{"common"}
	for _, g := range groupOrder {
		if g != "common" {
			ordered = append(ordered, g)
		}
	}

	fmt.Println("Available commands:")
	fmt.Println()
	for _, g := range ordered {
		cmds, ok := grouped[g]
		if !ok {
			continue
		}
		fmt.Println(color.Info(g + ":"))
		for _, cmd := range cmds {
			fmt.Printf("  %-10s %s\n", strings.ToLower(cmd.Name()), cmd.Description())
		}
		fmt.Println()
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

func parseOutputMode(fs *flag.FlagSet, flagArgs []string) string {
	minimal := fs.Bool("minimal", false, "show minimal output")
	fs.Parse(flagArgs)
	if *minimal {
		return "minimal"
	}
	return "normal"
}
