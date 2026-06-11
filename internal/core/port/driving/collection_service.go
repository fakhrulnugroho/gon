package driving

import "context"

type CollectionService interface {
	// Create scaffolds a collection (and any missing ancestor collections).
	Create(ctx context.Context, root string, collectionPath string) error
}
