package command

import (
	"flag"
	"fmt"
	"gon/internal/formatter"
	"gon/internal/httpclient"
	"net/http"
	"strings"
)

type PatchCommand struct{}

func (c PatchCommand) Name() string {
	return strings.ToLower(http.MethodPatch)
}

func (c PatchCommand) Group() string {
	return "http"
}

func (c PatchCommand) Description() string {
	return "Send an HTTP Patch request"
}

func (c PatchCommand) Execute(args []string) {
	var client = httpclient.NewClient()
	fs := flag.NewFlagSet(http.MethodPatch, flag.ExitOnError)
	body := fs.String("body", "", "request body")
	outputMode := parseOutputMode(fs, args[1:])
	fs.Parse(args[1:])

	if len(args) == 0 {
		fmt.Println("usage example: patch <url>")
		return
	}

	request := httpclient.NewRequestBuilder()

	request.Method(http.MethodPatch)
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
