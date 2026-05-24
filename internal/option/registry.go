package option

import "gon/internal/httpclient"

type Handler interface {
	Name() string
	ArgCount() int
	Apply(rb *httpclient.RequestBuilder, args []string) error
}

var registry = map[string]Handler{
	"--json":   JsonOption{},
	"--header": HeaderOption{},
}

func Find(token string) (Handler, bool) {
	h, ok := registry[token]
	return h, ok
}
