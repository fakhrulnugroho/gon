package driving

import (
	"context"

	"gon/internal/core/domain"
	"gon/internal/core/payload"
)

type HttpService interface {
	Execute(ctx context.Context, input *payload.HttpExecuteInput, env *domain.Environment) (*payload.HttpExecuteOutput, error)
}
