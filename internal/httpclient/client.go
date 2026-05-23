package httpclient

import (
	"fmt"
	"io"
	"net/http"
)

type Response struct {
	Body       string
	BodyBytes  []byte
	StatusCode int
}

func Execute(method string, url string) *Response {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println("request error:", err)
		return nil
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("response error:", err)
		return nil
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	return &Response{
		Body:       string(body),
		BodyBytes:  body,
		StatusCode: res.StatusCode,
	}

}
