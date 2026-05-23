package command

import (
	"fmt"
	"gon/internal/formatter"
	"gon/internal/httpclient"
	"net/http"
	"strings"
)

type DeleteCommand struct{}

func (c DeleteCommand) Name() string {
	return strings.ToLower(http.MethodDelete)
}

func (c DeleteCommand) Group() string {
	return "http"
}

func (c DeleteCommand) Description() string {
	return "Send an HTTP Delete request"
}

func (c DeleteCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("usage example: delete <url>")
		return
	}

	client := httpclient.NewClient()
	request, err := Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	request.Method(http.MethodDelete)
	request.URL(args[0])

	response := client.Execute(request.Build())

	if response == nil {
		fmt.Println("error")
		return
	}
	formatter.Formatter["minimal"].Format(response)
}
