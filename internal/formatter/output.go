package formatter

import (
	"fmt"
	"gon/internal/color"
	"gon/internal/httpclient"
	"net/http"
	"strings"
)

type OutputFormatter interface {
	Mode() string
	Format(data *httpclient.Result)
}

var Formatter = map[string]OutputFormatter{
	"minimal": NewMinimalFormatter(),
	"normal":  NewNormalFormatter(),
}

type MinimalFormatter struct{}
type NormalFormatter struct{}

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

func printBasicHeader(result *httpclient.Result) {
	response := result.Response

	fmt.Print("\n")
	fmt.Println(renderHttpStatus(response.StatusCode), fmt.Sprintf("(%s)", renderExecutionTime(response.ExecutionTime)))
}

func printBasicFooter(result *httpclient.Result) {
	response := result.Response

	fmt.Print("\n")
	fmt.Println(PrettyJSON(response.Body))
	fmt.Print("\n")
}

func NewMinimalFormatter() *MinimalFormatter {
	return &MinimalFormatter{}
}

func (m *MinimalFormatter) Format(data *httpclient.Result) {
	printBasicHeader(data)
	printBasicFooter(data)
}

func (m *MinimalFormatter) Mode() string {
	return "minimal"
}

func NewNormalFormatter() *NormalFormatter {
	return &NormalFormatter{}
}

func (n *NormalFormatter) Format(data *httpclient.Result) {
	var response = data.Response
	printBasicHeader(data)
	fmt.Print("\n")
	for header, values := range response.Header {
		for _, value := range values {
			fmt.Println(color.Info(header+":"), color.Secondary(value))
		}
	}
	printBasicFooter(data)
}

func (n *NormalFormatter) Mode() string {
	return "normal"
}
