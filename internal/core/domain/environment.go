package domain

import "regexp"

// Environment is a named, project-scoped set of variables plus a base URL.
// Requests reference its variables via {{name}} substitution.
type Environment struct {
	Name      string
	BaseURL   string
	Variables map[string]string
}

// placeholderPattern matches {{name}} with optional inner whitespace. Names may
// contain letters, digits, underscore, dot, and dash.
var placeholderPattern = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_.-]+)\s*\}\}`)

// Substitute replaces {{name}} placeholders in s with this environment's
// variables in a single pass. Unknown placeholders are left intact so callers
// can detect them.
func (e *Environment) Substitute(s string) string {
	if e == nil {
		return s
	}
	return placeholderPattern.ReplaceAllStringFunc(s, func(match string) string {
		name := placeholderPattern.FindStringSubmatch(match)[1]
		if v, ok := e.Variables[name]; ok {
			return v
		}
		return match
	})
}

// FindPlaceholders returns the variable names of every {{name}} placeholder in s,
// in order of appearance (duplicates included).
func FindPlaceholders(s string) []string {
	matches := placeholderPattern.FindAllStringSubmatch(s, -1)
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, m[1])
	}
	return names
}
