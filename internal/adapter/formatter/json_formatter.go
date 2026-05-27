package formatter

import (
	"gon/internal/core/port/driven"
	"gon/internal/utility"
)

type jsonOutput struct {
}

func NewJsonFormatter() driven.Formatter[[]byte] {
	return &jsonOutput{}
}

func (h *jsonOutput) Format(data []byte) string {
	return utility.PrettyJSON(data)
}
