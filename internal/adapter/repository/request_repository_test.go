package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

func TestRequestRepositoryLoad(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gon", "auth", "collection.yml"), "name: Auth\nconfig:\n  path: /auth\n")
	writeFile(t, filepath.Join(root, ".gon", "auth", "admin", "collection.yml"), "name: Admin\nconfig:\n  path: /admin\n")
	writeFile(t, filepath.Join(root, ".gon", "auth", "admin", "impersonate.yml"), "method: post\nurl: /impersonate\n")

	repo := NewRequestRepository()
	req, collections, err := repo.Load(context.Background(), root, "auth/admin/impersonate")

	require.NoError(t, err)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/impersonate", req.URL)
	require.Len(t, collections, 2)
	assert.Equal(t, "Admin", collections[0].Name) // nearest first
	assert.Equal(t, "Auth", collections[1].Name)
}

func TestRequestRepositoryLoadMissing(t *testing.T) {
	repo := NewRequestRepository()
	_, _, err := repo.Load(context.Background(), t.TempDir(), "nope/missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request not found")
}

func TestRequestRepositoryLoadRejectsCollectionFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gon", "auth", "collection.yml"), "name: Auth\n")
	repo := NewRequestRepository()
	_, _, err := repo.Load(context.Background(), root, "auth/collection")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reserved")
}

func TestRequestRepositorySaveAndExists(t *testing.T) {
	root := t.TempDir()
	repo := NewRequestRepository()

	exists, err := repo.Exists(context.Background(), root, "auth/login")
	require.NoError(t, err)
	assert.False(t, exists)

	err = repo.Save(context.Background(), root, "auth/login", domainGetRequest())
	require.NoError(t, err)

	exists, err = repo.Exists(context.Background(), root, "auth/login")
	require.NoError(t, err)
	assert.True(t, exists)

	_, err = os.Stat(filepath.Join(root, ".gon", "auth", "login.yml"))
	require.NoError(t, err)
}

func TestRequestRepositoryLoadYamlExtension(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gon", "ping.yaml"), "method: get\nurl: /ping\n")
	repo := NewRequestRepository()
	req, _, err := repo.Load(context.Background(), root, "ping")
	require.NoError(t, err)
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "/ping", req.URL)
}

func domainGetRequest() domain.Request {
	return domain.Request{Method: "GET", URL: "/login"}
}
