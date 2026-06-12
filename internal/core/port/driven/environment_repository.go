package driven

import (
	"context"

	"gon/internal/core/domain"
)

// EnvironmentRepository persists environment definition files
// (environments/<name>.yml) and the locally-selected active environment
// (.gon/active-env, gitignored).
type EnvironmentRepository interface {
	Save(ctx context.Context, root string, environment domain.Environment) error
	Load(ctx context.Context, root string, name string) (*domain.Environment, error)
	List(ctx context.Context, root string) ([]string, error)
	Exists(ctx context.Context, root string, name string) (bool, error)
	// ReadActive returns the active environment name, or "" if none is set.
	ReadActive(ctx context.Context, root string) (string, error)
	WriteActive(ctx context.Context, root string, name string) error
}
