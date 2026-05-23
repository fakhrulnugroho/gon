package command

import (
	"fmt"
	"gon/internal/formatter"
	"gon/internal/httpclient"
	"net/http"
	"strings"
)

type GetCommand struct{}

func (c GetCommand) Name() string {
	return strings.ToLower(http.MethodGet)
}

func (c GetCommand) Group() string {
	return "http"
}

func (c GetCommand) Description() string {
	return "Send an HTTP GET request"
}

func (c GetCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("usage example: get <url>")
		return
	}

	client := httpclient.NewClient()
	request, err := Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	request.Method(http.MethodGet)
	request.URL(args[0])

	response := client.Execute(request.Build())

	if response == nil {
		fmt.Println("error")
		return
	}
	formatter.Formatter["minimal"].Format(response)
}
