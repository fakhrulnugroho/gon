package command

import (
	"context"
	"fmt"
	"gon/internal/formatter"
	"gon/internal/httpclient"
	"gon/internal/option"
	"strings"
)

// command/http.go
type httpCommand struct {
	method      string
	description string
}

func (c httpCommand) Name() string        { return strings.ToLower(c.method) }
func (c httpCommand) Group() string       { return "http" }
func (c httpCommand) Description() string { return c.description }
func (c httpCommand) Execute(args []string) {
	if len(args) == 0 {
		fmt.Printf("usage: %s <url> [options]\n", c.Name())
		return
	}
	client := httpclient.NewClient()
	request, err := parse(args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}
	request.Method(c.method).URL(args[0])
	ctx := context.Background()
	result, err := client.Execute(ctx, request.Build())
	if err != nil {
		fmt.Println(err)
		return
	}
	if result == nil {
		fmt.Println("error executing request")
		return
	}
	formatter.Formatter["minimal"].Format(result)
}

func parse(args []string) (*httpclient.RequestBuilder, error) {
	rb := httpclient.NewRequestBuilder()

	i := 0

	for i < len(args) {

		token := args[i]

		handler, exists := option.Find(token)

		if !exists {
			return nil, fmt.Errorf("unknown handler: %s", token)
		}

		argCount := handler.ArgCount()

		start := i + 1
		end := start + argCount

		if end > len(args) {
			return nil, fmt.Errorf("not enough args for %s", token)
		}

		err := handler.Apply(rb, args[start:end])

		if err != nil {
			return nil, err
		}

		i = end
	}

	return rb, nil
}
