package service

import (
	"context"
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingCollectionRepo records Save calls and answers Exists from a set.
type recordingCollectionRepo struct {
	existing map[string]bool
	saved    []string
}

func (r *recordingCollectionRepo) Save(ctx context.Context, root, collectionPath string, c domain.Collection) error {
	r.saved = append(r.saved, collectionPath)
	if r.existing == nil {
		r.existing = map[string]bool{}
	}
	r.existing[collectionPath] = true
	return nil
}
func (r *recordingCollectionRepo) Exists(ctx context.Context, root, collectionPath string) (bool, error) {
	return r.existing[collectionPath], nil
}

func TestCollectionServiceCreate(t *testing.T) {
	t.Run("creates nested collections including ancestors", func(t *testing.T) {
		repo := &recordingCollectionRepo{}
		svc := NewCollectionService(repo)

		err := svc.Create(context.Background(), "/root", "auth/admin")

		require.NoError(t, err)
		assert.Equal(t, []string{"auth", "auth/admin"}, repo.saved)
	})

	t.Run("errors when target already exists", func(t *testing.T) {
		repo := &recordingCollectionRepo{existing: map[string]bool{"auth": true}}
		svc := NewCollectionService(repo)

		err := svc.Create(context.Background(), "/root", "auth")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}
