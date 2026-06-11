package utility

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyJSON(t *testing.T) {
	t.Run("empty input returns empty string", func(t *testing.T) {
		assert.Equal(t, "", PrettyJSON([]byte("")))
	})

	t.Run("whitespace-only input returns empty string", func(t *testing.T) {
		assert.Equal(t, "", PrettyJSON([]byte("  \n\t ")))
	})

	t.Run("invalid JSON is returned unchanged", func(t *testing.T) {
		raw := []byte("not json at all")
		assert.Equal(t, "not json at all", PrettyJSON(raw))
	})

	t.Run("valid JSON is indented", func(t *testing.T) {
		got := PrettyJSON([]byte(`{"a":1,"b":2}`))
		// json.Indent uses two-space indentation.
		assert.Contains(t, got, "\n  ")
	})

	t.Run("valid JSON contains color codes", func(t *testing.T) {
		got := PrettyJSON([]byte(`{"name":"gon"}`))
		// Key color (#88C0D0 -> 136;192;208) and string color (#A3BE8C -> 163;190;140).
		assert.Contains(t, got, "38;2;136;192;208")
		assert.Contains(t, got, "38;2;163;190;140")
	})
}

func TestHighlightJSON(t *testing.T) {
	t.Run("number coloring", func(t *testing.T) {
		got := highlightJSON("42")
		assert.Contains(t, got, "38;2;180;142;173") // #B48EAD number color
		assert.Contains(t, got, "42")
	})

	t.Run("boolean true", func(t *testing.T) {
		got := highlightJSON("true")
		assert.Contains(t, got, "38;2;208;135;112") // #D08770 bool color
		assert.Contains(t, got, "true")
	})

	t.Run("boolean false", func(t *testing.T) {
		got := highlightJSON("false")
		assert.Contains(t, got, "38;2;208;135;112")
		assert.Contains(t, got, "false")
	})

	t.Run("null", func(t *testing.T) {
		got := highlightJSON("null")
		assert.Contains(t, got, "38;2;191;97;106") // #BF616A null color
		assert.Contains(t, got, "null")
	})

	t.Run("key vs string distinction", func(t *testing.T) {
		got := highlightJSON(`{"key": "value"}`)
		// Key color and string color must both be present and differ.
		assert.Contains(t, got, "38;2;136;192;208") // key
		assert.Contains(t, got, "38;2;163;190;140") // string value
	})

	t.Run("punctuation is colored", func(t *testing.T) {
		got := highlightJSON("{}")
		assert.Contains(t, got, "38;2;216;222;233") // #D8DEE9 punctuation
	})
}

func TestReadJSONString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		start     int
		wantToken string
		wantNext  int
	}{
		{
			name:      "simple string",
			input:     `"abc"`,
			start:     0,
			wantToken: `"abc"`,
			wantNext:  5,
		},
		{
			name:      "string with escaped quote",
			input:     `"a\"b"`,
			start:     0,
			wantToken: `"a\"b"`,
			wantNext:  6,
		},
		{
			name:      "unterminated string returns rest",
			input:     `"abc`,
			start:     0,
			wantToken: `"abc`,
			wantNext:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, next := readJSONString(tt.input, tt.start)
			assert.Equal(t, tt.wantToken, token)
			assert.Equal(t, tt.wantNext, next)
		})
	}
}

func TestReadJSONNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		start int
		want  int
	}{
		{"integer terminated by comma", "123,", 0, 3},
		{"float", "1.5 ", 0, 3},
		{"exponent", "1e10}", 0, 4},
		{"number to end of input", "42", 0, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, readJSONNumber(tt.input, tt.start))
		})
	}
}

func TestIsJSONNumberByte(t *testing.T) {
	for _, c := range "0123456789-+.eE" {
		assert.True(t, isJSONNumberByte(byte(c)), "expected %q to be a number byte", c)
	}
	for _, c := range "abx{}\"" {
		assert.False(t, isJSONNumberByte(byte(c)), "expected %q not to be a number byte", c)
	}
}

func TestIsJSONKey(t *testing.T) {
	t.Run("followed by colon is a key", func(t *testing.T) {
		assert.True(t, isJSONKey(`: 1`, 0))
	})

	t.Run("whitespace before colon is a key", func(t *testing.T) {
		assert.True(t, isJSONKey("  \t: 1", 0))
	})

	t.Run("followed by comma is not a key", func(t *testing.T) {
		assert.False(t, isJSONKey(`, "next"`, 0))
	})

	t.Run("end of input is not a key", func(t *testing.T) {
		assert.False(t, isJSONKey(strings.Repeat(" ", 3), 0))
	})
}
