package formatter

import (
	"fmt"
	"gon/internal/color"
	"gon/internal/httpclient"
	"net/http"
	"strings"
)

func renderHttpStatus(statusCode int) string {
	text := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	switch {
	case statusCode >= 500:
		return color.Danger(text)
	case statusCode >= 400:
		return color.Warning(text)
	case statusCode >= 300:
		return color.Info(text)
	default:
		return color.Success(text)
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
		return color.Danger(trimExecutionTime(executionTime))
	} else if executionTime >= 100 {
		return color.Warning(trimExecutionTime(executionTime))
	} else {
		return color.Success(trimExecutionTime(executionTime))

	}
}

func HttpCall(result *httpclient.Result, output string) {
	response := result.Response

	fmt.Print("\n")
	fmt.Println(renderHttpStatus(response.StatusCode), fmt.Sprintf("(%s)", renderExecutionTime(response.ExecutionTime)))
	if output == "normal" {
		fmt.Print("\n")
		for header, values := range response.Header {
			for _, value := range values {
				fmt.Println(color.Info(header+":"), color.Secondary(value))
			}
		}
	}
	fmt.Print("\n")
	fmt.Println(PrettyJSON(response.Body))
	fmt.Print("\n")
}
