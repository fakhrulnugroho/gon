package driven

import (
	"context"
	"gon/hexagonal/core/payload"
)

type HttpService interface {
	Execute(ctx context.Context, input *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error)
}
