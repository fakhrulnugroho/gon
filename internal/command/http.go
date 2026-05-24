package command

import (
	"fmt"
	"gon/internal/formatter"
	"gon/internal/httpclient"
	"strings"
)

// command/http.go
type httpCommand struct {
	method      string
	description string
}

func (c httpCommand) Name() string        { return strings.ToLower(c.method) }
func (c httpCommand) Group() string       { return "http" }
func (c httpCommand) Description() string { return c.description }
func (c httpCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Printf("usage: %s <url> [options]\n", c.Name())
		return
	}
	client := httpclient.NewClient()
	request, err := Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}
	request.Method(c.method).URL(args[0])
	result := client.Execute(request.Build())
	if result == nil {
		fmt.Println("error executing request")
		return
	}
	formatter.Formatter["minimal"].Format(result)
}
