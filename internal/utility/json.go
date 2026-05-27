package utility

import (
	"bytes"
	"encoding/json"
	"gon/internal/color"
	"strings"
)

func PrettyJSON(input []byte) string {
	input = bytes.TrimSpace(input)
	if len(input) == 0 {
		return ""
	}

	var formatted bytes.Buffer
	if err := json.Indent(&formatted, input, "", "  "); err != nil {
		return string(input)
	}

	return highlightJSON(formatted.String())
}

func highlightJSON(input string) string {
	var builder strings.Builder
	builder.Grow(len(input))

	for i := 0; i < len(input); {
		switch input[i] {
		case '"':
			token, next := readJSONString(input, i)
			if isJSONKey(input, next) {
				builder.WriteString(color.JSONKey(token))
			} else {
				builder.WriteString(color.JSONString(token))
			}
			i = next
		case '{', '}', '[', ']', ':', ',':
			builder.WriteString(color.JSONPunctuation(input[i : i+1]))
			i++
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			next := readJSONNumber(input, i)
			builder.WriteString(color.JSONNumber(input[i:next]))
			i = next
		default:
			switch {
			case strings.HasPrefix(input[i:], "true"):
				builder.WriteString(color.JSONBool("true"))
				i += len("true")
			case strings.HasPrefix(input[i:], "false"):
				builder.WriteString(color.JSONBool("false"))
				i += len("false")
			case strings.HasPrefix(input[i:], "null"):
				builder.WriteString(color.JSONNull("null"))
				i += len("null")
			default:
				builder.WriteByte(input[i])
				i++
			}
		}
	}

	return builder.String()
}

func readJSONString(input string, start int) (string, int) {
	for i := start + 1; i < len(input); i++ {
		switch input[i] {
		case '\\':
			i++
		case '"':
			return input[start : i+1], i + 1
		}
	}

	return input[start:], len(input)
}

func readJSONNumber(input string, start int) int {
	for i := start + 1; i < len(input); i++ {
		if !isJSONNumberByte(input[i]) {
			return i
		}
	}

	return len(input)
}

func isJSONNumberByte(char byte) bool {
	return char == '-' || char == '+' || char == '.' || char == 'e' || char == 'E' || (char >= '0' && char <= '9')
}

func isJSONKey(input string, start int) bool {
	for i := start; i < len(input); i++ {
		switch input[i] {
		case ' ', '\t', '\n', '\r':
			continue
		case ':':
			return true
		default:
			return false
		}
	}

	return false
}
