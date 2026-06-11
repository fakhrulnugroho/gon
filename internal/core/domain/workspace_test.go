package domain

import (
	"testing"

	"gon/internal/core/payload"

	"github.com/stretchr/testify/assert"
)

func TestWorkspaceResolveURL(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		configPath string
		path       string
		want       string
	}{
		{
			name:    "relative path is prefixed with base URL",
			baseURL: "https://api.example.com",
			path:    "/users",
			want:    "https://api.example.com/users",
		},
		{
			name:       "config path is inserted between base URL and request path",
			baseURL:    "https://api.example.com",
			configPath: "/v1",
			path:       "/users",
			want:       "https://api.example.com/v1/users",
		},
		{
			name:       "config path is ignored for absolute URLs",
			baseURL:    "https://api.example.com",
			configPath: "/v1",
			path:       "https://other.com/api",
			want:       "https://other.com/api",
		},
		{
			name:    "empty path returns base URL",
			baseURL: "https://api.example.com",
			path:    "",
			want:    "https://api.example.com",
		},
		{
			name:    "absolute https URL is returned unchanged",
			baseURL: "https://api.example.com",
			path:    "https://other.com/api",
			want:    "https://other.com/api",
		},
		{
			name:    "absolute http URL is returned unchanged",
			baseURL: "https://api.example.com",
			path:    "http://other.com/api",
			want:    "http://other.com/api",
		},
		{
			name:    "relative path without leading slash",
			baseURL: "https://api.example.com/",
			path:    "users",
			want:    "https://api.example.com/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Workspace{BaseURL: tt.baseURL, Config: Config{Path: tt.configPath}}
			assert.Equal(t, tt.want, w.ResolveURL(tt.path))
		})
	}
}

func TestWorkspaceApplyDefaults(t *testing.T) {
	t.Run("injects default header and query when absent", func(t *testing.T) {
		w := &Workspace{Config: Config{
			Headers: map[string]string{"Authorization": "Bearer token"},
			Query:   map[string]string{"debug": "1"},
		}}
		input := &payload.HttpExecuteInput{}

		w.ApplyDefaults(input)

		assert.Equal(t, []string{"Bearer token"}, input.Headers["Authorization"])
		assert.Equal(t, []string{"1"}, input.Query["debug"])
	})

	t.Run("per-request header wins over default and is not duplicated", func(t *testing.T) {
		w := &Workspace{Config: Config{
			Headers: map[string]string{"authorization": "Bearer default"},
		}}
		input := &payload.HttpExecuteInput{
			Headers: map[string][]string{"Authorization": {"Bearer override"}},
		}

		w.ApplyDefaults(input)

		assert.Equal(t, []string{"Bearer override"}, input.Headers["Authorization"])
	})

	t.Run("per-request query wins over default", func(t *testing.T) {
		w := &Workspace{Config: Config{
			Query: map[string]string{"page": "1"},
		}}
		input := &payload.HttpExecuteInput{
			Query: map[string][]string{"page": {"5"}},
		}

		w.ApplyDefaults(input)

		assert.Equal(t, []string{"5"}, input.Query["page"])
	})

	t.Run("empty config is a no-op", func(t *testing.T) {
		w := &Workspace{}
		input := &payload.HttpExecuteInput{}

		w.ApplyDefaults(input)

		assert.Empty(t, input.Headers)
		assert.Empty(t, input.Query)
	})
}
