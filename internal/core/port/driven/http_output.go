package driven

import (
	"gon/internal/core/enums"
	"gon/internal/core/payload"
)

type HttpOutput interface {
	Format(input *payload.HttpExecuteInput, output *payload.HttpExecuteOutput, mode enums.DisplayMode)
}
