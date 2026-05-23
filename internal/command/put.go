package command

import (
	"flag"
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
	var client = httpclient.NewClient()
	fs := flag.NewFlagSet(http.MethodPut, flag.ExitOnError)
	body := fs.String("body", "", "request body")
	outputMode := parseOutputMode(fs, args[1:])
	fs.Parse(args[1:])

	if len(args) == 0 {
		fmt.Println("usage example: put <url>")
		return
	}

	request := httpclient.NewRequestBuilder()

	request.Method(http.MethodPut)
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
