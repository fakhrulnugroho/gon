package output

import (
	"fmt"
	"gon/internal/adapter/formatter"
	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
	"gon/internal/utility"
	"net/http"
	"strings"
)

type httpOutput struct {
	jsonFormatter    formatter.Formatter[[]byte]
	keyPairFormatter formatter.Formatter[map[string]string]
}

func NewHttpOutput(jsonFormatter formatter.Formatter[[]byte], keyPairFormatter formatter.Formatter[map[string]string]) driven.HttpOutput {
	return &httpOutput{jsonFormatter: jsonFormatter, keyPairFormatter: keyPairFormatter}
}

func (h *httpOutput) Format(input *payload.HttpExecuteInput, output *payload.HttpExecuteOutput, mode int) {
	fmt.Println()
	if mode > 1 {
		if input.URL != "" && input.Method != "" {
			fmt.Println(utility.ColorInfo(input.Method), utility.ColorSecondary(input.URL))
		}
		if len(input.Headers) > 0 {
			fmt.Print("\n")
		}
		for header, values := range input.Headers {
			for _, value := range values {
				fmt.Println(utility.ColorInfo(header+":"), utility.ColorSecondary(value))
			}
		}
		if input.Body != nil {
			fmt.Print("\n")
			fmt.Println(h.jsonFormatter.Format(input.Body))
			fmt.Print("\n")
		}
		fmt.Println("------------------------")
		fmt.Println()
	}

	if mode > 0 {
		headers := make(map[string]string)
		for k, v := range output.Headers {
			if len(v) == 0 {
				continue
			}
			headers[k] = strings.Join(v, ", ")
		}
		fmt.Println(h.keyPairFormatter.Format(headers))
	}

	fmt.Println(renderHttpStatus(output.StatusCode), fmt.Sprintf("(%s)", renderExecutionTime(output.Metadata.ExecutionTime.Milliseconds())))
	fmt.Println()
	fmt.Println(h.jsonFormatter.Format(output.Body))
}

func renderHttpStatus(statusCode int) string {
	text := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	switch {
	case statusCode >= 500:
		return utility.ColorDanger(text)
	case statusCode >= 400:
		return utility.ColorWarning(text)
	case statusCode >= 300:
		return utility.ColorInfo(text)
	default:
		return utility.ColorSuccess(text)
	}
}

func trimExecutionTime(executionTime int64) string {
	if executionTime >= 1000 {
		seconds := float64(executionTime) / 1000
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", seconds), "0"), ".") + "s"
	}

	return fmt.Sprintf("%dms", executionTime)
}

func renderExecutionTime(executionTime int64) string {
	if executionTime >= 500 {
		return utility.ColorDanger(trimExecutionTime(executionTime))
	} else if executionTime >= 100 {
		return utility.ColorWarning(trimExecutionTime(executionTime))
	} else {
		return utility.ColorSuccess(trimExecutionTime(executionTime))

	}
}
