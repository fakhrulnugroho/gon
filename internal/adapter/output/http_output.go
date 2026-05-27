package output

import (
	"fmt"
	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
)

type httpOutput struct {
	jsonFormatter driven.Formatter[[]byte]
}

func NewHttpOutput(jsonFormatter driven.Formatter[[]byte]) driven.HttpOutput {
	return &httpOutput{jsonFormatter: jsonFormatter}
}

func (h *httpOutput) Format(input *payload.HttpExecuteInput, output *payload.HttpExecuteOutput, mode int) {
	fmt.Println(h.jsonFormatter.Format(output.Body))
}
