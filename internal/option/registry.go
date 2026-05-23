package option

import "gon/internal/httpclient"

type OptionHandler interface {
	Name() string
	ArgCount() int
	Apply(rb *httpclient.RequestBuilder, args []string) error
}

var Registry = map[string]OptionHandler{
	"--json":   JsonOption{},
	"--header": HeaderOption{},
}
