package httpclient

import (
	"bytes"
	"context"
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
}

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Execute(ctx context.Context, request *Request) (*Result, error) {
	start := time.Now()

	var requestBody io.Reader

	if request.Body != nil {
		requestBody = bytes.NewReader(request.Body)
	}

	req, err := http.NewRequestWithContext(ctx, request.Method, request.URL, requestBody)

	if err != nil {
		return nil, fmt.Errorf("error building request : %w", err)
	}

	req.Header = request.Header.Clone()

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("response error: %w", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	result := Result{
		Request: Request{
			Method: strings.ToUpper(request.Method),
			URL:    request.URL,
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

	return &result, nil
}
