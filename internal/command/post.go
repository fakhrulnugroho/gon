package command

import (
	"fmt"
	"gon/internal/formatter"
	"gon/internal/httpclient"
	"net/http"
	"strings"
)

type PostCommand struct{}

func (c PostCommand) Name() string {
	return strings.ToLower(http.MethodPost)
}

func (c PostCommand) Group() string {
	return "http"
}

func (c PostCommand) Description() string {
	return "Send an HTTP POST request"
}

func (c PostCommand) Execute(args []string) {

	if len(args) == 0 {
		fmt.Println("usage example: post <url>")
		return
	}

	client := httpclient.NewClient()
	request, err := Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	request.Method(http.MethodPost)
	request.URL(args[0])

	response := client.Execute(request.Build())

	if response == nil {
		fmt.Println("error")
		return
	}
	formatter.Formatter["minimal"].Format(response)
}
