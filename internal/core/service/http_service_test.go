package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gon/internal/core/domain"
	"gon/internal/core/payload"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpServiceExecuteSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	svc := NewHttpService(nil, server.Client())
	out, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method: http.MethodGet,
		URL:    server.URL,
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, out.StatusCode)
	assert.Equal(t, `{"ok":true}`, string(out.Body))
	assert.Equal(t, "application/json", out.Metadata.ContentType)
	assert.Equal(t, []string{"application/json"}, out.Headers["Content-Type"])
	assert.GreaterOrEqual(t, out.Metadata.ExecutionTime.Nanoseconds(), int64(0))
}

func TestHttpServiceExecuteResolvesWorkspaceURL(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ws := &domain.Workspace{BaseURL: server.URL}
	svc := NewHttpService(ws, server.Client())

	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method: http.MethodGet,
		URL:    "/users",
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, "/users", gotPath)
}

func TestHttpServiceExecuteAbsoluteURLBypassesWorkspace(t *testing.T) {
	var hit bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Workspace base URL points elsewhere; an absolute URL must override it.
	ws := &domain.Workspace{BaseURL: "https://should-not-be-used.invalid"}
	svc := NewHttpService(ws, server.Client())

	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method: http.MethodGet,
		URL:    server.URL + "/abs",
	}, nil)

	require.NoError(t, err)
	assert.True(t, hit)
}

func TestHttpServiceExecuteAppliesWorkspaceDefaults(t *testing.T) {
	var gotHeader, gotQuery, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		gotQuery = r.URL.Query().Get("debug")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ws := &domain.Workspace{
		BaseURL: server.URL,
		Config: domain.Config{
			Path:    "/v1",
			Headers: map[string]string{"Authorization": "Bearer default"},
			Query:   map[string]string{"debug": "1"},
		},
	}
	svc := NewHttpService(ws, server.Client())

	input := &payload.HttpExecuteInput{Method: http.MethodGet, URL: "/users"}
	_, err := svc.Execute(context.Background(), input, nil)

	require.NoError(t, err)
	assert.Equal(t, "Bearer default", gotHeader)
	assert.Equal(t, "1", gotQuery)
	assert.Equal(t, "/v1/users", gotPath)
	// input is mutated so --full output echoes the merged request.
	assert.Equal(t, []string{"Bearer default"}, input.Headers["Authorization"])
	assert.Equal(t, []string{"1"}, input.Query["debug"])
}

func TestHttpServiceExecuteRequestOverridesWorkspaceDefaults(t *testing.T) {
	var gotHeaders []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Values("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ws := &domain.Workspace{
		BaseURL: server.URL,
		Config:  domain.Config{Headers: map[string]string{"Authorization": "Bearer default"}},
	}
	svc := NewHttpService(ws, server.Client())

	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method:  http.MethodGet,
		URL:     "/",
		Headers: map[string][]string{"Authorization": {"Bearer override"}},
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, []string{"Bearer override"}, gotHeaders)
}

func TestHttpServiceExecuteForwardsHeaders(t *testing.T) {
	var gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Token")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewHttpService(nil, server.Client())
	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method:  http.MethodGet,
		URL:     server.URL,
		Headers: map[string][]string{"X-Token": {"secret"}},
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, "secret", gotHeader)
}

func TestHttpServiceExecuteEncodesQuery(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("page")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewHttpService(nil, server.Client())
	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method: http.MethodGet,
		URL:    server.URL,
		Query:  map[string][]string{"page": {"2"}},
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, "2", gotQuery)
}

func TestHttpServiceExecuteSendsBody(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	svc := NewHttpService(nil, server.Client())
	out, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method: http.MethodPost,
		URL:    server.URL,
		Body:   []byte(`{"name":"gon"}`),
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, `{"name":"gon"}`, string(gotBody))
	assert.Equal(t, http.StatusCreated, out.StatusCode)
}

func TestHttpServiceExecuteNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Empty(t, body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewHttpService(nil, server.Client())
	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method: http.MethodGet,
		URL:    server.URL,
		Body:   nil,
	}, nil)

	require.NoError(t, err)
}

func TestHttpServiceExecuteInvalidMethod(t *testing.T) {
	svc := NewHttpService(nil, http.DefaultClient)
	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method: "INVALID METHOD",
		URL:    "https://example.com",
	}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error building request")
}

func TestHttpServiceExecuteTransportError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	client := server.Client()
	url := server.URL
	server.Close() // close so the connection is refused

	svc := NewHttpService(nil, client)
	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method: http.MethodGet,
		URL:    url,
	}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "response error")
}

func TestHttpServiceExecuteSubstitutesVariables(t *testing.T) {
	var gotPath, gotHeader, gotQuery string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHeader = r.Header.Get("Authorization")
		gotQuery = r.URL.Query().Get("uid")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	env := &domain.Environment{
		Name:      "dev",
		BaseURL:   server.URL,
		Variables: map[string]string{"token": "secret", "uid": "42"},
	}
	svc := NewHttpService(nil, server.Client())

	input := &payload.HttpExecuteInput{
		Method:  http.MethodPost,
		URL:     "/users/{{uid}}",
		Headers: map[string][]string{"Authorization": {"Bearer {{token}}"}},
		Query:   map[string][]string{"uid": {"{{uid}}"}},
		Body:    []byte(`{"id":"{{uid}}"}`),
	}
	_, err := svc.Execute(context.Background(), input, env)

	require.NoError(t, err)
	assert.Equal(t, "/users/42", gotPath)
	assert.Equal(t, "Bearer secret", gotHeader)
	assert.Equal(t, "42", gotQuery)
	assert.Equal(t, `{"id":"42"}`, string(gotBody))
}

func TestHttpServiceExecuteFailsOnUnresolvedVariable(t *testing.T) {
	env := &domain.Environment{Name: "dev", BaseURL: "https://api.example.com"}
	svc := NewHttpService(nil, http.DefaultClient)

	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method:  http.MethodGet,
		URL:     "/users",
		Headers: map[string][]string{"Authorization": {"Bearer {{token}}"}},
	}, env)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unresolved variable")
	assert.Contains(t, err.Error(), "token")
}
