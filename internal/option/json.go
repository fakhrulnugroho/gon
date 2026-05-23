package option

import (
	"fmt"
	"gon/internal/httpclient"
)

type JsonOption struct{}

func (o JsonOption) Name() string {
	return "--json"
}

func (o JsonOption) ArgCount() int {
	return 1
}

func (o JsonOption) Apply(rb *httpclient.RequestBuilder, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("--json requires value")
	}

	rb.AddHeader("Content-Type", "application/json")
	rb.Body([]byte(args[0]))

	return nil
}
