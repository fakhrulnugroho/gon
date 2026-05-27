package utility

import "fmt"

const (
	Reset = "\033[0m"
)

func formatColor(hex string, text string) string {
	var r, g, b int

	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)

	return fmt.Sprintf(
		"\033[38;2;%d;%d;%dm%s%s",
		r, g, b, text, Reset,
	)
}

func ColorSecondary(text string) string {
	return formatColor("#d8dee9", text)
}

func ColorSuccess(text string) string {
	return formatColor("#A3BE8C", text)
}

func ColorDanger(text string) string {
	return formatColor("#BF616A", text)
}

func ColorWarning(text string) string {
	return formatColor("#EBCB8B", text)
}

func ColorInfo(text string) string {
	return formatColor("#81A1C1", text)
}
