package command

var commands = map[string]Command{}

type Command interface {
	Name() string
	Group() string
	Description() string
	Execute(args []string)
}

func Register(cmd Command) {
	commands[cmd.Name()] = cmd
}

func Find(name string) (Command, bool) {
	cmd, ok := commands[name]
	return cmd, ok
}

func Init() {
	Register(HelpCommand{})
	Register(ClearCommand{})
	Register(ExitCommand{})
	Register(GetCommand{})
}
