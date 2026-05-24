package httpclient

import (
	"net/http"
)

type RequestBuilder struct {
	method  string
	url     string
	headers map[string]string
	body    []byte
}

func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{
		headers: make(map[string]string),
	}
}

func (b *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	b.headers[key] = value
	return b
}

func (b *RequestBuilder) Method(method string) *RequestBuilder {
	b.method = method
	return b
}

func (b *RequestBuilder) URL(url string) *RequestBuilder {
	b.url = url
	return b
}

func (b *RequestBuilder) Headers(headers map[string]string) *RequestBuilder {
	b.headers = headers
	return b
}

func (b *RequestBuilder) Body(body []byte) *RequestBuilder {
	b.body = body
	return b
}

func (b *RequestBuilder) Build() *Request {
	headers := make(http.Header)
	for key, value := range b.headers {
		headers.Set(key, value)
	}
	return &Request{
		Method: b.method,
		URL:    b.url,
		Header: headers,
		Body:   b.body,
	}
}
