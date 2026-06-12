package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"gon/internal/core/domain"
	"gon/internal/core/payload"
	"gon/internal/core/port/driving"
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

func (s *httpService) Execute(ctx context.Context, input *payload.HttpExecuteInput, env *domain.Environment) (*payload.HttpExecuteOutput, error) {
	start := time.Now()

	base := ""
	configPath := ""
	if env != nil {
		base = env.BaseURL
	}
	if s.workspace != nil {
		s.workspace.ApplyDefaults(input)
		configPath = s.workspace.Config.Path
		if base == "" {
			base = s.workspace.BaseURL // deprecated fallback
		}
	}

	url := domain.ResolveURL(base, configPath, input.URL)
	url = substituteInput(url, input, env)

	if missing := unresolvedVariables(url, input); len(missing) > 0 {
		return nil, fmt.Errorf("unresolved variables: %s", strings.Join(missing, ", "))
	}

	var requestBody io.Reader
	if input.Body != nil {
		requestBody = bytes.NewReader(input.Body)
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

// substituteInput resolves {{var}} placeholders in the URL and, in place, across
// every header value, query value, and the body. It returns the resolved URL.
func substituteInput(url string, input *payload.HttpExecuteInput, env *domain.Environment) string {
	if env == nil {
		return url
	}
	url = env.Substitute(url)
	for _, values := range input.Headers {
		for i := range values {
			values[i] = env.Substitute(values[i])
		}
	}
	for _, values := range input.Query {
		for i := range values {
			values[i] = env.Substitute(values[i])
		}
	}
	if input.Body != nil {
		input.Body = []byte(env.Substitute(string(input.Body)))
	}
	return url
}

// unresolvedVariables returns the sorted, de-duplicated names of any {{var}}
// placeholders still present in the URL, headers, query, or body.
func unresolvedVariables(url string, input *payload.HttpExecuteInput) []string {
	seen := map[string]struct{}{}
	add := func(s string) {
		for _, name := range domain.FindPlaceholders(s) {
			seen[name] = struct{}{}
		}
	}
	add(url)
	for _, values := range input.Headers {
		for _, v := range values {
			add(v)
		}
	}
	for _, values := range input.Query {
		for _, v := range values {
			add(v)
		}
	}
	if input.Body != nil {
		add(string(input.Body))
	}
	if len(seen) == 0 {
		return nil
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
