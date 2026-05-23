package httpclient

type RequestBuilder struct {
	method  string
	url     string
	headers map[string]string
	body    []byte
}

func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{}
}

func (b *RequestBuilder) Method(method string) {
	b.method = method
}

func (b *RequestBuilder) URL(url string) {
	b.url = url
}

func (b *RequestBuilder) Headers(headers map[string]string) {
	b.headers = headers
}

func (b *RequestBuilder) Body(body []byte) {
	b.body = body
}

func (b *RequestBuilder) Build() *RequestBuilder {
	return b
}
