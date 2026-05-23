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

type Client struct {
	Headers http.Header
}

func NewClient() *Client {
	return &Client{Headers: http.Header{}}
}

func (c *Client) SetHeader(key, value string) {
	c.Headers.Set(key, value)
}

func (c *Client) Execute(method string, url string, body []byte) *Result {
	start := time.Now()

	var requestBody io.Reader

	if body != nil {
		requestBody = bytes.NewReader(body)
		c.SetHeader("Content-Type", "application/json")
	}

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		fmt.Println("request error:", err)
		return nil
	}
	req.Header = c.Headers

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
			Method: strings.ToUpper(method),
			URL:    url,
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
