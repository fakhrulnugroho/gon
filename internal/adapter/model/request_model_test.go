package model

import (
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRequestModelToDomain(t *testing.T) {
	t.Run("maps fields, canonicalizes headers, parses json body", func(t *testing.T) {
		raw := []byte(`
name: Login
description: Authenticate a user
method: post
url: /login
headers:
  accept: application/json
query:
  remember: "true"
body:
  json:
    email: a@b.com
`)
		var m RequestModel
		require.NoError(t, yaml.Unmarshal(raw, &m))

		req, err := m.ToDomain()

		require.NoError(t, err)
		assert.Equal(t, "Login", req.Name)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "/login", req.URL)
		assert.Equal(t, []string{"application/json"}, req.Headers["Accept"])
		assert.Equal(t, []string{"true"}, req.Query["remember"])
		assert.Equal(t, domain.BodyJSON, req.Body.Kind)
	})

	t.Run("rejects more than one body kind", func(t *testing.T) {
		var m RequestModel
		require.NoError(t, yaml.Unmarshal([]byte("body:\n  raw: hi\n  form:\n    a: b\n"), &m))

		_, err := m.ToDomain()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "only one body")
	})

	t.Run("no body yields BodyNone", func(t *testing.T) {
		var m RequestModel
		require.NoError(t, yaml.Unmarshal([]byte("method: get\nurl: /ping\n"), &m))

		req, err := m.ToDomain()

		require.NoError(t, err)
		assert.Equal(t, domain.BodyNone, req.Body.Kind)
	})

	t.Run("round-trips through FromDomain", func(t *testing.T) {
		req := domain.Request{
			Name:    "Get User",
			Method:  "GET",
			URL:     "/users/1",
			Headers: map[string][]string{"Accept": {"application/json"}},
		}
		m := NewRequestModelFromDomain(req)
		back, err := m.ToDomain()
		require.NoError(t, err)
		assert.Equal(t, "GET", back.Method)
		assert.Equal(t, []string{"application/json"}, back.Headers["Accept"])
	})
}
