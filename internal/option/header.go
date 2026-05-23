package option

import (
	"fmt"
	"gon/internal/httpclient"
)

type HeaderOption struct{}

func (o HeaderOption) Name() string {
	return "--header"
}
func (o HeaderOption) ArgCount() int {
	return 2
}
func (o HeaderOption) Apply(rb *httpclient.RequestBuilder, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("--header requires key value")
	}

	rb.AddHeader(args[0], args[1])

	return nil
}
