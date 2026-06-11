package service

import (
	"bytes"
	"context"
	"fmt"
	"gon/internal/core/domain"
	"gon/internal/core/payload"
	"gon/internal/core/port/driving"
	"io"
	"net/http"
	"time"
)

type httpService struct {
	workspace  *domain.Workspace
	httpClient *http.Client
}

func NewHttpService(workspace *domain.Workspace, httpClient *http.Client) driving.HttpService {
	return &httpService{
		workspace:  workspace,
		httpClient: httpClient,
	}
}

func (s *httpService) Execute(ctx context.Context, input *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error) {
	start := time.Now()

	var requestBody io.Reader

	if input.Body != nil {
		requestBody = bytes.NewReader(input.Body)
	}

	url := input.URL
	if s.workspace != nil {
		s.workspace.ApplyDefaults(input)
		url = s.workspace.ResolveURL(url)
	}

	req, err := http.NewRequestWithContext(ctx, input.Method, url, requestBody)

	if err != nil {
		return nil, fmt.Errorf("error building request : %w", err)
	}

	for key, values := range input.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if len(input.Query) > 0 {
		q := req.URL.Query()
		for key, values := range input.Query {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

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
		Headers:    map[string][]string(res.Header),
		Metadata: payload.Metadata{
			ExecutionTime: time.Since(start),
			ContentType:   res.Header.Get("Content-Type"),
			ContentLength: res.ContentLength,
		},
	}

	return &result, nil
}
