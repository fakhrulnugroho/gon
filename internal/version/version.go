package version

import "runtime"

var (
	Version = "dev"
	OS      = runtime.GOOS
	Arch    = runtime.GOARCH
)
