package driven

import (
	"context"

	"gon/internal/core/domain"
)

type CollectionRepository interface {
	Save(ctx context.Context, root string, collectionPath string, collection domain.Collection) error
	Exists(ctx context.Context, root string, collectionPath string) (bool, error)
}
