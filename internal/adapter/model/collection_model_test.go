package model

import (
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCollectionModel(t *testing.T) {
	raw := []byte(`
name: Auth
description: Authentication endpoints
config:
  path: /auth
  headers:
    X-Client: gon
`)
	var m CollectionModel
	require.NoError(t, yaml.Unmarshal(raw, &m))

	c := m.ToDomain()

	assert.Equal(t, "Auth", c.Name)
	assert.Equal(t, "/auth", c.Config.Path)
	assert.Equal(t, "gon", c.Config.Headers["X-Client"])

	back := NewCollectionModelFromDomain(*c)
	assert.Equal(t, "Auth", back.Name)
	assert.Equal(t, "/auth", back.Config.Path)
	_ = domain.Collection{}
}
