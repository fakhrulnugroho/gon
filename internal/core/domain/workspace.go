package domain

import "strings"

type Workspace struct {
	Name    string
	BaseURL string
	Config  Config
}

func (w *Workspace) ResolveURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return w.BaseURL + path
}
