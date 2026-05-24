package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Request struct {
	Method string
	URL    string
	Body   []byte
	Header http.Header
}

type Response struct {
	StatusCode    int
	Body          []byte
	Header        http.Header
	ContentType   string
	ContentLength int64
	ExecutionTime int64
}

type Result struct {
	Request  Request
	Response Response
	Error    error
}

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Execute(request *RequestBuilder) *Result {
	start := time.Now()

	var requestBody io.Reader

	if request.body != nil {
		requestBody = bytes.NewReader(request.body)
	}

	req, err := http.NewRequest(request.method, request.url, requestBody)

	if err != nil {
		fmt.Println("error request building request :", err)
		return nil
	}

	for key, value := range request.headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("response error:", err)
		return nil
	}
	defer res.Body.Close()

	responseBody, _ := io.ReadAll(res.Body)

	result := Result{
		Request: Request{
			Method: strings.ToUpper(request.method),
			URL:    request.url,
			Header: req.Header,
		},
		Response: Response{
			Body:          responseBody,
			StatusCode:    res.StatusCode,
			ExecutionTime: time.Since(start).Milliseconds(),
			Header:        res.Header,
			ContentType:   res.Header.Get("Content-Type"),
			ContentLength: res.ContentLength,
		},
	}

	return &result
}
