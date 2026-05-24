package color

import "fmt"

const (
	Reset = "\033[0m"
)

func color(hex string, text string) string {
	var r, g, b int

	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)

	return fmt.Sprintf(
		"\033[38;2;%d;%d;%dm%s%s",
		r, g, b, text, Reset,
	)
}

func Secondary(text string) string {
	return color("#d8dee9", text)
}

func Success(text string) string {
	return color("#A3BE8C", text)
}

func Danger(text string) string {
	return color("#BF616A", text)
}

func Warning(text string) string {
	return color("#EBCB8B", text)
}

func Info(text string) string {
	return color("#81A1C1", text)
}

func JSONKey(text string) string {
	return color("#88C0D0", text)
}

func JSONString(text string) string {
	return color("#A3BE8C", text)
}

func JSONNumber(text string) string {
	return color("#B48EAD", text)
}

func JSONBool(text string) string {
	return color("#D08770", text)
}

func JSONNull(text string) string {
	return color("#BF616A", text)
}

func JSONPunctuation(text string) string {
	return color("#D8DEE9", text)
}
