package adapter

import (
	"bytes"
	"context"
	"fmt"
	"gon/hexagonal/core/payload"
	"io"
	"net/http"
	"time"
)

type HttpService struct {
	httpClient *http.Client
}

func NewHttpService(httpClient *http.Client) *HttpService {
	return &HttpService{httpClient: httpClient}
}

func (s *HttpService) Execute(ctx context.Context, input *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error) {
	start := time.Now()

	var requestBody io.Reader

	if input.Body != nil {
		requestBody = bytes.NewReader(input.Body)
	}

	req, err := http.NewRequestWithContext(ctx, input.Method, input.URL, requestBody)

	if err != nil {
		return nil, fmt.Errorf("error building request : %w", err)
	}

	req.Header = input.Headers

	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("response error: %w", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	result := payload.HttpExecuteOutput{
		Body:       responseBody,
		StatusCode: res.StatusCode,
		Metadata: payload.Metadata{
			ExecutionTime: time.Since(start),
			ContentType:   res.Header.Get("Content-Type"),
			ContentLength: res.ContentLength},
	}

	return &result, nil
}
