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

func TestCollectionRepositorySaveAndExists(t *testing.T) {
	root := t.TempDir()
	repo := NewCollectionRepository()

	exists, err := repo.Exists(context.Background(), root, "auth")
	require.NoError(t, err)
	assert.False(t, exists)

	err = repo.Save(context.Background(), root, "auth", domain.Collection{Name: "auth"})
	require.NoError(t, err)

	exists, err = repo.Exists(context.Background(), root, "auth")
	require.NoError(t, err)
	assert.True(t, exists)

	_, err = os.Stat(filepath.Join(root, ".gon", "auth", "collection.yml"))
	require.NoError(t, err)
}
