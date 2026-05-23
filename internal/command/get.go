package command

import (
	"flag"
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
	fs := flag.NewFlagSet(http.MethodGet, flag.ExitOnError)
	outputMode := parseOutputMode(fs, args[1:])
	client := httpclient.NewClient()

	request := httpclient.NewRequestBuilder()

	request.Method(http.MethodGet)
	request.URL(args[0])

	response := client.Execute(request)
	if response == nil {
		fmt.Println("error")
		return
	}

	formatter.Formatter[outputMode].Format(response)
}
