package payload

import "time"

type HttpExecuteInput struct {
	Method  string
	URL     string
	Headers map[string][]string
	Body    []byte
}

type Metadata struct {
	ContentType   string
	ContentLength int64
	ExecutionTime time.Duration
}

type HttpExecuteOutput struct {
	Body       []byte
	Headers    map[string][]string
	StatusCode int
	Metadata   Metadata
}
