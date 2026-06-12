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

func TestEnvironmentRepositorySaveLoadExistsList(t *testing.T) {
	root := t.TempDir()
	repo := NewEnvironmentRepository()
	ctx := context.Background()

	exists, err := repo.Exists(ctx, root, "dev")
	require.NoError(t, err)
	assert.False(t, exists)

	env := domain.Environment{
		Name:      "dev",
		BaseURL:   "https://api.dev.example.com",
		Variables: map[string]string{"token": "abc123"},
	}
	require.NoError(t, repo.Save(ctx, root, env))

	_, err = os.Stat(filepath.Join(root, "environments", "dev.yml"))
	require.NoError(t, err)

	exists, err = repo.Exists(ctx, root, "dev")
	require.NoError(t, err)
	assert.True(t, exists)

	got, err := repo.Load(ctx, root, "dev")
	require.NoError(t, err)
	assert.Equal(t, env, *got)

	require.NoError(t, repo.Save(ctx, root, domain.Environment{Name: "prod", BaseURL: "https://api.example.com"}))
	names, err := repo.List(ctx, root)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"dev", "prod"}, names)
}

func TestEnvironmentRepositoryListEmpty(t *testing.T) {
	root := t.TempDir()
	repo := NewEnvironmentRepository()

	names, err := repo.List(context.Background(), root)
	require.NoError(t, err)
	assert.Empty(t, names)
}

func TestEnvironmentRepositoryActiveState(t *testing.T) {
	root := t.TempDir()
	repo := NewEnvironmentRepository()
	ctx := context.Background()

	active, err := repo.ReadActive(ctx, root)
	require.NoError(t, err)
	assert.Equal(t, "", active)

	require.NoError(t, repo.WriteActive(ctx, root, "dev"))

	_, err = os.Stat(filepath.Join(root, ".gon", "active-env"))
	require.NoError(t, err)

	active, err = repo.ReadActive(ctx, root)
	require.NoError(t, err)
	assert.Equal(t, "dev", active)
}
