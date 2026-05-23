package command

import (
	"flag"
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
	var client = httpclient.NewClient()
	fs := flag.NewFlagSet(http.MethodPost, flag.ExitOnError)
	body := fs.String("body", "", "request body")
	outputMode := parseOutputMode(fs, args[1:])
	fs.Parse(args[1:])

	if len(args) == 0 {
		fmt.Println("usage example: post <url>")
		return
	}

	request := httpclient.NewRequestBuilder()

	request.Method(http.MethodPost)
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
