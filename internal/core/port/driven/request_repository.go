package driven

import (
	"context"

	"gon/internal/core/domain"
)

// RequestRepository loads and persists request files. Load also returns the
// chain of collections from the request's own folder up to the project root,
// ordered nearest-first (the request's folder first, the outermost folder last).
type RequestRepository interface {
	Load(ctx context.Context, root string, requestPath string) (*domain.Request, []domain.Collection, error)
	Save(ctx context.Context, root string, requestPath string, request domain.Request) error
	Exists(ctx context.Context, root string, requestPath string) (bool, error)
}
