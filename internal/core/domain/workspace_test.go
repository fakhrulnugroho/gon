package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspaceResolveURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		path    string
		want    string
	}{
		{
			name:    "relative path is prefixed with base URL",
			baseURL: "https://api.example.com",
			path:    "/users",
			want:    "https://api.example.com/users",
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
			w := &Workspace{BaseURL: tt.baseURL}
			assert.Equal(t, tt.want, w.ResolveURL(tt.path))
		})
	}
}
