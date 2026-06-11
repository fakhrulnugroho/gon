package domain

import (
	"strings"

	"gon/internal/core/payload"
)

type Workspace struct {
	Name    string
	BaseURL string
	Config  Config
}

func (w *Workspace) ResolveURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return w.BaseURL + w.Config.Path + path
}

// ApplyDefaults merges the workspace's configured default headers and query
// parameters into input. Per-request values always win.
func (w *Workspace) ApplyDefaults(input *payload.HttpExecuteInput) {
	w.Config.ApplyDefaults(input)
}
