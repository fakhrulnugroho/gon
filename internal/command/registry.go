package command

var commands = map[string]Command{
	"help":   HelpCommand{},
	"clear":  ClearCommand{},
	"exit":   ExitCommand{},
	"get":    GetCommand{},
	"post":   PostCommand{},
	"patch":  PatchCommand{},
	"put":    PutCommand{},
	"delete": DeleteCommand{},
}

type Command interface {
	Name() string
	Group() string
	Description() string
	Execute(args []string)
}

func Find(name string) (Command, bool) {
	cmd, ok := commands[name]
	return cmd, ok
}
