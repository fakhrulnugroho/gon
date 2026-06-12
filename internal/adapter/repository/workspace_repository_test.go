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

func sampleWorkspace() domain.Workspace {
	return domain.Workspace{
		Name:    "my-api",
		BaseURL: "https://api.example.com",
		Config: domain.Config{
			Path:    "/v1",
			Query:   map[string]string{"page": "1"},
			Headers: map[string]string{"Accept": "application/json"},
		},
	}
}

func TestWorkspaceRepositorySave(t *testing.T) {
	dir := t.TempDir()
	repo := NewWorkspaceRepository()

	err := repo.Save(context.Background(), dir, sampleWorkspace())
	require.NoError(t, err)

	path := filepath.Join(dir, "workspace.yml")
	require.FileExists(t, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "name: my-api")
	assert.Contains(t, content, "base_url: https://api.example.com")
}

func TestWorkspaceRepositoryRoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo := NewWorkspaceRepository()
	original := sampleWorkspace()

	require.NoError(t, repo.Save(context.Background(), dir, original))

	loaded, err := repo.Load(context.Background(), dir)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, original, *loaded)
}

func TestWorkspaceRepositoryLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	repo := NewWorkspaceRepository()

	loaded, err := repo.Load(context.Background(), dir)
	require.Error(t, err)
	assert.Nil(t, loaded)
}

func TestWorkspaceRepositoryLoadCorruptedYAML(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "workspace.yml"), []byte("::: not valid yaml :::"), 0644))

	repo := NewWorkspaceRepository()
	loaded, err := repo.Load(context.Background(), dir)
	require.Error(t, err)
	assert.Nil(t, loaded)
}
