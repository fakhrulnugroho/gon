package command

import "net/http"

var commands = map[string]Command{
	"help":    HelpCommand{},
	"clear":   ClearCommand{},
	"exit":    ExitCommand{},
	"version": VersionCommand{},
	"get":     httpCommand{method: http.MethodGet, description: "Send an HTTP GET request"},
	"post":    httpCommand{method: http.MethodPost, description: "Send an HTTP POST request"},
	"patch":   httpCommand{method: http.MethodPatch, description: "Send an HTTP PATCH request"},
	"put":     httpCommand{method: http.MethodPut, description: "Send an HTTP PUT request"},
	"delete":  httpCommand{method: http.MethodDelete, description: "Send an HTTP DELETE request"},
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
