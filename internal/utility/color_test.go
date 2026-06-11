package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatColor(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		text string
		want string
	}{
		{
			name: "white text",
			hex:  "#ffffff",
			text: "hello",
			want: "\033[38;2;255;255;255mhello\033[0m",
		},
		{
			name: "black text",
			hex:  "#000000",
			text: "x",
			want: "\033[38;2;0;0;0mx\033[0m",
		},
		{
			name: "mixed channels",
			hex:  "#81A1C1",
			text: "abc",
			want: "\033[38;2;129;161;193mabc\033[0m",
		},
		{
			name: "empty text still wrapped",
			hex:  "#A3BE8C",
			text: "",
			want: "\033[38;2;163;190;140m\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatColor(tt.hex, tt.text))
		})
	}
}

func TestColorHelpers(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string) string
		want string
	}{
		{"secondary", ColorSecondary, "\033[38;2;216;222;233mtext\033[0m"},
		{"success", ColorSuccess, "\033[38;2;163;190;140mtext\033[0m"},
		{"danger", ColorDanger, "\033[38;2;191;97;106mtext\033[0m"},
		{"warning", ColorWarning, "\033[38;2;235;203;139mtext\033[0m"},
		{"info", ColorInfo, "\033[38;2;129;161;193mtext\033[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn("text")
			assert.Equal(t, tt.want, got)
			// All helpers must reset at the end.
			assert.Contains(t, got, Reset)
		})
	}
}
