package driven

import (
	"context"
	"gon/internal/core/payload"
)

type HttpService interface {
	Execute(ctx context.Context, input *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error)
}
