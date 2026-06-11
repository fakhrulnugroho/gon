package model

import (
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigModelFromDomain(t *testing.T) {
	cfg := domain.Config{
		Path:    "/v1",
		Query:   map[string]string{"page": "1"},
		Headers: map[string]string{"Accept": "application/json"},
	}

	m := NewConfigModelFromDomain(cfg)

	assert.Equal(t, "/v1", m.Path)
	assert.Equal(t, cfg.Query, m.Query)
	assert.Equal(t, cfg.Headers, m.Headers)
}

func TestConfigModelToDomain(t *testing.T) {
	m := &ConfigModel{
		Path:    "/v2",
		Query:   map[string]string{"q": "x"},
		Headers: map[string]string{"X-Token": "abc"},
	}

	cfg := m.ToDomain()

	assert.Equal(t, "/v2", cfg.Path)
	assert.Equal(t, m.Query, cfg.Query)
	assert.Equal(t, m.Headers, cfg.Headers)
}

func TestConfigModelRoundTrip(t *testing.T) {
	original := domain.Config{
		Path:    "/api",
		Query:   map[string]string{"a": "1", "b": "2"},
		Headers: map[string]string{"H": "v"},
	}

	got := NewConfigModelFromDomain(original).ToDomain()

	assert.Equal(t, original, got)
}

func TestConfigModelEmpty(t *testing.T) {
	got := NewConfigModelFromDomain(domain.Config{}).ToDomain()
	assert.Equal(t, domain.Config{}, got)
}
