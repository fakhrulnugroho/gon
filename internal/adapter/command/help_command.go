package command

import (
	"context"
	"fmt"
	"gon/internal/utility"

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
		fmt.Println(utility.ColorInfo(group.Name + ":"))
		for _, cmd := range group.Commands {
			printEntry(cmd.Name, cmd.ArgsUsage, cmd.Usage)
			for _, f := range cmd.Flags {
				printFlag(f)
			}
		}
	}
}

func printEntry(name, args, desc string) {
	nameCol := utility.ColorSuccess(fmt.Sprintf("  %-8s", name))
	argsCol := utility.ColorSecondary(fmt.Sprintf("%-6s", args))
	fmt.Printf("%s %s  %s\n", nameCol, argsCol, desc)
}

func printFlag(f cli.Flag) {
	name := "--" + f.Names()[0]
	usage := flagUsage(f)
	fmt.Printf("  %s  %s\n", utility.ColorWarning(fmt.Sprintf("  %-10s", name)), utility.ColorSecondary(usage))
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
