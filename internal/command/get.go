package command

import (
	"fmt"
	"gon/internal/formatter"
	"gon/internal/httpclient"
)

type GetCommand struct{}

var client = httpclient.NewClient()

func (c GetCommand) Name() string {
	return "get"
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
	response := client.Execute("GET", args[0])

	if response == nil {
		fmt.Println("error")
		return
	}
	formatter.HttpCall(&response.Response, "minimal")
}
