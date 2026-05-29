package formatter

import (
	"gon/internal/utility"
	"sort"
	"strings"
)

type keyPairFormatter struct {
}

func NewKeyPairFormatter() Formatter[map[string]string] {
	return &keyPairFormatter{}
}

func (h *keyPairFormatter) Format(data map[string]string) string {
	var builder strings.Builder
	keys := make([]string, 0, len(data))

	for key := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for _, key := range keys {
		value := data[key]
		builder.WriteString(utility.ColorInfo(key))
		builder.WriteString(utility.ColorInfo(": "))
		builder.WriteString(utility.ColorSecondary(value))
		builder.WriteString("\n")
	}
	return builder.String()
}
