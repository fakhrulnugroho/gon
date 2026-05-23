package command

import (
	"flag"
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
	var client = httpclient.NewClient()
	fs := flag.NewFlagSet(http.MethodDelete, flag.ExitOnError)
	body := fs.String("body", "", "request body")
	outputMode := parseOutputMode(fs, args[1:])
	fs.Parse(args[1:])

	if len(args) == 0 {
		fmt.Println("usage example: delete <url>")
		return
	}

	request := httpclient.NewRequestBuilder()

	request.Method(http.MethodDelete)
	request.URL(args[0])
	request.Body([]byte(*body))
	request.Headers(map[string]string{"Content-Type": "application/json"})

	response := client.Execute(request.Build())

	if response == nil {
		fmt.Println("error")
		return
	}
	formatter.Formatter[outputMode].Format(response)
}
