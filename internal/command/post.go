package command

import (
	"flag"
	"fmt"
	"gon/internal/formatter"
	"gon/internal/httpclient"
)

type PostCommand struct{}

func (c PostCommand) Name() string {
	return "post"
}

func (c PostCommand) Group() string {
	return "http"
}

func (c PostCommand) Description() string {
	return "Send an HTTP GET request"
}

func (c PostCommand) Execute(args []string) {
	var client = httpclient.NewClient()
	fs := flag.NewFlagSet("post", flag.ExitOnError)
	minimal := fs.Bool("minimal", false, "test")
	body := fs.String("body", "", "request body")

	fs.Parse(args[1:])

	outputMode := "normal"

	if *minimal {
		outputMode = "minimal"
	}

	if len(args) == 0 {
		fmt.Println("usage example: post <url>")
		return
	}

	response := client.Execute("POST", args[0], []byte(*body))

	if response == nil {
		fmt.Println("error")
		return
	}
	formatter.HttpCall(response, outputMode)
}
