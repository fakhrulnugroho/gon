package service

import "strings"

// toSlash normalizes OS path separators to forward slashes so path.Dir and
// path.Base behave consistently regardless of platform.
func toSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}
