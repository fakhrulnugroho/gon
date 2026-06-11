package domain

import (
	"net/textproto"
	"strings"

	"gon/internal/core/payload"
)

type Workspace struct {
	Name    string
	BaseURL string
	Config  Config
}

func (w *Workspace) ResolveURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return w.BaseURL + w.Config.Path + path
}

// ApplyDefaults merges the workspace's configured default headers and query
// parameters into input. Per-request values always win: a default is added only
// when input does not already supply that key.
func (w *Workspace) ApplyDefaults(input *payload.HttpExecuteInput) {
	for key, value := range w.Config.Headers {
		canonical := textproto.CanonicalMIMEHeaderKey(key)
		if _, ok := input.Headers[canonical]; ok {
			continue
		}
		if input.Headers == nil {
			input.Headers = make(map[string][]string)
		}
		input.Headers[canonical] = append(input.Headers[canonical], value)
	}

	for key, value := range w.Config.Query {
		if _, ok := input.Query[key]; ok {
			continue
		}
		if input.Query == nil {
			input.Query = make(map[string][]string)
		}
		input.Query[key] = append(input.Query[key], value)
	}
}
