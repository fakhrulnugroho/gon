package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionServiceGetVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{"semver", "1.0.0", "linux", "amd64", "gon 1.0.0 (linux/amd64)"},
		{"dev default", "dev", "darwin", "arm64", "gon dev (darwin/arm64)"},
		{"empty values", "", "", "", "gon  (/)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewVersionService(tt.version, tt.os, tt.arch)
			assert.Equal(t, tt.want, svc.GetVersion())
		})
	}
}
