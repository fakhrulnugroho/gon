package command

import (
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
	if len(args) == 0 {
		fmt.Println("usage example: patch <url>")
		return
	}

	client := httpclient.NewClient()
	request, err := Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	request.Method(http.MethodPatch)
	request.URL(args[0])

	response := client.Execute(request.Build())

	if response == nil {
		fmt.Println("error")
		return
	}
	formatter.Formatter["minimal"].Format(response)
}
