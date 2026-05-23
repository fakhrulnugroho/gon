package version

import "runtime"

var (
	Version = "latest"
	OS      = runtime.GOOS
	Arch    = runtime.GOARCH
)
