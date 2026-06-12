package domain

import (
	"strings"

	"gon/internal/core/payload"
)

type Workspace struct {
	Name string
	// BaseURL is a deprecated fallback. The active Environment's base_url is the
	// source of truth; BaseURL is only consulted when no environment supplies one
	// (e.g. older workspaces, or when no environments exist yet).
	BaseURL string
	Config  Config
}

// ResolveURL builds the absolute request URL from a base URL, a config path, and
// the request path. An absolute http(s) request path bypasses resolution.
func ResolveURL(baseURL, configPath, requestPath string) string {
	if strings.HasPrefix(requestPath, "http://") || strings.HasPrefix(requestPath, "https://") {
		return requestPath
	}
	return baseURL + configPath + requestPath
}

func (w *Workspace) ResolveURL(path string) string {
	return ResolveURL(w.BaseURL, w.Config.Path, path)
}

// ApplyDefaults merges the workspace's configured default headers and query
// parameters into input. Per-request values always win.
func (w *Workspace) ApplyDefaults(input *payload.HttpExecuteInput) {
	w.Config.ApplyDefaults(input)
}
