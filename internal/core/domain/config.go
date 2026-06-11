package domain

import (
	"net/textproto"

	"gon/internal/core/payload"
)

type Config struct {
	Path    string
	Query   map[string]string
	Headers map[string]string
}

// ApplyDefaults merges this config's default headers and query parameters into
// input. It is additive: a default is added only when input does not already
// supply that key, so more specific values always win.
func (c *Config) ApplyDefaults(input *payload.HttpExecuteInput) {
	for key, value := range c.Headers {
		canonical := textproto.CanonicalMIMEHeaderKey(key)
		if _, ok := input.Headers[canonical]; ok {
			continue
		}
		if input.Headers == nil {
			input.Headers = make(map[string][]string)
		}
		input.Headers[canonical] = append(input.Headers[canonical], value)
	}

	for key, value := range c.Query {
		if _, ok := input.Query[key]; ok {
			continue
		}
		if input.Query == nil {
			input.Query = make(map[string][]string)
		}
		input.Query[key] = append(input.Query[key], value)
	}
}
