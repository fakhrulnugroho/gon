package formatter

import (
	"gon/internal/core/port/driven"
	"gon/internal/utility"
)

type jsonFormatter struct {
}

func NewJsonFormatter() driven.Formatter[[]byte] {
	return &jsonFormatter{}
}

func (h *jsonFormatter) Format(data []byte) string {
	return utility.PrettyJSON(data)
}
