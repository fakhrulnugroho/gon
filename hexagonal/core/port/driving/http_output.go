package driving

import "gon/hexagonal/core/payload"

type HttpOutput interface {
	Format(input *payload.HttpExecuteInput, output *payload.HttpExecuteOutput)
}
