package driving

import (
	"context"

	"gon/internal/core/domain"
)

type EnvironmentService interface {
	// Create scaffolds a new environment file. Requires an initialized workspace.
	Create(ctx context.Context, root string, name string) error
	// List returns all environment names and the active environment name ("" if none).
	List(ctx context.Context, root string) (names []string, active string, err error)
	// Use marks name as the active environment for this project (local state).
	Use(ctx context.Context, root string, name string) error
	// Resolve returns the active environment. Precedence: override (--env flag) >
	// persisted active state > the sole environment if exactly one exists.
	// Returns (nil, nil) when no environments exist. Returns an error when
	// multiple exist and none is selected, or when a named environment is missing.
	Resolve(ctx context.Context, root string, override string) (*domain.Environment, error)
}
