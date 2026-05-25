package command

import (
	"context"
	"fmt"
	"gon/internal/color"

	"github.com/urfave/cli/v3"
)

type CommandGroup struct {
	Name     string
	Commands []*cli.Command
}

func HelpCommand(groups []CommandGroup) *cli.Command {
	return &cli.Command{
		Name:  "help",
		Usage: "Show available commands",
		Action: func(ctx context.Context, c *cli.Command) error {
			printHelp(groups)
			return nil
		},
	}
}

func printHelp(groups []CommandGroup) {
	for _, group := range groups {
		fmt.Println()
		fmt.Println(color.Info(group.Name + ":"))
		for _, cmd := range group.Commands {
			printEntry(cmd.Name, cmd.ArgsUsage, cmd.Usage)
			for _, f := range cmd.Flags {
				printFlag(f)
			}
		}
	}
}

func printEntry(name, args, desc string) {
	nameCol := color.Success(fmt.Sprintf("  %-8s", name))
	argsCol := color.Secondary(fmt.Sprintf("%-6s", args))
	fmt.Printf("%s %s  %s\n", nameCol, argsCol, desc)
}

func printFlag(f cli.Flag) {
	name := "--" + f.Names()[0]
	usage := flagUsage(f)
	fmt.Printf("  %s  %s\n", color.Warning(fmt.Sprintf("  %-10s", name)), color.Secondary(usage))
}

func flagUsage(f cli.Flag) string {
	switch v := f.(type) {
	case *cli.StringFlag:
		return v.Usage
	case *cli.StringSliceFlag:
		return v.Usage
	case *cli.BoolFlag:
		return v.Usage
	case *cli.IntFlag:
		return v.Usage
	}
	return ""
}
