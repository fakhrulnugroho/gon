package formatter

import (
	"gon/internal/utility"
)

type jsonFormatter struct {
}

func NewJsonFormatter() Formatter[[]byte] {
	return &jsonFormatter{}
}

func (h *jsonFormatter) Format(data []byte) string {
	return utility.PrettyJSON(data)
}
