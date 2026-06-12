package model

import (
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentModelRoundTrip(t *testing.T) {
	env := domain.Environment{
		Name:      "dev",
		BaseURL:   "https://api.dev.example.com",
		Variables: map[string]string{"token": "abc123"},
	}

	m := NewEnvironmentModelFromDomain(env)
	assert.Equal(t, "dev", m.Name)
	assert.Equal(t, "https://api.dev.example.com", m.BaseURL)
	assert.Equal(t, "abc123", m.Variables["token"])

	got := m.ToDomain()
	assert.Equal(t, env, *got)
}
