package driving

import (
	"context"

	"gon/internal/core/payload"
)

type RequestService interface {
	// Run executes a saved request. overrides (may be nil) carries per-execution
	// header/query/body values that take precedence over the request file.
	Run(ctx context.Context, root string, requestPath string, overrides *payload.HttpExecuteInput) (*payload.HttpExecuteInput, *payload.HttpExecuteOutput, error)
	// Create scaffolds a new request file with the given method.
	Create(ctx context.Context, root string, requestPath string, method string) error
}
