package option

import "gon/internal/httpclient"

type Handler interface {
	Name() string
	ArgCount() int
	Apply(rb *httpclient.RequestBuilder, args []string) error
}

var Registry = map[string]Handler{
	"--json":   JsonOption{},
	"--header": HeaderOption{},
}
