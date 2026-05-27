package driven

import "gon/internal/core/payload"

type HttpOutput interface {
	Format(input *payload.HttpExecuteInput, output *payload.HttpExecuteOutput, mode int)
}
