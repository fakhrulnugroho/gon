package model

import (
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkspaceModelFromDomain(t *testing.T) {
	ws := domain.Workspace{
		Name:    "my-api",
		BaseURL: "https://api.example.com",
		Config: domain.Config{
			Path:    "/v1",
			Headers: map[string]string{"Accept": "application/json"},
		},
	}

	m := NewWorkspaceModelFromDomain(ws)

	assert.Equal(t, "my-api", m.Name)
	assert.Equal(t, "https://api.example.com", m.BaseURL)
	assert.Equal(t, "/v1", m.Config.Path)
	assert.Equal(t, ws.Config.Headers, m.Config.Headers)
}

func TestWorkspaceModelToDomain(t *testing.T) {
	m := &WorkspaceModel{
		Name:    "svc",
		BaseURL: "https://svc.local",
		Config: ConfigModel{
			Path:  "/p",
			Query: map[string]string{"k": "v"},
		},
	}

	ws := m.ToDomain()

	assert.NotNil(t, ws)
	assert.Equal(t, "svc", ws.Name)
	assert.Equal(t, "https://svc.local", ws.BaseURL)
	assert.Equal(t, "/p", ws.Config.Path)
	assert.Equal(t, m.Config.Query, ws.Config.Query)
}

func TestWorkspaceModelRoundTrip(t *testing.T) {
	original := domain.Workspace{
		Name:    "round-trip",
		BaseURL: "https://example.org",
		Config: domain.Config{
			Path:    "/api",
			Query:   map[string]string{"page": "1"},
			Headers: map[string]string{"X-Trace": "on"},
		},
	}

	got := NewWorkspaceModelFromDomain(original).ToDomain()

	assert.Equal(t, original, *got)
}

func TestWorkspaceModelToDomainReturnsPointer(t *testing.T) {
	m := &WorkspaceModel{Name: "a"}
	ws := m.ToDomain()
	assert.NotNil(t, ws)
	assert.IsType(t, &domain.Workspace{}, ws)
}
