package command

import (
	"fmt"
	"gon/internal/formatter"
	"gon/internal/httpclient"
	"net/http"
	"strings"
)

type PutCommand struct{}

func (c PutCommand) Name() string {
	return strings.ToLower(http.MethodPut)
}

func (c PutCommand) Group() string {
	return "http"
}

func (c PutCommand) Description() string {
	return "Send an HTTP PUT request"
}

func (c PutCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Println("usage example: put <url>")
		return
	}

	client := httpclient.NewClient()
	request, err := Parse(args[1:])

	if err != nil {
		fmt.Println(err)
		return
	}

	request.Method(http.MethodPut)
	request.URL(args[0])

	response := client.Execute(request.Build())

	if response == nil {
		fmt.Println("error")
		return
	}
	formatter.Formatter["minimal"].Format(response)
}
