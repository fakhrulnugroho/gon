package output

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"gon/internal/adapter/formatter"
	"gon/internal/core/enums"
	"gon/internal/core/payload"
	"gon/internal/utility"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderHttpStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		color      func(string) string
	}{
		{"200 success", 200, utility.ColorSuccess},
		{"299 success", 299, utility.ColorSuccess},
		{"300 info", 300, utility.ColorInfo},
		{"399 info", 399, utility.ColorInfo},
		{"400 warning", 400, utility.ColorWarning},
		{"499 warning", 499, utility.ColorWarning},
		{"500 danger", 500, utility.ColorDanger},
		{"503 danger", 503, utility.ColorDanger},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderHttpStatus(tt.statusCode)
			// The colored output should match the expected color wrapper applied
			// to the "<code> <text>" string.
			plain := fmt.Sprintf("%d %s", tt.statusCode, http.StatusText(tt.statusCode))
			assert.Equal(t, tt.color(plain), got)
		})
	}
}

func TestTrimExecutionTime(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		want string
	}{
		{"sub-second milliseconds", 250, "250ms"},
		{"zero", 0, "0ms"},
		{"just under a second", 999, "999ms"},
		{"exactly one second trims to whole", 1000, "1s"},
		{"one and a half seconds", 1500, "1.5s"},
		{"two decimal places", 1230, "1.23s"},
		{"trailing zero trimmed", 1200, "1.2s"},
		{"large value", 12340, "12.34s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, trimExecutionTime(tt.ms))
		})
	}
}

func TestRenderExecutionTime(t *testing.T) {
	tests := []struct {
		name  string
		ms    int64
		color func(string) string
	}{
		{"under 100ms is success", 50, utility.ColorSuccess},
		{"exactly 100ms is warning", 100, utility.ColorWarning},
		{"between thresholds is warning", 300, utility.ColorWarning},
		{"exactly 500ms is danger", 500, utility.ColorDanger},
		{"above 500ms is danger", 800, utility.ColorDanger},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderExecutionTime(tt.ms)
			assert.Equal(t, tt.color(trimExecutionTime(tt.ms)), got)
		})
	}
}

// captureStdout redirects os.Stdout for the duration of fn and returns what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	defer func() { os.Stdout = orig }()

	fn()

	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

func newTestOutput() *httpOutput {
	return &httpOutput{
		jsonFormatter:    formatter.NewJsonFormatter(),
		keyPairFormatter: formatter.NewKeyPairFormatter(),
	}
}

func sampleIO() (*payload.HttpExecuteInput, *payload.HttpExecuteOutput) {
	input := &payload.HttpExecuteInput{
		Method:  "POST",
		URL:     "https://api.example.com/users",
		Headers: map[string][]string{"X-Trace": {"abc"}},
		Body:    []byte(`{"name":"gon"}`),
	}
	output := &payload.HttpExecuteOutput{
		Body:       []byte(`{"id":1}`),
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		StatusCode: 200,
		Metadata:   payload.Metadata{ExecutionTime: 42 * time.Millisecond},
	}
	return input, output
}

func TestFormatMinimal(t *testing.T) {
	input, output := sampleIO()
	out := captureStdout(t, func() {
		newTestOutput().Format(input, output, enums.DisplayModeMinimal)
	})

	// Status and response body always printed.
	assert.Contains(t, out, "200")
	assert.Contains(t, out, "id")
	// Minimal must NOT print request echo or response headers.
	assert.NotContains(t, out, "POST")
	assert.NotContains(t, out, "Content-Type")
}

func TestFormatNormal(t *testing.T) {
	input, output := sampleIO()
	out := captureStdout(t, func() {
		newTestOutput().Format(input, output, enums.DisplayModeNormal)
	})

	assert.Contains(t, out, "200")
	// Normal adds response headers...
	assert.Contains(t, out, "Content-Type")
	// ...but not the request echo.
	assert.NotContains(t, out, "X-Trace")
}

func TestFormatFull(t *testing.T) {
	input, output := sampleIO()
	out := captureStdout(t, func() {
		newTestOutput().Format(input, output, enums.DisplayModeFull)
	})

	// Full echoes the request method, URL and headers.
	assert.Contains(t, out, "POST")
	assert.Contains(t, out, "api.example.com/users")
	assert.Contains(t, out, "X-Trace")
	// Plus the separator and response details.
	assert.Contains(t, out, "------------------------")
	assert.Contains(t, out, "200")
}
